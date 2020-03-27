package main

import (
	"github.com/webjohny/qaphantom/config"
)

func main() {
	conf := config.Create()

	// Connect to MongoDB
	mongoDb := MongoConn{
		conf: conf,
	}
	mongoDb.CreateConnection()

	// Run routes
	routes := Routes{
		mongo: mongoDb,
		conf: conf,
	}
	routes.Run()
}