package main

import (
	"encoding/json"
	"fmt"
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

func (m *MysqlDb) SetExtra(extra ConfigExtra) error {
	extraJson, err := json.Marshal(extra)
	if err != nil {
		log.Println("MysqlConfig.SetExtra.HasError", err)
		return err
	}
	sqlQuery := "UPDATE `config` SET `extra` = :extra"
	data := map[string]interface{}{
		"extra": extraJson,
	}

	res, err := m.db.NamedExec(sqlQuery, data)
	fmt.Println(res)

	if err != nil {
		log.Println("MysqlConfig.SetExtra.HasError.2", err)
		return err
	}

	return nil
}

func (c *MysqlConfig) GetVariants() []string {
	var result []string
	if c.Variants.Valid {
		result = strings.Split(c.Variants.String, ";")
	}
	return result
}

func (c *MysqlConfig) GetExtra() ConfigExtra {
	Extra := ConfigExtra{}

	var extra map[string]interface{}
	_ = json.Unmarshal([]byte(c.Extra.String), &extra)
	if v, ok := extra["deep_paa"] ; ok {
		Extra.DeepPaa = v.(bool)
	}
	if v, ok := extra["redirect_method"] ; ok {
		Extra.RedirectMethod = v.(bool)
	}
	if v, ok := extra["count_streams"] ; ok {
		Extra.CountStreams = int(v.(float64))
	}
	if v, ok := extra["limit_streams"] ; ok {
		Extra.LimitStreams = int(v.(float64))
	}
	if v, ok := extra["cmd_streams"] ; ok {
		Extra.CmdStreams = v.(string)
	}

	return Extra
}