package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Run(h *ScanHandler, host string, port uint) error {
	r := gin.Default()
	endpoint := fmt.Sprintf("%s:%d", host, port)

	r.POST("/scan", h.CreateScan)
	r.GET("/result/:scan_id", h.GetResult)
	r.GET("/stats", h.GetStats)
	r.GET("/history", h.GetHistory)

	fmt.Println("try to listening http server API on ", endpoint)
	return r.Run(endpoint)
}
