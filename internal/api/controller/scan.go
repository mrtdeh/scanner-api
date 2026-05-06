package controller

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/mrtdeh/scanners-management/internal/api/repositories"
)

type ScanController struct {
	ss repositories.ScannerRepository
}

func NewScanController() *ScanController {
	return &ScanController{}
}

type ScanRequest struct {
	Name        string         `form:"name"`
	Description string         `form:"description"`
	File        multipart.File `form:"file"`
}

func (ctr *ScanController) CreateScan(c *gin.Context) {
	var req ScanRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	defer req.File.Close()

}
func (ctr *ScanController) GetResult(c *gin.Context)
func (ctr *ScanController) GetStats(c *gin.Context)
func (ctr *ScanController) GetHistory(c *gin.Context)
