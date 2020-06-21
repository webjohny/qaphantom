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
	"github.com/chromedp/cdproto/target"
	"log"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosimple/slug"
	"github.com/webjohny/chromedtp"
)

type JobHandler struct {
	Task MysqlFreeTask

	NumberInits int
	SearchHtml string
	IsStart bool
	BrowserContextID cdp.BrowserContextID

	CancelBrowser context.CancelFunc
	CancelTimeout context.CancelFunc
	CancelLogger context.CancelFunc

	proxy Proxy
	
	interceptionID fetch.RequestID
	networkRequestID network.RequestID
	targetID target.ID
	sessionID target.SessionID

	ctx context.Context
	taskCtx context.Context
}

type QaSetting struct {
	H int
	Fast bool
	Question string
	Text string
	Html string
	Link string
	LinkTitle string
	Date string
	Ved string
	Length int
	Viewed bool
	Clicked bool
}

// Счётчик типов вопросов
type QaStats struct {
	All int // всего вопросов найдено
	Yt int // вопросы с youtube видео
	S int // простые текстовые ответы

	Qac int // кол-во блоков
	Size int // длина контента
	Length int // текущая длина контента по факту
	CycExit int // защита от бесконечного цикла

	Wqc int // полное количество блоков с вопросами
}

type QaTotalPage struct {
	Url string
	Title string
	Content string
	CatId int
	PhotoId int
}

type QaImageResult struct {
	Id int
	Url string
	UrlMedium string
	Author string
	ShortLink string
}

func (j *JobHandler) InitBrowser() bool {
	proxyScheme := j.proxy.LocalIp

	opts := append(
		chromedtp.DefaultExecAllocatorOptions[:],
		chromedtp.DisableGPU,
		chromedtp.NoSandbox,
		chromedtp.Flag("headless", false),
		chromedtp.Flag("ignore-certificate-errors", true),
	)

	if proxyScheme != "" {
		j.Task.SetLog("Подключаем прокси к браузеру (" + proxyScheme + ")")
		opts = append(opts, chromedtp.ProxyServer(proxyScheme))
	}

	// Запускаем контекст браузера
	allocCtx, cancel := chromedtp.NewExecAllocator(context.Background(), opts...)
	j.CancelBrowser = cancel

	// Устанавливаем собственный logger
	taskCtx, cancel := chromedtp.NewContext(allocCtx)
	j.CancelLogger = cancel

	// Ставим таймер на отключение если зависнет
	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	j.CancelTimeout = cancel

	j.ctx = taskCtx

	pattern := fetch.RequestPattern{}
	pattern.URLPattern = "*"

	var reqPatterns []*fetch.RequestPattern
	reqPatterns = append(reqPatterns, &pattern)

	fetchEnabled := fetch.Enable().WithPatterns(reqPatterns).WithHandleAuthRequests(true)

	if err := chromedtp.Run(taskCtx,
		network.Enable(),
		performance.Enable(),
		page.SetLifecycleEventsEnabled(true),
		security.SetIgnoreCertificateErrors(true),
		emulation.SetTouchEmulationEnabled(false),
		network.SetCacheDisabled(true),
		fetchEnabled,
		chromedtp.ActionFunc(func (ctx context.Context) error {
			j.ListenForNetworkEvent(ctx)
			return nil
		}),
	); err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (j *JobHandler) CancelJob() () {
	if j.CancelBrowser != nil {
		j.CancelBrowser()
	}
	if j.CancelLogger != nil {
		j.CancelLogger()
	}
	if j.CancelTimeout != nil {
		j.CancelTimeout()
	}
}

func (j *JobHandler) OpenPaa() () {

}

func (j *JobHandler) ListenForNetworkEvent(ctx context.Context) {
	chromedtp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *fetch.EventAuthRequired:
			c := chromedtp.FromContext(ctx)
			buf, _ := json.Marshal(map[string]interface{}{
				"requestId": ev.RequestID.String(),
				"authChallengeResponse": map[string]string{
					"response": "ProvideCredentials",
					"username": j.proxy.Login,
					"password": j.proxy.Password,
				},
			})
			cmd := &cdproto.Message{
				ID:        20,
				SessionID: c.Target.SessionID,
				Method:    cdproto.MethodType("Fetch.continueWithAuth"),
				Params:    buf,
			}
			err := c.Browser.Conn.Write(ctx, cmd)
			fmt.Println("fetch.EventAuthRequired", err)

		case *fetch.EventRequestPaused:
			j.interceptionID = ev.RequestID
			c := chromedtp.FromContext(ctx)
			buf, _ := json.Marshal(map[string]string{"requestId":ev.RequestID.String()})
			cmd := &cdproto.Message{
				ID:        20,
				SessionID: c.Target.SessionID,
				Method:    cdproto.MethodType("Fetch.continueRequest"),
				Params:    buf,
			}
			err := c.Browser.Conn.Write(ctx, cmd)
			fmt.Println("fetch.EventRequestPaused", err)
		}
	})
}

func (j *JobHandler) Run(parser int) (status bool, msg string) {
	if !j.IsStart {
		return false, "Задача закрыта"
	}
	var taskId int

	j.proxy = Proxy{}

	var fast QaSetting

	// Инициализация контроллера для управление парсингом
	if parser < 1 {
		parser = 1000
	}

	// Берём свободную задачу в работу
	task := mysql.GetFreeTask([]string{})
	if task.Id < 1 {
		return false, "Свободных задач нет в наличии"
	}
	taskId = task.Id
	j.Task = task

	if task.TryCount == 5 {
		task.SetLog("5-я неудавшаяся попытка парсинга. Исключаем ключевик")
		task.SetFinished(2, "Исключён после 5 попыток парсинга")
		return false, "Исключён после 5 попыток парсинга"
	}

	task.SetTimeout(parser)

	stats := QaStats{}
	stats.Wqc = task.QaCountFrom + task.QaCountTo

	if task.From != 0 && task.To != 0 {
		stats.Size = rand.Intn((task.To - task.From) + task.From)
	}else if task.QaCountFrom != 0 && task.QaCountTo != 0 {
		stats.Qac = rand.Intn((task.QaCountTo - task.QaCountFrom) + task.QaCountFrom)
	}

	var searchHtml string

	for i := 1; i < 2; i++ {
		// Подключаемся к прокси
		j.proxy.NewProxy()

		if j.proxy.Id < 1 {
			task.SetLog("Нет свободных прокси. Выход")
			task.FreeTask()
			return false, "Нет свободных прокси. Выход"
		}

		j.proxy.SetTimeout(parser)

		j.InitBrowser()

		// Запускаемся
		task.SetLog("Открываем страницу (попытка №" + strconv.Itoa(i) + "): https://www.google.com/search?hl=en&q=" + url.QueryEscape(task.Keyword))

		if err := chromedtp.Run(j.ctx,
			// Устанавливаем страницу для парсинга
			chromedtp.Navigate("https://www.google.com/search?hl=en&q=" + task.Keyword),
			chromedtp.WaitVisible("body", chromedtp.ByQuery),
			// Вытащить html на проверку каптчи
			chromedtp.OuterHTML("body", &searchHtml, chromedtp.ByQuery),
		); err != nil {
			//j.CancelJob()
			task.SetLog("Попытка №" + strconv.Itoa(i) + " провалилась.")
			log.Println(err)
		}else{
			break
		}
	}

	defer j.CancelJob()

	if j.CheckCaptcha(searchHtml) {
		task.SetError("Отсутствует PAA. Есть каптча...")
		j.proxy.FreeProxy()
		utils.ErrorHandler(chromedtp.Cancel(j.ctx))
		return false, "Отсутствует PAA. Есть каптча..."
	}

	if searchHtml == "" || !j.CheckPaa(searchHtml) {
		task.SetError("Отсутствует PAA.")
		j.proxy.FreeProxy()
		utils.ErrorHandler(chromedtp.Cancel(j.ctx))
		return false, "Отсутствует PAA."
	}

	task.SetLog("Блоки загружены")
	task.SetLog("Начинаем обработку PAA")

	// Загружаем HTML документ в GoQuery пакет который организует облегчённую работу с HTML селекторами
	j.SetFastAnswer(searchHtml)

	searchHtml = ""
	if err := chromedtp.Run(j.ctx,
		// Кликаем сразу на первый вопрос
		chromedtp.Click(".related-question-pair:first-child .cbphWd"),
		// Ждём 0.3 секунды чтобы открылся вопрос
		chromedtp.Sleep(time.Millisecond * 300),
	); err != nil {
		log.Println(err)
	}

	// Запускаем функцию перебора вопросов
	settings := j.ParsingPaa(&stats)
	task.SetLog("Обнаружено блоков: " + strconv.Itoa(stats.All))

	//if stats.Yt < stats.S && stats.S >= task.QaCountFrom {
	if stats.All == stats.Wqc {
		if stats.Yt > stats.S {
			var msg string
			if stats.Yt > stats.S {
				msg = "Кол-во блоков с видео (" + strconv.Itoa(stats.Yt) + ") превышает кол-во текстов (" + strconv.Itoa(stats.S) + ")."
			} else if stats.S < task.QaCountFrom { //task.QaCountFrom
				msg = "Кол-во блоков с текстом (" + strconv.Itoa(stats.S) + ") меньше значения S в настройках (" + strconv.Itoa(task.QaCountFrom) + ")."
			}

			// Завершение работы скрипта
			task.SetError(msg)
			j.proxy.FreeProxy()
			utils.ErrorHandler(chromedtp.Cancel(j.ctx))
			return false, msg

		} else if stats.S <= task.QaCountFrom {}
	}

	var mainEntity []map[string]interface{}

	microMarking := map[string]interface{}{
		"@context" : "https://schema.org",
		"@type" : "FAQPage",
		"mainEntity" : &mainEntity,
	}

	symb := task.GetRandSymb()

	for _, setting := range settings {
		if _, err := mysql.AddResult(map[string]interface{}{
			"a" : setting.Text,
			"q" : setting.Question,
			"task_id" : strconv.Itoa(task.Id),
			"link" : setting.Link,
			"link_title" : setting.LinkTitle,
		}); err != nil {
			log.Println(err)
		}

		if task.SavingAvailable {
			if _, err := mysql.AddTask(map[string]interface{}{
				"site_id" : strconv.Itoa(task.SiteId),
				"cat_id" : strconv.Itoa(task.CatId),
				"parent_id" : strconv.Itoa(task.Id),
				"keyword" : setting.Question,
				"parser" : "",
				"error" : "",
			}); err != nil {
				log.Println(err)
			}
		}

		text := setting.Text

		reg := regexp.MustCompile(`\s+`)
		text = reg.ReplaceAllString(text, ` `)

		matches := utils.PregMatch(`(?P<sen>.+?\.)`, text)
		if matches["sen"] != "" {
			text = matches["sen"]
		}else{
			text = setting.Text
		}
		text += "<a href='{{link}}#qa-" + slug.Make(setting.Question) + "'>" + task.GetRandTag() + "</a>"

		name := setting.Question
		if symb != "" {
			name = symb + name
		}
		mainEntity = append(mainEntity, map[string]interface{}{
			"@type" : "Question",
			"name" : name,
			"acceptedAnswer" : map[string]string{
				"@type" : "Answer",
				"text" : text,
			},
		})
	}

	if task.ParseSearch4 < 1 {
		qaTotalPage := QaTotalPage{}
		wp := Wordpress{}
		wp.Connect(`https://` + task.Domain + `/xmlrpc2.php`, task.Login, task.Password, 1)
		if !wp.CheckConn() {
			task.SetError("Не получилось подключится к wp xmlrpc (https://" + task.Domain + "/xmlrpc2.php - " + task.Login + " / " + task.Password + ")")
			task.SetError(wp.err.Error())
			j.proxy.FreeProxy()
			utils.ErrorHandler(chromedtp.Cancel(j.ctx))
			return false, "Не получилось подключится к wp xmlrpc (https://" + task.Domain + "/xmlrpc2.php - " + task.Login + " / " + task.Password + ")"
		}

		list := "ol"
		lists := map[string]string{"ul": "ol", "ol": "ul"}
		ch3 := 0

		var qaQs []QaSetting
		// Если есть быстрый ответ, ставим его первым
		if task.ParseFast > 0 && fast.Question != "" && task.H1 < 1 {
			qaQs = append(qaQs, fast)
		}
		for _, setting := range settings {
			qaQs = append(qaQs, setting)
		}

		// Пробегаем по блокам
		for k, q := range qaQs {
			// Чередуем типы списков
			if strings.Contains(q.Text, "<ul>") ||
				strings.Contains(q.Text, "<ol>"){
				q.Text = strings.ReplaceAll(q.Text, "<" + list + ">", "<" + lists[list] + ">")
				q.Text = strings.ReplaceAll(q.Text, "</" + list + ">", "</" + lists[list] + ">")
				list = lists[list]
			}

			// Если это первый блок в списке
			if k < 1 {
				q.H = 2
			} else if qaQs[k - 1].Fast { // Если предыдущи блок был быстрым ответом
				q.H = 2
				ch3 = 0
			}

			// Если есть подзаголовок
			if strings.Contains(q.Text, "<h3>") {
				q.H = 2
			}

			// Если размер заголовка ещё не определён
			if k > 0 && q.H < 1 {
				if qaQs[k - 1].H == 2 {
					q.H = 3
					ch3 = 1
				} else if ch3 < 2 {
					q.H = 3
					ch3 = 2
				} else if ch3 == 2 {
					if utils.RandBool() {
						q.H = 3
						ch3 = 3
					}else {
						q.H = 2
						ch3 = 0
					}
				} else if ch3 == 3 {
					q.H = 2
					ch3 = 0
				}
			}
		}

		// Вычисляем кол-во блоков
		qaCount := len(qaQs)
		var vCount int
		var vStep int

		// Если в настройках не задан шаг расстановки видео
		if task.VideoStep < 1 {
			// Вычисляем кол-во видео
			vCount = int(math.Floor(float64(stats.Length / 500)))
			if vCount < 1 {
				vCount = 1
			}
			// Шаг расстановки видео
			vStep = int(math.Floor(float64((qaCount - 2) / vCount)))
			if vStep < 1 {
				vStep = 1
			}
		} else {
			// Если в настройках задан шаг расстановки видео
			vStep = task.VideoStep
			vCount = int(math.Floor(float64((qaCount - 1) / vStep)))
		}

		task.SetLog("Парсим видео")

		// Парсим видео
		var videosHtml string
		if err := chromedtp.Run(j.ctx,
			// Устанавливаем страницу для парсинга
			chromedtp.Navigate("https://www.youtube.com/results?search_query=" + task.Keyword),
			// Ждём блоков с видео
			chromedtp.WaitVisible(`a.ytd-thumbnail`, chromedtp.ByQueryAll),
			// Вытащить html со списком
			chromedtp.OuterHTML("#contents.ytd-section-list-renderer", &videosHtml, chromedtp.ByQuery),
		); err != nil {
			log.Println(err)
		}

		var videos []string
		var lastVideo string

		if videosHtml != "" {
			videoReader := strings.NewReader(videosHtml)
			doc, err := goquery.NewDocumentFromReader(videoReader)
			if err != nil {
				log.Println(err)
			}

			// Начинаем перебор блоков с видео
			doc.Find("a.ytd-thumbnail").Each(func(i int, s *goquery.Selection) {
				if len(videos) != vCount {
					link, _ := s.Attr("href")
					videos = append(videos, utils.YoutubeEmbed(link))
					task.SetLog(link)
				}
			})

			if len(videos) > 0 {
				lastVideo, videos = videos[len(videos)-1], videos[:len(videos)-1]
			}
			task.SetLog("Парсинг видео. Готово")
		}

		// Заголовок
		//toDo $variants = $this->fconfig->get_variants();
		variants := []string{"Question: ", "Quick answer: ", ""}

		var h1 string
		if task.H1 < 1 || len(qaQs) < 1 {
			h1 = task.Keyword
		}else if len(qaQs) > 0 && qaQs[0].Question != ""{
			h1 = qaQs[0].Question
		}
		tmp := strings.Split(h1, " ")
		if len(tmp) > 0 {
			for k, v := range tmp {
				tmp[k] = strings.Title(v)
			}
		}
		rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
		variant := variants[rand.Intn(len(variants))]

		qaTotalPage.Title = variant + strings.Join(tmp, " ")

		var photo QaImageResult
		var mainImg string

		if task.PubImage < 1 {
			task.SetLog("Парсинг фото отключён настройками")
		}else{
			// Парсинг только по ключу
			if task.ImageKey == 2 {
				//$photo = empty($task->keyword)? null : (empty($task->image_source)? $this->_parse_photos($task->site_id, $task->keyword, false) : $this->_parse_photos($task->site_id, $task->keyword, false, $page));
			} else if task.ImageKey == 1 { // Парсинг только по теме
				//$photo = empty($task->theme)? null : (empty($task->image_source)? $this->_parse_photos($task->site_id, $task->theme, true) : $this->_parse_photos($task->site_id, $task->theme, true, $page));
			} else { // Парсинг сначала по ключу, потом по теме
				//$photo = empty($task->keyword)? null : (empty($task->image_source)? $this->_parse_photos($task->site_id, $task->keyword, false) : $this->_parse_photos($task->site_id, $task->keyword, false, $page));
			}

			if photo.Url == "" {
				//$photo = empty($task->theme)? null : (empty($task->image_source)? $this->_parse_photos($task->site_id, $task->theme, true) : $this->_parse_photos($task->site_id, $task->theme, true, $page));
			}

			// Добавляем фото в Вордпресс
			if photo.Url == "" {
				task.SetLog("Фото не найдено")
			} else {
				task.SetLog(photo.Url)

				// Загружаем фото в Вордпресс
				//TOdO $tmp = $this->xmlrpc->photo($photo['url'], $task->keyword);
				res := wp.UploadFile("", "", "", "", 0)

				task.SetLog("Фото загружено на сайт")

				// Обрабатываем результат добавления фото в Вордпресс
				qaTotalPage.PhotoId = res["id"].(int)
				photo.Url = res["url"].(string)
				photo.UrlMedium = res["url_medium"].(string)

				// Готовим код вставки фото в текст
				if task.PubImage >= 2 {
					mainImg = `<p><img class="alignright size-medium" src="` + photo.UrlMedium + `"></p>` + "\n"
				}

			}
		}
		// Пробегаемся по всем блокам
		for k, item := range qaQs{
			// Подзаголовок
			if task.ShFormat > 0 {
				item.Text = strings.ReplaceAll(item.Text, "<h3>", "<strong>")
				item.Text = strings.ReplaceAll(item.Text, "</h3>", "</strong>")
			}

			//	// Вставляем видео в текст
			if task.VideoStep < 1 {
				if k == (qaCount - 2) {
					if lastVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + lastVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>` + "\n"
					}
				}else if len(videos) > 0 && k > 0 && k < (qaCount - 2) && k % vStep == 0 {
					firstVideo, _ := videos[0], videos[1:]
					if firstVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + firstVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>` + "\n"
					}
				}
			} else {
				if k == qaCount - 1 {
					if lastVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + lastVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>` + "\n"
					}
				}else if len(videos) > 0 && k > 0 && k < (qaCount - 1) && k % vStep == 0 {
					firstVideo, _ := videos[0], videos[1:]
					if firstVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + firstVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>` + "\n"
					}
				}
			}

			// Заголовок
			if item.Question != "" {
				qaTotalPage.Content += `<span id="qa-` + slug.Make(item.Question) + `"></span>`
				if task.H1 < 1 || k > 0 {
					if task.ShOrder < 1 {
						qaTotalPage.Content += `<h` + strconv.Itoa(item.H) + `>` + item.Question + `</h` + strconv.Itoa(item.H) + ">\n"
					} else {
						qaTotalPage.Content += `<h2>` + item.Question + "</h2>\n"
					}
				}
			}

			// Если ответ первый
			if k < 1 {
				// Вставляем фото
				qaTotalPage.Content += mainImg
				// Ответ разбиваем по предложениям
				if !strings.Contains(item.Text, "<ul>") && !strings.Contains(item.Text, "<ol>") && !strings.Contains(item.Text, "<h3>") {
					formattedText := utils.StripTags(item.Text)
					splittedText := utils.SentenceSplit(formattedText)
					qaTotalPage.Content += "<p>" + strings.Join(splittedText, ".</p><p>") + ".</p>\n"
				} else {
					// Просто ставим ответ
					qaTotalPage.Content += item.Text + "\n"
				}
			} else {
				// Просто ставим ответ
				qaTotalPage.Content += item.Text + "\n"
			}

			// Дата
			if task.ParseDates > 0 && item.Date != "" {
				qaTotalPage.Content += `<div id="qa_date">Date: ` + item.Date + "</div>\n"
			}

			// Ссылка
			if task.Linking > 0 && item.Link != "" {
				if task.Linking == 2 {
					qaTotalPage.Content += `<p>Source: <a href="` + item.Link + `" target="_blank">` + item.LinkTitle + "</a></p>\n"
				} else {
					qaTotalPage.Content += `<p>Source: <code>` + item.Link + "</code></p>\n"
				}
			}
		}

		// Добавляем копирайт автора фото в конце статьи
		if photo.Author != "" || photo.ShortLink != "" {
			qaTotalPage.Content += "<p>"
			if photo.Author != "" {
				qaTotalPage.Content += `Photo in the article by “` + photo.Author + `” `
			}
			if photo.ShortLink != "" {
				qaTotalPage.Content += `<code>` + photo.ShortLink + `</code>`
			}
			qaTotalPage.Content += "</p>\n"
		}

		task.SetLog("Текст статьи подготовлен")

		// Сохраняем текст
		task.SetLog("Текст статьи сохранён в БД")
		if (task.QaCountFrom > 0 && len(qaQs) < task.QaCountFrom) || (task.From > 0 && utf8.RuneCountInString(qaTotalPage.Content) < task.From) {
			task.SetError("Снята с публикации — слишком короткая статья получилась")
			j.proxy.FreeProxy()
			utils.ErrorHandler(chromedtp.Cancel(j.ctx))
			return false, "Снята с публикации — слишком короткая статья получилась"
		}

		// Определяем ID категории
		qaTotalPage.CatId = wp.CatIdByName(task.Cat)
		if qaTotalPage.CatId < 1 {
			task.SetError("Проблема с размещением в рубрику")
			task.SetError(wp.err.Error())
			j.proxy.FreeProxy()
			utils.ErrorHandler(chromedtp.Cancel(j.ctx))
			return false, "Проблема с размещением в рубрику"
		}

		// Отправляем заметку на сайт
		postId := wp.NewPost(qaTotalPage.Title, qaTotalPage.Content, qaTotalPage.CatId, qaTotalPage.PhotoId)
		var fault bool
		if postId > 0 {
			post := wp.GetPost(postId)
			if post.Id > 0 {
				jsonMarking, _ := json.Marshal(microMarking)
				qaTotalPage.Content += `<script type="application/ld+json">`
				qaTotalPage.Content += strings.ReplaceAll(string(jsonMarking), "{{link}}", post.Link)
				qaTotalPage.Content += `</script>`

				wp.EditPost(postId, qaTotalPage.Title, qaTotalPage.Content)
			}else{
				fault = true
			}
		}else{
			fault = true
		}

		if fault {
			task.SetLog("Не получилось разместить статью на сайте")
			j.proxy.FreeProxy()
			task.SetError(wp.err.Error())
			utils.ErrorHandler(chromedtp.Cancel(j.ctx))
			return false, "Не получилось разместить статью на сайте"
		}

		task.SetLog("Статья размещена на сайте")
	}else{
		task.SetLog(`Данные сохранены в "Search for"`)
	}
	task.SetFinished(1, "")
	j.proxy.FreeProxy()
	fmt.Println(taskId)

	return true, "Задача #" + strconv.Itoa(taskId) + " была успешно выполнена"
}

func (j *JobHandler) ParsingPaa(stats *QaStats) map[string]QaSetting {
	var paaHtml string
	settings := map[string]QaSetting{}

	// Вытягиваем html код PAA для парсинга вопросов
	if err := chromedtp.Run(j.ctx,
		chromedtp.OuterHTML(`.kno-kp .ifM9O`, &paaHtml, chromedtp.ByQuery),
	); err != nil {
		log.Println(err)
	}

	// Загружаем HTML документ в GoQuery пакет который организует облегчённую работу с HTML селекторами
	paaReader := strings.NewReader(paaHtml)
	doc, err := goquery.NewDocumentFromReader(paaReader)
	if err != nil {
		log.Println(err)
	}

	var tasks chromedtp.Tasks
	clicked := map[string]bool{}
	// Начинаем перебор блоков с вопросами
	doc.Find(".related-question-pair").Each(func(i int, s *goquery.Selection) {
		question := s.Find(".cbphWd").Text()
		link, _ := s.Find(".g a").Attr("href")

		// Ищем дату в блоке, она может быть или в div (если вне текста) или в span (если внутри текста)
		date := s.Find(".kX21rb").Text()
		if date == "" {
			date = s.Find(".Od5Jsd").Text()
		}
		text := strings.Replace(s.Find(".mod").Text(), date, "", -1)
		txtTtml, _ := s.Find(".mod").Html()

		if j.Task.ParseDoubles > 0 || !mysql.GetResultByQAndA(question, text).Id.Valid {
			// Берём уникальный идентификатор для вопроса
			stats.All++
			ved, _ := s.Find(".cbphWd").Attr("data-ved")
			if question != "" {
				qa := QaSetting{}
				qa.Question = question
				qa.Text = text
				qa.Html = txtTtml
				qa.Link = link
				qa.LinkTitle = s.Find(".g a").Text()
				qa.Date = date
				qa.Length = utf8.RuneCountInString(text) + utf8.RuneCountInString(question)
				qa.Ved = ved
				qa.Viewed = true

				stats.Length += qa.Length

				// Собираем задачи для кликинга по вопросам
				if _, ok := settings[ved]; !ok {
					tasks = append(tasks, chromedtp.Click(".cbphWd[data-ved=\""+ved+"\"]"))
					tasks = append(tasks, chromedtp.Sleep(time.Millisecond*500))
					clicked[ved] = true
				}

				if strings.Contains(txtTtml, "youtube.com/watch") || strings.Contains(txtTtml, "Suggested clip") {
					stats.Yt++
				}else{
					stats.S++
					settings[ved] = qa
				}
			}
		}
	})

	// Проверяем есть ли уже достаточное количество вопросов или всё таки нужно продолжить кликинг по блокам
	if stats.All < stats.Wqc && len(tasks) > 0 {
		if err := chromedtp.Run(j.ctx, tasks); err != nil {
			log.Println(err)
		}
		for k, v := range clicked {
			if setting, ok := settings[k]; ok {
				setting.Clicked = v
			}
		}
		// Продолжаем рекурсию
		return j.ParsingPaa(stats)
	}

	return settings
}

func (j *JobHandler) SetFastAnswer(html string) {
	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		log.Println(err)
	}

	fastSelector := doc.Find(".kp-blk.c2xzTb")
	fastHtml, _ := fastSelector.Html()
	if !strings.Contains("youtube.com", fastHtml) {

	}
}

func (j *JobHandler) CheckPaa(html string) bool {
	return strings.Contains(html,"JolIg") && strings.Contains(html,"related-question-pair")
}

func (j *JobHandler) CheckCaptcha(html string) bool {
	return strings.Contains(html,"g-recaptcha") && strings.Contains(html,"data-sitekey")
}