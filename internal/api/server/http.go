package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Run(h *ScanHandler, host string, port uint) error {
	r := gin.Default()
	endpoint := fmt.Sprintf("http://%s:%d", host, port)

	r.POST("/scan", h.CreateScan)
	r.GET("/result/:job_id", h.GetResult)
	r.GET("/stats", h.GetStats)
	r.GET("/history", h.GetHistory)

	return r.Run(endpoint)
}
