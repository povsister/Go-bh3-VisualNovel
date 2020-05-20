package main

import (
	"bh3-visualNovel"
	"flag"
	"io"
	"log"
	"net/http"
)

var (
	addr      = flag.String("addr", "127.0.0.1:1551", "http server listen address:port")
	workerNum = flag.Int("worker", 1, "set the worker number")
	wPool     *WorkerPool
)

func main() {
	// 解析配置参数
	flag.Parse()
	// worker pool
	wPool = buildWorkerPool()
	// 启动http服务器
	startHTTP()
}

func buildWorkerPool() *WorkerPool {
	// 见 workerPool
	return initializeWorkerPool(*workerNum)
}

func ioWriteString(w http.ResponseWriter, s string) {
	n, err := io.WriteString(w, s)
	if err != nil {
		log.Println("Response Writer error:", err, ". Wrote:", n)
	}
}

func startHTTP() {
	// 默认路由
	http.HandleFunc("/", indexPage)
	// 注册路由
	// status
	http.HandleFunc("/vn/task/state/", vn_taskState)
	// VN ROUTE
	http.HandleFunc("/vn/gf/antiEntropy/", gf_antiEntropy)
	http.HandleFunc("/vn/gf/Durandal/", gf_Durandal)
	http.HandleFunc("/vn/gf/7-Sword/", gf_SevenSword)
	// start
	log.Printf("Starting Web server on %s", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func indexPage(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	ioWriteString(w, "404 not found")
}

func gf_antiEntropy(w http.ResponseWriter, req *http.Request) {
	t := NewAntiEntropyGF(vn.GetTaskIdFromPath(req), req)
	// 验证任务有效性
	msg, achievedIDs, progress, total, ok := t.valid(wPool.libAchievement)
	if ok {
		// 创建一个任务状态
		wPool.taskStatus.newTaskState(t.getTaskID(), progress, total)
		// 将已经完成的成就列表保存起来
		wPool.taskStatus.setAchievedIDs(t.getTaskID(), achievedIDs)
		// 加入到处理队列中
		wPool.taskQueue.put(t)
	}
	ioWriteString(w, msg)
}

func gf_Durandal(w http.ResponseWriter, req *http.Request) {
	t := NewDurandalGF(vn.GetTaskIdFromPath(req), req)

	msg, achievedIDs, progress, total, ok := t.valid(wPool.libAchievement)
	if ok {
		wPool.taskStatus.newTaskState(t.getTaskID(), progress, total)
		wPool.taskStatus.setAchievedIDs(t.getTaskID(), achievedIDs)
		wPool.taskQueue.put(t)
	}
	ioWriteString(w, msg)
}

func gf_SevenSword(w http.ResponseWriter, req *http.Request) {
	t := NewSevenSwordGF(vn.GetTaskIdFromPath(req), req)
	msg, achievedIDs, progress, total, ok := t.valid(wPool.libAchievement)
	if ok {
		wPool.taskStatus.newTaskState(t.getTaskID(), progress, total)
		wPool.taskStatus.setAchievedIDs(t.getTaskID(), achievedIDs)
		wPool.taskQueue.put(t)
	}
	ioWriteString(w, msg)
}

func vn_taskState(w http.ResponseWriter, req *http.Request) {
	t := vn.GetTaskIdFromPath(req)
	s := wPool.taskStatus.getStateJSON(t)
	ioWriteString(w, s)
}
