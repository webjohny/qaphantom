package main

import (
	"context"
	"fmt"
	"github.com/webjohny/qaphantom/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
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
	Status int
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

	results := GetQuestions(1, 1)
	if ! reflect.DeepEqual(results, Question{}) {
		fmt.Println(results)
		//results
	}
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

func GetQuestions(limit int64, offset int64) []Question {
	coll := mongoDb.Collection("questions")

	//find records
	//pass these options to the Find method
	findOptions := options.Find()
	//Set the limit of the number of record to find
	if limit != 0 {
		findOptions.SetLimit(limit)
	}
	if offset != 0 {
		findOptions.SetSkip(offset)
	}
	//Define an array in which you can store the decoded documents
	var results []Question

	//Passing the bson.D{{}} as the filter matches  documents in the collection
	cur, err := coll.Find(context.TODO(), bson.D{{}}, findOptions)
	if err !=nil {
		log.Fatal(err)
	}
	//Finding multiple documents returns a cursor
	//Iterate through the cursor allows us to decode documents one at a time

	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var elem Question
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem)

	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	//Close the cursor once finished
	cur.Close(context.TODO())

	return results
}

func SetQuestions(question Question, id int) *mongo.InsertOneResult {
	coll := mongoDb.Collection("questions")

	result, _ := coll.InsertOne(
		context.Background(),
		question)

	return result
}

func CheckQuestionByKeyword(keyword map[string]interface{}) {

}

func CheckQuestionsByKeywords(keywords map[string]interface{}) {

}
