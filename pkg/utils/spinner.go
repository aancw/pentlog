package utils

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	mu       sync.Mutex
	stopChan chan struct{}
	active   bool
	frames   []string
	delay    time.Duration
	message  string
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		stopChan: make(chan struct{}),
		frames:   []string{"▘", "▝", "▗", "▖"}, // TL, TR, BR, BL
		delay:    200 * time.Millisecond,
		message:  message,
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.stopChan:
				return
			default:
				fmt.Printf("\r%s %s", s.frames[i%len(s.frames)], s.message)
				i++
				time.Sleep(s.delay)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	close(s.stopChan)
	fmt.Printf("\r\033[K") // Clear the line
}
