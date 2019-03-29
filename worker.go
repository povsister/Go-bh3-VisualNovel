package main

import (
	"log"
	"strconv"
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
				// 获取任务开始时间
				tS := time.Now().Unix()
				log.Printf("Delayed worker %d is handling task\n", w.id)
				w.handleTask(*task)
				// 获取任务结束时间
				tE := time.Now().Unix()
				// 判断是否需要额外等待
				tI := tE - tS
				secWait := int64(180)
				if tI < secWait {
					tStr := strconv.FormatInt(secWait-tI, 10)
					tDur, _ := time.ParseDuration(tStr + "s")
					time.Sleep(tDur)
				}
			case quitSignal := <-w.quit:
				if quitSignal {
					log.Printf("Delayed worker %d stopped\n", w.id)
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
			// 加入到延时队列中
			w.pool.taskStatus.updateTaskState(t.getTaskID(), "failedFrequent")
			w.pool.delayedTaskQueue.put(t)
		} else {
			w.pool.taskStatus.updateTaskState(t.getTaskID(), "failed")
		}
	}

}
