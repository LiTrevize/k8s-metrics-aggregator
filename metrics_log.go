package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type MetricsLog struct {
	Name  string                 `json:"name"`
	Tag   map[string]string      `json:"tag"`
	Val   interface{}            `json:"val"`
	Time  time.Time              `json:"time"`
	Field map[string]interface{} `json:"field"`
}

func NewMetricsLog(name string) *MetricsLog {
	mlog := &MetricsLog{
		Name: name,
		Tag:  make(map[string]string, 3)}
	return mlog
}

func (ml *MetricsLog) Log() {
	b, err := json.Marshal(ml)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}

func (ml *MetricsLog) ParseString(log string) error {
	err := json.Unmarshal([]byte(log), &ml)
	return err
}

func (ml *MetricsLog) ParseBytes(log []byte) error {
	err := json.Unmarshal(log, &ml)
	return err
}
