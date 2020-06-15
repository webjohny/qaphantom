package main

import (
	"database/sql"
	"log"
	"time"
)

func (m *MysqlDb) GetResultByQAndA(q string, a string) MysqlResult {
	var result MysqlResult
	sqlQuery := "SELECT * FROM `results` WHERE `q` = '" + q + "' AND `a` = '" + a + "' LIMIT 1"

	err := m.db.Get(&result, sqlQuery)
	if err != nil {
		log.Println(err)
	}

	return result
}

func (m *MysqlDb) AddResult(item map[string]interface{}) (sql.Result, error) {
	t := time.Now()
	now := t.Format("2006-01-02 15:04:05")

	if _, ok := item["create_date"]; !ok {
		item["create_date"] = now
	}

	if _, ok := item["qa_date"]; !ok {
		item["qa_date"] = now
	}

	sqlQuery := "INSERT INTO `results` SET "
	sqlQuery += "`task_id` = :task_id, " +
		"`q` = :q, " +
		"`a` = :a, " +
		"`link` = :link, " +
		"`link_title` = :link_title, " +
		"`create_date` = :create_date, " +
		"`qa_date` = :qa_date"

	res, err := m.db.NamedExec(sqlQuery, item)

	return res, err
}
