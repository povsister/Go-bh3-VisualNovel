package main

type TaskQueue struct {
	elements []Task
	taskChan *chan Task
}

func (entry *TaskQueue) isEmpty() bool {
	if len(entry.elements) == 0 {
		return true
	}
	return false
}

func (entry *TaskQueue) put(e Task) {
	if len(*entry.taskChan) == 0 {
		*entry.taskChan <- e
		return
	}
	entry.elements = append(entry.elements, e)
}

func (entry *TaskQueue) pop() Task {
	if entry.isEmpty() {
		return nil
	}
	first := entry.elements[0]
	entry.elements = entry.elements[1:]
	return first
}

func (entry *TaskQueue) size() int {
	return len(entry.elements)
}

func (entry *TaskQueue) clear() {
	entry.elements = entry.elements[0:0]
}
