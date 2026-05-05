package controller

import "github.com/gin-gonic/gin"

type ScanController struct{}

func NewScanController() *ScanController {
	return &ScanController{}
}

func (ctr *ScanController) CreateScan(c *gin.Context)
func (ctr *ScanController) GetResult(c *gin.Context)
func (ctr *ScanController) GetStats(c *gin.Context)
func (ctr *ScanController) GetHistory(c *gin.Context)
