package main

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type Stream struct {
	state       bool
	job         JobHandler
	Proxy       Proxy
	browser     Browser
	cmd         string
	ctxTimer    context.Context
	cancelTimer context.CancelFunc
}

func (s *Stream) StartTaskTimer(streamId int, limit int64) {
	cmd := s.cmd

	if limit < 1 {
		limit = 360
	}

	s.ctxTimer, s.cancelTimer = context.WithTimeout(context.Background(), time.Second*time.Duration(limit))
	defer s.cancelTimer()

	if cmd != "" {
		//php -f /var/www/html/cron.php parser cron sleeping 5
		go exec.CommandContext(s.ctxTimer, "bash", "-c", cmd).Output()
	} else if s.browser.isOpened {
		fmt.Println("Start job", limit)
		s.browser.limit = limit
		s.browser.streamId = streamId
		// for {
		// 	if s.browser.ctx == nil {
		// 		s.browser.Reload()
		// 		time.Sleep(time.Minute * 5)
		// 	} else {
		// 		break
		// 	}
		// }
		s.job.Browser = s.browser
		s.job.isFinished = make(chan bool)
		s.job.IsStart = true
		go s.job.Run(streamId)
	}

	select {
	case <-s.ctxTimer.Done():
		fmt.Println("Timeout job")
		if s.cancelTimer != nil {
			s.cancelTimer()
		}
		s.job.Cancel()

	case <-s.job.isFinished:
		fmt.Println("End job")
		if s.cancelTimer != nil {
			s.cancelTimer()
		}
	}
}

func (s *Streams) StartStreams(count int, limit int, cmd string) {
	fmt.Println("Started")
	for i := 1; i <= count; i++ {
		stream := STREAMS.Add(i)
		if cmd != "" {
			stream.cmd = cmd + " " + strconv.Itoa(i)
		} else {
			stream.job = JobHandler{}
		}
		fmt.Println(i)
		go stream.Start(i, int64(limit))
	}
}

func (s *Streams) ReStartStreams(count int, limit int, cmd string) {
	fmt.Println("Restarted streams")
	STREAMS.StopAllWithoutClean()
	time.Sleep(time.Second * 40)
	s.StartStreams(count, limit, cmd)
}

func (s *Streams) StartLoop(count int, limit int, cmd string) {
	s.StopAll()

	var restartFunc func()

	restartFunc = func() {
		if STREAMS.isStarted {
			s.ReStartStreams(count, limit, cmd)
			time.AfterFunc(time.Second*2000, restartFunc)
		}
	}
	time.AfterFunc(time.Second*2000, restartFunc)

	STREAMS.isStarted = true
	go s.StartStreams(count, limit, cmd)
}

func (s *Stream) Start(streamId int, limit int64) {
	s.state = true

	time.Sleep(time.Millisecond * time.Duration(streamId*500))

	if s.cmd == "" {
		for {
			s.browser.streamId = streamId
			if !s.browser.Init() {
				s.browser.Cancel()
				time.Sleep(time.Minute)
			} else {
				break
			}
		}
	}

	for {
		if !s.state {
			s.browser.Cancel()
			break
		}

		secs := time.Second * time.Duration(int64(rand.Intn(15)))

		fmt.Println("Start stream #", streamId, s.cmd)
		s.StartTaskTimer(streamId, limit)
		fmt.Println("End stream #", streamId, secs)

		time.Sleep(secs)
	}
}

func (s *Stream) Stop() {
	s.state = false
	if s.cancelTimer != nil {
		s.cancelTimer()
	}
	s.browser.Cancel()
	s.job.IsStart = false
	go s.job.Cancel()
}

type Streams struct {
	isStarted bool
	items     map[int]*Stream
	mu        sync.RWMutex
}

func (s *Streams) Get(id int) *Stream {
	if stream, ok := s.items[id]; ok {
		return stream
	} else {
		return nil
	}
}

func (s *Streams) Add(id int) *Stream {
	s.mu.Lock()
	if len(s.items) < 1 {
		s.items = map[int]*Stream{}
	}
	s.items[id] = &Stream{}
	s.mu.Unlock()
	return s.items[id]
}

func (s *Streams) Stop(id int) {
	stream := s.Get(id)
	if stream != nil {
		stream.Stop()
	}
	delete(s.items, id)
}

func (s *Streams) StopAll() {
	fmt.Println("Stopped")
	if len(s.items) > 0 {
		for _, stream := range s.items {
			stream.Stop()
		}
		s.items = map[int]*Stream{}
	}
	s.StopAllInstances()
}

func (s *Streams) StopAllWithoutClean() {
	fmt.Println("Stopped without clean")
	if len(s.items) > 0 {
		for _, stream := range s.items {
			stream.Stop()
		}
	}
	//kill -9 $(pgrep -f chromium)
	s.StopAllInstances()
}

func (s *Streams) StopAllInstances() {
	_, err := exec.CommandContext(context.TODO(), "bash", "-c", "kill -9 $(pgrep -f chromium)").Output()
	if err != nil {
		fmt.Println("StopAllWithoutClean", err)
	}
}

func (s *Streams) Count() int {
	return len(s.items)
}
