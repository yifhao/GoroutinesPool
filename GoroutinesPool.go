package main

import (
	"sync"
	"time"
	"errors"
	"sync/atomic"
)

//32位int的最大值
const (
	maxInteger = 0x7fffffff
)

//worker数据结构, worker代表一个活动的协程
type Worker struct {
	//完成的任务数
	completedTasks int32
	//这个worker在生成时赋予的任务
	firstRunnable  func()
}

//worker的启动方法
func (worker *Worker) Run(executor *GoroutinesPoolExecutor) {
	go func() {
		worker.firstRunnable()
		atomic.AddInt32(&(worker.completedTasks), 1)
		//如果worker具有固定的空闲时间, 那么worker在规定时间获取不到任务, 就退出
		if executor.GoroutinesAliveSecond > 0 {
			WORK:for {
				timer := time.NewTimer(time.Second * time.Duration(executor.GoroutinesAliveSecond))
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
			//worker可以一直阻塞, 直到获取到任务
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


type GoroutinesPoolExecutor struct {
	mutex                 sync.Mutex
	maxGoroutines         int
	workers               map[*Worker]bool
	work                  chan func()
	wg                    sync.WaitGroup
	state                 int32
	GoroutinesAliveSecond int32
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

//创建一个Goroutine池,goroutineNums指定池的最大协程数,channelSize指定任务队列的长度,goroutineALiveSecond指定协程在空闲多久时,会自动退出
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
		state:                 0,
		workers:               make(map[*Worker]bool),
		mutex:                 sync.Mutex{},
		maxGoroutines:         int(goroutineNums),
		work:                  workChannel,
		wg:                    wg,
		GoroutinesAliveSecond: goroutineALiveSecond,
	}, nil
}
