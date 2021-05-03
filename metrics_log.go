package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type MetricsLog struct {
	Name string            `json:"name"`
	Tag  map[string]string `json:"tag"`
	Val  interface{}       `json:"val"`
	Time time.Time         `json:"time"`
}

func (ml *MetricsLog) Log() {
	b, err := json.Marshal(ml)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}

func (ml *MetricsLog) Parse(log string) error {
	err := json.Unmarshal([]byte(log), &ml)
	return err
}
