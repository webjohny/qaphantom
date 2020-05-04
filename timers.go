package main

import (
	"fmt"
	"time"
)

var chk bool = false

func setInterval(someFunc func(), milliseconds int, async bool) chan bool {

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

func run() {
	fmt.Println("Start async timer")
	for {
		if chk {
			break
		}
		time.Sleep(time.Second * 3)
	}
}

func stop() {
	fmt.Println("Stop async timer")
	chk = true
}

func gorout() {
	stop()
	time.Sleep(time.Second * 3)
	go run()
}

func timers() {

	go run()
	setInterval(func() {
		gorout()
	}, 20000, false)

	fmt.Println("Async timer")

	time.Sleep(time.Second * 50)
}