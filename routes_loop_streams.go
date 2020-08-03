package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (rt *Routes) StartLoopStreams(w http.ResponseWriter, r *http.Request) {
	count := utils.toInt(r.FormValue("count"))
	limit := utils.toInt(r.FormValue("limit"))
	cmd := r.FormValue("cmd")

	if limit < 1 {
		limit = 10
	}

	config := mysql.GetConfig()
	extra := config.GetExtra()
	extra.CmdStreams = cmd
	extra.LimitStreams = limit
	extra.CountStreams = count
	_ = mysql.SetExtra(extra)

	streams.StartLoop(count, limit, cmd)

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		log.Println("Routes.StartLoopStreams.HasError", err)
	}
}

func (rt *Routes) StopLoopStreams(w http.ResponseWriter, r *http.Request) {
	streams.isStarted = false
	go streams.StopAll()

	err := json.NewEncoder(w).Encode(map[string]bool{
		"status": true,
	})
	if err != nil {
		log.Println("Routes.StopLoopStreams.HasError", err)
	}
}

func (rt *Routes) CountLoopStreams(w http.ResponseWriter, r *http.Request) {
	count := streams.Count()

	_, err := fmt.Fprintln(w, strconv.Itoa(count))
	if err != nil {
		log.Println("Routes.CountLoopStreams.HasError", err)
	}
}
