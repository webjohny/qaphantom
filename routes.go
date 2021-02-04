package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Routes struct {}

func (rt *Routes) cmdTimer(w http.ResponseWriter, r *http.Request) {
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

func (rt *Routes) stopStream(w http.ResponseWriter, r *http.Request) {
	id := UTILS.toInt(r.FormValue("id"))

	STREAMS.Stop(id)
}

func (rt *Routes) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/get/cats", rt.getCats).Methods("POST")
	r.HandleFunc("/get/task-stats", rt.getTasksForStat).Methods("POST")
	r.HandleFunc("/get/tasks", rt.getTasks).Methods("POST")
	r.HandleFunc("/get/free-task", rt.getFreeTask).Methods("GET")

	r.HandleFunc("/cmd-timer", rt.cmdTimer).Methods("POST")

	r.HandleFunc("/loop-streams/count", rt.countLoopStreams).Methods("GET")
	r.HandleFunc("/loop-streams/start", rt.startLoopStreams).Methods("POST")
	r.HandleFunc("/loop-streams/stop", rt.stopLoopStreams).Methods("GET")

	r.HandleFunc("/run/job", rt.runJob).Methods("GET")
	r.HandleFunc("/stream/stop", rt.stopStream).Methods("POST")

	r.HandleFunc("/test/proxy", rt.testProxy).Methods("GET")

	log.Fatal(http.ListenAndServe(":" + CONF.Port, r))
}