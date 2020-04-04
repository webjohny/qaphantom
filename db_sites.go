package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
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