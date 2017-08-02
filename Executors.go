package main

//没有队列的执行器
func NewNoQueuePoolExecutor(maxGoroutines int32) (*GoroutinesPoolExecutor, error) {
	return newGoroutinesPoolExecutor(maxGoroutines, 0, -1)
}
//有固定大小队列的执行器
func NewBoundQueuePoolExecutor(maxGoroutines int32,queueSize int32) (*GoroutinesPoolExecutor, error) {
	return newGoroutinesPoolExecutor(maxGoroutines, queueSize, -1)
}
//队列无限大的执行器
func NewUnBoundQueuePoolExecutor(maxGoroutines int32) (*GoroutinesPoolExecutor, error) {
	return newGoroutinesPoolExecutor(maxGoroutines, 0, -1)
}
//单个协程, 队列无限大的执行器
func NewSingleGoroutinesPoolExecutor()(*GoroutinesPoolExecutor, error) {
	return newGoroutinesPoolExecutor(1,maxInteger,-1)
}