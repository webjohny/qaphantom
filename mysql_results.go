package main

import (
	"database/sql"
	"strconv"
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

func (m *MysqlDb) GetResultByQ(q string) MysqlResult {
	var result MysqlResult
	sqlQuery := "SELECT * FROM `results` WHERE `q` = ? LIMIT 1"

	err := m.db.Get(&result, sqlQuery, q)
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
		"`text` = :text, " +
		"`html` = :html, " +
		"`link` = :link, " +
		"`create_date` = :create_date, " +
		"`qa_date` = :qa_date"

	res, err := m.db.NamedExec(sqlQuery, item)

	return res, err
}

func (m *MysqlDb) UpdateResult(data map[string]interface{}, id int) (sql.Result, error) {
	sqlQuery := "UPDATE `results` SET "

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

func (m *MysqlDb) InsertOrUpdateResult(item map[string]interface{}) (sql.Result, error) {
	result := m.GetResultByQ(item["q"].(string))
	var res sql.Result
	var err error
	if !result.Id.Valid {
		res, err = m.AddResult(item)
	}else {
		res, err = m.UpdateResult(item, int(result.Id.Int64))
	}

	return res, err
}
