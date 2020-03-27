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
}

type StatusResponse struct {
	status bool
	msg string
}

func (th *Routes) CmdTimer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	commandExec := r.FormValue("cmd")

	result := StatusResponse{
		status: true,
	}

	var limit int64 = 1000

	if len(r.FormValue("limit")) > 0 {
		limit, _ = strconv.ParseInt(r.FormValue("limit"), 0, 64)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1000 * limit) * time.Millisecond)
	defer cancel()
	//php -f /var/www/html/cron.php parser cron sleeping 5
	out, err := exec.CommandContext(ctx, "bash", "-c", commandExec).Output()

	if err != nil {
		// This will fail after 100 milliseconds. The 5 second sleep
		// will be interrupted.
		result.status = false
		fmt.Println(err)
	}
	fmt.Println(string(out))

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		fmt.Println(err)
	}
}

func (th *Routes) CheckQuestion(w http.ResponseWriter, r *http.Request) {

}

func (th *Routes) CheckQuestions(w http.ResponseWriter, r *http.Request) {

}

func (th *Routes) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	data := r.FormValue("data")
	fmt.Println(id)
	fmt.Println(data)

	//th.mongo.UpdateQuestion(bson.M{}, id)
}

func (th *Routes) InsertQuestion(w http.ResponseWriter, r *http.Request) {

}

func (th *Routes) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/cmd-timer", th.CmdTimer).Methods("POST")
	r.HandleFunc("/check/question", th.CheckQuestion).Methods("POST")
	r.HandleFunc("/check/questions", th.CheckQuestions).Methods("POST")
	r.HandleFunc("/update/question", th.UpdateQuestion).Methods("POST")
	r.HandleFunc("/insert/question", th.InsertQuestion).Methods("POST")

	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(th.conf.Port), r))
}