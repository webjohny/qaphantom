package main

import (
	"os"
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

	// Run routes

	if CONF.Env == "local" {
		go func() {
			job := JobHandler{}
			job.IsStart = true
			if job.Init() {
				job.Run(0)
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