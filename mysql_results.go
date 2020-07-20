package main

import (
	"database/sql"
	"time"
)

func (m *MysqlDb) GetResultByQAndA(q string, a string) MysqlResult {
	var result MysqlResult
	sqlQuery := "SELECT * FROM `results` WHERE `q` = ? AND `a` = ? LIMIT 1"

	err := m.db.Get(&result, sqlQuery, q, a)
	if err != nil {
		//log.Println("MysqlDb.GetResultByQAndA.HasError", err)
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
		"`link_title` = :link_title, " +
		"`site_id` = :site_id, " +
		"`cat_id` = :cat_id, " +
		"`domain` = :domain, " +
		"`cat` = :cat, " +
		"`q` = :q, " +
		"`link` = :link, " +
		"`create_date` = :create_date, " +
		"`qa_date` = :qa_date"

	res, err := m.db.NamedExec(sqlQuery, item)

	return res, err
}
