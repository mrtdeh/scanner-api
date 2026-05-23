package responses

import (
	"time"

	scannerpb "github.com/mrtdeh/scanners-management/internal/scanner/pb"
)

type CreateScanResponse struct {
	FilesStates []FileStateResponse `json:"files_states"`
}

type FileStateResponse struct {
	FileID string `json:"file_id"`
	Cached bool   `json:"cached"`
}

type ScanRequestResultResponse struct {
	ScanID      string               `json:"scan_id"`
	Status      string               `json:"status"`
	Result      []ScanResultResponse `json:"result"`
	StartedAt   time.Time            `json:"started_at"`
	CompletedAt time.Time            `json:"completed_at"`
}

type ScanResultResponse struct {
	FileName   string                  `json:"file_name"`
	FileStatus string                  `json:"status"`
	Results    []ScannerResultResponse `json:"results"`
}

type ScannerResultResponse struct {
	Engine      string    `json:"engine"`
	Status      string    `json:"status"`
	Output      string    `json:"output"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// ====================================== Converters

func ConvertScanResultProtoToResponse(s *scannerpb.ScanResponse) ScanRequestResultResponse {
	var scanResults []ScanResultResponse

	for _, sr := range s.Results {
		var scannerResults []ScannerResultResponse
		for _, r := range sr.Result {
			scannerResults = append(scannerResults, ScannerResultResponse{
				Engine:      r.Engine,
				Status:      r.Status,
				Output:      r.Output,
				StartedAt:   r.StartedAt.Timestamp.AsTime(),
				CompletedAt: r.CompletedAt.Timestamp.AsTime(),
			})
		}
		scanResults = append(scanResults, ScanResultResponse{
			FileName:   sr.FileName,
			FileStatus: sr.FileStatus, // Status is not included in proto response, so we can set it to empty string or you can add it to proto response if needed
			Results:    scannerResults,
		})
	}

	return ScanRequestResultResponse{
		ScanID:      s.ScanId,
		Status:      s.Status,
		Result:      scanResults,
		StartedAt:   s.StartedAt.Timestamp.AsTime(),
		CompletedAt: s.CompletedAt.Timestamp.AsTime(),
	}
}
