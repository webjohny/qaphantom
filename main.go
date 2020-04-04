package main

func main() {
	conf := Configuration{}
	conf.Create()

	// Connect to MongoDB
	mongoDb := MongoDb{
		conf: conf,
	}
	mongoDb.CreateConnection()

	// Run routes
	routes := Routes{
		mongo: mongoDb,
		conf: conf,
		streams: Streams{},
	}

	routes.Run()
}