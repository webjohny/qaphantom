package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

func (s *MysqlSite) GetInfo() map[string]interface{} {
	var result map[string]interface{}

	if s.Info.String != "" && s.Info.String != "[]" {
		if err := json.Unmarshal([]byte(s.Info.String), &result); err != nil {
			fmt.Println(err)
		}
	}
	return result
}

func (m *MysqlDb) GetSites(params map[string]interface{}, postData map[string]interface{}) []MysqlSite {
	var results []MysqlSite

	//fmt.Println(task.ParseDate.String)
	sqlQuery := "SELECT * FROM `sites`"

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
		log.Println("MysqlSites.GetSites.HasError", err)
	}

	return results
}

func (m *MysqlDb) UpdateSite(data map[string]interface{}, id int) (sql.Result, error) {
	sqlQuery := "UPDATE `sites` SET "

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
