package services

import (
	"time"

	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
)

type CreateScanRequest struct {
	ScanID string
	Files  []FileInfoRequest
}

type FileInfoRequest struct {
	ID     string
	Status string
	Hash   string
	Name   string
	Size   int64
}

type FileInfo struct {
	ID       string
	ScanID   string
	Status   string
	Hash     string
	Name     string
	Size     int64
	ResultID string
	FilePath string
}

type CreateScanResponse struct {
	FilesStates []FileStateResponse `json:"files_states"`
}

type FileStateResponse struct {
	FileID string `json:"file_id"`
	Cached bool   `json:"cached"`
}

//==================================================================

type ScanRequestResultResponse struct {
	ScanID      string               `json:"scan_id"`
	Result      []ScanResultResponse `json:"result"`
	StartedAt   time.Time            `json:"started_at"`
	CompletedAt time.Time            `json:"completed_at"`
}

type ScanResultResponse struct {
	FileName string                  `json:"file_name"`
	Status   string                  `json:"status"`
	Results  []ScannerResultResponse `json:"results"`
}

type ScannerResultResponse struct {
	Engine      string    `json:"engine"`
	Status      string    `json:"status"`
	Output      string    `json:"output"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

func (s *ScannerServerService) convertRequestScanToResponse(rs domains.ScanRequest) ScanRequestResultResponse {
	var scanResults []ScanResultResponse
	for _, f := range rs.Files {

		var scannerResults []ScannerResultResponse

		if f.ResultID != "" {
			// If file has result id, then get scan result by this id and convert to response model
			sr, _ := s.resRepo.GetResultByID(f.ResultID)
			if sr != nil {
				for _, r := range sr.Results {
					scannerResults = append(scannerResults, ScannerResultResponse{
						Engine:      r.Engine,
						Status:      string(r.Status),
						Output:      r.Output,
						StartedAt:   r.StartedAt,
						CompletedAt: r.CompletedAt,
					})
				}
			}
		}

		if f.Status == "completed" && len(scannerResults) == 0 {
			// If file status is completed but no scanner result found, then set status to failed for this file
			f.Status = "failed"
		}

		scanResults = append(scanResults, ScanResultResponse{
			FileName: f.Name,
			Status:   f.Status,
			Results:  scannerResults,
		})

	}

	return ScanRequestResultResponse{
		ScanID:      rs.ScanID,
		Result:      scanResults,
		StartedAt:   rs.StartedAt,
		CompletedAt: rs.CompletedAt,
	}
}
