package repositories

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/mrtdeh/scanners-management/internal/model"
	"github.com/mrtdeh/scanners-management/proto/scannerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ScannerRepository interface {
	CreateScan(s model.ScanRequest) (string, error)
	SendFile(scanID, fileReqID string, filePath string) error
	GetScanResultByID(scanID string) (*model.ScanResponse, error)
	GetHistory() ([]model.ScanResponse, error)
}

type scannerRepo struct {
	client scannerpb.ScannerServiceClient
	ctx    context.Context
	cancel context.CancelFunc
}

func NewScannerRepository(addr string) (ScannerRepository, error) {
	// Connect to scanner service via gRPC
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := scannerpb.NewScannerServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())

	return &scannerRepo{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (r *scannerRepo) CreateScan(s model.ScanRequest) (string, error) {
	var files []*scannerpb.RequestFile
	for _, f := range s.Files {
		files = append(files, &scannerpb.RequestFile{
			Name: f.Name,
			Size: f.Size,
		})
	}
	res, err := r.client.CreateScan(r.ctx, &scannerpb.CreateScanRequest{
		ScanId: s.ScanID,
		Files:  files,
	})
	if err != nil {
		return "", err
	}

	if !res.Success {
		return "", errors.New(res.Message)
	}

	return res.ScanId, nil
}

func (r *scannerRepo) SendFile(scanID string, fileReqID string, filepath string) error {
	var err error
	var frec *os.File
	var stat fs.FileInfo
	// open selected file
	frec, err = os.Open(filepath)
	if err != nil {
		log.Fatal("cannot open txt file: ", err)
	}
	defer frec.Close()

	stat, err = frec.Stat()
	if err != nil {
		log.Fatal("file read failed : ", err)
	}
	// failed when is directory
	if stat.IsDir() {
		return fmt.Errorf("you select a dir, but must select a file")
	}

	reader := bufio.NewReader(frec)
	buffer := make([]byte, 1024)

	// connnect to RPC function
	s, err := r.client.SendFile(r.ctx)
	if err != nil {
		fmt.Printf("http api -> Error Sending Info: %v", err)
	}

	// send file info to server
	s.Send(&scannerpb.SendFileRequest{
		ScanId:    scanID,
		RequestId: fileReqID,
		Data: &scannerpb.SendFileRequest_Info{
			Info: &scannerpb.FileInfo{
				Name: stat.Name(),
				Size: uint32(stat.Size()),
			},
		},
	})

	// send file chunk data to server
	for {
		// read file buffer
		n, err := reader.Read(buffer)
		// check if end of file
		if err == io.EOF {
			break
		}
		// check error
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		s.Send(&scannerpb.SendFileRequest{
			Data: &scannerpb.SendFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		})
	}

	return nil
}

func (r *scannerRepo) GetScanResultByID(scanID string) (*model.ScanResponse, error) {
	res, err := r.client.GetScanResultByID(r.ctx, &scannerpb.GetScanResultByIDRequest{
		ScanId: scanID,
	})
	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, errors.New("scan not found : " + scanID)
	}

	scan := unmarshalScanResponseProto(res.Scan)

	return &scan, nil
}

func (r *scannerRepo) GetHistory() ([]model.ScanResponse, error) {
	res, err := r.client.GetHistory(r.ctx, &scannerpb.GetHistoryRequest{})
	if err != nil {
		return nil, err
	}

	var scans []model.ScanResponse
	for _, s := range res.Scans {
		scan := unmarshalScanResponseProto(s)
		scans = append(scans, scan)
	}

	return scans, nil
}
