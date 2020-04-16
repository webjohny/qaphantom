package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
	"strconv"
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

func (m *MongoDb) GetQuestions(params map[string]interface{}) []map[string]interface{} {
	coll := m.db.Collection("questions")

	results := make([]map[string]interface{}, 0)

	ctx := context.Background()

	findOptions := options.Find()

	if len(params) > 0{
		if params["limit"] != 0 {
			findOptions.SetLimit(int64(params["limit"].(int)))
		}
		if params["offset"] != 0 {
			findOptions.SetSkip(int64(params["limit"].(int)))
		}
	}

	//Passing the bson.D{{}} as the filter matches  documents in the collection
	cur, err := coll.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	//Finding multiple documents returns a cursor
	//Iterate through the cursor allows us to decode documents one at a time

	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var elem map[string]interface{}
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


func (m *MongoDb) LoopCollectStats() {
	if ! checkLoopCollect {
		checkLoopCollect = true
		for {
			count := m.GetCountQuestions(map[string]interface{}{})
			fmt.Println(count)
			if count < 10000 {
				break
			}
			m.CollectStats()
			time.Sleep(3 * time.Minute)
		}
	}
}


func (m *MongoDb) CollectStats() map[int]map[string]interface{} {
	params := make(map[string]interface{})

	limit := 5000
	offset := 0
	stat := map[int]map[string]interface{}{}

	count := m.GetCountQuestions(map[string]interface{}{})
	var parts int

	if limit > count {
		parts = 1
	} else {
		parts = int(math.Ceil(float64(count) / float64(limit)))
	}

	cats := m.GetCats(map[string]interface{}{}, map[string]interface{}{})
	sites := m.GetSites(map[string]interface{}{}, map[string]interface{}{})

	//notCorrectData := make([]interface{}, 0)
	if false {
		for i := 0; i < parts; i++ {
			params["limit"] = limit
			params["offset"] = offset
			params["isStat"] = true
			questions := m.GetQuestions(params)
			if true {
				for _, question := range questions {
					if question["site_id"] == nil {
						//notCorrectData = append(notCorrectData, question)
						continue
					}

					for _, cat := range cats {
						if question["cat_id"] == cat["_id"] {
							question["cat_info"] = cat
						}
					}

					for _, site := range sites {
						if question["site_id"] == site["id"] {
							question["site_info"] = site
						}
					}

					siteId := int(question["site_id"].(int32))

					var status int
					switch question["status"].(type) {
					case string:
						status, _ = strconv.Atoi(question["status"].(string))
					default:
						status = int(question["status"].(int32))
					}

					switch question["cat_id"].(type) {
					case primitive.ObjectID:
						catIdObj := question["cat_id"].(primitive.ObjectID)
						catId := catIdObj.Hex()
						site := map[string]interface{}{}

						if item, ok := stat[siteId]; ok {
							site = item
						}

						if _, ok := question["site_info"]; ok {
							siteInfo := question["site_info"].(map[string]interface{})
							site["domain"] = siteInfo["domain"].(string)
						}

						if _, ok := site["ready"]; ! ok {
							site["ready"] = 0
						}

						if _, ok := site["error"]; ! ok {
							site["error"] = 0
						}

						if _, ok := site["total"]; ! ok {
							site["total"] = 0
						}

						cats := map[string]interface{}{}
						cat := map[string]interface{}{}

						_, ok := site["cats"]
						if ok && len(site["cats"].(map[string]interface{})) > 0 {
							cats = site["cats"].(map[string]interface{})

							_, ok := cats[catId]
							if ok && len(cats[catId].(map[string]interface{})) > 0 {
								cat = cats[catId].(map[string]interface{})
							}
						}

						if _, ok := question["cat_info"]; ok {
							catInfo := question["cat_info"].(map[string]interface{})
							cat["title"] = catInfo["title"].(string)
						}

						if _, ok := cat["ready"]; ! ok {
							cat["ready"] = 0
						}

						if _, ok := cat["error"]; ! ok {
							cat["error"] = 0
						}

						if _, ok := cat["total"]; ! ok {
							cat["total"] = 0
						}

						if status == 2 {
							site["error"] = site["error"].(int) + 1
							cat["error"] = cat["error"].(int) + 1
						} else if status == 1 {
							site["ready"] = site["ready"].(int) + 1
							cat["ready"] = cat["ready"].(int) + 1
						}

						site["total"] = site["total"].(int) + 1
						cat["total"] = cat["total"].(int) + 1

						cats[catId] = cat
						site["cats"] = cats

						stat[siteId] = site
						//notCorrectData = append(notCorrectData, question)
					default:
						//notCorrectData = append(notCorrectData, question)
						continue
					}
				}
			}


			if count < offset {
				offset = offset - (offset - count)
			}else{
				offset = offset + limit
			}
			time.Sleep(time.Second)
		}

		if true {
			for k, v := range stat {
				coll := m.db.Collection("sites")

				filter := bson.M{"id": bson.M{"$eq": k}}
				result, _ := coll.UpdateOne(
					context.Background(),
					filter, bson.M{
						"$set": bson.M{
							"info": v,
						},
					})
				fmt.Println(result)
			}
		}
	}
	return stat
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
