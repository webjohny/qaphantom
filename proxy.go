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

func (p *Proxy) NewProxy() {
	proxy := mysql.GetFreeProxy()
	p.Id = int(proxy.Id.Int64)
	p.Host = proxy.Host.String
	p.Port = proxy.Port.String
	p.Login = proxy.Login.String
	p.Password = proxy.Password.String
	p.Agent = proxy.Agent.String
	p.LocalIp = p.Host + ":" + p.Port
}

func (p Proxy) SetTimeout(parser int, minutes int) sql.Result {
	now := time.Now().Local().Add(time.Minute * time.Duration(minutes))
	formattedDate := now.Format("2006-01-02 15:04:05")

	data := map[string]interface{}{}
	data["parser"] = strconv.Itoa(parser)
	data["timeout"] = formattedDate

	res, err := mysql.UpdateProxy(data, p.Id)
	if err != nil {
		log.Println("Proxy.SetTimeout.HasError", err)
	}

	return res
}

func (p Proxy) FreeProxy() {
	now := time.Now().Local().Add(time.Minute * time.Duration(2))
	formattedDate := now.Format("2006-01-02 15:04:05")

	data := map[string]interface{}{}
	data["parser"] = "NULL"
	data["timeout"] = formattedDate

	_, err := mysql.UpdateProxy(data, p.Id)
	if err != nil {
		log.Println("Proxy.FreeProxy.HasError", err)
	}
}