package qaphantom

func main() {
	conf := Configuration{}
	conf.Create()

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