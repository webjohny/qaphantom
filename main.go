package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

var (
	utils Utils
	mysql MysqlDb
	conf Configuration
	streams Streams
)

var LocalTest = false

func main() {
	utils = Utils{}

	conf = Configuration{}
	conf.Create()

	// Connect to MysqlDB
	mysql = MysqlDb{}
	mysql.CreateConnection()

	streams = Streams{}

	// Run routes
	routes := Routes{}

	if LocalTest {

		TestScreen()
		log.Fatal("")

		job := JobHandler{}
		job.IsStart = true
		if job.Browser.Init() {
			//job.taskId = 529235

			fmt.Println("Stop")
			fmt.Println(job.Run(0))
			job.Run(0)
			//job.Run(0)
			//job.Run(0)
			//job.Run(0)
		}
	}else if mysql.CountWorkingTasks() > 0 {
		config := mysql.GetConfig()
		extra := config.GetExtra()
		if extra.CountStreams > 0 {
			streams.StartLoop(extra.CountStreams, extra.LimitStreams, extra.CmdStreams)
		}
	}

	routes.Run()

	time.Sleep(time.Minute)
}

func TestScreen() {

	host := "89.191.225.148"
	port := "45785"
	login := "phillip"
	password := "I2n9BeJ"

	proxy := Proxy{
		Id:       1,
		Host:     host,
		Port:     port,
		Login:    login,
		Password: password,
		LocalIp:  host + ":" + port,
	}
	browser := Browser{}
	browser.Proxy = proxy
	browser.Init()

	ctx, cancel := context.WithTimeout(browser.ctx, time.Second * 15)
	browser.CancelTimeout = cancel
	browser.ctx = ctx
	defer browser.Cancel()

	status, buffer := browser.ScreenShot("https://www.google.com/search?hl=en&gl=us&q=what+is+my+ip")
	if !status {
		log.Fatal(string(buffer))
	}

	if len(buffer) > 0 {
		fmt.Println("dadsa")
	}
}