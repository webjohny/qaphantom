package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

func (m *MysqlDb) GetCountTasks(params map[string]interface{}) int {
	rows, _ := m.db.Query("SELECT COUNT(*) as count FROM `tasks`")
	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			fmt.Println(err)
		}
	}
	return count
}


func (m *MysqlDb) GetTasks(params map[string]interface{}) []MysqlTask {
	var results []MysqlTask

	//fmt.Println(task.ParseDate.String)
	sqlQuery := "SELECT * FROM `tasks`"

	sqlQuery = sqlQuery + " ORDER BY `id`"

	if len(params) > 0{
		if params["limit"] != 0 {
			sqlQuery = sqlQuery + " LIMIT " + strconv.Itoa(params["limit"].(int))
			if params["offset"] != 0 {
				sqlQuery = sqlQuery + ", " + strconv.Itoa(params["offset"].(int))
			}
		}
	}

	err := m.db.Select(&results, sqlQuery)
	if err != nil {
		panic(err)
	}

	return results
}


func (m *MysqlDb) InsertTask(question Question) {

}

func (m *MysqlDb) UpdateTask(data map[string]interface{}, id int) (sql.Result, error) {
	sqlQuery := "UPDATE `tasks` SET "

	if len(data) > 0 {
		updateQuery := ""
		for k, _ := range data {
			updateQuery += "`" + k + "` = :" + k
		}
		sqlQuery += updateQuery
	}

	sqlQuery += " WHERE `id` = " + strconv.Itoa(id)

	res, err := m.db.NamedExec(sqlQuery, data)

	return res, err
}

func (m *MysqlDb) LoopCollectStats() {
	if ! checkLoopCollect {
		checkLoopCollect = true
		for {
			count := m.GetCountTasks(map[string]interface{}{})
			fmt.Println(count)
			if count < 10000 {
				break
			}
			m.CollectStats()
			time.Sleep(5 * time.Minute)
		}
	}
}


func (m *MysqlDb) CollectStats() map[int64]map[string]interface{} {
	params := make(map[string]interface{})

	limit := 5000
	offset := 0
	stat := map[int64]map[string]interface{}{}

	count := m.GetCountTasks(map[string]interface{}{})
	var parts int

	if limit > count {
		parts = 1
	} else {
		parts = int(math.Ceil(float64(count) / float64(limit)))
	}
	//fmt.Println(parts)
	//parts = 1

	cats := m.GetCats(map[string]interface{}{}, map[string]interface{}{})
	sites := m.GetSites(map[string]interface{}{}, map[string]interface{}{})

	//notCorrectData := make([]interface{}, 0)
	if true {
		for i := 0; i < parts; i++ {
			params["limit"] = limit
			params["offset"] = offset
			params["isStat"] = true
			tasks := m.GetTasks(params)
			if true {
				for _, task := range tasks {
					SiteId := task.SiteId.Int64
					CatId := task.CatId.Int64
					Status := task.Status.Int32

					var Site MysqlSite
					var Cat MysqlCat

					if SiteId == 0 {
						//notCorrectData = append(notCorrectData, question)
						continue
					}

					for _, cat := range cats {
						if CatId == cat.Id.Int64 {
							Cat = cat
						}
					}

					for _, site := range sites {
						if SiteId == site.Id.Int64 {
							Site = site
						}
					}

					if CatId != 0 {
						site := map[string]interface{}{}

						if item, ok := stat[SiteId]; ok {
							site = item
						}

						if Site.Domain.Valid {
							site["domain"] = Site.Domain.String
						}

						if _, ok := site["ready"]; ! ok {
							site["ready"] = 0
						}

						if _, ok := site["error"]; ! ok {
							site["error"] = 0
						}

						if _, ok := site["total"]; ! ok {
							site["total"] = 0
						}

						cetegors := map[int64]interface{}{}
						cat := map[string]interface{}{}

						_, ok := site["cats"]
						if ok && len(site["cats"].(map[int64]interface{})) > 0 {
							cetegors = site["cats"].(map[int64]interface{})

							_, ok := cetegors[CatId]
							if ok && len(cetegors[CatId].(map[string]interface{})) > 0 {
								cat = cetegors[CatId].(map[string]interface{})
							}
						}

						cat["title"] = Cat.Title.String

						if _, ok := cat["ready"]; ! ok {
							cat["ready"] = 0
						}

						if _, ok := cat["error"]; ! ok {
							cat["error"] = 0
						}

						if _, ok := cat["total"]; ! ok {
							cat["total"] = 0
						}

						if Status == 2 {
							site["error"] = site["error"].(int) + 1
							cat["error"] = cat["error"].(int) + 1
						} else if Status == 1 {
							site["ready"] = site["ready"].(int) + 1
							cat["ready"] = cat["ready"].(int) + 1
						}

						site["total"] = site["total"].(int) + 1
						cat["total"] = cat["total"].(int) + 1

						cetegors[CatId] = cat
						site["cats"] = cetegors

						stat[SiteId] = site
					}
				}

				if count < offset {
					offset = offset - (offset - count)
				} else {
					offset = offset + limit
				}
				time.Sleep(time.Second)
			}
		}

		if count > 10000 {
			for k, v := range stat {
				info, err := json.Marshal(v)
				if err != nil {
					fmt.Println(err)
				}
				item := map[string]interface{}{
					"info": info,
				}
				res, err := m.UpdateSite(item, int(k))
				fmt.Println(res)
				fmt.Println(err)
			}
		}
	}
	return stat
}


func (m *MysqlDb) GetFreeTask(ids []string) map[string]interface{} {
	var result map[string]interface{}

	return result
}