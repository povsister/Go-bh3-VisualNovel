package vn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AchievementHelper struct {
	// constant
	URL_BASE        string
	URL_ACHIEVEMENT string
	COOKIE_NAME     map[string]string
	QUERY_STRING    string

	HttpClient *MyAJAX
}

func (ah *AchievementHelper) GetUserProgress() ([]string, int, float64, float64) {
	req := buildRequest("POST", getFullURL(ah.URL_ACHIEVEMENT, ah.QUERY_STRING), ah.getAchievementPostBody())
	addRefererHeader(ah.URL_BASE, ah.QUERY_STRING, req)

	respBody := ah.HttpClient.simulateAJAX(req)

	ret := ah.readAchievementJSON(respBody)

	var achievedIDs []string

	if ret.Retcode != 1 {
		return achievedIDs, 0, 0, ret.Retcode
	}

	// 坑逼的 json
	var num float64
	switch tp := ret.Progress.(type) {
	case string:
		t := string(tp)
		num, _ = strconv.ParseFloat(t, 64)
		break
	case float64:
		t := float64(tp)
		num = t
	}

	progress := num
	achieved := ret.Achievement
	achievedNum := 0
	if len(achieved) > 0 {
		// 获取已经完成的成就数量
		achievedNum = len(achieved)
		// 获取已经完成的成就 ID
		for _, v := range achieved {
			achievedIDs = append(achievedIDs, v["achievement"])
		}
	}
	return achievedIDs, achievedNum, progress, ret.Retcode
}

func (ah *AchievementHelper) SubmitAchievement(achieveCode achievementCode, secondsWait int) string {
	// 每次都变换随机数种子
	rand.Seed(time.Now().UnixNano())
	/* 判断是否是递归操作
	   0  : 否
	   !0 : 是
	*/
	var timeSleepInSec int
	if secondsWait == 0 {
		timeSleepInSec = rand.Intn(5) + 3
	} else {
		timeSleepInSec = secondsWait
	}
	dur, _ := time.ParseDuration(strconv.Itoa(timeSleepInSec) + "s")
	time.Sleep(dur)

	// 随机 增减 action 字段
	action, _ := strconv.Atoi(achieveCode.action)
	if action > 20 {
		for {
			randNum := rand.Intn(15) - 30
			newAction := action + randNum
			if newAction > 0 {
				achieveCode.action = strconv.Itoa(newAction)
				break
			}
		}
	}

	req := buildRequest("POST", getFullURL(ah.URL_ACHIEVEMENT, ah.QUERY_STRING), ah.getSubmitAchievementPostBody(achieveCode))
	addRefererHeader(ah.URL_BASE, ah.QUERY_STRING, req)
	req = ah.addAchievementCookies(achieveCode, req)

	respBody := ah.HttpClient.simulateAJAX(req)

	ret := ah.readAchievementSubmittedJSON(respBody)

	/*
		Retcode: 1     成功 插入记录
		Retcode: 0     记录已存在
		Retcode: -0.6  操作过于频繁
	*/
	if ret.Retcode == -0.6 {
		return ah.SubmitAchievement(achieveCode, timeSleepInSec*2)
	}

	msg := fmt.Sprintf("已提交第%s章-场景%s(对话%s)处的成就记录. code:%s", achieveCode.chapter, achieveCode.scene, achieveCode.action, achieveCode.code)
	log.Println(msg)

	return msg
}

func (ah *AchievementHelper) addAchievementCookies(achieveCode achievementCode, req *http.Request) *http.Request {

	cks := []http.Cookie{
		{Name: ah.COOKIE_NAME["chapter"], Value: achieveCode.chapter},
		{Name: ah.COOKIE_NAME["scene"], Value: achieveCode.scene},
		{Name: ah.COOKIE_NAME["action"], Value: achieveCode.action},
	}
	for _, v := range cks {
		req.AddCookie(&v)
	}

	return req
}

func (ah *AchievementHelper) getAchievementPostBody() io.Reader {
	t := url.Values{
		"achievement": {"LOAD"},
		"chapter":     {"1"},
		"scene":       {"-1"},
	}
	return strings.NewReader(t.Encode())
}

func (ah *AchievementHelper) getSubmitAchievementPostBody(achieveCode achievementCode) io.Reader {
	t := url.Values{
		"achievement": {achieveCode.code},
		"chapter":     {achieveCode.chapter},
		"scene":       {achieveCode.scene},
	}
	return strings.NewReader(t.Encode())
}

func (ah *AchievementHelper) readAchievementJSON(data []byte) Achievement {
	var ret Achievement

	body := bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	e := json.Unmarshal(body, &ret)
	if e != nil {
		log.Println("Read achievement JSON error:", e)
	}

	//log.Printf("%+v", ret)

	return ret
}

func (ah *AchievementHelper) readAchievementSubmittedJSON(data []byte) AchievementSubmitted {
	var ret AchievementSubmitted

	body := bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	e := json.Unmarshal(body, &ret)
	if e != nil {
		log.Println("Read achievement JSON submitted error:", e)
	}

	log.Printf("%+v", ret)

	return ret
}
