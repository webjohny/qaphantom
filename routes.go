package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	status := stream.StartTaskTimer(commandExec, limit)

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": status,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) GetFreeTask(w http.ResponseWriter, r *http.Request) {
	dataIds := r.FormValue("ids")

	var ids []string

	if dataIds != "" {
		ids = strings.Split(dataIds, ",")
	}
	fmt.Println(ids)

	question := rt.mysql.GetFreeTask(ids)

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

	cats := rt.mysql.GetCats(params, postData)

	err := json.NewEncoder(w).Encode(cats)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) GetTasksForStat(w http.ResponseWriter, r *http.Request) {

	count := rt.mysql.GetCountTasks(map[string]interface{}{})

	stat := map[int64]map[string]interface{}{}

	//stat = rt.mysql.CollectStats()

	go rt.mysql.LoopCollectStats()

	if count > 10000 {
		sites := rt.mysql.GetSites(map[string]interface{}{}, map[string]interface{}{})
		if len(sites) > 0 {
			for _, site := range sites {
				info := site.GetInfo()
				if info != nil {
					stat[site.Id.Int64] = info
				}
			}
		}
	}else{
		stat = rt.mysql.CollectStats()
	}

	err := json.NewEncoder(w).Encode(stat)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) GetTasks(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})

	tasks := rt.mysql.GetTasks(params)

	err := json.NewEncoder(w).Encode(tasks)
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) StartStreams(count int, limit int, cmd string) {
	fmt.Println("Started")
	for i := 1; i <= count; i++ {
		stream := rt.streams.Add(i)
		stream.cmd = cmd + " " + strconv.Itoa(i)
		go stream.Start(i, int64(limit*1000))
	}
}

func (rt *Routes) ReStartStreams(count int, limit int, cmd string) {
	fmt.Println("Restarted streams")
	rt.streams.StopAllWithoutClean()
	time.Sleep(time.Minute * 5)
	rt.StartStreams(count, limit, cmd)
}

func (rt *Routes) StartLoopStreams(w http.ResponseWriter, r *http.Request) {
	count := rt.utils.toInt(r.FormValue("count"))
	limit := rt.utils.toInt(r.FormValue("limit"))
	cmd := r.FormValue("cmd")

	if limit < 1 {
		limit = 10
	}

	rt.streams.StopAll()

	var intval chan bool

	intval = rt.utils.SetInterval(func() {
		if !rt.streams.isStarted {
			intval <- true
			fmt.Println("Stopped interval")
		}else {
			rt.ReStartStreams(count, limit, cmd)
		}
	}, 6200000, true)

	rt.streams.isStarted = true
	go rt.StartStreams(count, limit, cmd)

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) StopLoopStreams(w http.ResponseWriter, r *http.Request) {
	rt.streams.isStarted = false
	go rt.streams.StopAll()

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (rt *Routes) CountLoopStreams(w http.ResponseWriter, r *http.Request) {
	count := rt.streams.Count()

	_, err := fmt.Fprintln(w, strconv.Itoa(count))
	if err != nil {
		panic(err)
	}
}

func (rt *Routes) StopStream(w http.ResponseWriter, r *http.Request) {
	id := rt.utils.toInt(r.FormValue("id"))

	rt.streams.Stop(id)
}

func (rt *Routes) Run() {
	rt.utils = Utils{}

	r := mux.NewRouter()

	r.HandleFunc("/get/cats", rt.GetCats).Methods("POST")
	r.HandleFunc("/get/task-stats", rt.GetTasksForStat).Methods("POST")
	r.HandleFunc("/get/tasks", rt.GetTasks).Methods("POST")
	r.HandleFunc("/get/free-task", rt.GetFreeTask).Methods("GET")

	r.HandleFunc("/cmd-timer", rt.CmdTimer).Methods("POST")

	r.HandleFunc("/loop-streams/count", rt.CountLoopStreams).Methods("GET")
	r.HandleFunc("/loop-streams/start", rt.StartLoopStreams).Methods("POST")
	r.HandleFunc("/loop-streams/stop", rt.StopLoopStreams).Methods("GET")

	//r.HandleFunc("/stream/start", rt.StartStream).Methods("POST")
	r.HandleFunc("/stream/stop", rt.StopStream).Methods("POST")

	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(rt.conf.Port), r))
}