package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (rt *Routes) GetFreeTask(w http.ResponseWriter, r *http.Request) {
	dataIds := r.FormValue("ids")

	var ids []string

	if dataIds != "" {
		ids = strings.Split(dataIds, ",")
	}
	fmt.Println(ids)

	question := mysql.GetFreeTask(ids)

	err := json.NewEncoder(w).Encode(question)
	if err != nil {
		log.Println(err)
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

	cats := mysql.GetCats(params, postData)

	err := json.NewEncoder(w).Encode(cats)
	if err != nil {
		log.Println(err)
	}
}

func (rt *Routes) GetTasksForStat(w http.ResponseWriter, r *http.Request) {

	count := mysql.GetCountTasks(map[string]interface{}{})

	stat := map[int64]map[string]interface{}{}

	//stat = mysql.CollectStats()

	go mysql.LoopCollectStats()

	if count > 10000 {
		sites := mysql.GetSites(map[string]interface{}{}, map[string]interface{}{})
		if len(sites) > 0 {
			for _, site := range sites {
				info := site.GetInfo()
				if info != nil {
					stat[site.Id.Int64] = info
				}
			}
		}
	}else{
		stat = mysql.CollectStats()
	}

	err := json.NewEncoder(w).Encode(stat)
	if err != nil {
		log.Println(err)
	}
}

func (rt *Routes) GetTasks(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})

	tasks := mysql.GetTasks(params)

	err := json.NewEncoder(w).Encode(tasks)
	if err != nil {
		log.Println(err)
	}
}
