package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const AllMetrics string = `node_cpu_usage_nano_cores
node_memory_usage_bytes
pod_cpu_usage_nano_cores
pod_memory_usage_bytes
gpu_power
gpu_temperature
gpu_util
gpu_memory_util
gpu_device_info
`

func main() {
	ma := NewMetricsAggregator()
	go ma.Aggregate(time.Minute)
	// http server
	// 1.创建路由
	r := gin.Default()
	// 2.绑定路由规则，执行的函数
	// gin.Context，封装了request和response
	r.GET("/metric", func(c *gin.Context) {
		c.String(http.StatusOK, AllMetrics)
	})
	r.GET("/metric/:name", func(c *gin.Context) {
		name := c.Param("name")
		start := c.DefaultQuery("start", "")
		filterStr := c.DefaultQuery("filter", "")
		var filter []string
		if filterStr != "" {
			filter = strings.Split(filterStr, ",")
		}
		if start == "" {
			mlogs := ma.IC.queryMetricsLatest(name, filter...)
			c.JSON(http.StatusOK, map[string][]*MetricsLog{"data": mlogs})
		} else {
			mlogs := ma.IC.queryMetricsRange(name, start, filter...)
			fmt.Printf("%+v\n", mlogs)
			c.JSON(http.StatusOK, map[string][]*MetricsSeries{"data": mlogs})
		}
	})
	// 3.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(":8087")
}
