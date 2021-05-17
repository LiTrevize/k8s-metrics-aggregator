package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	ma := NewMetricsAggregator()
	fmt.Println(ma.LastAggregateTimeForNode("ubuntu16"))
	mlogs := ma.IC.queryMetricsRange("pod_memory_usage_bytes", "-1d")
	for _, mlog := range mlogs {
		fmt.Printf("%+v\n", mlog)
	}

	// http server
	// 1.创建路由
	r := gin.Default()
	// 2.绑定路由规则，执行的函数
	// gin.Context，封装了request和response
	r.GET("/metrics/:name", func(c *gin.Context) {
		name := c.Param("name")
		start := c.DefaultQuery("start", "")
		if start == "" {
			mlogs := ma.IC.queryMetricsLatest(name)
			c.JSON(http.StatusOK, map[string][]*MetricsLog{"data": mlogs})
		} else {
			mlogs := ma.IC.queryMetricsRange(name, start)
			c.JSON(http.StatusOK, map[string][]*MetricsSeries{"data": mlogs})
		}
	})
	// 3.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(":8087")
}
