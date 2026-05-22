package services

import responses "github.com/mrtdeh/scanners-management/internal/models"

type FormFile struct {
	Name     string
	TempPath string
	Size     int64
}
type CreateScanRequest struct {
	ScanID    string
	FormFiles []FormFile
}

type CreateScanResponse struct {
	ScanID string `json:"scan_id"`
}

type GetHistoryRequest struct{}

type GetHistoryResponse struct {
	Results []responses.ScanRequestResultResponse `json:"results"`
}

type GetResultByIDRequest struct {
	ScanID string `json:"scan_id"`
}

type GetResultByIDResponse struct {
	Result responses.ScanRequestResultResponse `json:"result"`
}

type GetStatsResponse struct {
	Running   int `json:"running"`
	Completed int `json:"completed"`
}
