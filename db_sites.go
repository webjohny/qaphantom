package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
)

var sites map[int]map[string]interface{}

func (m *MongoDb) GetSite(id int) map[string]interface{} {
	coll := m.db.Collection("sites")

	if sites == nil {
		sites = make(map[int]map[string]interface{})
	}

	var site map[string]interface{}

	if item, ok := sites[id]; ok {
		site = item
	}else{
		err := coll.FindOne(context.TODO(), bson.D{{"id", id}}).Decode(&site)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}

	sites[id] = site

	return site
}

func (m *MongoDb) GetSites(params map[string]interface{}, postData map[string]interface{}) []map[string]interface{} {
	coll := m.db.Collection("sites")

	findOptions := options.Find()

	if params["limit"] != nil {
		limit, _ := strconv.Atoi(params["limit"].(string))
		findOptions.SetLimit(int64(limit))
	}

	if params["offset"] != nil {
		offset, _ := strconv.Atoi(params["offset"].(string))
		findOptions.SetSkip(int64(offset))
	}

	cur, err := coll.Find(context.TODO(), postData, findOptions)
	if err != nil {
		fmt.Println(err)
	}
	//Finding multiple documents returns a cursor
	//Iterate through the cursor allows us to decode documents one at a time

	var output []map[string]interface{}

	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var elem map[string]interface{}
		err := cur.Decode(&elem)
		if err != nil {
			fmt.Println(err)
		}else {
			output = append(output, elem)
		}
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
	}

	//Close the cursor once finished
	err = cur.Close(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	return output
}