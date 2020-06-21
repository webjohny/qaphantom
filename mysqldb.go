package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"os/exec"
	"time"
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
	conn.SetMaxIdleConns(20)
	conn.SetConnMaxLifetime(time.Minute * 2)
	conn.SetMaxOpenConns(100)
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

func (m *MysqlDb) Restart() {
	cmd := exec.Command("service", "mysql restart")
	log.Printf("Mysql restarting and waiting for it to finish...")
	err := cmd.Run()
	log.Printf("Command finished with error: %v", err)
	time.Sleep(time.Second * 5)
}