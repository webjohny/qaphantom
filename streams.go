package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"time"
)

type Stream struct {
	state bool
	job JobHandler
	cmd string
	ctxTimer context.Context
	cancelTimer context.CancelFunc
}

func (s *Stream) StartTaskTimer(streamId int, limit int64) bool {
	var status bool = true
	cmd := s.cmd

	s.ctxTimer, s.cancelTimer = context.WithTimeout(context.Background(), time.Duration(1000 * limit) * time.Millisecond)
	defer s.cancelTimer()

	var err error
	if cmd != "" {
		//php -f /var/www/html/cron.php parser cron sleeping 5
		_, err = exec.CommandContext(s.ctxTimer, "bash", "-c", cmd).Output()
	} else {
		fmt.Println("Start job")
		s.job.IsStart = true
		s.job.Run(streamId)
	}

	if err != nil {
		status = false
		log.Println(err)
	}

	return status
}

func (s *Stream) Start(streamId int, limit int64) {
	s.state = true

	time.Sleep(time.Millisecond * time.Duration(streamId * 500))

	for {
		if ! s.state {
			break
		}

		randSecs := time.Second * time.Duration(int64(rand.Intn(20)))

		fmt.Println("Start stream #", streamId, s.cmd)
		s.StartTaskTimer(streamId, limit)
		fmt.Println("End stream #", streamId, randSecs)

		time.Sleep(randSecs)
	}
}

func (s *Stream) Stop() {
	s.state = false
	if s.cancelTimer != nil {
		s.cancelTimer()
	}
	if s.job.IsStart {
		s.job.Cancel()
		s.job.IsStart = false
	}
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