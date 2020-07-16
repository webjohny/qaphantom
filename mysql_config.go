package main

import (
	"log"
	"strings"
)

func (m *MysqlDb) GetConfig() MysqlConfig {
	var result MysqlConfig
	sqlQuery := "SELECT * FROM `config` LIMIT 1"

	err := m.db.Get(&result, sqlQuery)
	if err != nil {
		log.Println("MysqlConfig.GetConfig.HasError", err)
	}

	return result
}


func (c *MysqlConfig) GetVariants() []string {
	var result []string
	if c.Variants.Valid {
		result = strings.Split(c.Variants.String, ";")
	}
	return result
}