package domains

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

type ScanResult struct {
	ID        string          `bson:"id" json:"id"`
	FileHash  string          `bson:"file_hash" json:"file_hash"`   // use for finding scan result by file content
	Results   []ScannerResult `bson:"results" json:"results"`       // scan result for multiple scanner
	CreatedAt time.Time       `bson:"created_at" json:"created_at"` // time of created the first scanner result
}

func (sr *ScanResult) IsExpired() bool {
	// Check if scan result is expired or not by created time and expired duration
	expiredDuration := 24 * time.Hour // Set expired duration for scan result (example: 24 hours)
	return time.Since(sr.CreatedAt) > expiredDuration
}

type ScannerResult struct {
	Engine      string        `bson:"engine" json:"engine"`             // engine name (must be unique)
	Status      ScannerStatus `bson:"status" json:"status"`             // pending , running , ...
	Output      string        `bson:"output" json:"output"`             // clean , malware , ...
	StartedAt   time.Time     `bson:"started_at" json:"started_at"`     // time of scanner starting for scan
	CompletedAt time.Time     `bson:"completed_at" json:"completed_at"` // time of scanner ended/failed scan
	Error       string        `bson:"error,omitempty" json:"error,omitempty"`
}
