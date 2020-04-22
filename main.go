package main

import "time"

func main() {
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