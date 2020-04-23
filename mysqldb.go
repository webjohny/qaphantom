package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MysqlDb struct {
	db *sqlx.DB
	conf Configuration
}

func (m *MysqlDb) CreateConnection() {
	conn, err := sqlx.Connect("mysql", m.conf.MysqlLogin + ":" + m.conf.MysqlPass + "@tcp(" + m.conf.MysqlHost + ")/" + m.conf.MysqlDb)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to MysqlDB!")

	m.db = conn
}

func (m *MysqlDb) Disconnect() {
	err := m.db.Close()

	if err != nil {
		panic(err)
	}
	fmt.Println("Connection to MySQL closed.")
}
