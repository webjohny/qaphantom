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

type StdResponse struct {
	status bool
	msg string
}

type InsertResponse struct {
	Status bool
	InsertedId interface{}
}

func (rt *Routes) CmdTimer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	commandExec := r.FormValue("cmd")

	result := StdResponse{
		status: true,
	}

	var limit int64 = 1000

	if len(r.FormValue("limit")) > 0 {
		limit, _ = strconv.ParseInt(r.FormValue("limit"), 0, 64)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1000 * limit) * time.Millisecond)
	defer cancel()
	//php -f /var/www/html/cron.php parser cron sleeping 5
	_, err := exec.CommandContext(ctx, "bash", "-c", commandExec).Output()

	if err != nil {
		// This will fail after 100 milliseconds. The 5 second sleep
		// will be interrupted.
		result.status = false
		fmt.Println(err)
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) CheckQuestion(w http.ResponseWriter, r *http.Request) {
	id := rt.utils.toInt(r.FormValue("id"))
	keyword := r.FormValue("keyword")

	rt.mongo.CheckQuestionByKeyword(keyword, id)
}

func (rt *Routes) CheckQuestions(w http.ResponseWriter, r *http.Request) {
}

func (rt *Routes) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	data := rt.utils.ParseFormCollection(r, "data")

	rt.mongo.UpdateQuestion(data, id)
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
	response := InsertResponse{
		Status: false,
	}

	if err != nil {
		fmt.Println(err)
	}else{
		response.InsertedId = res.InsertedID
		response.Status = true
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) Run() {
	rt.utils = Utils{}

	r := mux.NewRouter()

	r.HandleFunc("/cmd-timer", rt.CmdTimer).Methods("POST")
	r.HandleFunc("/check/question", rt.CheckQuestion).Methods("POST")
	r.HandleFunc("/check/questions", rt.CheckQuestions).Methods("POST")
	r.HandleFunc("/update/question", rt.UpdateQuestion).Methods("POST")
	r.HandleFunc("/insert/question", rt.InsertQuestion).Methods("POST")

	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(rt.conf.Port), r))
}