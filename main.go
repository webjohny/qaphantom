package main

import (
	"fmt"
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