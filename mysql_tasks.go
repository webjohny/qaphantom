package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var checkLoopCollect = false

func ShuffleSites(sites []MysqlSite) []MysqlSite {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(sites), func(i, j int) { sites[i], sites[j] = sites[j], sites[i] })
	return sites
}

func (m *MysqlDb) GetFreeTask(id int) MysqlFreeTask {
	var freeTask MysqlFreeTask
	var sites []MysqlSite

	sqlCount := "SELECT COUNT(*) FROM `tasks` WHERE `site_id` = s.id"
	sqlSelectSite := "s.id, s.extra, s.qsts_limit, s.more_tags, s.symb_micro_marking, s.language, s.theme, s.from, s.to, s.qa_count_from, s.qa_count_to, s.login, s.password, s.domain, s.h1, s.sh_format, s.sh_order, s.video_step, s.linking, s.parse_dates, s.parse_doubles, s.parse_fast, s.parse_search4, s.image_source, s.image_key, s.pub_image, (" + sqlCount + ") as count_rows"
	sqlSite := "SELECT " + sqlSelectSite + " FROM sites s"

	err := m.db.Select(&sites, sqlSite)
	if err != nil{
		log.Println("MysqlDb.GetFreeTask.HasError", err)
	}
	sites = ShuffleSites(sites)

	var site MysqlSite
	var siteId int64
	var siteCountTasks int64
	for _, item := range sites {
		if item.CountRows.Int64 > 0 {
			site = item
			siteId = item.Id.Int64
			siteCountTasks = item.CountRows.Int64
			break
		}
	}

	if siteId > 0 {
		freeTask.MergeSite(site)

		t := time.Now()
		now := t.Format("2006-01-02 15:04:05")

		randomOffset := int(siteCountTasks) - 1
		if randomOffset < 1 {
			return freeTask
		}
		randomOffset = rand.Intn(randomOffset)

		var sqlQuery string
		if id > 0 {
			sqlQuery = "SELECT t.id, t.keyword, t.try_count, c.title AS cat, t.site_id, t.cat_id FROM tasks t"
			sqlQuery += " LEFT JOIN cats c ON (c.id = t.cat_id)"
			sqlQuery += " AND t.id = " + strconv.Itoa(id)
		}else{
			sqlQuery = "SELECT t.id, t.keyword, t.try_count, c.title AS cat, t.site_id, t.cat_id FROM tasks t"
			sqlQuery += " LEFT JOIN cats c ON (c.id = t.cat_id)"
			sqlQuery += " WHERE t.site_id = "
			sqlQuery += strconv.Itoa(int(siteId))
			sqlQuery += " AND (t.try_count IS NULL OR t.try_count <= 5)"
			sqlQuery += " AND (t.status IS NULL OR t.status = 0) AND (t.timeout is NULL OR t.timeout < '"
			sqlQuery += now
			sqlQuery += "') ORDER BY RAND() LIMIT 1"
		}

		fmt.Println(sqlQuery)

		var task MysqlTask
		err := m.db.Get(&task, sqlQuery)
		if err != nil{
			log.Println("MysqlDb.GetFreeTask.2.HasError", err)
		}
		freeTask.MergeTask(task)
		freeTask.SavingAvailable = freeTask.QstsLimit > freeTask.CountRows
	}
	return freeTask
}

func (m *MysqlDb) GetTaskByKeyword(k string) MysqlTask {
	var result MysqlTask
	sqlQuery := "SELECT * FROM `tasks` WHERE `keyword` = ? LIMIT 1"

	err := m.db.Get(&result, sqlQuery, k)
	if err != nil {
		//log.Println("MysqlDb.GetTaskByKeyword.HasError", err)
	}

	return result
}

func (m *MysqlDb) GetCountTasks(params map[string]interface{}) int {
	rows, _ := m.db.Query("SELECT COUNT(*) as count FROM `tasks`")
	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Println("MysqlDb.GetCountTasks.HasError", err)
		}
	}
	return count
}

func (m *MysqlDb) GetTasks(params map[string]interface{}) []MysqlTask {
	var results []MysqlTask

	//fmt.Println(task.ParseDate.String)
	sqlQuery := "SELECT * FROM `tasks`"

	if len(params) > 0{
		if params["isStat"] != 0 {
			sqlQuery = "SELECT id, site_id, cat_id, status FROM `tasks`"
		}
	}
	sqlQuery = sqlQuery + " ORDER BY `id`"

	if len(params) > 0{
		if params["limit"] != 0 {
			if params["offset"] != 0 {
				sqlQuery = sqlQuery + "LIMIT " + strconv.Itoa(params["offset"].(int)) + ", " + strconv.Itoa(params["limit"].(int))
			}else{
				sqlQuery = sqlQuery + " LIMIT " + strconv.Itoa(params["limit"].(int))
			}
		}
	}

	err := m.db.Select(&results, sqlQuery)
	if err != nil {
		log.Println("MysqlDb.GetTasks.HasError", err)
	}

	return results
}

func (m *MysqlDb) UpdateTask(data map[string]interface{}, id int) (sql.Result, error) {
	sqlQuery := "UPDATE `tasks` SET "

	if len(data) > 0 {
		updateQuery := ""
		i := 0
		for k, v := range data {
			if i > 0 {
				updateQuery += ", "
			}
			updateQuery += "`" + k + "` = "
			if v == "NULL" {
				updateQuery += "NULL"
			}else{
				updateQuery += ":" + k
			}
			//data[k] = utils.MysqlRealEscapeString(v.(string))
			i++
		}
		sqlQuery += updateQuery
	}

	sqlQuery += " WHERE `id` = " + strconv.Itoa(id)

	res, err := m.db.NamedExec(sqlQuery, data)

	return res, err
}

func (m *MysqlDb) AddTask(item map[string]interface{}) (sql.Result, error) {
	sqlQuery := "INSERT INTO `tasks` SET "
	sqlQuery += "`site_id` = :site_id, " +
		"`cat_id` = :cat_id, " +
		"`keyword` = :keyword, " +
		"`parent_id` = :parent_id, " +
		"`parser` = NULL, " +
		"`error` = NULL"

	res, err := m.db.NamedExec(sqlQuery, item)

	return res, err
}

func (m *MysqlDb) LoopCollectStats() {
	if ! checkLoopCollect {
		checkLoopCollect = true
		for {
			count := m.GetCountTasks(map[string]interface{}{})
			fmt.Println(count)
			if count < 20000 {
				break
			}
			m.CollectStats()
			time.Sleep(2 * time.Minute)
		}
	}
}


func (m *MysqlDb) CollectStats() map[int64]map[string]interface{} {
	params := make(map[string]interface{})

	limit := 20000
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
			//tasks := []MysqlTask{}
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

		if count > 20000 {
			for k, v := range stat {
				info, err := json.Marshal(v)
				if err != nil {
					log.Println("MysqlDb.CollectStats.HasError", err)
				}
				item := map[string]interface{}{
					"info": info,
				}
				res, err := m.UpdateSite(item, int(k))
				fmt.Println(res)
				if err != nil {
					log.Println("MysqlDb.CollectStats.2.HasError", err)
				}
			}
		}
	}
	return stat
}