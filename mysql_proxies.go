package main

import (
	"database/sql"
	"fmt"
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
		fmt.Println(err)
	}

	proxy := MysqlProxy{}

	if len(proxies) > 0 {
		proxy = proxies[0]
		proxy.Mysql = m
	}

	return proxy
}

func (p MysqlProxy) SetTimeout(parser int) sql.Result {
	now := time.Now().Local().Add(time.Minute * time.Duration(5))
	formattedDate := now.Format("2006-01-02 15:04:05")

	data := map[string]interface{}{}
	data["parser"] = strconv.Itoa(parser)
	data["timeout"] = formattedDate

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

	sqlQuery += " WHERE `id` = " + strconv.Itoa(int(p.Id.Int64))

	res, err := p.Mysql.db.NamedExec(sqlQuery, data)
	if err != nil {
		fmt.Println(err)
	}

	return res
}

func (p MysqlProxy) FreeProxy() {
	data := map[string]interface{}{}
	data["parser"] = "NULL"
	data["timeout"] = "NULL"

	_, err := p.Mysql.UpdateProxy(data, int(p.Id.Int64))
	if err != nil {
		fmt.Println(err)
	}
}
