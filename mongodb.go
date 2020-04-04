package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type MongoDb struct {
	client *mongo.Client
	db *mongo.Database
	conf Configuration
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
