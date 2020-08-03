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

	// Подключаемся к прокси
	b.Proxy.NewProxy()
	if b.Proxy.Id < 1 {
		return false
	}

	//@todo Commented
	b.Proxy.SetTimeout(b.streamId, 5)

	proxyScheme := b.Proxy.LocalIp

	// Инициализация контроллера для управление парсингом
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)

	if LocalTest {
		opts = append(opts, chromedp.Flag("headless", false))
	}

	if proxyScheme != "" {
		opts = append(opts, chromedp.ProxyServer(proxyScheme))
	}

	if b.Proxy.Agent != "" {
		opts = append(opts, chromedp.UserAgent(b.Proxy.Agent))
	}

	// Запускаем контекст браузера
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	b.CancelBrowser = cancel

	// Устанавливаем собственный logger
	//, chromedp.WithDebugf(log.Printf)
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	b.CancelLogger = cancel

	if err := chromedp.Run(taskCtx,
		chromedp.ActionFunc(func (ctx context.Context) error {
			// Ставим таймер на отключение если зависнет
			//ctx, cancel = context.WithTimeout(ctx, time.Duration(b.limit) * time.Second)
			//b.CancelTimeout = cancel
			//
			b.ctx = ctx
			return nil
		}),
		chromedp.Sleep(time.Second),
		network.Enable(),
		performance.Enable(),
		page.SetLifecycleEventsEnabled(true),
		security.SetIgnoreCertificateErrors(true),
		emulation.SetTouchEmulationEnabled(false),
		network.SetCacheDisabled(true),
		chromedp.Authentication(b.Proxy.Login, b.Proxy.Password),
	); err != nil {
		log.Println("Browser.Init.HasError", err)
		return false
	}

	//taskCtx2, cancel := chromedp.NewContext(taskCtx)
	//b.CancelLogger = cancel
	//
	//if err := chromedp.Run(taskCtx2); err != nil {
	//	log.Println(err)
	//}

	b.isOpened = true
	return true
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