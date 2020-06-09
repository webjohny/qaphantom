package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/gosimple/slug"
	"github.com/webjohny/go-wordpress-xmlrpc"
)

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

type QaParsing struct {
	Task MysqlFreeTask
	Proxy MysqlProxy
	Ctx context.Context
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

func (qa *QaParsing) ParsingPAA(length int, stats QaStats) map[string]QaSetting {
	var paaHtml string
	settings := map[string]QaSetting{}

	// Вытягиваем html код PAA для парсинга вопросов
	if err := chromedp.Run(qa.Ctx,
		chromedp.WaitVisible(`.kno-kp .ifM9O`, chromedp.ByQueryAll),
		chromedp.OuterHTML(`.kno-kp .ifM9O`, &paaHtml, chromedp.ByQueryAll),
	); err != nil {
		log.Fatal(err)
	}

	// Загружаем HTML документ в GoQuery пакет который организует облегчённую работу с HTML селекторами
	paaReader := strings.NewReader(paaHtml)
	doc, err := goquery.NewDocumentFromReader(paaReader)
	if err != nil {
		log.Fatal(err)
	}

	var tasks chromedp.Tasks
	var clicked map[string]bool
	// Начинаем перебор блоков с вопросами
	doc.Find(".related-question-pair").Each(func(i int, s *goquery.Selection) {
		stats.All++
		question := s.Find(".cbphWd").Text()
		link, _ := s.Find(".g a").Attr("href")

		// Ищем дату в блоке, она может быть или в div (если вне текста) или в span (если внутри текста)
		date := s.Find(".kX21rb").Text()
		if date == "" {
			date = s.Find(".Od5Jsd").Text()
		}
		text := strings.Replace(s.Find(".mod").Text(), date, "", -1)
		txtTtml, _ := s.Find(".mod").Html()

		// Берём уникальный идентификатор для вопроса
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
				tasks = append(tasks, chromedp.Click(".cbphWd[data-ved=\""+ved+"\"]"))
				tasks = append(tasks, chromedp.Sleep(time.Millisecond*500))
				clicked[ved] = true
			}

			if strings.Contains(txtTtml, "youtube.com/watch") {
				stats.Yt++
			}else{
				stats.S++
			}
			settings[ved] = qa
		}
	})

	// Проверяем есть ли уже достаточное количество вопросов или всё таки нужно продолжить кликинг по блокам
	if stats.All < length && len(tasks) > 0 {
		if err := chromedp.Run(qa.Ctx, tasks); err != nil {
			log.Fatal(err)
		}
		for k, v := range clicked {
			if setting, ok := settings[k]; ok {
				setting.Clicked = v
			}
		}
		// Продолжаем рекурсию
		return qa.ParsingPAA(length, stats)
	}

	return settings
}

func (qa *QaParsing) RunStream(rt *Routes, question string) {


	//fmt.Println("The end", stats)
	//res, _ := json.Marshal(settings)
	//fmt.Println(string(res))


}

func (rt *Routes) RunJob(w http.ResponseWriter, r *http.Request) {
	//htmlString := "<div class=\"mod\" data-md=\"83\"><!--m--><div class=\"di3YZe\"><div class=\"co8aDb gsrt\" aria-level=\"3\" role=\"heading\"><b>More specifically they must have:</b></div><div class=\"RqBzHd\"><ul class=\"i8Z77e\"><li class=\"TrT0Xe\">An excellent knowledge of and interest in science, particularly the science of oral health.</li><li class=\"TrT0Xe\">Good eyesight.</li><li class=\"TrT0Xe\">Good manual dexterity and a steady hand.</li><li class=\"TrT0Xe\">The ability to concentrate for long periods of time.</li><li class=\"TrT0Xe\">The ability to use specialist equipment.</li></ul><div class=\"ZGh7Vc\"><a class=\"truncation-information\" href=\"https://myjobsearch.com/careers/dentist.html\" data-ved=\"2ahUKEwi4o77Mp-fpAhXXvosKHc_2D1UQnLoEMAR6BAgfEBI\" ping=\"/url?sa=t&amp;source=web&amp;rct=j&amp;url=https://myjobsearch.com/careers/dentist.html&amp;ved=2ahUKEwi4o77Mp-fpAhXXvosKHc_2D1UQnLoEMAR6BAgfEBI\" target=\"_blank\" rel=\"noopener\">Ещё</a></div></div></div><!--n--></div>"
	//qa := QaParsing{}
	//qa.Format(htmlString)

	//items := []string{
	//	"1", "2", "3", "4",
	//}
	//el, items := items[len(items) - 1], items[:len(items) - 1]
	//fmt.Println(el, items)

	//text := utils.StripTags(htmlString)
	//
	//reg := regexp.MustCompile(`\s+`)
	//text = reg.ReplaceAllString(text, ` `)
	//
	//matches := utils.PregMatch(`(?P<sen>.+?\.)`, text)
	//fmt.Println(matches["sen"])
	//task := rt.mysql.GetFreeTask([]string{})
	//c, err := xmlrpc.NewClient("https://" + task.Domain + "/xmlrpc2.php", xmlrpc.UserInfo{
	//	task.Login,
	//	task.Password,
	//})
	//if err != nil {
	//	log.Fatalln(err)
	//}
	c, err := xmlrpc.NewClient(`https://vanguarddentalclinics.com/xmlrpc2.php`, xmlrpc.UserInfo{
		`Jekyll1911`,
		`ghjcnjgfhjkm`,
	})
	if err != nil {
		log.Fatalln(err)
	}

	wpCnf := []interface{}{
		1, c.Username, c.Password,
	}

	var result interface{}
	err = c.Client.Call(`wp.getTerms`, append(
		wpCnf, "category",
	), &result)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)

	//fmt.Println(qa.YoutubeEmbed("https://www.youtube.com/watch?v=PbWHNRHi5B8&first=1&first=1&first=1"))
	//qa.RunStream(rt, `https://www.google.com/search?hl=en&q=How+do+you+use+Clear+TV+HD+Black+Box`)
	//go runPage(rt, "screen2.png")
}

func (rt *Routes) RunJob2(w http.ResponseWriter, r *http.Request) {
	var taskId int
	var proxy MysqlProxy
	var fast QaSetting

	// Инициализация контроллера для управление парсингом
	qa := QaParsing{}
	parser := 1

	// Берём свободную задачу в работу
	task := rt.mysql.GetFreeTask([]string{})
	if task.Id < 1 {
		err := json.NewEncoder(w).Encode(map[string]bool{
			"status": false,
		})
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	taskId = task.Id
	qa.Task = task

	if task.TryCount == 5 {
		task.SetLog("5-я неудавшаяся попытка парсинга. Исключаем ключевик")
		task.SetFinished(2, "Исключён после 5 попыток парсинга")
		return
	}

	task.SetTimeout(parser)

	// Подключаемся к прокси
	proxy = rt.mysql.GetFreeProxy()

	if !proxy.Id.Valid {
		task.SetLog("Нет свободных прокси. Выход")
		task.FreeTask()
		return
	}

	qa.Proxy = proxy

	proxy.SetTimeout(parser)
	task.SetLog("Загружен прокси №" + strconv.Itoa(int(proxy.Id.Int64)) + " (" + proxy.Host.String + ")")

	stats := QaStats{}
	stats.Qac = 10

	if task.From != 0 && task.To != 0 {
		stats.Size = rand.Intn((task.To - task.From) + task.From)
	}else if task.QaCountFrom != 0 && task.QaCountTo != 0 {
		stats.Qac = rand.Intn((task.QaCountTo - task.QaCountFrom) + task.QaCountFrom)
	}

	login  := os.Getenv(proxy.Login.String)
	password := os.Getenv(proxy.Password.String)
	proxyAddr := os.Getenv(proxy.Host.String) //127.0.0.1:1080
	schema := os.Getenv("http://") //socks5:// or http://

	proxyScheme := fmt.Sprintf("%s%s:%s@%s", schema, login, password, proxyAddr)

	task.SetLog("Подключаем прокси к браузеру (" + proxyScheme + ")")
	chromedp.ProxyServer(proxyScheme)

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("ignore-certificate-errors", true),
	)

	// Запускаем контекст браузера
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Устанавливаем собственный logger
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Ставим таймер на отключение если зависнет
	taskCtx, cancel = context.WithTimeout(taskCtx, 100*time.Second)
	defer cancel()

	qa.Ctx = taskCtx

	task.SetLog("Запускаем браузер")
	if err := chromedp.Run(qa.Ctx); err != nil {
		panic(err)
	}

	var paaHtml string
	var existsHtml string

	// Запускаемся
	task.SetLog("Открываем страницу: https://www.google.com/search?hl=en&q=" + url.QueryEscape(task.Keyword))
	if err := chromedp.Run(qa.Ctx,
		// Устанавливаем страницу для парсинга
		chromedp.Navigate("https://www.google.com/search?hl=en&q=" + task.Keyword),
		// Ждём блока с PAA
		chromedp.WaitVisible(`.kno-kp .ifM9O`, chromedp.ByQuery),
		// Вытащить html на проверку каптчи
		chromedp.OuterHTML(".ifM9O", &existsHtml),
		// Вытащить html блока PAA для проверки на наличие
		chromedp.OuterHTML(".kno-kp .ifM9O", &paaHtml),
		// Кликаем сразу на первый вопрос
		chromedp.Click(".related-question-pair:first-child .cbphWd"),
		// Ждём 0.3 секунды чтобы открылся вопрос
		chromedp.Sleep(time.Millisecond * 300),
	); err != nil {
		log.Fatal(err)
	}

	if existsHtml == "" {
		task.SetLog("Отсутствует PAA. Ищем капчу")
	}

	if paaHtml == "" {
		task.SetLog("Не удалось загрузить PAA")
	}
	task.SetLog("Блоки загружены")
	task.SetLog("Начинаем обработку PAA")

	// Запускаем функцию перебора вопросов

	settings := qa.ParsingPAA(stats.Wqc, stats)
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
			task.SetLog(msg)

			// Завершение работы скрипта

			return
		} else if stats.S <= task.QaCountFrom {

		}
	}

	if task.ParseDoubles > 0 {

	}

	var mainEntity []map[string]interface{}

	microMarking := map[string]interface{}{
		"@context" : "https://schema.org",
		"@type" : "FAQPage",
		"mainEntity" : mainEntity,
	}
	fmt.Println(microMarking)

	symb := task.GetRandSymb()
	fmt.Println(symb)
	for _, setting := range settings {
		if _, err := rt.mysql.AddResult(map[string]string{
			"a" : setting.Text,
			"q" : setting.Question,
			"task_id" : strconv.Itoa(task.Id),
			"link" : setting.Link,
			"link_title" : setting.LinkTitle,
		}); err != nil {
			fmt.Println(err)
		}

		if task.SavingAvailable {
			if _, err := rt.mysql.AddTask(map[string]string{
				"site_id" : strconv.Itoa(task.SiteId),
				"cat_id" : strconv.Itoa(task.CatId),
				"parent_id" : strconv.Itoa(task.Id),
				"keyword" : setting.Question,
				"parser" : "",
				"error" : "",
			}); err != nil {
				fmt.Println(err)
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
		text += "<a href='{{link}}#" + slug.Make(setting.Question) + "'>" + task.GetRandTag() + "</a>"

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
		c, err := xmlrpc.NewClient("https://" + task.Domain + "/xmlrpc2.php", xmlrpc.UserInfo{
			task.Login,
			task.Password,
		})
		if err != nil {
			log.Fatalln(err)
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
		if err := chromedp.Run(qa.Ctx,
			// Устанавливаем страницу для парсинга
			chromedp.Navigate("https://www.youtube.com/results?search_query=" + task.Keyword),
			// Ждём блоков с видео
			chromedp.WaitVisible(`a.ytd-thumbnail`, chromedp.ByQueryAll),
			// Вытащить html со списком
			chromedp.OuterHTML("#contents.ytd-section-list-renderer", &videosHtml),
		); err != nil {
			log.Fatal(err)
		}

		videoReader := strings.NewReader(videosHtml)
		doc, err := goquery.NewDocumentFromReader(videoReader)
		if err != nil {
			log.Fatal(err)
		}

		var videos []string
		// Начинаем перебор блоков с видео
		doc.Find("a.ytd-thumbnail").Each(func(i int, s *goquery.Selection) {
			if len(videos) != vCount {
				link, _ := s.Attr("href")
				videos = append(videos, utils.YoutubeEmbed(link))
				task.SetLog(link)
			}
		})

		lastVideo, videos := videos[len(videos) - 1], videos[:len(videos) - 1]
		task.SetLog("Парсинг видео. Готово")

		// Заголовок
		//toDo $variants = $this->fconfig->get_variants();
		variants := []string{"first", "second"}

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

		qaTotalPage.Title = variant + strings.Join(tmp, " ") + "?"

		var photo QaImageResult
		var image string

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

				task.SetLog("Фото загружено на сайт")

				// Готовим код вставки фото в текст
				if task.PubImage >= 2 {
					image = `<p><img class="alignright size-medium" src="` + photo.UrlMedium + `"></p>\n`
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
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + lastVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>\n`
					}
				}else if k > 0 && k < (qaCount - 2) && k % vStep == 0 {
					firstVideo, _ := videos[0], videos[1:]
					if firstVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + firstVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>\n`
					}
				}
			} else {
				if k == qaCount - 1 {
					if lastVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + lastVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>\n`
					}
				}else if k > 0 && k < (qaCount - 1) && k % vStep == 0 {
					firstVideo, _ := videos[0], videos[1:]
					if firstVideo != "" {
						qaTotalPage.Content += `<div class="mb-5"><iframe src="` + firstVideo + `" width="740" height="520" frameborder="0" allowfullscreen="allowfullscreen"></iframe></div>\n`
					}
				}
			}

			// Заголовок
			if item.Question != "" {
				qaTotalPage.Content += `<div id="` + slug.Make(item.Question) + `"></div>`
				if task.H1 < 1 || k > 0 {
					if task.ShOrder < 1 {
						qaTotalPage.Content += `<h` + strconv.Itoa(item.H) + `>` + item.Question + `</h` + strconv.Itoa(item.H) + `>\n`
					} else {
						qaTotalPage.Content += `<h2>` + slug.Make(item.Question) + `</h2>\n`
					}
				}
			}

			// Если ответ первый
			if k < 1 {
				// Вставляем фото
				qaTotalPage.Content += image
				// Ответ разбиваем по предложениям
				if !strings.Contains(item.Text, "<ul>") && !strings.Contains(item.Text, "<ol>") && !strings.Contains(item.Text, "<h3>") {
					formattedText := utils.StripTags(item.Text)
					splittedText := utils.SentenceSplit(formattedText)
					qaTotalPage.Content += "<p>" + strings.Join(splittedText, "</p><p>") + "</p>\n"
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
				qaTotalPage.Content += `<div id="qa_date">Date: ` + item.Date + `</div>\n`
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
			proxy.FreeProxy()
			utils.ErrorHandler(chromedp.Cancel(qa.Ctx))
			return
		}

		// Определяем ID категории


		fmt.Println(c)
		fmt.Println(qaTotalPage)
	}

	fmt.Println(settings)
	fmt.Println(taskId)
}