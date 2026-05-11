package model

import (
	"time"
)

type ScannerStatus string

const (
	ScannerStatusPending   ScannerStatus = "pending"
	ScannerStatusRunning   ScannerStatus = "running"
	ScannerStatusCompleted ScannerStatus = "completed"
	ScannerStatusFailed    ScannerStatus = "failed"
	ScannerStatusTimeout   ScannerStatus = "timeout"
)

type ScanJob struct {
	ScanID   string
	FileID   string
	FilePath string
}

type ScannerResult struct {
	Engine      string        `bson:"engine" json:"engine"`             // engine name (must be unique)
	Status      ScannerStatus `bson:"status" json:"status"`             // pending , running , ...
	Output      string        `bson:"output" json:"output"`             // clean , malware , ...
	StartedAt   time.Time     `bson:"started_at" json:"started_at"`     // time of scanner starting for scan
	CompletedAt time.Time     `bson:"completed_at" json:"completed_at"` // time of scanner ended/failed scan
	Error       string        `bson:"error,omitempty" json:"error,omitempty"`
}

type ScanResult struct {
	ScanFinished bool            `bson:"scan_finished" json:"scan_finished"`
	FileName     string          `bson:"file_name" json:"file_name"`     // only for view
	FileHash     string          `bson:"file_hash" json:"file_hash"`     // use for finding scan result by file content
	Results      []ScannerResult `bson:"results" json:"results"`         // scan result for multiple scanner
	CreatedAt    time.Time       `bson:"created_at" json:"created_at"`   // time of created the first scanner result
	FinishedAt   time.Time       `bson:"finished_at" json:"finished_at"` // time of finished the last scanner result
}

//=========================  Scan Request from API ==================================

type ScanRequest struct {
	ScanID      string            `bson:"scan_id" json:"scan_id"`           // id of each scan request
	Files       []ScanRequestFile `bson:"files" json:"files"`               // contains files info in request that must be scan
	StartedAt   time.Time         `bson:"started_at" json:"started_at"`     // time of start scan for first file
	CompletedAt time.Time         `bson:"completed_at" json:"completed_at"` // time of end/failed scan for last file
	Error       string            `bson:"error,omitempty" json:"error,omitempty"`
}

type ScanRequestFile struct {
	RequestID string `bson:"request_id" json:"request_id"` // use for make sure uploded
	Name      string `bson:"name" json:"name"`
	Size      int64  `bson:"size" json:"size"`
	Hash      string `bson:"hash" json:"hash"` // use for file scan result discovery
	Received  bool   `bson:"received" json:"received"`
}

//=========================  Scan Response to Request ==================================

// this struct used only for API response to client
type ScanResponse struct {
	ScanID      string       `json:"scan_id"`
	Result      []ScanResult `json:"result"`
	StartedAt   time.Time    `json:"started_at"`   // time of start scan for first file
	CompletedAt time.Time    `json:"completed_at"` // time of end/failed scan for last file
}
