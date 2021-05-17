package main

import (
	"fmt"
)

func main() {
	ma := NewMetricsAggregator()
	fmt.Println(ma.LastAggregateTimeForNode("ubuntu16"))
	ma.IC.queryMetricsExample()
	// ma.Aggregate(time.Minute)
	// ic := NewInfluxdbClient()
	// ic.writeMetricsExample()
	// ic.WriteMetricsFromLog(`{"name":"cpu_usage_nano_cores","tag":{"node":"ubuntu16"},"val":195703672,"time":"2021-05-06T14:07:13.234253206+08:00"}`)
	// ic.WriteMetricsFromLog(`{"name":"gpu_device_info","tag":{"GPU":"0","UUID":"GPU-25be9bfa-ba12-0ab4-5fee-b1c4763a8eef","model":"GeForce GTX 1080 Ti","node":"ubuntu16"},"val":null,"time":"2021-05-06T14:07:13.234253206+08:00","field":{"bandwidth":"15760 MB/s","clock_cores":"1911 MHz","clock_memory":"5505 MHz","memory":"11178 MiB","power":"250 W"}}`)
	// ms := ic.queryMetricsRange("cpu_usage_nano_cores", "-1d")
	// fmt.Printf("%+v\n", ms)
	// ml := ic.queryMetricsLatest("gpu_device_info")
	// fmt.Printf("%+v\n", ml)
}
