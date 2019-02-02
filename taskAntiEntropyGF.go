package main

import (
	"bh3-visualNovel"
	"net/http"
)

const (
	URL_ANTIENTROPY             = "https://event.bh3.com/avgAntiEntropy/indexAntiEntropy.php"
	URL_ANTIENTROPY_ACHIEVEMENT = "https://event.bh3.com/avgAntiEntropy/utils/achievement.php"
	URL_ANTIENTROPY_XML         = "https://event.bh3.com/avgAntiEntropy/lang_CN/xml/ch%s.xml?sid=%s"
	URL_ANTIENTROPY_LATEST      = "https://event.bh3.com/avgAntiEntropy/lang_CN/xml/date_url.xml?sid=%s"
	URL_ANTIENTROPY_EXHIBITION  = "https://event.bh3.com/avgAntiEntropy/lang_CN/xml/exhibition_list.xml?sid=%s"
	// cookie name
	COOKIE_AE_CHAPTER = "_2017_anti_entropy_now_galgame"
	COOKIE_AE_SCENE   = "_2017_anti_entropy_now_scene"
	COOKIE_AE_ACTION  = "_2017_anti_entropy_now_action"
)

type AntiEntropyGF struct {
	// 任务ID
	id          string
	vNo         int
	req         *http.Request
	queryString string
	httpClient  *vn.MyAJAX
	// URLS
	URL_BASE        string
	URL_CHAPTER_XML string
	URL_ACHIEVEMENT string
	URL_LATEST      string
	URL_EXHIBITION  string
	// helper
	xmlHelper     *vn.XMLHelper
	achieveHelper *vn.AchievementHelper
}

func (t AntiEntropyGF) process(worker *Worker) {
	vnA := worker.pool.libAchievement.GetNovelAchievements(t.vNo)

	// 玩家已完成成就
	achieved := worker.pool.taskStatus.task[t.getTaskID()].achievedIDs
	// 全部成就
	allAchieve := vnA.Achieves
	// 去除已达成的成就
	for _, v := range achieved {
		delete(allAchieve, v)
	}
	// 提交
	for _, v := range allAchieve {
		thisLog := t.achieveHelper.SubmitAchievement(v, 0)
		worker.pool.taskStatus.updateTaskState(t.getTaskID(), thisLog)
	}
}

func (t AntiEntropyGF) getTaskID() string {
	return t.id
}

func (t AntiEntropyGF) valid(libAchieve *vn.LIBAchievement) (string, []string, int, bool) {
	// 检查成就库的更新
	libAchieve.SetNovelAchievements(t.vNo, t.xmlHelper.UpdateAchievementLib(libAchieve.GetNovelAchievements(t.vNo)))
	achievedIDs, achievedNum, percent, retcode := t.achieveHelper.GetUserProgress()
	var msg string
	code := 1
	success := false
	if retcode == 1 {
		if len(achievedIDs) >= len(libAchieve.GetNovelAchievements(t.vNo).Achieves) {
			msg = "成就已经全部达成，无需处理"
			code = 0
		} else {
			msg = "成功加入处理队列"
			success = true
		}
	} else if retcode == -0.6 {
		msg = "请稍后重试"
		code = -2
	} else {
		msg = "无效链接"
		code = -1
	}
	respJSON := RespJSON{
		Retcode: code,
		Msg: msg,
		Progress: achievedNum,
		Percent: percent,
	}


	return respJSON.toString(), achievedIDs, achievedNum, success
}


func NewAntiEntropyGF(id string, req *http.Request) AntiEntropyGF {
	t := AntiEntropyGF{
		id:          id,
		vNo:         vn.ANTIENTROPY,
		req:         req,
		queryString: vn.PrepareQueryString(req),
		httpClient: &vn.MyAJAX{
			Client: http.Client{},
		},
		// URL
		URL_BASE:        URL_ANTIENTROPY,
		URL_CHAPTER_XML: URL_ANTIENTROPY_XML,
		URL_ACHIEVEMENT: URL_ANTIENTROPY_ACHIEVEMENT,
		URL_LATEST:      URL_ANTIENTROPY_LATEST,
		URL_EXHIBITION:  URL_ANTIENTROPY_EXHIBITION,
		// helper
	}
	xmlHelper := vn.XMLHelper{
		URL_BASE:        t.URL_BASE,
		URL_CHAPTER_XML: t.URL_CHAPTER_XML,
		URL_EXHIBITION:  t.URL_EXHIBITION,
		URL_LATEST:      t.URL_LATEST,
		QUERY_STRING:    t.queryString,
		HttpClient:      t.httpClient,
	}
	achieveHelper := vn.AchievementHelper{
		URL_BASE:        t.URL_BASE,
		URL_ACHIEVEMENT: t.URL_ACHIEVEMENT,
		COOKIE_NAME: map[string]string{
			"chapter": COOKIE_AE_CHAPTER,
			"scene":   COOKIE_AE_SCENE,
			"action":  COOKIE_AE_ACTION,
		},
		QUERY_STRING: t.queryString,
		HttpClient:   t.httpClient,
	}
	t.xmlHelper = &xmlHelper
	t.achieveHelper = &achieveHelper
	return t
}