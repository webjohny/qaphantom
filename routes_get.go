package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (rt *Routes) GetFreeTask(w http.ResponseWriter, r *http.Request) {
	dataIds := r.FormValue("ids")

	var ids []string

	if dataIds != "" {
		ids = strings.Split(dataIds, ",")
	}
	fmt.Println(ids)

	question := MYSQL.GetFreeTask(0)

	err := json.NewEncoder(w).Encode(question)
	if err != nil {
		log.Println("RoutesGet.GetFreeTask.HasError", err)
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

	cats := MYSQL.GetCats(params, postData)

	err := json.NewEncoder(w).Encode(cats)
	if err != nil {
		log.Println("RoutesGet.GetCats.HasError", err)
	}
}

func (rt *Routes) GetTasksForStat(w http.ResponseWriter, r *http.Request) {

	count := MYSQL.GetCountTasks(map[string]interface{}{})

	stat := map[int64]map[string]interface{}{}

	//stat = MYSQL.CollectStats()

	go MYSQL.LoopCollectStats()

	if count > 10000 {
		sites := MYSQL.GetSites(map[string]interface{}{}, map[string]interface{}{})
		if len(sites) > 0 {
			for _, site := range sites {
				info := site.GetInfo()
				if info != nil {
					stat[site.Id.Int64] = info
				}
			}
		}
	}else{
		stat = MYSQL.CollectStats()
	}

	err := json.NewEncoder(w).Encode(stat)
	if err != nil {
		log.Println("RoutesGet.GetTasksForStat.HasError", err)
	}
}

func (rt *Routes) GetTasks(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})

	tasks := MYSQL.GetTasks(params)

	err := json.NewEncoder(w).Encode(tasks)
	if err != nil {
		log.Println("RoutesGet.GetTasks.HasError", err)
	}
}

func (rt *Routes) TestProxy(w http.ResponseWriter, r *http.Request) {
	var errMsg string

	host := r.FormValue("host")
	port := r.FormValue("port")
	login := r.FormValue("login")
	password := r.FormValue("password")

	//host := "89.191.225.148"
	//port := "45785"
	//login := "phillip"
	//password := "I2n9BeJ"

	if host == "" {
		errMsg = "undefined host"
	} else if port == "" {
		errMsg = "undefined port"
	} else {
		proxy := &Proxy{
			Id:       1,
			Host:     host,
			Port:     port,
			Login:    login,
			Password: password,
			LocalIp:  host + ":" + port,
		}
		browser := Browser{}
		browser.Proxy = proxy
		browser.Init()

		ctx, cancel := context.WithTimeout(browser.ctx, time.Second*15)
		browser.CancelTimeout = cancel
		browser.ctx = ctx
		defer browser.Cancel()

		status, buffer := browser.ScreenShot("https://www.google.com/search?hl=en&gl=us&q=what+is+my+ip")
		if !status {
			errMsg = string(buffer)
		} else {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Length", strconv.Itoa(len(buffer)))
			if _, err := w.Write(buffer); err != nil {
				errMsg = "unable to write image"
			}else{
				return
			}
		}
	}
	_, _ = w.Write([]byte(errMsg))
}
