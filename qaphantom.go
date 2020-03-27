package main

import (
	"github.com/webjohny/qaphantom/config"
	"github.com/webjohny/qaphantom/pkg/middleware/connect_mongodb"
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