package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/mrtdeh/scanners-management/internal/api/grpc_client"
	responses "github.com/mrtdeh/scanners-management/internal/models"
	scannerpb "github.com/mrtdeh/scanners-management/internal/scanner/pb"
)

type ScanService struct {
	client *grpc_client.Client
}

func NewScanService(client *grpc_client.Client) *ScanService {
	return &ScanService{client: client}
}

func (ss *ScanService) CreateScan(sr CreateScanRequest) (*CreateScanResponse, error) {
	var files []*scannerpb.FileRequest
	var filesMap map[string]FormFile
	filesMap = make(map[string]FormFile)
	for _, sf := range sr.FormFiles {

		fileId := uuid.NewString()
		filesMap[fileId] = sf
		// Calculate hash from file content for integration
		hash, err := calculateSHA256(sf.TempPath)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash :%v", err.Error())
		}

		files = append(files, &scannerpb.FileRequest{
			FileId: fileId,
			Hash:   hash,
			Name:   sf.Name,
			Size:   sf.Size,
		})

	}

	res, err := ss.client.CreateScan(context.Background(), &scannerpb.CreateScanRequest{
		ScanId: sr.ScanID,
		Files:  files,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create scan from scanner :%v", err.Error())
	}

	if !res.Success {
		return nil, fmt.Errorf("failed to create scan: %s", res.Message)
	}

	// TODO: send files need to monitoring ,check, retry
	go func() {
		for _, state := range res.States {
			if !state.Cached {
				ff := filesMap[state.FileId]
				err := ss.client.SendFile(context.Background(), res.ScanId, state.FileId, ff.TempPath)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	return &CreateScanResponse{ScanID: res.ScanId}, nil
}

func (ss *ScanService) GetHistory() (*GetHistoryResponse, error) {
	ctx := context.Background()
	results, err := ss.client.GetHistory(ctx)
	if err != nil {
		return nil, err
	}
	var historyResults []responses.ScanRequestResultResponse
	for _, r := range results {
		fmt.Println("r.status:", r.Status)
		res := responses.ConvertScanResultProtoToResponse(r)
		historyResults = append(historyResults, res)
	}
	return &GetHistoryResponse{Results: historyResults}, nil

}

func (ss *ScanService) GetResultByID(req GetResultByIDRequest) (*GetResultByIDResponse, error) {
	ctx := context.Background()
	result, err := ss.client.GetScanResultByID(ctx, req.ScanID)
	if err != nil {
		return nil, err
	}

	return &GetResultByIDResponse{
		Result: responses.ConvertScanResultProtoToResponse(result),
	}, nil
}

func (ss *ScanService) GetStats() (*GetStatsResponse, error) {
	// TODO : get job stats from scanner server
	return &GetStatsResponse{
		Running:   3,
		Completed: 10,
	}, nil
}
