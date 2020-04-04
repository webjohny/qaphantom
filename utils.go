package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"sync"
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


// Counters - work with mutex
type Counters struct {
	mx sync.Mutex
	m map[string]int
}

func (c *Counters) Load(key string) (int, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	val, ok := c.m[key]
	return val, ok
}

func (c *Counters) Store(key string, value int) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m[key] = value
}

func NewCounters() *Counters {
	return &Counters{
		m: make(map[string]int),
	}
}