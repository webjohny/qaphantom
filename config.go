package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	Port      int
	ApiHost   string
	ApiLink   string
	ApiKey    string
	ApiSecret string
	MongoUrl  string
	MongoDb   string
}

func (conf *Configuration) Create() {
	var filename = "./config.json"

	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)
}