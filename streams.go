package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type Stream struct {
	state bool
	cmd string
	ctxTimer context.Context
	cancelTimer context.CancelFunc
}

func (s *Stream) StartTaskTimer(cmd string, limit int64) bool {
	var status bool = true

	s.ctxTimer, s.cancelTimer = context.WithTimeout(context.Background(), time.Duration(1000 * limit) * time.Millisecond)
	defer s.cancelTimer()

	//php -f /var/www/html/cron.php parser cron sleeping 5
	_, err := exec.CommandContext(s.ctxTimer, "bash", "-c", cmd).Output()

	if err != nil {
		status = false
		fmt.Println(err)
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

		fmt.Println("Start stream #", streamId, s.cmd)
		s.StartTaskTimer(s.cmd, limit)
		fmt.Println("End stream #", streamId, time.Second*15)

		time.Sleep(time.Second * 15)
	}
}

func (s *Stream) Stop() {
	s.state = false
	if s.cancelTimer != nil {
		s.cancelTimer()
	}
}


type Streams struct {
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
	if len(s.items) > 0 {
		for _, stream := range s.items {
			stream.Stop()
		}
		s.items = map[int]*Stream{}
	}
}

func (s *Streams) Count() int {
	return len(s.items)
}