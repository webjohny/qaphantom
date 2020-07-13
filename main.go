package main

import (
	"fmt"
	"time"
)

var (
	utils Utils
	mysql MysqlDb
	conf Configuration
)

var LocalTest = false

func main() {
	utils = Utils{}

	conf = Configuration{}
	conf.Create()

	// Connect to MysqlDB
	mysql = MysqlDb{
		conf: conf,
	}
	mysql.CreateConnection()

	// Run routes
	routes := Routes{
		mysql: mysql,
		conf: conf,
		streams: Streams{},
	}

	if LocalTest {
		job := JobHandler{}
		if job.Browser.Init() {
			//job.taskId = 529235
			job.IsStart = true
			fmt.Println("Stop")
			fmt.Println(job.Run(0))
			job.Run(0)
			//job.Run(0)
			//job.Run(0)
			//job.Run(0)
		}
	}

	routes.Run()

	time.Sleep(time.Minute)
}