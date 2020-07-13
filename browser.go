package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/cdproto/security"
	"github.com/webjohny/chromedtp"
	"log"
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
		chromedtp.DefaultExecAllocatorOptions[:],
		chromedtp.DisableGPU,
		chromedtp.NoSandbox,
	)

	if LocalTest {
		opts = append(opts, chromedtp.Flag("headless", false))
	}

	if proxyScheme != "" {
		opts = append(opts, chromedtp.ProxyServer(proxyScheme))
	}

	if b.Proxy.Agent != "" {
		opts = append(opts, chromedtp.UserAgent(b.Proxy.Agent))
	}

	// Запускаем контекст браузера
	allocCtx, cancel := chromedtp.NewExecAllocator(context.Background(), opts...)
	b.CancelBrowser = cancel

	// Устанавливаем собственный logger
	//, chromedtp.WithDebugf(log.Printf)
	taskCtx, cancel := chromedtp.NewContext(allocCtx)
	b.CancelLogger = cancel

	if err := chromedtp.Run(taskCtx,
		chromedtp.ActionFunc(func (ctx context.Context) error {
			// Ставим таймер на отключение если зависнет
			//ctx, cancel = context.WithTimeout(ctx, time.Duration(b.limit) * time.Second)
			//b.CancelTimeout = cancel
			//
			b.ctx = ctx
			return nil
		}),
		network.Enable(),
		performance.Enable(),
		page.SetLifecycleEventsEnabled(true),
		security.SetIgnoreCertificateErrors(true),
		emulation.SetTouchEmulationEnabled(false),
		network.SetCacheDisabled(true),
		chromedtp.ActionFunc(func (ctx context.Context) error {
			if b.Proxy.Login != "" && b.Proxy.Password != "" {
				err := fetch.Enable().WithPatterns([]*fetch.RequestPattern{{"*", "", ""}}).WithHandleAuthRequests(true).Do(ctx)
				if err != nil {
					fmt.Println("Fetch.enable", err)
					return err
				}
				b.ListenForNetworkEvent(ctx)
			}
			return nil
		}),
	); err != nil {
		log.Println(err)
		b.Reload()
		return false
	}

	//taskCtx2, cancel := chromedtp.NewContext(taskCtx)
	//b.CancelLogger = cancel
	//
	//if err := chromedtp.Run(taskCtx2); err != nil {
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

func (b *Browser) Reload() {
	b.Cancel()
	b.Init()
	b.isOpened = true
}

func (b *Browser) ChangeTab() {

}

func (b *Browser) ListenForNetworkEvent(ctx context.Context) {
	c := chromedtp.FromContext(ctx)
	id := 5000
	chromedtp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *fetch.EventAuthRequired:
			buf, _ := json.Marshal(map[string]interface{}{
				"requestId": ev.RequestID.String(),
				"authChallengeResponse": map[string]string{
					"response": "ProvideCredentials",
					"username": b.Proxy.Login,
					"password": b.Proxy.Password,
				},
			})
			cmd := &cdproto.Message{
				ID:        int64(id + 1),
				SessionID: c.Target.SessionID,
				Method:    cdproto.MethodType("Fetch.continueWithAuth"),
				Params:    buf,
			}
			err := c.Browser.Conn.Write(ctx, cmd)
			if err != nil {
				fmt.Println("Fetch.continueWithAuth", err)
			}

		case *fetch.EventRequestPaused:
			b.interceptionID = ev.RequestID
			buf, _ := json.Marshal(map[string]string{"requestId":ev.RequestID.String()})
			cmd := &cdproto.Message{
				ID:        int64(id + 1),
				SessionID: c.Target.SessionID,
				Method:    cdproto.MethodType("Fetch.continueRequest"),
				Params:    buf,
			}
			err := c.Browser.Conn.Write(ctx, cmd)
			if err != nil {
				fmt.Println("Fetch.continueRequest", err)
			}
		}
	})
}