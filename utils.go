package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type Utils struct {}

func (u *Utils) SetInterval(someFunc func(), milliseconds int, async bool) chan bool {

	// How often to fire the passed in function
	// in milliseconds
	interval := time.Duration(milliseconds) * time.Millisecond

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(interval)
	clear := make(chan bool)

	// Put the selection in a go routine
	// so that the for loop is none blocking
	go func() {
		for {

			select {
			case <-ticker.C:
				if async {
					// This won't block
					go someFunc()
				} else {
					// This will block
					someFunc()
				}
			case <-clear:
				ticker.Stop()
				return
			}

		}
	}()

	// We return the channel so we can pass in
	// a value to it to clear the interval
	return clear

}

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