package main

func main() {
	ic := NewInfluxdbClient()
	ic.writeMetricsExample()
	ic.queryMetricsExample()
	ic.WriteMetricsFromLog(`{"name":"cpu_usage_nano_cores","tag":{"node":"ubuntu16"},"val":195703672,"time":"2021-05-03T00:16:09Z"}`)
}
