package repositories

import "github.com/mrtdeh/scanners-management/proto/scannerpb"

func unmarshalScanProto(scanProto *scannerpb.Scan) Scan {
	var scannerResults []ScannerResult
	for _, rr := range scanProto.File.Result {
		scannerResults = append(scannerResults, ScannerResult{
			ScannerName: rr.ScannerName,
			Result:      rr.Result,
		})
	}

	var result = FileScanResult{
		ScanFinished: scanProto.File.ScanFinished,
		FileName:     scanProto.File.FileName,
		Results:      scannerResults,
	}

	return Scan{
		ID:       scanProto.Id,
		ScanName: scanProto.Name,
		Result:   result,
	}
}
