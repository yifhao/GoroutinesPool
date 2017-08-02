package main

import (
	"sync"
	"time"
	"errors"
	"sync/atomic"
)

const (
	maxInteger = 0x7fffffff
)

type Worker struct {
	completedTasks int32
	firstRunnable  func()
}

func (worker *Worker) Run(executor *GoroutinesPoolExecutor) {
	go func() {
		worker.firstRunnable()
		atomic.AddInt32(&(worker.completedTasks), 1)
		if executor.aliveSecond > 0 {
			WORK:for {
				timer := time.NewTimer(time.Second * time.Duration(executor.aliveSecond))
				select {
				case <-timer.C:
					break WORK
				case job := <-executor.work:
					{
						job()
						atomic.AddInt32(&(worker.completedTasks), 1)
					}
				}
			}
		} else{
			for job := range executor.work {
				job()
				atomic.AddInt32(&(worker.completedTasks), 1)
			}
		}
		delete(executor.workers,worker)
		executor.wg.Done()
	}()
}

type ExecutorService interface {
	execute(runnable func())
	shutdown(waitSecond int32) bool
}

type GoroutinesPoolExecutor struct {
	mutex         sync.Mutex
	maxGoroutines int
	workers       map[*Worker]bool
	work          chan func()
	wg            sync.WaitGroup
	state         int32
	aliveSecond   int32
}

func (executor *GoroutinesPoolExecutor) execute(runnable func()) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	if len(executor.workers) < executor.maxGoroutines {
		worker := &Worker{
			completedTasks: 0,
			firstRunnable:  runnable,
		}
		executor.workers[worker] = true
		worker.Run(executor)
	} else {
		executor.work <- runnable
	}
}

func (executor *GoroutinesPoolExecutor) awaitTermination(waitSecond int32) bool {
	timer := time.NewTimer(time.Duration(waitSecond) * time.Second)
	end := make(chan bool)
	go func() {
		executor.wg.Wait()
		end <- true
	}()
	select {
	case <-end:
		return true
	case <-timer.C:
		return false
	}
}

func (executor *GoroutinesPoolExecutor) shutdown() {
	atomic.StoreInt32(&executor.state, -1)
	close(executor.work)
}


func newGoroutinesPoolExecutor(goroutineNums int32, channelSize int32, goroutineALiveSecond int32) (*GoroutinesPoolExecutor, error) {
	if goroutineNums <= 0 {
		return nil, errors.New("go routineNums should larger than 0")
	}
	if channelSize < 0 {
		return nil, errors.New("channelSize should larger than or equals to 0")
	}
	var workChannel chan func()
	if channelSize == 0 {
		workChannel = make(chan func())
	} else {
		workChannel = make(chan func(), goroutineNums)
	}
	wg:=sync.WaitGroup{}
	wg.Add(int(goroutineNums))
	return &GoroutinesPoolExecutor{
		state:         0,
		workers:       make(map[*Worker]bool),
		mutex:         sync.Mutex{},
		maxGoroutines: int(goroutineNums),
		work:          workChannel,
		wg:            wg,
		aliveSecond:   goroutineALiveSecond,
	}, nil
}
