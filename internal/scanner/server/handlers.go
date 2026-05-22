package grpcserver

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
	scannerpb "github.com/mrtdeh/scanners-management/internal/scanner/pb"
	"github.com/mrtdeh/scanners-management/internal/scanner/services"
)

type ScannerHandler struct {
	scannerpb.UnimplementedScannerServiceServer
	ss *services.ScannerServerService
}

func NewScannerServerHandler(scannerService *services.ScannerServerService) *ScannerHandler {
	return &ScannerHandler{
		ss: scannerService,
	}
}

func (sc *ScannerHandler) CreateScan(ctx context.Context, req *scannerpb.CreateScanRequest) (*scannerpb.CreateScanResponse, error) {
	if req.ScanId == "" {
		// if Scan ID is not prepare from API, Generate a new and return in response
		req.ScanId = uuid.NewString()
	}

	var files []services.FileInfoRequest

	// Convert DTO to business model
	for _, rf := range req.Files {

		f := services.FileInfoRequest{
			ID:     rf.FileId,
			Name:   rf.Name,
			Size:   rf.Size,
			Hash:   rf.Hash,
			Status: "pending",
		}

		files = append(files, f)
	}
	m := services.CreateScanRequest{
		ScanID: req.ScanId,
		Files:  files,
	}

	res, err := sc.ss.CreateScan(m)
	if err != nil {
		return nil, err
	}

	var states []*scannerpb.FileScanResultState
	for _, s := range res.FilesStates {
		states = append(states, &scannerpb.FileScanResultState{
			FileId: s.FileID,
			Cached: s.Cached,
		})
	}

	return &scannerpb.CreateScanResponse{
		Success: true,
		ScanId:  req.ScanId,
		States:  states,
	}, nil
}

func (sc *ScannerHandler) GetHistory(ctx context.Context, req *scannerpb.GetHistoryRequest) (*scannerpb.GetHistoryResponse, error) {
	scans, err := sc.ss.GetHistory()
	if err != nil {
		return nil, err
	}

	var res []*scannerpb.ScanResponse
	for _, s := range scans {
		r := convertScanToProtoResponse(s)
		res = append(res, r)
	}

	return &scannerpb.GetHistoryResponse{
		Scans: res,
	}, nil
}

func (sc *ScannerHandler) GetScanResultByID(ctx context.Context, req *scannerpb.GetScanResultByIDRequest) (*scannerpb.GetScanResultByIDResponse, error) {
	scan, err := sc.ss.GetResultByScanID(req.ScanId)
	if err != nil {
		return nil, err
	}

	r := convertScanToProtoResponse(*scan)

	return &scannerpb.GetScanResultByIDResponse{Scan: r}, nil
}

func (sc *ScannerHandler) SendFile(stream scannerpb.ScannerService_SendFileServer) error {
	var file *os.File
	var bytesWriten int64
	var fileSize int64
	var scanId, fileId string
	var finfo *services.FileInfo

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.New("error in receive file : " + err.Error())
		}

		if info := req.GetInfo(); info != nil {
			scanId = info.ScanId
			fileId = info.FileId

			scanIsExists := sc.ss.IsScanRequestExists(scanId)
			if !scanIsExists {
				return errors.New("scan request id not found: " + scanId)
			}

			finfo, err = sc.ss.GetFileInfo(scanId, fileId)
			if err != nil {
				return err
			}
			// name := path.Base(finfo.Name)
			fileSize = int64(finfo.Size)

			file, err = os.Create(finfo.FilePath)
			if err != nil {
				log.Println("error in open file : ", err.Error())
				return errors.New("error in open file : " + err.Error())
			}
		} else if chunk := req.GetChunkData(); chunk != nil {
			n, err := file.Write(chunk)
			if err != nil {
				log.Println("error in write to file : ", err.Error())
				return errors.New("error in write to file : " + err.Error())
			}
			bytesWriten += int64(n)
			if bytesWriten == fileSize {
				file.Close()
				stream.SendAndClose(&scannerpb.SendFileResponse{})
				break
			}
		}
	}

	// Add job for recieved file
	err := sc.ss.AddFileToScanQueue(finfo)
	if err != nil {
		return errors.New("error in add file to scan request job : " + err.Error())
	}

	return nil
}
