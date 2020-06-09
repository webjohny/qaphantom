package main

import "time"

var utils Utils

func main() {
	utils = Utils{}

	conf := Configuration{}
	conf.Create()

	// Connect to MysqlDB
	mysqlDb := MysqlDb{
		conf: conf,
	}
	mysqlDb.CreateConnection()

	// Run routes
	routes := Routes{
		mysql: mysqlDb,
		conf: conf,
		streams: Streams{},
	}

	routes.Run()
	time.Sleep(time.Minute)
}