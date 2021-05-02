package main

func main() {
	ic := NewInfluxdbClient()
	ic.writeMetrics()
	ic.writeMetrics()
	ic.queryMetrics()
}
