package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (rt *Routes) ReStartStreams(count int, limit int, cmd string) {
	fmt.Println("Restarted streams")
	rt.streams.StopAllWithoutClean()
	time.Sleep(time.Second * 600)
	go rt.StartStreams(count, limit, cmd)
}

func (rt *Routes) StartLoopStreams(w http.ResponseWriter, r *http.Request) {
	count := utils.toInt(r.FormValue("count"))
	limit := utils.toInt(r.FormValue("limit"))
	cmd := r.FormValue("cmd")

	if limit < 1 {
		limit = 10
	}

	rt.streams.StopAll()

	var restartFunc func()

	restartFunc = func() {
		if rt.streams.isStarted {
			rt.ReStartStreams(count, limit, cmd)
			time.AfterFunc(time.Second * 3600, restartFunc)
		}
	}
	time.AfterFunc(time.Second * 3600, restartFunc)

	rt.streams.isStarted = true
	go rt.StartStreams(count, limit, cmd)

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		log.Println("Routes.StartLoopStreams", err)
	}
}

func (rt *Routes) StopLoopStreams(w http.ResponseWriter, r *http.Request) {
	rt.streams.isStarted = false
	go rt.streams.StopAll()

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		log.Println("Routes.StopLoopStreams", err)
	}
}

func (rt *Routes) CountLoopStreams(w http.ResponseWriter, r *http.Request) {
	count := rt.streams.Count()

	_, err := fmt.Fprintln(w, strconv.Itoa(count))
	if err != nil {
		log.Println("Routes.CountLoopStreams", err)
	}
}
