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
	VNO             int
	URL_BASE        string
	URL_ACHIEVEMENT string
	COOKIE_NAME     map[string]string
	QUERY_STRING    string
	// for V2
	URL_REFERER string

	HttpClient *MyAJAX
}

func (ah *AchievementHelper) GetUserProgress() (map[string]int, int, float64, float64) {
	req := buildRequest("POST", getFullURL(ah.URL_ACHIEVEMENT, ah.QUERY_STRING), ah.getAchievementPostBody())
	addRefererHeader(ah.URL_BASE, ah.QUERY_STRING, req)

	respBody := ah.HttpClient.simulateAJAX(req)

	ret := ah.readAchievementJSON(respBody)

	achievedIDs := make(map[string]int)

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
			achievedIDs[v["achievement"]] = 1
		}
	}
	return achievedIDs, achievedNum, progress, ret.Retcode
}

func (ah *AchievementHelper) GetUserProgressV2() (map[string]int, int, int) {
	req := buildRequestV2("GET", getFullURL(ah.URL_BASE, ah.QUERY_STRING), nil)
	addRefererHeader(ah.URL_REFERER, ah.QUERY_STRING, req)

	respBody := ah.HttpClient.simulateAJAX(req)

	ret, isLogin := ah.readAchievementJSONV2(respBody)

	achievedIDs := make(map[string]int)

	if isLogin != 1 {
		return achievedIDs, 0, isLogin
	}

	achieved := ret
	achievedNum := len(achieved)
	if len(achieved) > 0 {
		// 获取已经完成的成就 ID
		for _, v := range achieved {
			achievedIDs[v.UniqueKey] = 1
		}
	}
	return achievedIDs, achievedNum, isLogin
}

func (ah *AchievementHelper) SubmitAchievement(achieveCode achievementCode, secondsWait int) (string, bool, bool) {
	// 每次都变换随机数种子
	rand.Seed(time.Now().UnixNano())
	/* 判断是否是递归操作
	   0  : 否
	   !0 : 是
	*/
	// 每次随机等待时间 10 - 20秒
	var timeSleepInSec int
	if secondsWait == 0 {
		timeSleepInSec = rand.Intn(10) + 10
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

	msg := fmt.Sprintf("已提交第%s章-场景%s(对话%s)处的成就记录. Msg:%s", achieveCode.chapter, achieveCode.scene, achieveCode.action, ret.Msg)

	//log.Println(msg)

	/*
		Retcode: 1     成功 插入记录
		Retcode: 0     记录已存在
		Retcode: -0.6  Your operation is too frequent. Retcode:-0.6
		Retcode: -1    Your operation is too frequent. Retcode:-1
		Retcode: -1    Illegal Operation. Retcode:-1
	*/
	if ret.Retcode < 0 {
		if ret.Retcode == -1 || ret.Retcode == -0.6 {
			// too frequent
			if strings.Index(ret.Msg, "frequent") != -1 {
				// Msg, failed?, frequent?
				return msg, true, true
			}
			return msg, true, false
			// return ah.SubmitAchievement(achieveCode, timeSleepInSec*2)
		}
		return msg, true, false
	}
	if ret.Retcode == 0 && ret.Msg == "" {
		return "Request failed. Re-try ...", true, true
	}
	return msg, false, false
}

func (ah *AchievementHelper) SubmitAchievementV2(achieveCode achievementCode, secondsWait int) (string, bool, bool) {
	// 每次都变换随机数种子
	rand.Seed(time.Now().UnixNano())
	/* 判断是否是递归操作
	   0  : 否
	   !0 : 是
	*/
	// 每次随机等待时间 10 - 20秒
	var timeSleepInSec int
	if secondsWait == 0 {
		timeSleepInSec = rand.Intn(10) + 10
	} else {
		timeSleepInSec = secondsWait
	}
	dur, _ := time.ParseDuration(strconv.Itoa(timeSleepInSec) + "s")
	time.Sleep(dur)

	req := buildRequest("POST", getReferURL(ah.URL_ACHIEVEMENT, ah.QUERY_STRING), ah.getSubmitAchievementPostBodyV2(achieveCode))
	addRefererHeader(ah.URL_REFERER, ah.QUERY_STRING, req)

	respBody := ah.HttpClient.simulateAJAX(req)

	ret := ah.readAchievementSubmittedJSONV2(respBody)

	msg := fmt.Sprintf("已提交%s:%s-场景%s(对话%s)处的成就记录. Msg:%s", achieveCode.chapter, achieveCode.name, achieveCode.scene, achieveCode.action, ret.Msg)

	//log.Println(msg)

	/*
		Retcode: 0     成功 插入记录
		Retcode: -1005     记录已存在
		Retcode: -0.6  Your operation is too frequent. Retcode:-0.6
		Retcode: -1    Your operation is too frequent. Retcode:-1
		Retcode: -1    Illegal Operation. Retcode:-1
	*/
	if ret.RetCode < 0 {
		if ret.RetCode == -1005 {
			return msg, false, false
		} else {
			// too frequent
			if strings.Index(ret.Msg, "frequent") != -1 {
				// Msg, failed?, frequent?
				return msg, true, true
			}
		}
		return msg, true, false
	}
	if ret.RetCode == 0 && ret.Msg == "" {
		return "Request failed. Re-try ...", true, true
	}
	return msg, false, false
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
	//var achieveStr string
	//if ah.VNO == DURANDAL {
	//	achieveStr = "GET_AWARD_CN"
	//} else {
	//	achieveStr = "LOAD"
	//}
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

func (ah *AchievementHelper) getSubmitAchievementPostBodyV2(achieveCode achievementCode) io.Reader {
	t := url.Values{
		"achievement_key": {achieveCode.code},
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

func (ah *AchievementHelper) readAchievementJSONV2(data []byte) ([]SwordAchievement, int) {
	ret := SwordIndex{}
	body := bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	e := json.Unmarshal(body, &ret)
	if e != nil {
		log.Println("Read achievement 7Sword JSON error:", e)
	}

	//log.Printf("%+v", ret)

	return ret.Data.Achievements, ret.Data.IsLogin
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

type SevenSwordSubmitted struct {
	Data    interface{} `json:"data"`
	Msg     string      `json:"msg"`
	RetCode int         `json:"retcode"`
}

func (ah *AchievementHelper) readAchievementSubmittedJSONV2(data []byte) SevenSwordSubmitted {
	ret := SevenSwordSubmitted{}
	body := bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	e := json.Unmarshal(body, &ret)
	if e != nil {
		log.Println("Read achievement 7Sword JSON submitted error:", e)
	}

	log.Printf("%+v", ret)

	return ret
}
