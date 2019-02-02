package vn

import (
	"net/http"
	"strconv"
	"strings"
)

func PrepareQueryString(req *http.Request) string {
	return req.URL.RawQuery
}

func GetTaskIdFromPath(req *http.Request) string {
	path := req.URL.Path
	pathArr := strings.Split(path, "/")
	return pathArr[len(pathArr)-1]
}

func float64ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', 15, 64)
}
