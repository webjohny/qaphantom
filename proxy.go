package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"
)

type Proxy struct {
	Id int
	Host string
	Port string
	Login string
	Password string

	IsUsing bool

	Ip string
	LocalIp string
	Log []string
}

func (p *Proxy) NewProxy() {
	proxy := mysql.GetFreeProxy()
	p.Id = int(proxy.Id.Int64)
	p.Host = proxy.Host.String
	p.Port = proxy.Port.String
	p.Login = proxy.Login.String
	p.Password = proxy.Password.String
	p.LocalIp = p.Host + ":" + p.Port
}

func (p Proxy) SetTimeout(parser int) sql.Result {
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

	sqlQuery += " WHERE `id` = " + strconv.Itoa(p.Id)

	res, err := mysql.db.NamedExec(sqlQuery, data)
	if err != nil {
		log.Println(err)
	}

	return res
}

func (p Proxy) FreeProxy() {
	data := map[string]interface{}{}
	data["parser"] = "NULL"
	data["timeout"] = "NULL"

	_, err := mysql.UpdateProxy(data, p.Id)
	if err != nil {
		log.Println(err)
	}
}
