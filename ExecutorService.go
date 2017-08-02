package main
//执行池接口
type ExecutorService interface {
	execute(runnable func())
	shutdown()
	awaitTermination(waitSecond int32) bool
}