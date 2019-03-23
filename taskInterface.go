package main

import "bh3-visualNovel"

type Task interface {
	process(worker *Worker) (bool, bool)
	getTaskID() string
	valid(libAchieve *vn.LIBAchievement) (string, []string, int, int, bool)
}
