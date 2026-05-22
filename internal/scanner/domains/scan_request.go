package domains

import (
	"time"
)

// Store scan requests with requested files
type ScanRequest struct {
	ScanID      string            `bson:"scan_id" json:"scan_id"`           // id of each scan request
	Status      string            `bson:"status" json:"status"`             // created , started , completed , failed
	Files       []ScanRequestFile `bson:"files" json:"files"`               // contains files info in request that must be scan
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`     // time of creation of scan request
	StartedAt   time.Time         `bson:"started_at" json:"started_at"`     // time of start scan for first file
	CompletedAt time.Time         `bson:"completed_at" json:"completed_at"` // time of end/failed scan for last file
	Error       string            `bson:"error,omitempty" json:"error,omitempty"`
}

// Embedded model for store files in scan request
type ScanRequestFile struct {
	ID       string `bson:"id" json:"id"`
	Status   string `bson:"status" json:"status"`                           // pending , processing , completed , failed , cached
	ResultID string `bson:"result_id,omitempty" json:"result_id,omitempty"` // Id of scan result that related to this file.
	Hash     string `bson:"hash" json:"hash"`                               // use for file scan result discovery
	Name     string `bson:"name" json:"name"`
	Size     int64  `bson:"size" json:"size"`
	FilePath string `bson:"file_path" json:"file_path"` // local path of file that used for scanning
}
