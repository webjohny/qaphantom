package main

import (
	"log"
	"os"
	"qaphantom/services"
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

	// Connect to MysqlDB
	MYSQL.CreateConnection(CONF.MysqlHost, CONF.MysqlDb, CONF.MysqlLogin, CONF.MysqlPass)

	//aafprocesses.com
	//jekyll1911 / ghjcnjgfhjkm
	wp := services.Wordpress{}
	log.Fatal(wp.Connect(`https://magazineoptionscarrieres.com`, "jekyll1911", "ghjcnjgfhjkm", 1))
	log.Println(wp.CatIdByName("QA"))
	log.Println(wp.GetPost(1))
	log.Fatal(wp.NewPost("Test article", "Test article", 1, 0))


	if CONF.Env == "local" {
		task := MYSQL.GetFreeTask(564805)
		task.SetTimeout(2)

		go func() {
			job := JobHandler{}
			job.IsStart = true
			if job.Browser.Init() {
				job.Run(2)
				job.Run(1)
				//job.Run(1)
			}
		}()

		time.Sleep(100)

		//go func() {
		//	job := JobHandler{}
		//	job.IsStart = true
		//	if job.Browser.Init() {
		//		job.Run(1)
		//		job.Run(0)
		//		job.Run(2)
		//	}
		//}()
		//
		//time.Sleep(100)
		//
		//go func() {
		//	job := JobHandler{}
		//	job.IsStart = true
		//	if job.Browser.Init() {
		//		job.Run(0)
		//		job.Run(2)
		//		job.Run(1)
		//	}
		//}()
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