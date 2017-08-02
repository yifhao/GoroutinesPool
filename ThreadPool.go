package main

import (
	"sync"
	"time"
)

type ExecutorService interface {
	run(runnable func())
	shutdown(waitSecond int32) bool
}

type GoroutinesPoolExecutor struct {
	work chan func()
	wg   sync.WaitGroup
}


func (executor *GoroutinesPoolExecutor) run(runnable func()) {
	executor.work <- runnable
}

func (executor *GoroutinesPoolExecutor) shutdown(waitSecond int32) bool {
	close(executor.work)
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

func NewFixedExecutorService(maxGoroutines int) *GoroutinesPoolExecutor {
	p := GoroutinesPoolExecutor{
		work: make(chan func()),
		wg:   sync.WaitGroup{},
	}
	p.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range p.work {
				w()
			}
			p.wg.Done()
		}()
	}
	return &p
}

func main() {
	executor:=NewFixedExecutorService(1)
	test:=func() {
		time.Sleep(time.Second*10)
	}
	executor.run(test)
	result:=executor.shutdown(4)
	println(result)
}
