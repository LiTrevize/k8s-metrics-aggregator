package main

import (
	"fmt"
	"net/http"

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
	fmt.Println(ma.LastAggregateTimeForNode("ubuntu16"))
	mlogs := ma.IC.queryMetricsRange("pod_memory_usage_bytes", "-5d")
	for _, mlog := range mlogs {
		fmt.Printf("%+v\n", mlog)
	}

	// http server
	// 1.创建路由
	r := gin.Default()
	// 2.绑定路由规则，执行的函数
	// gin.Context，封装了request和response
	r.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, AllMetrics)
	})
	r.GET("/metrics/:name", func(c *gin.Context) {
		name := c.Param("name")
		start := c.DefaultQuery("start", "")
		if start == "" {
			mlogs := ma.IC.queryMetricsLatest(name)
			c.JSON(http.StatusOK, map[string][]*MetricsLog{"data": mlogs})
		} else {
			mlogs := ma.IC.queryMetricsRange(name, start)
			fmt.Printf("%+v\n", mlogs)
			c.JSON(http.StatusOK, map[string][]*MetricsSeries{"data": mlogs})
		}
	})
	// 3.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(":8087")
}
