package main

import (
	"fmt"
	"log"
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
