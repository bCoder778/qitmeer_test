package pool

import (
	"errors"
	"sync"
)

type Task struct {
	id     uint64
	params map[string]interface{}
	f      func(map[string]interface{}) (interface{}, error)
}

func NewTask(id uint64, params map[string]interface{}, f func(map[string]interface{}) (interface{}, error)) *Task {
	return &Task{id, params, f}
}

func (t *Task) Run(flagCh chan bool) {
	flagCh <- true
	go func() {
		t.f(t.params)
		<-flagCh
	}()
}

type Pool struct {
	mutex             sync.RWMutex
	maxGoroutineCount uint32
	worksCh           chan *Task
	readyCh           chan *Task
	flagCh            chan bool
	wg                sync.WaitGroup
}

func NewPool(maxCount uint32) *Pool {
	return &Pool{
		maxGoroutineCount: maxCount,
		worksCh:           make(chan *Task, maxCount),
		readyCh:           make(chan *Task, maxCount),
		flagCh:            make(chan bool, maxCount),
	}
}

func (p *Pool) Run() {
	go p.worksRun()
	go p.readyRun()
}

func (p *Pool) worksRun() {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.worksCh:
			if !ok {
				return
			}
			task.Run(p.flagCh)
		}
	}
}

func (p *Pool) readyRun() {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.readyCh:
			if !ok {
				return
			}
			p.worksCh <- task
		}
	}
}

func (p *Pool) AddTask(task *Task) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if uint32(len(p.readyCh)) == p.maxGoroutineCount {
		return errors.New("the pool is full, please wait")
	}
	p.readyCh <- task
	return nil
}

func (p *Pool) Close() {
	close(p.worksCh)
	close(p.readyCh)
	close(p.flagCh)
	p.wg.Wait()
}
