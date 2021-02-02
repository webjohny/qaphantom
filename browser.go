package main

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/cdproto/security"
	"github.com/webjohny/chromedp"
	"log"
	"time"
)

type Browser struct {
	interceptionID fetch.RequestID

	BrowserContextID cdp.BrowserContextID

	CancelBrowser context.CancelFunc
	CancelTimeout context.CancelFunc
	CancelLogger context.CancelFunc
	Proxy Proxy
	ctx context.Context

	isOpened bool
	streamId int
	limit int64
}

func (b *Browser) Init() bool {
	if b.isOpened {
		return true
	}

	if b.limit < 1 {
		b.limit = 60
	}

	if b.Proxy.Host == "" {
		// Подключаемся к прокси
		b.Proxy.NewProxy()
		if !b.checkProxy(b.Proxy) && b.Proxy.Id < 1 {
			return false
		}

		//@todo Commented
		b.Proxy.SetTimeout(b.streamId, 5)
	}

	options := b.setOpts(b.Proxy)


	if LocalTest {
		options = append(options, chromedp.Flag("headless", false))
	}

	// Запускаем контекст браузера
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	b.CancelBrowser = cancel

	// Устанавливаем собственный logger
	//, chromedp.WithDebugf(log.Printf)
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	b.CancelLogger = cancel

	if err := chromedp.Run(taskCtx,
		chromedp.ActionFunc(func (ctx context.Context) error {
			b.ctx = ctx
			return nil
		}),
		chromedp.Sleep(time.Second),
		b.setProxyToContext(b.Proxy),
	); err != nil {
		log.Println("Browser.Init.HasError", err)
		return false
	}

	b.isOpened = true
	return true
}

func (b *Browser) checkProxy(proxy Proxy) bool {
	options := b.setOpts(proxy)

	// Запускаем контекст браузера
	allocCtx, cancelBrowser := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancelBrowser()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	taskCtx.Deadline()

	var searchHtml string

	if err := chromedp.Run(taskCtx,
		b.setProxyToContext(proxy),
		b.runWithTimeOut(&taskCtx, 10, chromedp.Tasks{
			chromedp.Navigate("https://www.google.com/search?q=whats+my+ip"),
			chromedp.Sleep(3),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			// Вытащить html на проверку каптчи
			chromedp.OuterHTML("body", &searchHtml, chromedp.ByQuery),
		}),
	); err != nil {
		log.Println("Browser.Init.HasError", err)
		return false
	}

	return searchHtml != ""
}


func (b *Browser) setProxyToContext(proxy Proxy) chromedp.Tasks {
	return chromedp.Tasks{
		network.Enable(),
		performance.Enable(),
		page.SetLifecycleEventsEnabled(true),
		security.SetIgnoreCertificateErrors(true),
		emulation.SetTouchEmulationEnabled(false),
		network.SetCacheDisabled(true),
		chromedp.Authentication(proxy.Login, proxy.Password),
	}
}

func (b *Browser) setOpts(proxy Proxy) []chromedp.ExecAllocatorOption {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)

	proxyScheme := proxy.LocalIp

	if proxyScheme != "" {
		opts = append(opts, chromedp.ProxyServer(proxyScheme))
	}

	if proxy.Agent != "" {
		opts = append(opts, chromedp.UserAgent(proxy.Agent))
	}
	return opts
}

func (b *Browser) runWithTimeOut(ctx *context.Context, timeout time.Duration, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		timeoutContext, cancel := context.WithTimeout(ctx, timeout * time.Second)
		defer cancel()
		return tasks.Do(timeoutContext)
	}
}

func (b *Browser) Cancel() {
	if b.CancelBrowser != nil {
		b.CancelBrowser()
	}
	if b.CancelLogger != nil {
		b.CancelLogger()
	}
	if b.CancelTimeout != nil {
		b.CancelTimeout()
	}
	if b.Proxy.LocalIp != "" {
		b.Proxy.FreeProxy()
	}
	b.isOpened = false
}

func (b *Browser) Reload() bool {
	b.Cancel()
	return b.Init()
}

func (b *Browser) ChangeTab() {

}

func (b *Browser) ScreenShot(url string) (bool, []byte) {

	if url == "" {
		return false, []byte("undefined url")
	}

	var buf []byte
	if err := chromedp.Run(b.ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctxt context.Context) error {
			_, viewLayout, contentRect, err := page.GetLayoutMetrics().Do(ctxt)
			if err != nil {
				return err
			}

			v := page.Viewport{
				X:      contentRect.X,
				Y:      contentRect.Y,
				Width:  viewLayout.ClientWidth, // or contentRect.Width,
				Height: viewLayout.ClientHeight,
				Scale:  1,
			}
			log.Printf("Capture %#v", v)
			buf, err = page.CaptureScreenshot().WithClip(&v).Do(ctxt)
			if err != nil {
				return err
			}
			return nil
		}),
	); err != nil {
		log.Println("Browser.ScreenShotSave.HasError", err)
		return false, []byte(err.Error())
	}

	if len(buf) < 1 {
		return false, []byte("undefined image")
	}

	return true, buf
}
