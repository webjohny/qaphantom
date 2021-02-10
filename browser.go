package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/cdproto/security"
	"github.com/webjohny/chromedp"
	"log"
	"strings"
	"time"
)

type Browser struct {
	interceptionID fetch.RequestID

	BrowserContextID cdp.BrowserContextID

	Proxy *Proxy

	ctx context.Context
	allocCtx context.Context
	taskCtx context.Context
	cancelTask context.CancelFunc

	isOpened bool
	streamId int
	limit int64
}

func (b *Browser) Open(proxy *Proxy) (bool, context.Context, context.CancelFunc) {
	options := b.setOpts(proxy)
	if CONF.Env == "local" {
		options = append(options, chromedp.Flag("headless", false))
	}

	// Запускаем контекст браузера
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	taskCtx, cancelTask := chromedp.NewContext(allocCtx)

	var task chromedp.Action
	if proxy != nil {
		task = b.setProxyToContext(proxy)
	}

	if err := chromedp.Run(taskCtx,
		task,
	); err != nil {
		log.Println("Browser.checkProxy.HasError", err)
		return false, nil, nil
	}

	return true, taskCtx, cancelTask
}

func (b *Browser) Init() bool {
	if b.isOpened {
		return true
	}

	if b.limit < 1 {
		b.limit = 60
	}

	if b.Proxy == nil || b.Proxy.Host == "" {
		proxy := &Proxy{}
		// Подключаемся к прокси
		if !proxy.newProxy(){
			return false
		}

		if !b.checkProxy(proxy) {
			return false
		}

		//@todo Commented
		proxy.setTimeout(b.streamId, 5)
		b.Proxy = proxy
	}

	if b.ctx == nil {
		check, ctx, cancel := b.Open(b.Proxy)
		if !check {
			return false
		}
		fmt.Println("NEW INSTANCE")
		b.ctx = ctx
		b.cancelTask = cancel
	}

	fmt.Println("RETURN TRUE")

	b.isOpened = true
	return true
}


func (b *Browser) CheckCaptcha(html string) bool {
	return strings.Contains(html,"g-recaptcha") && strings.Contains(html,"data-sitekey")
}

func (b *Browser) setProxyToContext(proxy *Proxy) chromedp.Tasks {
	fmt.Print(proxy.Login, proxy.Password)
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

func (b *Browser) setOpts(proxy *Proxy) []chromedp.ExecAllocatorOption {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)

	if proxy != nil {
		proxyScheme := proxy.LocalIp

		if proxyScheme != "" {
			opts = append(opts, chromedp.ProxyServer(proxyScheme))
		}

		if proxy.Agent != "" {
			opts = append(opts, chromedp.UserAgent(proxy.Agent))
		}
	}
	return opts
}

func (b *Browser) runWithTimeOut(timeout time.Duration, isStrict bool, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var check bool
		time.AfterFunc(timeout * time.Second, func(){
			if !check {
				if isStrict {
					b.cancelTask()
				} else {
					b.Reload(true)
				}
			}
		})

		err := tasks.Do(ctx)
		if err != nil {
			if "page load error net::ERR_ABORTED" != err.Error() {
				fmt.Println("ERR.Browser.runWithTimeOut", err)
				b.cancelTask()
				return err
			}
			fmt.Println(err)
		}
		if !isStrict {
			check = true
		}
		fmt.Println("RUN_WITH_TIMEOUT")
		return nil
	}
}

func (b *Browser) Cancel() {
	if b.cancelTask != nil {
		b.cancelTask()
	}

	if b.Proxy.LocalIp != "" {
		b.Proxy.freeProxy()
	}
	b.isOpened = false
}

func (b *Browser) Reload(hasProxy bool) bool {
	if hasProxy  {
		if b.cancelTask != nil {
			b.cancelTask()
		}
		b.isOpened = false
	} else {
		b.Cancel()
	}
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
