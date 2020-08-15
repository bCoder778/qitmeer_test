package timer

import (
	"fmt"
	"github.com/bCoder778/log"
	"sync"
	"time"
)

type Timer struct {
	mutex   sync.RWMutex
	stop    chan bool
	stopped chan bool
}

func New() *Timer {
	return &Timer{stop: make(chan bool), stopped: make(chan bool)}
}

func (t *Timer) Start(f func(), timestamp int64, interval int64) {
	c := time.NewTicker(time.Millisecond * 1).C
	id := 1
	for {
		select {
		case _, ok := <-t.stop:
			if !ok {
				log.Infof("Stop timer")
				t.stopped <- true
				return
			}
		case <-c:
			if (time.Now().Unix()-timestamp)%interval == 0 {
				log.Mail(fmt.Sprintf("Start test func %d", id), time.Now().String())
				f()
				id++
				time.Sleep(time.Second)
			}
		}
	}
}

func (t *Timer) Stop() {
	close(t.stop)
	<-t.stopped
}
