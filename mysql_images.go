package main

import (
	"database/sql"
	"log"
	"strconv"
)

func (m *MysqlDb) GetImgTakeFree(siteId int, keyword string, s bool) MysqlImage {
	var image MysqlImage
	var source string
	if s {
		source = "1"
	}else{
		source = "0"
	}

	sqlQuery := "SELECT * FROM `images` WHERE `site_id` = '" + strconv.Itoa(siteId) + "' "
	sqlQuery += "AND `keyword` = '" + keyword + "' "
	sqlQuery += "AND `source` = '" + source + "' "
	sqlQuery += "AND `status` = 1 "
	sqlQuery += "LIMIT 1"

	err := m.db.Get(&image, sqlQuery)
	if err != nil {
		log.Println("MysqlImages.GetImgTakeFree.HasError", err)
	}

	if image.Id.Int64 > 0 {
		_, err = m.UpdateImg(map[string]interface{}{
			"status": 1,
		}, int(image.Id.Int64))
	}

	return image
}

func (m *MysqlDb) UpdateImg(data map[string]interface{}, id int) (sql.Result, error) {
	sqlQuery := "UPDATE `images` SET "

	if len(data) > 0 {
		updateQuery := ""
		var i int
		for k, _ := range data {
			if i != 0 {
				updateQuery += ", "
			}
			updateQuery += "`" + k + "` = :" + k
			i++
		}
		sqlQuery += updateQuery
	}

	sqlQuery += " WHERE `id` = " + strconv.Itoa(id)

	res, err := m.db.NamedExec(sqlQuery, data)
	if err != nil {
		log.Println("MysqlImages.UpdateImg.HasError", err)
	}

	return res, err
}

func (m *MysqlDb) AddImg(data map[string]interface{}) (sql.Result, error) {
	sqlQuery := "INSERT INTO `images` SET "

	if len(data) > 0 {
		insertQuery := ""
		var i int
		for k, _ := range data {
			if i != 0 {
				insertQuery += ", "
			}
			insertQuery += "`" + k + "` = :" + k
			i++
		}
		sqlQuery += insertQuery
	}

	res, err := m.db.NamedExec(sqlQuery, data)
	if err != nil {
		log.Println("MysqlImages.AddImg.HasError", err)
	}

	return res, err
}
