package main

import (
	"context"
	"fmt"
	"github.com/webjohny/qaphantom/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// You will be using this Trainer type later in the program
type Question struct {
	Log string
	LogLast string
	SiteId int
	CatId int
	TryCount int
	ErrorsCount int
	Status bool
	Error string
	ParserId int
	Timeout time.Time
	Keyword string
	FastA string
	FastLink string
	FastLinkTitle string
	FastDate time.Time
}

var mongoClient mongo.Client
var mongoDb *mongo.Database

func main() {
	// Rest of the code will go here
	// Create client
	CreateConnection()

	coll := mongoDb.Collection("questions")
	result, _ := coll.InsertOne(
		context.Background(),
		bson.D{
			{"Error", "Without error"},
			{"Log", "First operation"},
			{"LogLast", "Last operation"},
			{"SiteId", 100},
			{"CatId", 100},
			{"TryCount", 100},
			{"ErrorsCount", 100},
			{"Status", true},
			{"ParserId", 100},
			{"Keyword", "simple keyword"},
			{"FastA", "fast a link"},
			{"FastLink", "https://www.mongodb.com/blog/post/mongodb-go-driver-tutorial"},
			{"FastLinkTitle", "fast link title"},
			{"Timeout", time.Now()},
			{"FastDate", time.Now()},
		})

	fmt.Println(result)

	var item Question
	err := coll.FindOne(context.TODO(), bson.D{}).Decode(&item)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(item)
	//var result Trainer

	//result, err := collection.Find(context.TODO(), bson.D{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("Found a single document: %+v\n", result)

	//err := collection.FindOne(context.TODO(), bson.D{}).Decode(&result)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(result.Name)

	Disconnect()
}

func CreateConnection() {
	conf := config.Create()

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" + conf.MongoUrl))
	if err != nil {
		log.Fatal(err)
	}
	mongoClient = *client

	// Create connect
	err = mongoClient.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	mongoDb = mongoClient.Database(conf.MongoDb)
}

func Disconnect() {
	err := mongoClient.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}

func GetQuestions(limit, offset, lastId int) {
	coll := mongoDb.Collection("questions")

	var item Question
	err := coll.FindOne(context.TODO(), bson.D{}).Decode(&item)
	if err != nil {
		log.Fatal(err)
	}


}

func SetQuestions(question *Question, id int) {
	coll := mongoDb.Collection("questions")
	result, _ := coll.InsertOne(
		context.Background(),
		question)
	fmt.Println(result)
}

func CheckQuestionByKeyword(keyword map[string]interface{}) {

}

func CheckQuestionsByKeywords(keywords map[string]interface{}) {

}
