package grpcserver

import (
	"time"

	responses "github.com/mrtdeh/scanners-management/internal/models"
	scannerpb "github.com/mrtdeh/scanners-management/internal/scanner/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func pbTime(t time.Time) *scannerpb.Time {
	return &scannerpb.Time{
		Timestamp: timestamppb.New(t),
	}
}

func convertScanToProtoResponse(s responses.ScanRequestResultResponse) *scannerpb.ScanResponse {
	var scanresultProto []*scannerpb.ScanResult

	for _, sr := range s.Result {
		var scannerResults []*scannerpb.ScannerResult
		for _, r := range sr.Results {
			scannerResults = append(scannerResults, &scannerpb.ScannerResult{
				Engine:      r.Engine,
				Output:      r.Output,
				Status:      r.Status,
				StartedAt:   pbTime(r.StartedAt),
				CompletedAt: pbTime(r.CompletedAt),
			})
		}
		scanresultProto = append(scanresultProto, &scannerpb.ScanResult{
			FileName:   sr.FileName,
			FileStatus: sr.FileStatus,
			Result:     scannerResults,
		})
	}

	return &scannerpb.ScanResponse{
		ScanId:      s.ScanID,
		Status:      s.Status,
		Results:     scanresultProto,
		StartedAt:   pbTime(s.StartedAt),
		CompletedAt: pbTime(s.CompletedAt),
	}
}
