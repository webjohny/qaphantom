package main

import (
	"database/sql"
	"log"
	"strconv"
)

func (m *MysqlDb) GetCats(params map[string]interface{}, postData map[string]interface{}) []MysqlCat {
	var results []MysqlCat

	sqlQuery := "SELECT * FROM `cats`"

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
		log.Println("MysqlCats.GetCats.HasError", err)
	}

	return results
}


func (m *MysqlDb) UpdateCats(data map[string]interface{}, id int) (sql.Result, error) {
	sqlQuery := "UPDATE `cats` SET "

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