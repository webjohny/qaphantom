package main

import (
	"time"
)

var (
	utils Utils
	mysql MysqlDb
	conf Configuration
)

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

	//job := JobHandler{}
	//job.IsStart = true
	//job.Run(0)

	routes.Run()

	time.Sleep(time.Minute)
}