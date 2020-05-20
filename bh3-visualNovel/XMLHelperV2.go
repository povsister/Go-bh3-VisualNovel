package vn

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

type XMLHelperV2 struct {
	URL_BASE     string
	URL_REFERER  string
	QUERY_STRING string

	HttpClient *MyAJAX
}

func (x *XMLHelperV2) getXML(urlGiven string) []byte {
	var respBody []byte
	req := buildRequestV2("GET", urlGiven, nil)
	addRefererHeader(x.URL_REFERER, x.QUERY_STRING, req)
	respBody = x.HttpClient.simulateAJAX(req)

	return respBody
}

func (x *XMLHelperV2) readXMLToken(decoder *xml.Decoder) xml.Token {
	token, err := decoder.Token()
	if err == io.EOF {
		//log.Println("Finish parsing XML")
		return nil
	}
	if err != nil {
		log.Println("XML parse error: ", err)
		return nil
	}
	return token
}

func (x *XMLHelperV2) parseXMLForAchievement(decoder *xml.Decoder, keyNeeded map[string]int, scene *string, action *int) string {
	for {
		token := x.readXMLToken(decoder)
		if token == nil {
			return ""
		}

		switch tp := token.(type) {
		case xml.StartElement:
			se := xml.StartElement(tp)
			key := strings.ToLower(se.Name.Local)
			// 更新 *scene
			if key == "scene" {
				if len(se.Attr) > 0 {
					for _, v := range se.Attr {
						if strings.ToLower(v.Name.Local) == "id" {
							*scene = v.Value
						}
					}
				}
			}
			// 更新 *action
			if key == "mono" || key == "dialog" {
				*action++
			}
			// 判断是否是需要的 Token 如果有 post 属性则返回
			if _, ok := keyNeeded[key]; ok {
				if len(se.Attr) > 0 {
					isAchievement := false
					for _, v := range se.Attr {
						if strings.ToLower(v.Name.Local) == "action" && strings.ToLower(v.Value) == "achievement" {
							isAchievement = true
						}
						if strings.ToLower(v.Name.Local) == "aid" && isAchievement {
							return v.Value
						}
					}
				}
			}
		}
	}
}

func (x *XMLHelperV2) getAchievementFromXML(data []byte, chapter string, name string) map[string]achievementCode {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	keyNeeded := map[string]int{
		"event": 1,
	}
	scene := ""
	action := 0

	ret := make(map[string]achievementCode)
	lastID := ""
	for {
		ach := x.parseXMLForAchievement(decoder, keyNeeded, &scene, &action)
		if ach != "" {
			// 检查是否存在重复的 ach
			if _, ok := ret[ach]; ok {
				continue
			}
			lastID = ach
			t := achievementCode{
				id:      lastID,
				chapter: chapter,
				scene:   scene,
				action:  strconv.Itoa(action),
				code:    ach,
				name:    name,
			}
			ret[lastID] = t
		} else {
			return ret
		}
	}
}

type SwordIndex struct {
	Data    SwordIndexData `json:"data"`
	Msg     string         `json:"msg"`
	RetCode int            `json:"retcode"`
}

type SwordIndexData struct {
	Achievements []SwordAchievement `json:"achievements"`
	Chapters     []SwordChapter     `json:"chapters"`
	ID           string             `json:"id"`
	Intro        string             `json:"intro"`
	IsFinished   string             `json:"is_finished"`
	IsLogin      int                `json:"is_login"`
	Name         string             `json:"name"`
	// Resources.xml
	XmlURL string `json:"xml_url"`
}

type SwordAchievement struct {
	ChapterID string   `json:"chapter_id"`
	ID        string   `json:"id"`
	Desc      string   `json:"desc"`
	ImgURLs   []string `json:"img_urls"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	UniqueKey string   `json:"unique_key"`
	Weight    string   `json:"weight"`
}

type SwordChapter struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Order     string        `json:"order"`
	Type      string        `json:"type"`
	StartTime string        `json:"start_time"`
	EndTime   string        `json:"end_time"`
	XmlURL    string        `json:"xml_url"`
	Tips      string        `json:"tips"`
	Parts     []interface{} `json:"parts"`
}

func (x *XMLHelperV2) getNovelVersion() string {
	respBody := x.getXML(getReferURL(x.URL_BASE, x.QUERY_STRING))
	novelVersion := SwordIndex{}
	err := json.Unmarshal(respBody, &novelVersion)
	if err != nil {
		fmt.Println("7Sword Version JSON Unmarshal ERROR:", err)
	}
	return strconv.Itoa(len(novelVersion.Data.Chapters))
}

func (x *XMLHelperV2) getAllChapters() map[string]SwordChapter {
	respBody := x.getXML(getReferURL(x.URL_BASE, x.QUERY_STRING))
	novels := SwordIndex{}
	err := json.Unmarshal(respBody, &novels)
	if err != nil {
		fmt.Println("7Sword All Chapters JSON Unmarshal ERROR:", err)
	}
	chapters := novels.Data.Chapters
	ret := make(map[string]SwordChapter)
	for _, chap := range chapters {
		ret[chap.ID] = chap
	}
	return ret
}

func (x *XMLHelperV2) generateAchievementLib() VnAchievements {
	version := x.getNovelVersion()

	chapters := x.getAllChapters()
	all := VnAchievements{
		version:  version,
		Achieves: make(map[string]achievementCode),
	}
	for index, chapter := range chapters {
		for k, v := range x.getAchievementFromXML(x.getXML(getFullURL(chapter.XmlURL, fmt.Sprintf("t=%s", float64ToString(rand.Float64())))), index, chapter.Name) {
			all.Achieves[k] = v
		}
	}
	return all
}

func (x *XMLHelperV2) UpdateAchievementLib(achievements VnAchievements) VnAchievements {
	newVersion := x.getNovelVersion()
	if newVersion != achievements.version && newVersion != "0" {
		return x.generateAchievementLib()
	}
	return achievements
}
