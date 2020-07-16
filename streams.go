package main

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

type Stream struct {
	state bool
	job JobHandler
	Proxy Proxy
	browser Browser
	cmd string
	ctxTimer context.Context
	CancelTimer context.CancelFunc
}

func (s *Stream) StartTaskTimer(streamId int, limit int64) {
	cmd := s.cmd

	if limit < 1 {
		limit = 360
	}

	s.ctxTimer, s.CancelTimer = context.WithTimeout(context.Background(), time.Second * time.Duration(limit))
	defer s.CancelTimer()

	if cmd != "" {
		//php -f /var/www/html/cron.php parser cron sleeping 5
		go exec.CommandContext(s.ctxTimer, "bash", "-c", cmd).Output()
	} else if s.browser.isOpened {
		fmt.Println("Start job", limit)
		s.browser.limit = limit
		s.job.Browser = s.browser
		s.job.isFinished = make(chan bool)
		s.job.IsStart = true
		go s.job.Run(streamId)
	}

	select {
	case <-s.ctxTimer.Done():
		fmt.Println("Timeout job")
		s.CancelTimer()
		s.job.Cancel()

	case <-s.job.isFinished:
		fmt.Println("End job")
		s.CancelTimer()
	}
}

func (s *Stream) Start(streamId int, limit int64) {
	s.state = true

	time.Sleep(time.Millisecond * time.Duration(streamId * 500))

	if s.cmd == "" {
		s.browser.Init()
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
	if s.CancelTimer != nil {
		s.CancelTimer()
	}
	s.browser.Cancel()
	s.job.IsStart = false
	go s.job.Cancel()
}


type Streams struct {
	isStarted bool
	items map[int]*Stream
}

func (s *Streams) Get(id int) *Stream{
	if stream, ok := s.items[id]; ok {
		return stream
	} else {
		return nil
	}
}

func (s *Streams) Add(id int) *Stream{
	if len(s.items) < 1 {
		s.items = map[int]*Stream{}
	}
	s.items[id] = &Stream{}
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
}

func (s *Streams) StopAllWithoutClean() {
	fmt.Println("Stopped without clean")
	if len(s.items) > 0 {
		for _, stream := range s.items {
			stream.Stop()
		}
	}
}

func (s *Streams) Count() int {
	return len(s.items)
}