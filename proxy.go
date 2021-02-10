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
	Agent string

	IsUsing bool

	Ip string
	LocalIp string
	Log []string
}

func NewProxy() *Proxy {
	proxy := MYSQL.GetFreeProxy()
	if proxy.Id.Valid {
		instance := &Proxy{}
		instance.Id = int(proxy.Id.Int64)
		instance.Host = proxy.Host.String
		instance.Port = proxy.Port.String
		instance.Login = proxy.Login.String
		instance.Password = proxy.Password.String
		instance.Agent = proxy.Agent.String
		instance.LocalIp = instance.Host + ":" + instance.Port

		return instance
	}
	return nil
}

func (p *Proxy) setTimeout(parser int, minutes int) sql.Result {
	now := time.Now().Local().Add(time.Minute * time.Duration(minutes))
	formattedDate := now.Format("2006-01-02 15:04:05")

	data := map[string]interface{}{}
	data["parser"] = strconv.Itoa(parser)
	data["timeout"] = formattedDate

	res, err := MYSQL.UpdateProxy(data, p.Id)
	if err != nil {
		log.Println("Proxy.SetTimeout.HasError", err)
	}

	return res
}

func (p *Proxy) freeProxy() {
	now := time.Now().Local().Add(time.Minute * time.Duration(2))
	formattedDate := now.Format("2006-01-02 15:04:05")

	data := map[string]interface{}{}
	data["parser"] = "NULL"
	data["timeout"] = formattedDate

	_, err := MYSQL.UpdateProxy(data, p.Id)
	if err != nil {
		log.Println("Proxy.freeProxy.HasError", err)
	}
	p.Id = 0
	p.Host = ""
}