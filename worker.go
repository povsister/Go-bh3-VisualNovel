package main

import (
	"log"
	"time"
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

func (w *Worker) startDelayed() {
	go func() {
		for {
			w.pool.delayedWorkerChan <- w
			log.Printf("Delayed worker %d is waiting for task\n", w.id)
			select {
			case task := <-w.task:
				log.Printf("Delayed Worker %d is handling task\n", w.id)
				w.handleTask(*task)
				// 等待 180 秒, 再进行下次尝试
				time.Sleep(180 * time.Second)
			case quitSignal := <-w.quit:
				if quitSignal {
					log.Printf("Delayed Worker %d stopped\n", w.id)
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
			// 加入到延时队列中  该队列中的worker每180秒工作一次
			w.pool.taskStatus.updateTaskState(t.getTaskID(), "failedFrequent")
			w.pool.delayedTaskQueue.put(t)
		} else {
			w.pool.taskStatus.updateTaskState(t.getTaskID(), "failed")
		}
	}

}
