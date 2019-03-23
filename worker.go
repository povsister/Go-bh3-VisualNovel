package main

import (
	"log"
)

type Worker struct {
	id   int
	pool *WorkerPool
	task chan *Task
	quit chan bool
}

func (w *Worker) start() {
	go func() {
		for {
			w.pool.workerChan <- w
			log.Printf("Worker %d is waiting for task\n", w.id)
			select {
			case task := <-w.task:
				log.Printf("Worker %d is handling task\n", w.id)
				w.handleTask(*task)
			case quitSignal := <-w.quit:
				if quitSignal {
					log.Printf("Worker %d stopped\n", w.id)
					return
				}

			}
		}
	}()
}

func (w *Worker) handleTask(t Task) {
	// 任务处理逻辑
	succeed, frequent := t.process(w)
	if succeed {
		// 更新任务状态
		w.pool.taskStatus.updateTaskState(t.getTaskID(), "end")
	} else {
		if frequent {
			// 重新加入队列中
			w.pool.taskStatus.updateTaskState(t.getTaskID(), "failedFrequent")
			w.pool.taskQueue.put(t)
		} else {
			w.pool.taskStatus.updateTaskState(t.getTaskID(), "failed")
		}
	}

}
