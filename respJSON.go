package main

import (
	"encoding/json"
	"log"
)

type RespJSON struct {
	Retcode  int
	Msg      string
	Progress int
	Percent  float64
	Total    int
}

type StateRespJSON struct {
	Retcode     int
	AddedTime   int64
	Processing  bool
	StartedTime int64
	Finished    bool
	EndedTime   int64
	Process     int
	Total       int
	Log         string
}

func respJSON() *RespJSON {
	t := RespJSON{
		Retcode:  1,
		Msg:      "success",
		Progress: 0,
		Percent:  0,
	}
	return &t
}

func (j *RespJSON) toString() string {
	respByte, err := json.Marshal(j)
	if err != nil {
		log.Println("JSON to string error:", err)
	}
	return string(respByte)
}

func (j *RespJSON) toByte() []byte {
	respByte, err := json.Marshal(j)
	if err != nil {
		log.Println("JSON to byte error:", err)
	}
	return respByte
}
