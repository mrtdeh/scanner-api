package services

import (
	"log"
	"time"

	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
)

type CreateScanRequest struct {
	ScanID string     `json:"scan_id"`
	Files  []FileInfo `json:"files"`
}

type FileInfo struct {
	UID    string `json:"uid"`
	Status string `json:"status"`
	Hash   string `json:"hash"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}

type CreateScanResponse struct {
	FilesStates []FileState `json:"files_states"`
}

type FileState struct {
	FileID string `json:"file_id"`
	Cached bool   `json:"cached"`
}

//==================================================================

type ScanRequestResult struct {
	ScanID      string       `json:"scan_id"`
	Result      []ScanResult `json:"result"`
	StartedAt   time.Time    `json:"started_at"`
	CompletedAt time.Time    `json:"completed_at"`
}

type ScanResult struct {
	FileName string          `json:"file_name"`
	Results  []ScannerResult `json:"results"`
}

type ScannerResult struct {
	Engine      string    `json:"engine"`
	Status      string    `json:"status"`
	Output      string    `json:"output"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

func (s *ScannerServerService) convertRequestScanToResponse(rs domains.ScanRequest) ScanRequestResult {
	var scanResults []ScanResult
	for _, f := range rs.Files {
		sr, err := s.resRepo.GetByFileHash(f.Hash)
		if err != nil {
			log.Println("error in get by file hash : ", err)
			continue
		}
		var scannerResults []ScannerResult
		for _, r := range sr.Results {
			scannerResults = append(scannerResults, ScannerResult{
				Engine:      r.Engine,
				Status:      string(r.Status),
				Output:      r.Output,
				StartedAt:   r.StartedAt,
				CompletedAt: r.CompletedAt,
			})
		}

		scanResults = append(scanResults, ScanResult{
			FileName: f.Name,
			Results:  scannerResults,
		})

	}

	return ScanRequestResult{
		ScanID:      rs.ScanID,
		Result:      scanResults,
		StartedAt:   rs.StartedAt,
		CompletedAt: rs.CompletedAt,
	}
}
