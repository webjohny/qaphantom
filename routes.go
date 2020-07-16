package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Routes struct {
	conf Configuration
	mysql MysqlDb
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
	stream.cmd = commandExec
	stream.StartTaskTimer(500, limit)

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		log.Println("Routes.CmdTimer.HasError", err)
	}
}

func (rt *Routes) StartStreams(count int, limit int, cmd string) {
	fmt.Println("Started")
	for i := 1; i <= count; i++ {
		stream := rt.streams.Add(i)
		if cmd != "" {
			stream.cmd = cmd + " " + strconv.Itoa(i)
		}else{
			stream.job = JobHandler{}
		}
		go stream.Start(i, int64(limit))
	}
}

func (rt *Routes) StopStream(w http.ResponseWriter, r *http.Request) {
	id := utils.toInt(r.FormValue("id"))

	rt.streams.Stop(id)
}

func (rt *Routes) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/get/cats", rt.GetCats).Methods("POST")
	r.HandleFunc("/get/task-stats", rt.GetTasksForStat).Methods("POST")
	r.HandleFunc("/get/tasks", rt.GetTasks).Methods("POST")
	r.HandleFunc("/get/free-task", rt.GetFreeTask).Methods("GET")

	r.HandleFunc("/cmd-timer", rt.CmdTimer).Methods("POST")

	r.HandleFunc("/loop-streams/count", rt.CountLoopStreams).Methods("GET")
	r.HandleFunc("/loop-streams/start", rt.StartLoopStreams).Methods("POST")
	r.HandleFunc("/loop-streams/stop", rt.StopLoopStreams).Methods("GET")

	r.HandleFunc("/run/job", rt.RunJob).Methods("GET")
	r.HandleFunc("/stream/stop", rt.StopStream).Methods("POST")

	log.Fatal(http.ListenAndServe(":" + rt.conf.Port, r))
}