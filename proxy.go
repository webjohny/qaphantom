package main

import (
	"database/sql"
	"github.com/webjohny/chromedp"
	"log"
	"strconv"
	"time"
)

type Proxy struct {
	Id int
	Host string
	Port string
	Login string
	Password string
	Agent string

	IsUsing bool

	Ip string
	LocalIp string
	Log []string
}

func (p *Proxy) newProxy() *Browser {
	proxy := MYSQL.GetFreeProxy()
	if proxy.Id.Valid {
		p.Id = int(proxy.Id.Int64)
		p.Host = proxy.Host.String
		p.Port = proxy.Port.String
		p.Login = proxy.Login.String
		p.Password = proxy.Password.String
		p.Agent = proxy.Agent.String
		p.LocalIp = p.Host + ":" + p.Port

		return p.checkProxy()
	}
	return nil
}

func (p *Proxy) checkProxy() *Browser {
	browser := &Browser{}

	check, ctx, cancel := browser.Open(p)
	if !check {
		return nil
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
	var searchHtml string
	var videosHtml string

	browser.cancelTask = cancel

	if err := chromedp.Run(ctx,
		browser.runWithTimeOut(10, false, chromedp.Tasks{
			//chromedp.Navigate("https://www.google.com/search?q=" + UTILS.ArrayRand(keyWords)),
			//chromedp.WaitVisible("body", chromedp.ByQuery),
			//// Вытащить html на проверку каптчи
			//chromedp.OuterHTML("body", &searchHtml, chromedp.ByQuery),
			// Устанавливаем страницу для парсинга
			chromedp.Navigate("https://www.google.com/search?source=lnms&tbm=vid&as_sitesearch=youtube.com&num=50&q=" + UTILS.ArrayRand(keyWords)),
			chromedp.WaitVisible("#rso",chromedp.ByQuery),
			chromedp.OuterHTML("#rso", &videosHtml, chromedp.ByQuery),
			chromedp.Sleep(2222 * time.Second),
		}),
	); err != nil {
		log.Println("Browser.checkProxy.HasError", err)
		return nil
	}

	if searchHtml != "" {
		if browser.CheckCaptcha(searchHtml) {
			return nil
		}
		browser.ctx = ctx
		return browser
	}

	return nil
}

func (p *Proxy) setTimeout(parser int, minutes int) sql.Result {
	now := time.Now().Local().Add(time.Minute * time.Duration(minutes))
	formattedDate := now.Format("2006-01-02 15:04:05")

	data := map[string]interface{}{}
	data["parser"] = strconv.Itoa(parser)
	data["timeout"] = formattedDate

	res, err := MYSQL.UpdateProxy(data, p.Id)
	if err != nil {
		log.Println("Proxy.SetTimeout.HasError", err)
	}

	return res
}

func (p *Proxy) freeProxy() {
	now := time.Now().Local().Add(time.Minute * time.Duration(2))
	formattedDate := now.Format("2006-01-02 15:04:05")

	data := map[string]interface{}{}
	data["parser"] = "NULL"
	data["timeout"] = formattedDate

	_, err := MYSQL.UpdateProxy(data, p.Id)
	if err != nil {
		log.Println("Proxy.freeProxy.HasError", err)
	}
	p.Id = 0
	p.Host = ""
}