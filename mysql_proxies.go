package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (m *MysqlDb) GetFreeProxy() MysqlProxy {
	var proxies []MysqlProxy

	t := time.Now()
	now := t.Format("2006-01-02 15:04:05")
	sqlQuery := "SELECT * FROM `proxy` WHERE (status is NULL OR status = 0) AND (timeout is NULL OR timeout < '" + now + "') ORDER BY RAND() LIMIT 1"
	fmt.Println(sqlQuery)

	err := m.db.Select(&proxies, sqlQuery)
	if err != nil {
		log.Println(err)
	}

	proxy := MysqlProxy{}

	if len(proxies) > 0 {
		proxy = proxies[0]
		proxy.Mysql = m
	}

	return proxy
}

func (m *MysqlDb) UpdateProxy(data map[string]interface{}, id int) (sql.Result, error) {
	sqlQuery := "UPDATE `proxy` SET "

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
			i++
		}
		sqlQuery += updateQuery
	}

	sqlQuery += " WHERE `id` = " + strconv.Itoa(id)

	res, err := m.db.NamedExec(sqlQuery, data)

	return res, err
}