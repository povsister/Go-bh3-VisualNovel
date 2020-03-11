package vn

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MyAJAX struct {
	Client http.Client
}

func (a *MyAJAX) simulateAJAX(req *http.Request) []byte {

	resp, err := a.Client.Do(req)
	if err != nil {
		log.Println("AJAX error: ", err)
		time.Sleep(5e9)
		return a.simulateAJAX(req)
	}
	defer resp.Body.Close()

	decompressed := resp.Body

	encodingHeader := resp.Header.Get("Content-Encoding")
	if encodingHeader == "gzip" {
		decompressed, err = gzip.NewReader(resp.Body)
		if err != nil {
			log.Println("AJAX Gzip decompress error:", err)
		}
	}

	respBody, _ := ioutil.ReadAll(decompressed)

	return respBody
}

func buildRequest(method string, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	//req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", req.URL.Scheme+"://"+req.URL.Host)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	} else {
		req.Header.Set("Accept", "application/xml, text/xml, */*; q=0.01")
	}
	return req
}

func addRefererHeader(refer string, query string, req *http.Request) {
	req.Header.Set("Referer", getReferURL(refer, query))
}

func getFullURL(url string, query string) string {
	t := []string{url, query}
	return strings.Join(t, "?")
}

func getReferURL(urlGiven string, query string) string {
	t, _ := url.ParseQuery(query)
	q := url.Values{
		"auth_key": t["auth_key"],
		"_time":    {"1"},
		"from":     {"bh3"},
		"sign":     t["sign"],
	}
	f := []string{urlGiven, q.Encode()}

	return strings.Join(f, "?")
}
