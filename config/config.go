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
}

func (conf *Configuration)GetItems() Configuration {
	var filename = "./prod.json"
	var config Configuration

	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func (conf *Configuration)Get(key string) string {
	items := conf.GetItems()

	return items.ApiHost
}