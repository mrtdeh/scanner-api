package services

import "time"

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
	Results []ScanView `json:"results"`
}

type GetResultByIDRequest struct {
	ScanID string `json:"scan_id"`
}

type GetResultByIDResponse struct {
	Result ScanView `json:"result"`
}

type GetStatsResponse struct {
	Running   int `json:"running"`
	Completed int `json:"completed"`
}

type ScanView struct {
	ScanID      string           `json:"scan_id"`
	Result      []ScanResultView `json:"result"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt time.Time        `json:"completed_at"`
}

type ScanResultView struct {
	FileName string              `json:"file_name"`
	Results  []ScannerResultView `json:"results"`
}

type ScannerResultView struct {
	Engine      string    `json:"engine"`
	Status      string    `json:"status"`
	Output      string    `json:"output"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Error       string    `json:"error,omitempty"`
}
