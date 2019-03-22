package main

import (
	"encoding/json"
	"log"
	"time"
)

type TaskStatus struct {
	task map[string]taskState
}

type taskState struct {
	addedTime  time.Time
	processing bool
	startTime  time.Time
	finished   bool
	endTime    time.Time
	log        string
	// 代表提交该任务时玩家的成就完成个数
	progress int
	// 总成就个数
	total       int
	achievedIDs []string
}

func (s *TaskStatus) newTaskState(taskId string, progress int, total int) {
	s.task[taskId] = taskState{
		addedTime:   time.Now(),
		processing:  false,
		startTime:   time.Now(),
		finished:    false,
		endTime:     time.Now(),
		log:         "任务正在排队等候处理",
		progress:    progress,
		total:       total,
		achievedIDs: make([]string, progress),
	}
}

func (s *TaskStatus) setAchievedIDs(taskId string, arr []string) {
	tS := s.task[taskId]
	tS.achievedIDs = arr
	s.task[taskId] = tS
}

func (s *TaskStatus) updateTaskState(taskId string, info string) {
	thisTask := s.task[taskId]

	switch info {
	case "start":
		thisTask.startTime = time.Now()
		thisTask.processing = true
		thisTask.log = "任务处理中"
		break
	case "end":
		thisTask.endTime = time.Now()
		thisTask.finished = true
		thisTask.processing = false
		break
	default:
		thisTask.log = info
		thisTask.progress++
	}
	s.task[taskId] = thisTask
}

func (s *TaskStatus) getStateJSON(taskId string) string {
	var sF interface{}
	thisTask := s.task[taskId]
	if _, ok := s.task[taskId]; !ok {
		type StateNotFond struct {
			Retcode int
			Log     string
		}
		sF = StateNotFond{
			Retcode: -1,
			Log:     "任务不存在",
		}
		goto Response
	}

	sF = StateRespJSON{
		Retcode:     0,
		AddedTime:   thisTask.addedTime.Unix(),
		Processing:  thisTask.processing,
		StartedTime: thisTask.startTime.Unix(),
		Finished:    thisTask.finished,
		EndedTime:   thisTask.endTime.Unix(),
		Progress:    thisTask.progress,
		Total:       thisTask.total,
		Log:         thisTask.log,
	}

Response:
	serialized, err := json.Marshal(sF)
	if err != nil {
		log.Println("Serialize state error:", err)
	}
	return string(serialized)

}
