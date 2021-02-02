package main

import (
	"net/http"
)

func (rt *Routes) RunJob(w http.ResponseWriter, r *http.Request) {
	job := JobHandler{}
	job.IsStart = true
	job.Run(0)
	//htmlString := "<div class=\"mod\" data-md=\"83\"><!--m--><div class=\"di3YZe\"><div class=\"co8aDb gsrt\" aria-level=\"3\" role=\"heading\"><b>More specifically they must have:</b></div><div class=\"RqBzHd\"><ul class=\"i8Z77e\"><li class=\"TrT0Xe\">An excellent knowledge of and interest in science, particularly the science of oral health.</li><li class=\"TrT0Xe\">Good eyesight.</li><li class=\"TrT0Xe\">Good manual dexterity and a steady hand.</li><li class=\"TrT0Xe\">The ability to concentrate for long periods of time.</li><li class=\"TrT0Xe\">The ability to use specialist equipment.</li></ul><div class=\"ZGh7Vc\"><a class=\"truncation-information\" href=\"https://myjobsearch.com/careers/dentist.html\" data-ved=\"2ahUKEwi4o77Mp-fpAhXXvosKHc_2D1UQnLoEMAR6BAgfEBI\" ping=\"/url?sa=t&amp;source=web&amp;rct=j&amp;url=https://myjobsearch.com/careers/dentist.html&amp;ved=2ahUKEwi4o77Mp-fpAhXXvosKHc_2D1UQnLoEMAR6BAgfEBI\" target=\"_blank\" rel=\"noopener\">Ещё</a></div></div></div><!--n--></div>"
	//qa := QaParsing{}
	//qa.Format(htmlString)

	//items := []string{
	//	"1", "2", "3", "4",
	//}
	//el, items := items[len(items) - 1], items[:len(items) - 1]
	//fmt.Println(el, items)

	//text := UTILS.StripTags(htmlString)
	//
	//reg := regexp.MustCompile(`\s+`)
	//text = reg.ReplaceAllString(text, ` `)
	//
	//matches := UTILS.PregMatch(`(?P<sen>.+?\.)`, text)
	//fmt.Println(matches["sen"])
	//task := MYSQL.GetFreeTask([]string{})
	//c, err := xmlrpc.NewClient("https://" + task.Domain + "/xmlrpc2.php", xmlrpc.UserInfo{
	//	task.Login,
	//	task.Password,
	//})
	//if err != nil {
	//	log.Println(err)
	//}
	//wp := Wordpress{}
	//wp.Connect(`https://vanguarddentalclinics.com/xmlrpc2.php`, `Jekyll1911`, `ghjcnjgfhjkm`, 1)
	//result := wp.EditPost(13625, "Which Bank Of America Has A Notary?", "")
	//result := wp.GetPost(1365)
	//result := wp.GetCats()

	//resp, err := http.Get("https://plambi.com.ua/image/catalog/1nashdom/banners/1.jpg")
	//if err != nil {
	//	log.Println(err)
	//}
	//defer resp.Body.Close()

	//bodyBytes, err := ioutil.ReadAll(resp.Body)
	//mime := mimetype.Detect(bodyBytes)
	//fmt.Println(mime)
	//serveFrames(bodyBytes)
	//fmt.Println(wp.UploadFile("test.png", "image/png", string(bodyBytes), "", 0))
	//fmt.Println(resp.Header["Content-Type"])

	//log.Println(result)

	//fmt.Println(qa.YoutubeEmbed("https://www.youtube.com/watch?v=PbWHNRHi5B8&first=1&first=1&first=1"))
	//qa.RunStream(rt, `https://www.google.com/search?hl=en&q=How+do+you+use+Clear+TV+HD+Black+Box`)
	//go runPage(rt, "screen2.png")
}
