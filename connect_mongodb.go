package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

// You will be using this Trainer type later in the program
type Trainer struct {
	Name string
	Age  int
	City string
}

func main() {
	// Rest of the code will go here
	// Create client
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Create connect
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("test").Collection("trainers")

	var result Trainer

	//result, err := collection.Find(context.TODO(), bson.D{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("Found a single document: %+v\n", result)

	err = collection.FindOne(context.TODO(), bson.D{}).Decode(&result)
	fmt.Println(result.Name)

	//err = client.Disconnect(context.TODO())
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("Connection to MongoDB closed.")
}
