package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// You will be using this Trainer type later in the program
type Question struct {
	Log string `db:"Log" json:"log"`
	LogLast string `db:"LogLast" json:"loglast"`
	SiteId int `db:"SiteId" json:"siteid"`
	CatId int `db:"CatId" json:"catid"`
	TryCount int `db:"TryCount" json:"trycount"`
	ErrorsCount int `db:"ErrorsCount" json:"errorscount"`
	Status int `db:"Status" json:"status"`
	Error string `db:"Error" json:"error"`
	ParserId int `db:"ParserId" json:"parserid"`
	Timeout time.Time `db:"Timeout" json:"timeout"`
	Keyword string `db:"Keyword" json:"keyword"`
	FastA string `db:"FastA" json:"fasta"`
	FastLink string `db:"FastLink" json:"fastlink"`
	FastLinkTitle string `db:"FastLinkTitle" json:"fastlinktitle"`
	FastDate time.Time `db:"FastDate" json:"fastdate"`
}

type MongoDb struct {
	client *mongo.Client
	db *mongo.Database
	conf Configuration
}

func (q *Question) Create() {

}

func (m *MongoDb) CreateConnection() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" + m.conf.MongoUrl))
	if err != nil {
		log.Fatal(err)
	}
	m.client = client

	// Create connect
	err = m.client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = m.client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	m.db = m.client.Database(m.conf.MongoDb)
}

func (m *MongoDb) Disconnect() {
	err := m.client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}

func (m *MongoDb) GetQuestions(limit int64, offset int64) []Question {
	coll := m.db.Collection("questions")

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
	err = cur.Close(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	return results
}

func (m *MongoDb) InsertQuestion(question Question) (*mongo.InsertOneResult, error) {
	coll := m.db.Collection("questions")

	result, err := coll.InsertOne(
		context.Background(),
		question)
	fmt.Println(result)

	return result, err
}

func (m *MongoDb) UpdateQuestion(data map[string]string, id string) (*mongo.UpdateResult, error) {
	coll := m.db.Collection("questions")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	filter := bson.M{"_id": bson.M{"$eq": objID}}
	result, err := coll.UpdateOne(
		context.Background(),
		filter, bson.M{
			"$set": data,
		})

	return result, err
}

func (m *MongoDb) CheckQuestionByKeyword(keyword string, siteId int) *Question {
	coll := m.db.Collection("questions")

	var result *Question

	err := coll.FindOne(context.TODO(), bson.M{
		"keyword": keyword,
		"siteid": siteId,
	}).Decode(&result)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return result
}

func (m *MongoDb) CheckQuestionsByKeywords(keywords []string, siteId int) []Question {
	coll := m.db.Collection("questions")

	findOptions := options.Find()
	//Define an array in which you can store the decoded documents
	var results []Question

	//Passing the bson.D{{}} as the filter matches  documents in the collection
	cur, err := coll.Find(context.TODO(), bson.D{
		{"keyword", bson.D{{"$in", keywords}}},
		{"siteid", siteId},
	}, findOptions)
	if err != nil {
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

	err = cur.Err()
	if err != nil {
		fmt.Println(err)
	}

	//Close the cursor once finished
	err = cur.Close(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	return results
}
