package main

import "bh3-visualNovel"

type Task interface {
	process(worker *Worker)
	getTaskID() string
	valid(libAchieve *vn.LIBAchievement) (string, []string, int, bool)
}
