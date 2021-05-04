package main

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxdbClient struct {
	Client   influxdb2.Client
	WriteAPI api.WriteAPIBlocking
	QueryAPI api.QueryAPI
}

func NewInfluxdbClient() *InfluxdbClient {
	ic := new(InfluxdbClient)
	url := "http://127.0.0.1:8086"
	token := "KhYF9SoMOHcb9m8DJnrwRH5LVwoRLy-YVGmZKSYegp6LC5CEiKoksGFF2KK4bZFx1gHiS4lCZKDP1V9kN2oUrw=="
	ic.Client = influxdb2.NewClient(url, token)
	org := "SJTU"
	bucket := "metrics"
	ic.WriteAPI = ic.Client.WriteAPIBlocking(org, bucket)
	ic.QueryAPI = ic.Client.QueryAPI(org)

	return ic
}

func (ic *InfluxdbClient) writeMetricsExample() {
	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45},
		time.Now())
	ic.WriteAPI.WritePoint(context.Background(), p)
}

func (ic *InfluxdbClient) WriteMetrics(ml *MetricsLog) {
	if ml.Field == nil {
		p := influxdb2.NewPoint(ml.Name, ml.Tag, map[string]interface{}{"val": ml.Val}, ml.Time)
		ic.WriteAPI.WritePoint(context.Background(), p)
	} else {
		p := influxdb2.NewPoint(ml.Name, ml.Tag, ml.Field, ml.Time)
		ic.WriteAPI.WritePoint(context.Background(), p)
	}
}

func (ic *InfluxdbClient) WriteMetricsFromLog(log string) {
	ml := MetricsLog{}
	ml.Parse(log)
	ic.WriteMetrics(&ml)
}

func (ic *InfluxdbClient) queryMetricsExample() {
	result, err := ic.QueryAPI.Query(context.Background(), `from(bucket:"metrics")
    |> range(start: -1h)`)
	fmt.Println(result)
	if err == nil {
		for result.Next() {
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			fmt.Printf("value: %v\n", result.Record().Value())
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}

}
