package services

import (
	responses "github.com/mrtdeh/scanners-management/internal/models"
	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
)

// ============================================= Request Models

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

// ============================================= Internal Models

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

// ============================================= Response Models

func (s *ScannerServerService) convertRequestScanToResponse(rs domains.ScanRequest) responses.ScanRequestResultResponse {
	var scanResults []responses.ScanResultResponse
	for _, f := range rs.Files {

		var scannerResults []responses.ScannerResultResponse

		if f.ResultID != "" {
			// If file has result id, then get scan result by this id and convert to response model
			sr, _ := s.resRepo.GetResultByID(f.ResultID)
			if sr != nil {
				for _, r := range sr.Results {
					scannerResults = append(scannerResults, responses.ScannerResultResponse{
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

		scanResults = append(scanResults, responses.ScanResultResponse{
			FileName: f.Name,
			Status:   f.Status,
			Results:  scannerResults,
		})

	}

	return responses.ScanRequestResultResponse{
		ScanID:      rs.ScanID,
		Result:      scanResults,
		StartedAt:   rs.StartedAt,
		CompletedAt: rs.CompletedAt,
	}
}
