package grpcserver

import (
	scannerpb "github.com/mrtdeh/scanners-management/internal/scanner/pb"
	"github.com/mrtdeh/scanners-management/internal/scanner/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertScanToProtoResponse(s services.ScanRequestResultResponse) *scannerpb.ScanResponse {
	var scanresultProto []*scannerpb.ScanResult

	for _, sr := range s.Result {
		var scannerResults []*scannerpb.ScannerResult
		for _, r := range sr.Results {
			scannerResults = append(scannerResults, &scannerpb.ScannerResult{
				Engine: r.Engine,
				Output: r.Output,
				Status: r.Status,
				StartedAt: &scannerpb.Time{
					Timestamp: timestamppb.New(r.StartedAt),
				},
				CompletedAt: &scannerpb.Time{
					Timestamp: timestamppb.New(r.CompletedAt),
				},
			})
		}
		scanresultProto = append(scanresultProto, &scannerpb.ScanResult{
			FileName: sr.FileName,
			Result:   scannerResults,
		})
	}

	return &scannerpb.ScanResponse{
		ScanId:  s.ScanID,
		Results: scanresultProto,
		StartedAt: &scannerpb.Time{
			Timestamp: timestamppb.New(s.StartedAt),
		},
		CompletedAt: &scannerpb.Time{
			Timestamp: timestamppb.New(s.CompletedAt),
		},
	}
}
