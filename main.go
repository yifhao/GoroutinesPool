package main

import "time"

func main() {
	executor, _ := NewNoQueuePoolExecutor(10)
	for i:=0;i<100;i++{
		test := func() {
			time.Sleep(time.Second * 1)
			println(time.Now().String())
		}
		executor.execute(test)
	}
	executor.shutdown()
	result := executor.awaitTermination(20)
	println(result)
}