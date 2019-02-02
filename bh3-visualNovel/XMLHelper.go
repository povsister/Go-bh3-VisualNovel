package vn

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

type XMLHelper struct {
	// constant
	URL_BASE        string
	URL_CHAPTER_XML string
	URL_EXHIBITION  string
	URL_LATEST      string
	QUERY_STRING    string

	HttpClient *MyAJAX
}

func (x *XMLHelper) getXML(urlGiven string) []byte {
	var respBody []byte
	req := buildRequest("GET", urlGiven, nil)
	addRefererHeader(x.URL_BASE, x.QUERY_STRING, req)
	respBody = x.HttpClient.simulateAJAX(req)

	return respBody
}

func (x *XMLHelper) getChapterXML(chapter string) []byte {

	return x.getXML(fmt.Sprintf(x.URL_CHAPTER_XML, chapter, float64ToString(rand.Float64())))
}

func (x *XMLHelper) readXMLToken(decoder *xml.Decoder) xml.Token {
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

func (x *XMLHelper) parseXMLForAchievement(decoder *xml.Decoder, keyNeeded map[string]int, scene *string, action *int) string {
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
			if key == "speak" || key == "text" {
				*action++
			}
			// 判断是否是需要的 Token 如果有 post 属性则返回
			if _, ok := keyNeeded[key]; ok {
				if len(se.Attr) > 0 {
					for _, v := range se.Attr {
						if strings.ToLower(v.Name.Local) == "post" {
							return v.Value
						}
					}
				}
			}
		}
	}
}

func (x *XMLHelper) formatAchievementID(chapter string, index string) string {
	cB := []rune(chapter)
	cS := chapter
	if len(cB) < 2 {
		cS = "0" + chapter
	}
	return "10" + cS + index
}

func (x *XMLHelper) getAchievementFromXML(data []byte, chapter string) map[string]achievementCode {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	keyNeeded := map[string]int{
		"remark": 1,
		"speak":  1,
		"text":   1,
		"end":    1,
	}
	scene := ""
	action := 0

	ret := make(map[string]achievementCode)
	index := 1
	lastID := ""
	for {
		ach := x.parseXMLForAchievement(decoder, keyNeeded, &scene, &action)
		if ach != "" {
			lastID = x.formatAchievementID(chapter, strconv.Itoa(index))
			t := achievementCode{
				id:      lastID,
				chapter: chapter,
				scene:   scene,
				action:  strconv.Itoa(action),
				code:    ach,
			}
			ret[lastID] = t
			index++
		} else {
			// 将最后一个 end 成就挪到第一个位置上
			if len(ret) > 1 {
				endA := ret[lastID]
				delete(ret, lastID)
				endID := string([]rune(lastID)[:4]) + "0"
				endA.id = endID
				ret[endID] = endA
			}
			return ret
		}
	}

}

func (x *XMLHelper) GetTotalChapterNum() int {
	respBody := x.getXML(fmt.Sprintf(x.URL_EXHIBITION, float64ToString(rand.Float64())))
	decoder := xml.NewDecoder(bytes.NewReader(respBody))
	var achLogs []string
	for {
		token := x.readXMLToken(decoder)
		if token == nil {
			break
		}
		switch tp := token.(type) {
		case xml.StartElement:
			se := xml.StartElement(tp)
			key := strings.ToLower(se.Name.Local)
			if key == "log" {
				if len(se.Attr) > 0 {
					for _, v := range se.Attr {
						if strings.ToLower(v.Name.Local) == "id" {
							achLogs = append(achLogs, v.Value)
						}
					}
				}
			}

		}
	}
	lastID := achLogs[len(achLogs)-1]
	t := []rune(lastID)
	num, _ := strconv.Atoi(string(t[2:4]))
	return num
}

func (x *XMLHelper) getNovelVersion() string {
	respBody := x.getXML(fmt.Sprintf(x.URL_LATEST, float64ToString(rand.Float64())))
	decoder := xml.NewDecoder(bytes.NewReader(respBody))
	versionInfo := ""
	for {
		token := x.readXMLToken(decoder)
		if token == nil {
			break
		}
		switch tp := token.(type) {
		case xml.StartElement:
			se := xml.StartElement(tp)
			key := strings.ToLower(se.Name.Local)
			if key == "log" {
				if len(se.Attr) > 0 {
					for _, v := range se.Attr {
						if v.Name.Local == "lastDate" {
							versionInfo = v.Value
						}
					}
				}
			}

		}
	}
	return versionInfo
}

func (x *XMLHelper) generateAchievementLib() VnAchievements {
	version := x.getNovelVersion()
	//if achievements.version == version {
	//	// no need to update
	//	return achievements
	//}

	chapterNum := x.GetTotalChapterNum()
	all := VnAchievements{
		version:  version,
		Achieves: make(map[string]achievementCode),
	}
	for i := 1; i <= chapterNum; i++ {
		for k, v := range x.getAchievementFromXML(x.getChapterXML(strconv.Itoa(i)), strconv.Itoa(i)) {
			all.Achieves[k] = v
		}
	}

	return all
}

func (x *XMLHelper) UpdateAchievementLib(achievements VnAchievements) VnAchievements {
	newVersion := x.getNovelVersion()
	if newVersion != achievements.version {
		return x.generateAchievementLib()
	}
	return achievements
}
