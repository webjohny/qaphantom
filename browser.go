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

func (b *Browser) Init() bool {
	if b.isOpened {
		return true
	}

	if b.limit < 1 {
		b.limit = 60
	}

	if b.Proxy == nil || b.Proxy.Host == "" {
		b.Proxy = &Proxy{}
		// Подключаемся к прокси
		if !b.Proxy.newProxy(){
			return false
		}

		if !b.checkProxy(b.Proxy) {
			return false
		}

		//@todo Commented
		b.Proxy.setTimeout(b.streamId, 5)
	}

	options := b.setOpts(b.Proxy)


	if CONF.Env == "local" {
		options = append(options, chromedp.Flag("headless", false))
	}

	if b.ctx == nil {
		fmt.Println("NEW INSTANCE")
		// Запускаем контекст браузера
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)

		// Устанавливаем собственный logger
		//, chromedp.WithDebugf(log.Printf)
		taskCtx, cancel := chromedp.NewContext(allocCtx)
		b.cancelTask = cancel

		if err := chromedp.Run(taskCtx,
			chromedp.Sleep(time.Second),
			b.setProxyToContext(b.Proxy),
		); err != nil {
			log.Println("Browser.Init.HasError", err)
			return false
		}
		b.ctx = taskCtx
	}

	fmt.Println("RETURN TRUE")

	b.isOpened = true
	return true
}

func (b *Browser) checkProxy(proxy *Proxy) bool {
	if proxy == nil {
		return false
	}

	options := b.setOpts(proxy)
	if CONF.Env == "local" {
		options = append(options, chromedp.Flag("headless", false))
	}

	keyWords := []string{
		"whats+my+ip",
		"ssh+run+command",
		"how+work+with+git",
		"bitcoin+price+2013+year",
		"онлайн+обменник+крипта+рубль",
		"где+купить+акции",
		"i+want+to+spend+crypto",
	}

	// Запускаем контекст браузера
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), options...)
	taskCtx, _ := chromedp.NewContext(allocCtx)

	if err := chromedp.Run(taskCtx); err != nil {
		log.Println("start:", err)
	}

	ctx1, cancelTask := context.WithCancel(taskCtx)

	var searchHtml string
	var videosHtml string

	b.taskCtx = taskCtx
	b.cancelTask = cancelTask

fmt.Println(UTILS.ArrayRand(keyWords))
	if err := chromedp.Run(ctx1,
		b.setProxyToContext(proxy),
		b.runWithTimeOut(10, false, chromedp.Tasks{
			//chromedp.Navigate("https://www.google.com/search?q=" + UTILS.ArrayRand(keyWords)),
			//chromedp.WaitVisible("body", chromedp.ByQuery),
			//// Вытащить html на проверку каптчи
			//chromedp.OuterHTML("body", &searchHtml, chromedp.ByQuery),
			// Устанавливаем страницу для парсинга
			chromedp.Navigate("https://www.google.com/search?source=lnms&tbm=vid&as_sitesearch=youtube.com&as_qdr=y&num=50&q=whats+my+ip"),
			chromedp.WaitVisible("#rso",chromedp.ByQuery),
			chromedp.OuterHTML("#rso", &videosHtml, chromedp.ByQuery),
			chromedp.Sleep(2222 * time.Second),
		}),
		b.runWithTimeOut(10, false, chromedp.Tasks{
			//chromedp.Navigate("https://www.google.com/search?q=" + UTILS.ArrayRand(keyWords)),
			//chromedp.WaitVisible("body", chromedp.ByQuery),
			//// Вытащить html на проверку каптчи
			//chromedp.OuterHTML("body", &searchHtml, chromedp.ByQuery),
			// Устанавливаем страницу для парсинга
			chromedp.Navigate("https://www.google.com/search?source=lnms&tbm=vid&as_sitesearch=youtube.com&as_qdr=y&num=50&q=whats+my+ip"),
			chromedp.WaitVisible("#rso",chromedp.ByQuery),
			chromedp.OuterHTML("#rso", &videosHtml, chromedp.ByQuery),
			chromedp.Sleep(2222 * time.Second),
		}),
	); err != nil {
		log.Println("Browser.checkProxy.HasError", err)
		return false
	}

	fmt.Println("HELLO")

	return false

	if searchHtml != "" {
		if b.CheckCaptcha(searchHtml) {
			return false
		}
		b.ctx = taskCtx
		return true
	}

	return false
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
					ctx2, _ := chromedp.NewContext(b.taskCtx)
					if err := chromedp.Run(ctx2,
						chromedp.Navigate("https://google.com"),
					); err != nil {
						fmt.Println("ERR.Browser.runWithTimeOut", err)
					}else {
						b.ctx = ctx
						b.cancelTask()
						//b.cancelTask = cancelTask
					}
					//err := chromedp.Tasks{
					//	chromedp.Stop(),
					//	chromedp.Navigate("https://google.com"),
					//}.Do(ctx)
					//if err != nil {
					//	fmt.Println("ERR.Browser.runWithTimeOut", err)
					//}
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
