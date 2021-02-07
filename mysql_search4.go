package main

import (
	"database/sql"
)

func AddSearchFor(item map[string]interface{}) (sql.Result, error) {
	sqlQuery := "INSERT INTO `search4` SET "
	sqlQuery += "`task_id` = :task_id, " +
		"`link_title` = :link_title, " +
		"`site_id` = :site_id, " +
		"`cat_id` = :cat_id, " +
		"`title` = :title, " +
		"`keyword` = :keyword"

	res, err := MYSQL.db.NamedExec(sqlQuery, item)

	return res, err
}
