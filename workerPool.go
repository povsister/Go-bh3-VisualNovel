package main

import "bh3-visualNovel"

type WorkerPool struct {
	workerChan     chan *Worker
	workerList     []*Worker
	taskChan       chan Task
	taskQueue      *TaskQueue
	taskStatus     *TaskStatus
	libAchievement *vn.LIBAchievement
}

func initializeWorkerPool(n int) *WorkerPool {
	// 分配队列
	tQueue := &TaskQueue{}
	// 分配任务状态
	tS := TaskStatus{
		task: make(map[string]taskState),
	}
	// 分配成就数据
	lA := vn.LIBAchievement{
		Lib: make(map[int]vn.VnAchievements),
	}
	wPool = &WorkerPool{
		workerChan:     make(chan *Worker, n),
		workerList:     make([]*Worker, n),
		taskChan:       make(chan Task, n),
		taskQueue:      tQueue,
		taskStatus:     &tS,
		libAchievement: &lA,
	}
	tQueue.taskChan = &wPool.taskChan
	wPool.startWorker(n)
	wPool.startDispatch()
	return wPool
}

func (wP *WorkerPool) startWorker(n int) {
	for i := 0; i < n; i++ {
		w := &Worker{
			id:   i,
			pool: wP,
			task: make(chan *Task),
			quit: make(chan bool),
		}
		wP.workerList = append(wP.workerList, w)
		w.start()
	}
}

func (wP *WorkerPool) startDispatch() {
	go func() {
		for {
			//log.Println("Waiting for task to dispatch...")
			select {
			case task := <-wP.taskChan:
				worker := <-wP.workerChan
				// 更新任务状态
				wP.taskStatus.updateTaskState(task.getTaskID(), "start")
				// 分配给worker
				worker.task <- &task
				// 判断是否需要将队列中的任务继续加入缓冲区
				if !wP.taskQueue.isEmpty() && len(wP.taskChan) < cap(wP.taskChan) {
					wP.taskChan <- wP.taskQueue.pop()
				}
			}
		}
	}()
}

//func (wP *WorkerPool) cleanOldTaskState() {
//	taskMap := wP.taskStatus.task
//	for k, v := range taskMap {}
//}