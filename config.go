package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Port      string
	ProxyApi  string
	ProxyKey  string
	ApiHost   string
	ApiLink   string
	ApiKey    string
	ApiSecret string
	MongoUrl  string
	MongoDb   string
	MysqlHost  string
	MysqlDb   string
	MysqlLogin   string
	MysqlPass   string
}

func (conf *Configuration) Create() {
	var filename = "./config.json"

	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		panic(err)
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)
}