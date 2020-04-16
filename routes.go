package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Routes struct {
	conf Configuration
	mongo MongoDb
	utils Utils
	streams Streams
}

func (rt *Routes) CmdTimer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	commandExec := r.FormValue("cmd")

	var limit int64 = 1000

	if len(r.FormValue("limit")) > 0 {
		limit, _ = strconv.ParseInt(r.FormValue("limit"), 0, 64)
	}

	stream := Stream{}
	status := stream.StartTaskTimer(commandExec, limit)

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": status,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) CheckQuestion(w http.ResponseWriter, r *http.Request) {
	siteId := rt.utils.toInt(r.FormValue("id"))
	keyword := r.FormValue("keyword")

	question := rt.mongo.CheckQuestionByKeyword(keyword, siteId)

	err := json.NewEncoder(w).Encode(question)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) CheckQuestions(w http.ResponseWriter, r *http.Request) {
	siteId := rt.utils.toInt(r.FormValue("id"))
	keywords := rt.utils.ParseFormCollection(r,"keywords")

	var arrKeywords []string
	for _, v := range keywords {
		arrKeywords = append(arrKeywords, v)
	}

	questions := rt.mongo.CheckQuestionsByKeywords(arrKeywords, siteId)

	err := json.NewEncoder(w).Encode(questions)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) GetFreeQuestion(w http.ResponseWriter, r *http.Request) {
	dataIds := r.FormValue("ids")

	var ids []string

	if dataIds != "" {
		ids = strings.Split(dataIds, ",")
	}
	fmt.Println(ids)

	question := rt.mongo.GetFreeQuestion(ids)

	err := json.NewEncoder(w).Encode(question)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) GetCats(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})

	params["limit"] = r.FormValue("limit")
	params["offset"] = r.FormValue("offset")
	postData := map[string]interface{}{}

	if r.FormValue("site_id") != "" {
		val, _ := strconv.Atoi(r.FormValue("site_id"))
		postData["site_id"] = val
	}
	if r.FormValue("title") != "" {
		postData["title"] = r.FormValue("title")
	}

	questions := rt.mongo.GetCats(params, postData)

	err := json.NewEncoder(w).Encode(questions)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) GetQuestionsForStat(w http.ResponseWriter, r *http.Request) {

	count := rt.mongo.GetCountQuestions(map[string]interface{}{})

	stat := map[int]map[string]interface{}{}

	fmt.Println(1)

	go rt.mongo.LoopCollectStats()

	if count > 10000 {
		sites := rt.mongo.GetSites(map[string]interface{}{}, map[string]interface{}{})
		if len(sites) > 0 {
			for _, site := range sites {
				if site["info"] != nil {
					stat[int(site["id"].(int32))] = site["info"].(map[string]interface{})
				}
			}
		}
	}else{
		stat = rt.mongo.CollectStats()
	}

	err := json.NewEncoder(w).Encode(stat)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) GetQuestions(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})

	questions := rt.mongo.GetQuestions(params)

	err := json.NewEncoder(w).Encode(questions)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	data := rt.utils.ParseFormCollection(r, "data")

	response := map[string]bool{
		"status": false,
	}

	_, err := rt.mongo.UpdateQuestion(data, id)
	if err != nil {
		fmt.Println(err)
	}else{
		response["status"] = true
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) InsertQuestion(w http.ResponseWriter, r *http.Request) {
	question := Question{}
	question.Log = r.FormValue("Log")
	question.LogLast = r.FormValue("LogLast")
	question.SiteId = rt.utils.toInt(r.FormValue("SiteId"))
	question.Cat = r.FormValue("Cat")
	if r.FormValue("CatId") != "" {
		question.CatId, _ = primitive.ObjectIDFromHex(r.FormValue("CatId"))
	}
	question.TryCount = rt.utils.toInt(r.FormValue("TryCount"))
	question.ErrorsCount = rt.utils.toInt(r.FormValue("ErrorsCount"))
	question.Status = rt.utils.toInt(r.FormValue("status"))
	question.Error = r.FormValue("Error")
	question.ParserId = rt.utils.toInt(r.FormValue("ParserId"))
	question.Timeout = time.Now()
	question.Keyword = r.FormValue("Keyword")
	question.FastA = r.FormValue("FastA")
	question.FastLink = r.FormValue("FastLink")
	question.FastLinkTitle = r.FormValue("FastLinkTitle")
	question.FastDate = time.Now()

	res, err := rt.mongo.InsertQuestion(question)

	if err != nil {
		fmt.Println(err)
	}else{
		question.Id = res.InsertedID

		err = json.NewEncoder(w).Encode(question)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (rt *Routes) StartLoopStreams(w http.ResponseWriter, r *http.Request) {
	count := rt.utils.toInt(r.FormValue("count"))
	limit := rt.utils.toInt(r.FormValue("limit"))
	cmd := r.FormValue("cmd")

	if limit < 1 {
		limit = 10
	}

	rt.streams.StopAll()

	go func() {
		for i := 1; i <= count; i++ {
			stream := rt.streams.Add(i)
			stream.cmd = cmd + " " + strconv.Itoa(i)
			go stream.Start(i, int64(limit*1000))
		}
	}()

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) StopLoopStreams(w http.ResponseWriter, r *http.Request) {
	go rt.streams.StopAll()

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) StopStream(w http.ResponseWriter, r *http.Request) {
	id := rt.utils.toInt(r.FormValue("id"))

	rt.streams.Stop(id)
}

func (rt *Routes) Run() {
	rt.utils = Utils{}

	r := mux.NewRouter()

	r.HandleFunc("/check/question", rt.CheckQuestion).Methods("POST")
	r.HandleFunc("/check/questions", rt.CheckQuestions).Methods("POST")
	r.HandleFunc("/update/question", rt.UpdateQuestion).Methods("POST")
	r.HandleFunc("/insert/question", rt.InsertQuestion).Methods("POST")
	r.HandleFunc("/get/cats", rt.GetCats).Methods("POST")
	r.HandleFunc("/get/questions-stat", rt.GetQuestionsForStat).Methods("POST")
	r.HandleFunc("/get/questions", rt.GetQuestions).Methods("POST")
	r.HandleFunc("/get/free-question", rt.GetFreeQuestion).Methods("GET")

	r.HandleFunc("/cmd-timer", rt.CmdTimer).Methods("POST")

	r.HandleFunc("/loop-streams/start", rt.StartLoopStreams).Methods("POST")
	r.HandleFunc("/loop-streams/stop", rt.StopLoopStreams).Methods("GET")

	//r.HandleFunc("/stream/start", rt.StartStream).Methods("POST")
	r.HandleFunc("/stream/stop", rt.StopStream).Methods("POST")

	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(rt.conf.Port), r))
}