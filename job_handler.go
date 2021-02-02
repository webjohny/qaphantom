package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"gopkg.in/masci/flickr.v2"
	"log"
	"math"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosimple/slug"
	"github.com/webjohny/chromedp"
)

type JobHandler struct {
	NumberInits int
	SearchHtml string
	IsStart bool

	taskId int
	config MysqlConfig
	task MysqlFreeTask
	proxy Proxy

	Browser Browser
	ctx context.Context
	isFinished chan bool

	CancelTimeout context.CancelFunc
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

type QaFast struct {
	H int
	A string
	Link string
	LinkTitle string
	Fast bool
	Date string
	Length int
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
	Id string
	Url string
	UrlMedium string
	Author string
	ShortLink string
	Encoded bool
}

type XmlPhotos struct {
	XMLName xml.Name `xml:"photos"`
	Photos []struct {
		Id  int `xml:"id,attr"`
		OwnerName  string `xml:"ownername,attr"`
		Url string `xml:"url_z,attr"`
	} `xml:"photo"`
}

func (j *JobHandler) Run(parser int) (status bool, msg string) {
	if !j.IsStart {
		go j.Cancel()
		return false, "Задача закрыта"
	}

	fmt.Println("Start task")
	j.SetTimeout(150)
	//if j.ctx.Err().Error() != "" {
	//	fmt.Println("Stop browser", j.ctx.Err().Error())
	//	return false, "Browser failed"
	//}

	var taskId int

	//var fast QaFast

	// Берём свободную задачу в работу
	var task MysqlFreeTask
	if j.taskId < 1 {
		task = MYSQL.GetFreeTask(0)
	}else{
		task = MYSQL.GetFreeTask(j.taskId)
	}

	//log.Fatal(task)

	if task.Id < 1 {
		go j.Cancel()
		return false, "Свободных задач нет в наличии"
	}
	taskId = task.Id
	task.Domain = task.GetRandDomain()
	//task.SetLog("Задача #" + strconv.Itoa(taskId) + " с запросом (" + task.Keyword + ") взята в работу")

	j.task = task

	if j.CheckFinished() {
		task.SetLog("Задача завершилась преждевременно из-за таймаута")
		return false, "Timeout"
	}

	if task.TryCount == 5 {
		task.SetLog("5-я неудавшаяся попытка парсинга. Исключаем ключевик")
		task.SetFinished(2, "Исключён после 5 попыток парсинга")
		go j.Cancel()
		return false, "Исключён после 5 попыток парсинга"
	}

	task.SetLog("Подключаем прокси #" + strconv.Itoa(j.Browser.Proxy.Id) + " к браузеру (" + j.Browser.Proxy.LocalIp + ")")

	task.SetTimeout(parser)

	stats := QaStats{}
	stats.Wqc = task.QaCountFrom + task.QaCountTo

	if task.From != 0 && task.To != 0 {
		stats.Size = rand.Intn((task.To - task.From) + task.From)
	}else if task.QaCountFrom != 0 && task.QaCountTo != 0 {
		stats.Qac = rand.Intn((task.QaCountTo - task.QaCountFrom) + task.QaCountFrom)
	}

	var searchHtml string
	var googleUrl string

	j.config = MYSQL.GetConfig()

	for i := 1; i < 2; i++ {

		if j.CheckFinished() {
			task.SetLog("Задача завершилась преждевременно из-за таймаута")
			return false, "Timeout"
		}

		// Запускаемся
		googleUrl = "https://www.google.com/search?hl=en&q=" + url.QueryEscape(task.Keyword)
		task.SetLog("Открываем страницу (попытка №" + strconv.Itoa(i) + "): " + googleUrl)

		if j.ctx != nil {
			if err := chromedp.Run(j.ctx,
				// Устанавливаем страницу для парсинга
				//chromedp.Sleep(time.Second * 60),
				chromedp.Navigate(googleUrl),
				chromedp.Sleep(time.Second*time.Duration(rand.Intn(10))),
				chromedp.WaitVisible("body", chromedp.ByQuery),
				// Вытащить html на проверку каптчи
				chromedp.OuterHTML("body", &searchHtml, chromedp.ByQuery),
			); err != nil {
				log.Println("JobHandler.Run.HasError", err)
				task.SetLog("Попытка №" + strconv.Itoa(i) + " провалилась. (" + err.Error() + ")")
				continue
			} else {
				if j.CheckCaptcha(searchHtml) {
					task.SetLog("Есть каптча для " + j.Browser.Proxy.LocalIp + "...")
					//j.proxy.SetTimeout(parser, 500)
					//j.proxy.LocalIp = ""
					//j.Cancel()
					//continue
					//key, _ := j.AntiCaptcha(googleUrl, searchHtml)
					key := ""
					if key != "" {
						task.SetLog("Anticaptcha: " + key)
						if err := chromedp.Run(j.ctx,
							chromedp.WaitVisible("captcha-form", chromedp.ByID),
							chromedp.WaitVisible("g-recaptcha-response", chromedp.ByID),
							chromedp.ActionFunc(func(ctx context.Context) error {
								fmt.Println("Yes")
								return nil
							}),
							chromedp.SetValue(`g-recaptcha-response`, key, chromedp.ByID),
							chromedp.Submit(`captcha-form`, chromedp.ByID),
							chromedp.Sleep(time.Second*5),
							chromedp.WaitVisible("body", chromedp.ByQuery),
							chromedp.OuterHTML("body", &searchHtml, chromedp.ByQuery),
						); err != nil {
							log.Println("JobHandler.Run.2.HasError", err)
							task.SetLog("Попытка №" + strconv.Itoa(i) + " провалилась. (" + err.Error() + ")")
							continue
						}
						if searchHtml != "" {
							f, err := os.Create("/var/www/example.txt")
							if err != nil {
								fmt.Println(err)
							}
							d2 := []byte(searchHtml)
							n2, err := f.Write(d2)
							if err != nil {
								fmt.Println("JobHandler.Run.3", n2, err)
								_ = f.Close()
							}
						}
					} else {
						task.SetLog("Антикаптча не сработала для " + j.Browser.Proxy.LocalIp + "...")
						return false, "Каптча"
					}
				}

				if !j.CheckPaa(searchHtml) {
					task.SetError("Отсутствует PAA.")
					go j.Cancel()
					return false, "Отсутствует PAA."
				}
				break
			}
		}else{
			task.SetLog("Браузер не был запущен. Задача пропускается.")
			go j.Cancel()
			return false, "Context undefined"
		}
	}

	if j.CheckFinished() {
		task.SetLog("Задача завершилась преждевременно из-за таймаута")
		j.Cancel()
		return false, "Timeout"
	}

	if searchHtml == "" {
		j.Cancel()
		task.SetLog("Контент не подгрузился, задачу закрываем")
		return
	}

	task.SetLog("Блоки загружены")
	task.SetLog("Начинаем обработку PAA")

	var fast QaFast
	if task.ParseFast > 0{
		// Загружаем HTML документ в GoQuery пакет который организует облегчённую работу с HTML селекторами
		fast = j.SetFastAnswer(searchHtml)
	}
	log.Println(fast)

	searchHtml = ""

	duration := time.Duration(rand.Intn(8))
	if !j.config.GetExtra().DeepPaa {
		duration = time.Duration(1)
	}

	if j.ctx != nil {
		if err := chromedp.Run(j.ctx,
			// Кликаем сразу на первый вопрос
			chromedp.Click(".related-question-pair:first-child .cbphWd"),
			// Ждём 0.3 секунды чтобы открылся вопрос
			chromedp.Sleep(time.Second*duration),
		); err != nil {
			log.Println("JobHandler.Run.4.HasError", err)
			task.SetError("Отсутствует PAA. (" + err.Error() + ")")
			go j.Cancel()
			return false, "Отсутствует PAA."
		}
	}else{
		task.SetLog("Браузер не был запущен. Задача пропускается.")
		go j.Cancel()
		return false, "Context undefined"
	}

	if j.CheckFinished() {
		task.SetLog("Задача завершилась преждевременно из-за таймаута")
		go j.Cancel()
		return false, "Timeout"
	}

	var settings map[string]QaSetting
	if j.ctx != nil {
		// Запускаем функцию перебора вопросов
		if j.config.GetExtra().RedirectMethod {
			settings = j.RedirectParsing(&stats)
		}else{
			settings = j.ClickParsing(&stats)
		}
	}else{
		task.SetLog("Браузер не был запущен. Задача пропускается.")
		go j.Cancel()
		return false, "Context undefined"
	}

	if j.CheckFinished() {
		j.task.SetLog("Задача завершилась преждевременно из-за таймаута")
		go j.Cancel()
		return false, "Timeout"
	}

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
		if _, err := MYSQL.AddResult(map[string]interface{}{
			//"a" : setting.Text,
			"cat_id" : task.CatId,
			"site_id" : task.SiteId,
			"cat" : task.Cat,
			"domain" : task.Domain,
			"q" : setting.Question,
			"task_id" : strconv.Itoa(task.Id),
			"link" : setting.Link,
			"link_title" : setting.LinkTitle,
		}); err != nil {
			log.Println("JobHandler.Run.5.HasError", err)
			task.SetLog("Не сохранился результат. (" + err.Error() + ")")
		}

		if task.SavingAvailable && !MYSQL.GetTaskByKeyword(setting.Question).Id.Valid {
			if _, err := MYSQL.AddTask(map[string]interface{}{
				"site_id" : strconv.Itoa(task.SiteId),
				"cat_id" : strconv.Itoa(task.CatId),
				"parent_id" : strconv.Itoa(task.Id),
				"keyword" : setting.Question,
				"parser" : strconv.Itoa(parser),
				"error" : "",
			}); err != nil {
				log.Println("JobHandler.Run.6.HasError", err)
				task.SetLog("Не добавилась новая задача. (" + err.Error() + ")")
			}
		}

		text := setting.Text

		reg := regexp.MustCompile(`\s+`)
		text = reg.ReplaceAllString(text, ` `)

		matches := UTILS.PregMatch(`(?P<sen>.+?\.)`, text)
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
			task.SetLog("Не получилось подключится к wp xmlrpc (https://" + task.Domain + "/xmlrpc2.php - " + task.Login + " / " + task.Password + ")")
			task.SetError(wp.err.Error())
			go j.Cancel()
			return false, "Не получилось подключится к wp xmlrpc (https://" + task.Domain + "/xmlrpc2.php - " + task.Login + " / " + task.Password + ")"
		}

		list := "ol"
		lists := map[string]string{"ul": "ol", "ol": "ul"}
		ch3 := 0

		var qaQs []QaSetting
		// Если есть быстрый ответ, ставим его первым
		//if task.ParseFast > 0 && setting.Question != "" && task.H1 < 1 {
		//	qaQs = append(qaQs, setting)
		//}
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
					if UTILS.RandBool() {
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

		// Заголовок
		variants := j.config.GetVariants()

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
		var variant string
		if len(variants) > 0 {
			variant = variants[rand.Intn(len(variants))]
		}

		qaTotalPage.Title = variant + strings.Join(tmp, " ")

		var photo QaImageResult
		var mainImg string

		if task.PubImage < 1 {
			task.SetLog("Парсинг фото отключён настройками")
		}else{
			// Парсинг только по ключу
			if task.ImageKey == 2 {
				task.SetLog("Парсинг фото только по ключу")
				if task.Keyword != "" {
					if task.ImageSource == 0 {
						photo = j.ParsePhotos(task.Theme, "flickr", false)
					} else if task.ImageSource == 1 {
						photo = j.ParsePhotos(task.Keyword, "google", false)
					}
				}
			} else if task.ImageKey == 1 { // Парсинг только по теме
				task.SetLog("Парсинг фото только по теме")
				if task.Theme != "" {
					if task.ImageSource == 0 {
						photo = j.ParsePhotos(task.Theme, "flickr", true)
					} else if task.ImageSource == 1 {
						photo = j.ParsePhotos(task.Theme, "google", true)
					}
				}
			} else { // Парсинг сначала по ключу, потом по теме
				task.SetLog("Парсинг фото сначала по ключу, потом по теме")
				if task.ImageSource == 0 {
					photo = j.ParsePhotos(task.Keyword, "flickr", false)
				}else if task.ImageSource == 1 {
					photo = j.ParsePhotos(task.Keyword, "google", false)
				}
				if photo.Id == "" {
					if task.ImageSource == 0 {
						photo = j.ParsePhotos(task.Theme, "flickr", true)
					}else if task.ImageSource == 1 {
						photo = j.ParsePhotos(task.Theme, "google", true)
					}
				}
			}

			// Добавляем фото в Вордпресс
			if photo.Url == "" {
				task.SetLog("Фото не найдено")
			} else {
				task.SetLog("Новое фото")

				// Загружаем фото в Вордпресс
				res, _ := wp.UploadFile(photo.Url, 0, photo.Encoded)

				log.Println(res)

				if res.Url != "" {
					task.SetLog("Фото загружено на сайт")

					// Обрабатываем результат добавления фото в Вордпресс
					qaTotalPage.PhotoId = res.Id
					photo.Url = res.Url
					photo.UrlMedium = res.UrlMedium

					// Готовим код вставки фото в текст
					if task.PubImage >= 2 {
						mainImg = `<p><img class="alignright size-medium" src="` + photo.UrlMedium + `"></p>` + "\n"
					}
				}else if photo.Url != "" {
					task.SetLog("Фото (" + photo.Url + ") не загрузилось на сайт")
				}else {
					task.SetLog("Фото не найдено для поста")
				}
			}
		}


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

		if j.CheckFinished() {
			task.SetLog("Задача завершилась преждевременно из-за таймаута")
			go j.Cancel()
			return false, "Timeout"
		}

		// Парсим видео
		var videosHtml string
		if j.ctx != nil {
			if err := chromedp.Run(j.ctx,
				chromedp.Sleep(time.Second*time.Duration(rand.Intn(10))),
				// Устанавливаем страницу для парсинга
				chromedp.Navigate("https://www.youtube.com/results?search_query="+task.Keyword),
				// Вытащить html со списком
				chromedp.Sleep(time.Second*4),
				chromedp.OuterHTML("body", &videosHtml, chromedp.ByQuery),
			); err != nil {
				log.Println("JobHandler.Run.7.HasError", err)
				task.SetLog("Видео не спарсилось. (" + err.Error() + ")")
			}
		}else{
			task.SetLog("Браузер не был запущен. Задача пропускается.")
			go j.Cancel()
			return false, "Context undefined"
		}

		if j.CheckFinished() {
			task.SetLog("Задача завершилась преждевременно из-за таймаута")
			go j.Cancel()
			return false, "Timeout"
		}

		var videos []string
		var lastVideo string

		if videosHtml != "" {
			videoReader := strings.NewReader(videosHtml)
			doc, err := goquery.NewDocumentFromReader(videoReader)
			if err != nil {
				log.Println("JobHandler.Run.8.HasError", err)
				task.SetLog("Неразборчивый код из youtube. (" + err.Error() + ")")
			}

			// Начинаем перебор блоков с видео
			doc.Find("#contents.ytd-section-list-renderer").Find("a.ytd-thumbnail").Each(func(i int, s *goquery.Selection) {
				if len(videos) != vCount {
					link, _ := s.Attr("href")
					videos = append(videos, UTILS.YoutubeEmbed(link))
					task.SetLog(link)
				}
			})

			if len(videos) > 0 {
				lastVideo, videos = videos[len(videos)-1], videos[:len(videos)-1]
			}
			task.SetLog("Парсинг видео. Готово")
		}


		// Пробегаемся по всем блокам
		for k, item := range qaQs{
			// Подзаголовок
			if task.ShFormat > 0 {
				item.Text = strings.ReplaceAll(item.Text, "<h3>", "<strong>")
				item.Text = strings.ReplaceAll(item.Text, "</h3>", "</strong>")
			}

			var firstVideo string
			// Вставляем видео в текст
			if task.VideoStep < 1 {
				if k == (qaCount - 2) {
					if lastVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + lastVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>` + "\n"
					}
				}else if len(videos) > 0 && k > 0 && k < (qaCount - 2) && k % vStep == 0 {
					firstVideo, videos = videos[0], videos[1:]
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
					firstVideo, videos = videos[0], videos[1:]
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
					formattedText := UTILS.StripTags(item.Text)
					splittedText := UTILS.SentenceSplit(formattedText)
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

		qaTotalPage.Content = strings.ReplaceAll(qaTotalPage.Content, "<p>…</p>", "")

		task.SetLog("Текст статьи подготовлен")

		// Сохраняем текст
		task.SetLog("Текст статьи сохранён в БД")
		if (task.QaCountFrom > 0 && len(qaQs) < task.QaCountFrom) || (task.From > 0 && utf8.RuneCountInString(qaTotalPage.Content) < task.From) {
			task.SetError("Снята с публикации — слишком короткая статья получилась")
			go j.Cancel()
			return false, "Снята с публикации — слишком короткая статья получилась"
		}

		if j.CheckFinished() {
			task.SetLog("Задача завершилась преждевременно из-за таймаута")
			return false, "Timeout"
		}

		// Определяем ID категории
		qaTotalPage.CatId = wp.CatIdByName(task.Cat)
		if qaTotalPage.CatId < 1 {
			task.SetLog("Проблема с размещением в рубрику")
			task.SetError(wp.err.Error())
			go j.Cancel()
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
			task.SetError(wp.err.Error())
			go j.Cancel()
			return false, "Не получилось разместить статью на сайте"
		}

		task.SetLog("Статья размещена на сайте")
	}else{
		task.SetLog(`Данные сохранены в "Search for"`)
	}
	task.SetFinished(1, "")
	fmt.Println(taskId)
	go j.Cancel()
	return true, "Задача #" + strconv.Itoa(taskId) + " была успешно выполнена"
}

func (j *JobHandler) RedirectParsing(stats *QaStats) map[string]QaSetting {
	var paaHtml string
	settings := map[string]QaSetting{}

	if j.CheckFinished() {
		j.IsStart = false
		j.task.SetLog("Задача завершилась преждевременно из-за таймаута")
		return settings
	}

	for i := 0; i < stats.Wqc; i++ {
		// Вытягиваем html код PAA для парсинга вопросов
		if err := chromedp.Run(j.ctx,
			chromedp.OuterHTML(`.kno-kp .ifM9O`, &paaHtml, chromedp.ByQuery),
		); err != nil {
			log.Println("JobHandler.RedirectParsing.HasError", err)
			return settings
		}

		// Загружаем HTML документ в GoQuery пакет который организует облегчённую работу с HTML селекторами
		paaReader := strings.NewReader(paaHtml)
		doc, err := goquery.NewDocumentFromReader(paaReader)
		if err != nil {
			log.Println("JobHandler.RedirectParsing.2.HasError", err)
			return settings
		}

		var lastQuestion string

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

			if j.task.ParseDoubles > 0 || !MYSQL.GetResultByQAndA(question, text).Id.Valid {
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

					lastQuestion = question

					stats.Length += qa.Length

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
		if stats.All < stats.Wqc {
			// Вытягиваем html код PAA для парсинга вопросов
			if err := chromedp.Run(j.ctx,
				chromedp.Navigate("https://www.google.com/search?hl=en&q=" + url.QueryEscape(lastQuestion)),
			); err != nil {
				log.Println("JobHandler.RedirectParsing.HasError", err)
				return settings
			}
		}else{
			break
		}
	}

	return settings
}

func (j *JobHandler) ClickParsing(stats *QaStats) map[string]QaSetting {
	var paaHtml string
	settings := map[string]QaSetting{}

	if j.CheckFinished() {
		j.IsStart = false
		j.task.SetLog("Задача завершилась преждевременно из-за таймаута")
		return settings
	}

	if j.ctx == nil {
		j.IsStart = false
		j.task.SetLog("Браузер не был запущен. Задача пропускается.")
		return settings
	}

	// Вытягиваем html код PAA для парсинга вопросов
	if err := chromedp.Run(j.ctx,
		chromedp.OuterHTML(`.kno-kp .ifM9O`, &paaHtml, chromedp.ByQuery),
	); err != nil {
		log.Println("JobHandler.ClickParsing.HasError", err)
		return settings
	}

	// Загружаем HTML документ в GoQuery пакет который организует облегчённую работу с HTML селекторами
	paaReader := strings.NewReader(paaHtml)
	doc, err := goquery.NewDocumentFromReader(paaReader)
	if err != nil {
		log.Println("JobHandler.ClickParsing.2.HasError", err)
		return settings
	}

	var tasks chromedp.Tasks
	clicked := map[string]bool{}
	// Начинаем перебор блоков с вопросами
	doc.Find(".related-question-pair").Each(func(i int, s *goquery.Selection) {
		question := s.Find(".cbphWd").Text()
		link, _ := s.Find(".g a").Attr("href")
		isExpanded :=  s.Find(".UgLoB").Length() > 0

		// Ищем дату в блоке, она может быть или в div (если вне текста) или в span (если внутри текста)
		date := s.Find(".kX21rb").Text()
		if date == "" {
			date = s.Find(".Od5Jsd").Text()
		}
		text := strings.Replace(s.Find(".mod").Text(), date, "", -1)
		txtTtml, _ := s.Find(".mod").Html()

		if j.task.ParseDoubles > 0 || !MYSQL.GetResultByQAndA(question, text).Id.Valid {
			// Берём уникальный идентификатор для вопроса
			stats.All++
			ved, _ := s.Find(".cbphWd").Attr("data-ved")
			if question != "" {
				qa := QaSetting{}
				qa.Question = question
				qa.Text = text
				qa.Html = txtTtml
				qa.Link = link
				qa.LinkTitle = s.Find(".g .LC20lb").Text()
				qa.Date = date
				qa.Length = utf8.RuneCountInString(text) + utf8.RuneCountInString(question)
				qa.Ved = ved
				qa.Viewed = true

				stats.Length += qa.Length

				// Собираем задачи для кликинга по вопросам
				if _, ok := settings[ved]; !ok {
					if isExpanded {
						tasks = append(tasks, chromedp.Click(".cbphWd[data-ved=\""+ved+"\"]"))
						tasks = append(tasks, chromedp.Sleep(time.Second*time.Duration(rand.Intn(5))))
						clicked[ved] = true
					}
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
	if stats.All < stats.Wqc && len(tasks) > 0 && j.config.GetExtra().DeepPaa {
		dest := make(chromedp.Tasks, len(tasks))
		perm := rand.Perm(len(tasks))
		for i, v := range perm {
			dest[v] = tasks[i]
		}
		if cap(dest) > 0 {
			if err := chromedp.Run(j.ctx, dest); err != nil {
				log.Println("JobHandler.ClickParsing.3.HasError", err)
				return settings
			}
			for k, v := range clicked {
				if setting, ok := settings[k]; ok {
					setting.Clicked = v
				}
			}
			// Продолжаем рекурсию
			return j.ClickParsing(stats)
		}
	}

	return settings
}

func (j *JobHandler) SetFastAnswer(html string) QaFast {
	var fast QaFast

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		log.Println("JobHandler.SetFastAnswer.HasError", err)
	}
	doc.Find(".bNg8Rb").Remove()

	fastSelector := doc.Find(".kp-blk.c2xzTb")
	fastHtml, _ := fastSelector.Html()
	if !strings.Contains("youtube.com", fastHtml) {

	}
	return fast
}

func (j *JobHandler) ParsePhotos(keyword string, imageSource string, cache bool) QaImageResult {
	var result QaImageResult

	img := MYSQL.GetImgTakeFree(j.task.SiteId, keyword, true)
	if img.Id.Valid {
		result.Url = img.Url.String
		result.Author = img.Author.String
		result.ShortLink = img.ShortUrl.String
		result.Id = img.SourceId.String
		return result
	}

	if imageSource == "flickr" {
		client := flickr.NewFlickrClient(j.config.FlickrKey.String, j.config.FlickrSecret.String)
		client.Init()
		client.Args.Set("method", "flickr.photos.search")
		client.Args.Set("text", keyword)
		client.Args.Set("media", "photos")
		client.Args.Set("extras", "url_z,owner_name")
		client.Args.Set("sort", "relevance")
		client.Args.Set("orientation", "landscape")
		client.Args.Set("license", "4")
		client.Args.Set("per_page", "5")
		client.Args.Set("page", "1")

		client.OAuthSign()
		response := &flickr.BasicResponse{}
		err := flickr.DoGet(client, response)

		if err != nil {
			fmt.Println("JobHandler.ParsePhotos.Flickr.HasError", err)
		} else {
			fmt.Println("Api response:", response.Extra)
		}

		var resp XmlPhotos

		err = xml.Unmarshal([]byte(response.Extra), &resp)
		if err != nil {
			log.Println("JobHandler.ParsePhotos.Flickr.HasError.1", err)
		}
		if len(resp.Photos) > 0 {
			for _, item := range resp.Photos {
				result.Id = strconv.Itoa(item.Id)
				result.Url = item.Url
				result.Author = item.OwnerName

				if cache {
					_, err := MYSQL.AddImg(map[string]interface{}{
						"site_id":   j.task.SiteId,
						"source_id": item.Id,
						"url":       item.Url,
						"status":    1,
						"source":    1,
						"keyword":   keyword,
						"short_url": "https://flic.kr/p/" + base58.Encode([]byte(strconv.Itoa(item.Id))),
						"author":    item.OwnerName,
					})
					if err != nil {
						fmt.Println("JobHandler.ParsePhotos.Flickr.HasError.2", err)
					}
				}

				return result
			}
		}
	}else if imageSource == "google" {
		j.task.SetLog("Парсим фото через Google")

		var (
			link string
			sourceId string
			author string
		)
		if err := chromedp.Run(j.ctx,
			chromedp.Click(".qs[data-sc=I]", chromedp.ByQuery),
			chromedp.Sleep(time.Second * 3),
			chromedp.Click(".islrc .isv-r:first-child .wXeWr", chromedp.ByQuery),
			chromedp.Sleep(time.Second * 7),
			chromedp.AttributeValue(".BIB1wf .n3VNCb", "src", &link, nil, chromedp.ByQuery),
			chromedp.AttributeValue(".islrc .isv-r:first-child", "data-id", &sourceId, nil, chromedp.ByQuery),
			chromedp.InnerHTML(".islrc .isv-r:first-child .fxgdke", &author, chromedp.ByQuery),
		); err != nil {
			log.Println("JobHandler.ParsePhotos.Google.HasError", err)
		}

		var encodedImg string
		if strings.Contains(link, ";base64,") {
			reg := regexp.MustCompile(`(.*?),(.*)`)
			result.Url = reg.ReplaceAllString(link, "$2")
			result.Encoded = true
		}else{
			result.Url = regexp.MustCompile(`^(.*?)\?.*`).ReplaceAllString(link, `$1`)
		}

		result.Id = sourceId
		result.Author = author

		if cache && encodedImg == "" {
			_, err := MYSQL.AddImg(map[string]interface{}{
				"site_id":   j.task.SiteId,
				"source_id": sourceId,
				"url":       result.Url,
				"status":    1,
				"source":    1,
				"keyword":   keyword,
				"short_url": "",
				"author":    author,
			})
			if err != nil {
				fmt.Println("JobHandler.ParsePhotos.Google.HasError.1", err)
			}
		}

	}
	return result
}

func (j *JobHandler) CheckPaa(html string) bool {
	return strings.Contains(html,"JolIg") && strings.Contains(html,"related-question-pair")
}

func (j *JobHandler) CheckCaptcha(html string) bool {
	return strings.Contains(html,"g-recaptcha") && strings.Contains(html,"data-sitekey")
}

func (j *JobHandler) CheckFinished() bool {
	select {
	case <-j.isFinished:
		return true
	default:
		return false
	}
}

func (j *JobHandler) AntiCaptcha(url string, html string) (string, error) {
	paaReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(paaReader)
	if err != nil {
		log.Println("JobHandler.AntiCaptcha.HasError", err)
		return "", err
	}

	siteKey, _ := doc.Find("#recaptcha").Attr("data-sitekey")
	sToken, _ := doc.Find("#recaptcha").Attr("data-s")

	c := &Captcha{
		j.config.Antigate.String,
		url,
		siteKey,
		sToken,
		"http",
		j.Browser.Proxy.Host,
		j.Browser.Proxy.Port,
		j.Browser.Proxy.Login,
		j.Browser.Proxy.Password,
		j.Browser.Proxy.Agent,
		time.Minute * 2,
	}

	key, err := c.SendRecaptcha()
	if err != nil {
		log.Println("JobHandler.AntiCaptcha.2.HasError", err)
	}
	return key, err
}

func (j *JobHandler) SetTimeout(secs int) bool {
	if j.Browser.ctx == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(j.Browser.ctx, time.Duration(secs) * time.Second)
	j.CancelTimeout = cancel
	j.ctx = ctx

	return true
}

func (j *JobHandler) Cancel() {
	if j.CancelTimeout != nil {
		j.CancelTimeout()
	}
	if !LocalTest {
		j.isFinished <- true
	}
}

func (j *JobHandler) Reload() {
	if j.CancelTimeout != nil {
		j.CancelTimeout()
	}
	j.Browser.Reload()
	j.SetTimeout(150)
}