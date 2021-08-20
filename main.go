package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"qaphantom/config"
)

var (
	UTILS   Utils
	MYSQL   MysqlDb
	CONF    config.Configuration
	STREAMS Streams
)

func main() {
	path, _ := os.Getwd()

	CONF.Create(path + "/config.json")

	log.Fatal(readFile())

	// Connect to MysqlDB
	MYSQL.CreateConnection(CONF.MysqlHost, CONF.MysqlDb, CONF.MysqlLogin, CONF.MysqlPass)

	if CONF.Env == "local" {
		task := MYSQL.GetFreeTask(564805)
		task.SetTimeout(2)

		go func() {
			//fmt.Println(MYSQL.InsertOrUpdateResult(map[string]interface{}{
			//	//"a" : setting.Text,
			//	"link": "",
			//	"text": "",
			//	"cat": task.Cat,
			//	"domain": task.Domain,
			//	"cat_id": task.CatId,
			//	"site_id": task.SiteId,
			//	"link_title": "",
			//	"q" : "Test query",
			//	"html" : "<div>34343a</div>",
			//	"task_id" : strconv.Itoa(task.Id),
			//}))
			job := JobHandler{}
			job.IsStart = true
			if job.Browser.Init() {
				job.Run(2)
				job.Run(1)
				//job.Run(1)
			}
		}()

		time.Sleep(100)
	}else if MYSQL.CountWorkingTasks() > 0 {
		conf := MYSQL.GetConfig()
		extra := conf.GetExtra()
		if extra.CountStreams > 0 {
			STREAMS.StartLoop(extra.CountStreams, extra.LimitStreams, extra.CmdStreams)
		}
	}

	routes := Routes{}
	routes.Run()

	time.Sleep(time.Minute)
}

func readFile() string{
	file, err := os.Open("html.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()


	b, err := ioutil.ReadAll(file)

	str := string(b)

	var re = regexp.MustCompile(`(?si)(\<script.*?\<\/script\>)`)
	s := re.ReplaceAllString(str, "")
	re = regexp.MustCompile(`(?si)(\<style.*?\<\/style\>)`)
	s = re.ReplaceAllString(s, "")

	f, err := os.OpenFile("html.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(s)
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	paaReader := strings.NewReader(s)
	doc, err := goquery.NewDocumentFromReader(paaReader)
	if err != nil {
		log.Println("JobHandler.RedirectParsing.2.HasError", err)
		return "settings"
	}

	fmt.Println(doc.Find("#rso").Text())

	return ""
}