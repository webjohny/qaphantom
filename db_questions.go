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

var SitesLookupStage = bson.D{{"$lookup", bson.D{{"from", "sites"}, {"localField", "site_id"}, {"foreignField", "id"}, {"as", "site_info"}}}}
var SitesUnwindStage = bson.D{{"$unwind", bson.D{{"path", "$site_info"}, {"preserveNullAndEmptyArrays", true}}}}

var CatsLookupStage = bson.D{{"$lookup", bson.D{{"from", "cats"}, {"localField", "cat_id"}, {"foreignField", "_id"}, {"as", "cat_info"}}}}
var CatsUnwindStage = bson.D{{"$unwind", bson.D{{"path", "$cat_info"}, {"preserveNullAndEmptyArrays", true}}}}

var checkLoopCollect bool = false

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

func (m *MongoDb) GetCountQuestions(params map[string]interface{}) int {
	coll := m.db.Collection("questions")

	count, _ := coll.CountDocuments(context.TODO(), bson.M{})
	return int(count)
}

func (m *MongoDb) GetFreeQuestion(ids []string) map[string]interface{} {
	coll := m.db.Collection("questions")

	var result map[string]interface{}

	findOptions := bson.M{
		"try_count": bson.D{{
			"$lte", 5,
		}},
		"status": 0,
		"timeout": bson.D{{
			"$lt", time.Now(),
		}},
	}

	if len(ids) > 0 {
		var objIds []primitive.ObjectID
		for _, v := range ids {
			objectId, err := primitive.ObjectIDFromHex(v)
			if err == nil {
				objIds = append(objIds, objectId)
			}
		}
		findOptions["_id"] = bson.M{
			"$not": bson.M{"$in": objIds},
		}
	}

	err := coll.FindOne(context.TODO(), findOptions).Decode(&result)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if len(result) > 0 {
		siteId := int(result["site_id"].(int32))

		site := m.GetSite(siteId)

		if len(site) > 0 {
			for k, v := range site {
				if k != "_id" && k != "id" {
					result[k] = v
				}
			}
		}

		return result
	}

	return nil
}

func (m *MongoDb) CheckQuestionByKeyword(keyword string, siteId int) *Question {
	coll := m.db.Collection("questions")

	var result *Question

	err := coll.FindOne(context.TODO(), bson.M{
		"keyword": keyword,
		"site_id": siteId,
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
		{"site_id", siteId},
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
