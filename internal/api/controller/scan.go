package controller

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mrtdeh/scanners-management/internal/api/services"
)

type ScanController struct {
	ss *services.ScannerClientService
}

func NewScanController(ss *services.ScannerClientService) *ScanController {
	return &ScanController{ss}
}

func (ctr *ScanController) CreateScan(c *gin.Context) {
	// scanName := c.PostForm("name")
	// realFielpath := c.PostForm("realFilepath")

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("get form err: %s", err.Error()),
		})
		return
	}

	filename := filepath.Base(file.Filename)

	scanid := uuid.New().String()
	err = ctr.ss.CreateScan(services.ScanRequest{
		ScanID: scanid,
		FormFiles: []services.FormFile{
			services.FormFile{
				Name:     filename,
				TempPath: file.Filename,
				Size:     file.Size,
			},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("create scan err: %s", err.Error()),
		})
		return
	}

}
func (ctr *ScanController) GetResult(c *gin.Context)
func (ctr *ScanController) GetStats(c *gin.Context)
func (ctr *ScanController) GetHistory(c *gin.Context)
