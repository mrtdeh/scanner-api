package httpserver

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mrtdeh/scanners-management/internal/api/services"
)

type ScanHandler struct {
	service *services.ScanService
}

func NewScanHandler(service *services.ScanService) *ScanHandler {
	return &ScanHandler{service}
}

func (h *ScanHandler) CreateScan(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("get form err: %s", err.Error()),
		})
		return
	}

	// Save form files to safe directory
	tmpDir := fmt.Sprintf("/tmp/%s", file.Filename)
	err = c.SaveUploadedFile(file, tmpDir, 0655)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("save file err: %s", err.Error()),
		})
		return
	}

	scanid := uuid.New().String()
	res, err := h.service.CreateScan(services.CreateScanRequest{
		ScanID: scanid,
		FormFiles: []services.FormFile{
			{
				Name:     file.Filename,
				TempPath: tmpDir,
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

	c.JSON(http.StatusCreated, res)
}

func (h *ScanHandler) GetResult(c *gin.Context) {
	scnaId := c.Param("scan_id")
	result, err := h.service.GetResultByID(services.GetResultByIDRequest{
		ScanID: scnaId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("get result err: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ScanHandler) GetHistory(c *gin.Context) {
	history, err := h.service.GetHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("get history err: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *ScanHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("get stats err: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, stats)
}
