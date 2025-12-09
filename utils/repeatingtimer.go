package utils

import (
	"time"
)

type CallbackFunc func() error

type RepeatingTimer struct {
	Ticker *time.Ticker
	Runner CallbackFunc
	stop   chan struct{}
}

func NewRepeatingTimer(interval int, f CallbackFunc) *RepeatingTimer {
	return &RepeatingTimer{
		Ticker: time.NewTicker(time.Duration(interval) * time.Second),
		Runner: f,
		stop:   make(chan struct{}),
	}
}

func (t *RepeatingTimer) Start() {
	go func() {
		for {
			select {
			case <-t.Ticker.C:
				t.Runner()
			case <-t.stop:
				close(t.stop)
				t.Ticker.Stop()
				return
			}
		}
	}()
}

func (t *RepeatingTimer) Stop() {
	t.stop <- struct{}{}
}
