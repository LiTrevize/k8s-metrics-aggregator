package main

import (
	"context"
	"fmt"
	"reflect"
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

func (ic *InfluxdbClient) WriteMetricsFromLog(log string) *MetricsLog {
	ml := MetricsLog{}
	err := ml.ParseString(log)
	if err != nil {
		return nil
	}
	ic.WriteMetrics(&ml)
	return &ml
}

func (ic *InfluxdbClient) WriteMetricsFromLogBytes(log []byte) *MetricsLog {
	ml := MetricsLog{}
	err := ml.ParseBytes(log)
	if err != nil {
		return nil
	}
	ic.WriteMetrics(&ml)
	return &ml
}

func (ic *InfluxdbClient) queryMetricsExample() {
	result, err := ic.QueryAPI.Query(context.Background(), `from(bucket:"metrics")
    |> range(start: -1d)`)
	if err == nil {
		for result.Next() {
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.Record().Measurement())
			}
			for k, v := range result.Record().Values() {
				if strings.HasPrefix(k, "_") || k == "result" || k == "table" {
					continue
				}
				fmt.Printf("%v: %v ", k, v)
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

func (ic *InfluxdbClient) makeFluxQuery(metric string, start interface{}, last bool, filter ...string) string {
	query := fmt.Sprintf(`from(bucket:"%s")
    |> range(start: %v)
	|> filter(fn: (r) =>
	r._measurement == "%s"`, ic.Bucket, start, metric)
	for i := 0; i < len(filter); i += 2 {
		query += fmt.Sprintf(` and
		r.%s == "%s"`, filter[i], filter[i+1])
	}
	query += "\n)"
	if last {
		query += "\n|> last()"
	}
	return query
}

func (ic *InfluxdbClient) queryMetricsLatest(name string, filter ...string) []*MetricsLog {
	lastTime := ic.queryLastTime("_measurement", name)
	lastTime = lastTime.Add(-time.Second * 5)
	mlogs := make([]*MetricsLog, 0, 5)
	result, err := ic.QueryAPI.Query(context.Background(), ic.makeFluxQuery(name, lastTime.Format(time.RFC3339), true, filter...))
	if err == nil {
		for result.Next() {
			mlog := NewMetricsLog(name)
			for k, v := range result.Record().Values() {
				if strings.HasPrefix(k, "_") || k == "result" || k == "table" {
					continue
				}
				mlog.Tag[k] = v.(string)
			}
			mlog.Val = result.Record().Value()
			mlog.Time = result.Record().Time()
			mlogs = append(mlogs, mlog)
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		fmt.Printf("query error: %s\n", err)
	}
	return mlogs
}

func (ic *InfluxdbClient) queryMetricsRange(name string, start string, filter ...string) []*MetricsSeries {
	mlogs := make([]*MetricsSeries, 0, 5)
	result, err := ic.QueryAPI.Query(context.Background(), ic.makeFluxQuery(name, start, false, filter...))
	if err == nil {
		ms := NewMetricsSeries(name)
		setTag := false
		for result.Next() {
			if !setTag {
				for k, v := range result.Record().Values() {
					if strings.HasPrefix(k, "_") || k == "result" || k == "table" {
						continue
					}
					ms.Tag[k] = v.(string)
				}
				setTag = true
			} else {
				newTag := make(map[string]string, 3)
				for k, v := range result.Record().Values() {
					if strings.HasPrefix(k, "_") || k == "result" || k == "table" {
						continue
					}
					newTag[k] = v.(string)
				}
				if !reflect.DeepEqual(ms.Tag, newTag) {
					mlogs = append(mlogs, ms)
					ms = NewMetricsSeries(name)
					ms.Tag = newTag
				}
			}
			ms.Values = append(ms.Values, result.Record().Value())
			ms.Time = append(ms.Time, result.Record().Time())
		}
		if len(ms.Values) > 0 {
			mlogs = append(mlogs, ms)
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		fmt.Printf("query error: %s\n", err)
	}
	return mlogs
}

func (ic *InfluxdbClient) queryLastTime(tagKey, tagVal string) time.Time {
	result, err := ic.QueryAPI.Query(context.Background(), fmt.Sprintf(`from(bucket:"%s")
	|> range(start: -7d)
	|> filter(fn: (r) => r.%s == "%s")
	|> last()`, ic.Bucket, tagKey, tagVal))
	if err == nil {
		for result.Next() {
			return result.Record().Time()
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		fmt.Printf("query error: %s\n", err)
	}
	return time.Now()
}

type MetricsSeries struct {
	Name   string            `json:"name"`
	Tag    map[string]string `json:"tag"`
	Values []interface{}     `json:"values"`
	Time   []time.Time       `json:"time"`
}

func NewMetricsSeries(name string) *MetricsSeries {
	return &MetricsSeries{name,
		make(map[string]string, 3),
		make([]interface{}, 0, 10),
		make([]time.Time, 0, 10)}
}
