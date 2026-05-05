package httpserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mrtdeh/scanners-management/internal/api/controller"
)

func Run(host string, port uint) error {
	r := gin.Default()
	endpoint := fmt.Sprintf("http://%s:%d", host, port)

	ctr := controller.NewScanController()

	r.POST("/scan", ctr.CreateScan)
	r.GET("/result/:job_id", ctr.GetResult)
	r.GET("/stats", ctr.GetStats)
	r.GET("/history", ctr.GetHistory)

	return r.Run(endpoint)
}
