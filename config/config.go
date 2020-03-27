package config

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

func (conf *Configuration)Get(key string) string {
	return conf.ApiHost
}

func Create() *Configuration {
	var filename = "./config/prod.json"
	var config *Configuration

	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}