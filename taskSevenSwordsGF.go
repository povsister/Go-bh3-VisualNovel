package main

import (
	vn "bh3-visualNovel"
	"fmt"
	"log"
	"net/http"
)

const (
	URL_7_SWORDS             = "https://api.mihoyo.com/bh3-event-e20190831sword/api/index"
	URL_7_SWORDS_ACHIEVEMENT = "https://api.mihoyo.com/bh3-event-e20190831sword/api/achievement"
	URL_7_SWORDS_REFERER     = "https://webstatic.mihoyo.com/bh3/event/novel-7swords/index.html"
)

type SevenSwordsGF struct {
	id          string
	vNo         int
	req         *http.Request
	queryString string
	httpClient  *vn.MyAJAX
	// URLS
	URL_BASE        string
	URL_ACHIEVEMENT string
	URL_REFERER     string
	// helper
	xmlHelper     *vn.XMLHelperV2
	achieveHelper *vn.AchievementHelper
}

func (t SevenSwordsGF) process(worker *Worker) (bool, bool) {
	vnA := worker.pool.libAchievement.GetNovelAchievements(t.vNo)

	// 玩家已完成成就
	achieved := worker.pool.taskStatus.task[t.getTaskID()].achievedIDs
	// 全部成就
	allAchieve := vnA.Achieves
	// 找出未达成的成就
	for k, v := range allAchieve {
		if _, ok := achieved[k]; !ok {
			// 提交未达成的成就
			thisLog, failed, frequent := t.achieveHelper.SubmitAchievementV2(v, 0)
			// 先更新log
			worker.pool.taskStatus.updateTaskState(t.getTaskID(), thisLog)
			if failed {
				// success?, frequent?
				return !failed, frequent
			} else {
				// 提交成功则更新已完成的任务ID
				worker.pool.taskStatus.updateTaskState(t.getTaskID(), "progress++")
				achieved[k] = 1
				worker.pool.taskStatus.setAchievedIDs(t.getTaskID(), achieved)
			}
			log.Println(fmt.Sprintf("cat:%d cpCount:%d id:%s log:%s worker:%d", t.vNo, len(worker.pool.taskStatus.task[t.getTaskID()].achievedIDs), t.id, thisLog, worker.id))
		}
	}
	return true, false
}

func (t SevenSwordsGF) getTaskID() string {
	return t.id
}

func (t SevenSwordsGF) valid(libAchieve *vn.LIBAchievement) (string, map[string]int, int, int, bool) {
	// 检查成就库的更新
	libAchieve.SetNovelAchievements(t.vNo, t.xmlHelper.UpdateAchievementLib(libAchieve.GetNovelAchievements(t.vNo)))
	achievedIDs, achievedNum, isLogin := t.achieveHelper.GetUserProgressV2()
	totalAchieves := len(libAchieve.GetNovelAchievements(t.vNo).Achieves)
	var msg string
	code := 1
	success := false
	if isLogin == 1 {
		if len(achievedIDs) >= totalAchieves {
			msg = "成就已经全部达成"
			code = 0
		} else {
			msg = "成功加入处理队列"
			success = true
		}
		//} else if retcode == -1 || retcode == -0.6 {
		//	msg = "你的帐号已被米忽悠限制，请半小时后重试"
		//	code = -1
	} else {
		msg = "未检测到游戏id，请从游戏内重新获取URL"
		code = -2
	}
	respJSON := RespJSON{
		Retcode:  code,
		Msg:      msg,
		Progress: achievedNum,
		Percent:  1.0,
		Total:    totalAchieves,
	}

	return respJSON.toString(), achievedIDs, achievedNum, totalAchieves, success
}

func NewSevenSwordGF(id string, req *http.Request) SevenSwordsGF {
	t := SevenSwordsGF{
		id:          id,
		vNo:         vn.SEVEN_SWORDS,
		req:         req,
		queryString: vn.PrepareQueryString(req),
		httpClient: &vn.MyAJAX{
			Client: http.Client{},
		},
		URL_BASE:        URL_7_SWORDS,
		URL_ACHIEVEMENT: URL_7_SWORDS_ACHIEVEMENT,
		URL_REFERER:     URL_7_SWORDS_REFERER,
	}
	xmlHelper := vn.XMLHelperV2{
		URL_BASE:     t.URL_BASE,
		URL_REFERER:  t.URL_REFERER,
		QUERY_STRING: t.queryString,
		HttpClient:   t.httpClient,
	}
	achieveHelper := vn.AchievementHelper{
		VNO:             t.vNo,
		URL_BASE:        t.URL_BASE,
		URL_ACHIEVEMENT: t.URL_ACHIEVEMENT,
		URL_REFERER:     t.URL_REFERER,
		QUERY_STRING:    t.queryString,
		HttpClient:      t.httpClient,
	}
	t.xmlHelper = &xmlHelper
	t.achieveHelper = &achieveHelper
	return t
}
