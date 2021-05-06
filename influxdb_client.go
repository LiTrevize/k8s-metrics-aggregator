package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxdbClient struct {
	Client   influxdb2.Client
	WriteAPI api.WriteAPIBlocking
	QueryAPI api.QueryAPI
	Org      string
	Bucket   string
}

func NewInfluxdbClient() *InfluxdbClient {
	ic := new(InfluxdbClient)
	url := "http://127.0.0.1:8086"
	token := "KhYF9SoMOHcb9m8DJnrwRH5LVwoRLy-YVGmZKSYegp6LC5CEiKoksGFF2KK4bZFx1gHiS4lCZKDP1V9kN2oUrw=="
	ic.Client = influxdb2.NewClient(url, token)
	ic.Org = "SJTU"
	ic.Bucket = "metrics"

	ic.WriteAPI = ic.Client.WriteAPIBlocking(ic.Org, ic.Bucket)
	ic.QueryAPI = ic.Client.QueryAPI(ic.Org)

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
    |> range(start: -1d)`)
	if err == nil {
		for result.Next() {
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.Record().Measurement())
			}
			fmt.Printf("%v: %v\n", result.Record().Field(), result.Record().Value())
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
}

func (ic *InfluxdbClient) queryMetricsLatest(name string) *MetricsLog {
	ml := &MetricsLog{Name: name,
		Tag:   make(map[string]string, 3),
		Field: make(map[string]interface{}, 5)}
	setTag := false
	result, err := ic.QueryAPI.Query(context.Background(), fmt.Sprintf(`from(bucket:"%s")
    |> range(start: -1d)
	|> filter(fn: (r) => r._measurement == "%s")
	|> last()`, ic.Bucket, name))
	if err == nil {
		for result.Next() {
			if !setTag {
				for k, v := range result.Record().Values() {
					if strings.HasPrefix(k, "_") || k == "result" || k == "table" {
						continue
					}
					ml.Tag[k] = v.(string)
				}
				setTag = true
			}
			ml.Field[result.Record().Field()] = result.Record().Value()
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		fmt.Printf("query error: %s\n", err)
	}
	return ml
}

func (ic *InfluxdbClient) queryMetricsRange(name string, start string) *MetricsSeries {
	ms := NewMetricsSeries(name)
	setTag := false
	result, err := ic.QueryAPI.Query(context.Background(), fmt.Sprintf(`from(bucket:"%s")
    |> range(start: %s)
	|> filter(fn: (r) => r._measurement == "%s")`, ic.Bucket, start, name))
	if err == nil {
		for result.Next() {
			if !setTag {
				for k, v := range result.Record().Values() {
					if strings.HasPrefix(k, "_") || k == "result" || k == "table" {
						continue
					}
					ms.Tag[k] = v
				}
				setTag = true
			}
			ms.Values = append(ms.Values, result.Record().Value())
			ms.Time = append(ms.Time, result.Record().Time())
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		fmt.Printf("query error: %s\n", err)
	}
	return ms
}

type MetricsSeries struct {
	Name   string                 `json:"name"`
	Tag    map[string]interface{} `json:"tag"`
	Values []interface{}          `json:"values"`
	Time   []time.Time            `json:"time"`
}

func NewMetricsSeries(name string) *MetricsSeries {
	return &MetricsSeries{name,
		make(map[string]interface{}, 3),
		make([]interface{}, 0, 10),
		make([]time.Time, 0, 10)}
}
