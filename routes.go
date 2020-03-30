package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

type Routes struct {
	conf Configuration
	mongo MongoDb
	utils Utils
}

var streams map[int]bool

func (rt *Routes) CmdTimer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	commandExec := r.FormValue("cmd")

	var limit int64 = 1000

	if len(r.FormValue("limit")) > 0 {
		limit, _ = strconv.ParseInt(r.FormValue("limit"), 0, 64)
	}

	status := startTaskTimer(commandExec, limit)

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": status,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func startTaskTimer(commandExec string, limit int64) bool {
	var status bool = true

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1000 * limit) * time.Millisecond)
	defer cancel()
	//php -f /var/www/html/cron.php parser cron sleeping 5
	_, err := exec.CommandContext(ctx, "bash", "-c", commandExec).Output()

	if err != nil {
		// This will fail after 100 milliseconds. The 5 second sleep
		// will be interrupted.
		status = false
		fmt.Println(err)
	}

	return status
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
	question.CatId = rt.utils.toInt(r.FormValue("CatId"))
	question.TryCount = rt.utils.toInt(r.FormValue("TryCount"))
	question.ErrorsCount = rt.utils.toInt(r.FormValue("ErrorsCount"))
	question.Status = rt.utils.toInt(r.FormValue("Status"))
	question.Error = r.FormValue("Error")
	question.ParserId = rt.utils.toInt(r.FormValue("ParserId"))
	question.Timeout = time.Now()
	question.Keyword = r.FormValue("Keyword")
	question.FastA = r.FormValue("FastA")
	question.FastLink = r.FormValue("FastLink")
	question.FastLinkTitle = r.FormValue("FastLinkTitle")
	question.FastDate = time.Now()

	res, err := rt.mongo.InsertQuestion(question)
	response := map[string]interface{}{
		"status": false,
	}

	if err != nil {
		fmt.Println(err)
	}else{
		response["insertedId"] = res.InsertedID
		response["status"] = true
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) StartLoopStreams(w http.ResponseWriter, r *http.Request) {

}

func (rt *Routes) StopLoopStreams(w http.ResponseWriter, r *http.Request) {

}

func (rt *Routes) StartStream(w http.ResponseWriter, r *http.Request) {
	streamId := rt.utils.toInt(r.FormValue("streamId"))
	streams[streamId] = true

	for {
		if ! streams[streamId] {
			break
		}
		startTaskTimer("php -f /var/www/html/cron.php parser cron sleeping 5", 1000)
	}
}

func (rt *Routes) StopStream(w http.ResponseWriter, r *http.Request) {
	streamId := rt.utils.toInt(r.FormValue("streamId"))
	if streams[streamId] {
		streams[streamId] = false
	}
}

func (rt *Routes) Run() {
	rt.utils = Utils{}

	r := mux.NewRouter()

	r.HandleFunc("/check/question", rt.CheckQuestion).Methods("POST")
	r.HandleFunc("/check/questions", rt.CheckQuestions).Methods("POST")
	r.HandleFunc("/update/question", rt.UpdateQuestion).Methods("POST")
	r.HandleFunc("/insert/question", rt.InsertQuestion).Methods("POST")

	r.HandleFunc("/cmd-timer", rt.CmdTimer).Methods("POST")

	r.HandleFunc("/loop-streams/start", rt.StartLoopStreams).Methods("POST")
	r.HandleFunc("/loop-streams/stop", rt.StopLoopStreams).Methods("POST")

	r.HandleFunc("/stream/start", rt.StartStream).Methods("POST")
	r.HandleFunc("/stream/stop", rt.StopStream).Methods("POST")

	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(rt.conf.Port), r))
}