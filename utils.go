package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)
type Utils struct {}

func (u *Utils) ParseFormCollection(r *http.Request, typeName string) map[string]string {
	result := make(map[string]string)
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
	}
	for key, values := range r.Form {
		re := regexp.MustCompile(typeName + "\\[(.+)\\]")
		matches := re.FindStringSubmatch(key)

		if len(matches) >= 2 {
			result[matches[1]] = values[0]
		}
	}
	return result
}

func (u *Utils) toInt(value string) int {
	var integer int = 0
	if value != "" {
		integer, _ = strconv.Atoi(value)
	}
	return integer
}