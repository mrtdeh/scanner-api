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

	"github.com/mrtdeh/scanners-management/proto/scannerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ScannerResult struct {
	ScannerName string
	Result      string
}
type ScanStatus = scannerpb.ScanStatus

type FileScanResult struct {
	ScanFinished bool
	FileName     string
	Results      []ScannerResult
}

type Scan struct {
	ID       string // alternative for job ID
	ScanName string
	Status   ScanStatus
	Result   FileScanResult
}

type ScannerRepository interface {
	CreateScan(s Scan) (string, error)
	SendFile(scanID string, filePath string) error
	GetScanResultByID(scanID string) (*Scan, error)
	GetHistory() ([]Scan, error)
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

func (r *scannerRepo) CreateScan(s Scan) (string, error) {
	res, err := r.client.CreateScan(r.ctx, &scannerpb.CreateScanRequest{
		Name: s.ScanName,
	})
	if err != nil {
		return "", err
	}

	if !res.Success {
		return "", errors.New(res.Message)
	}

	return res.ScanId, nil
}

func (r *scannerRepo) SendFile(scanID string, filepath string) error {
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
		ScanId: scanID,
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

	//
	// s.Send(&scannerpb.SendFileRequest{
	// 	Data: &scannerpb.SendFileRequest_Done{
	// 		Done: true,
	// 	},
	// })

	return nil
}

func (r *scannerRepo) GetScanResultByID(scanID string) (*Scan, error) {
	res, err := r.client.GetScanResultByID(r.ctx, &scannerpb.GetScanResultByIDRequest{
		ScanId: scanID,
	})
	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, errors.New("scan not found : " + scanID)
	}

	scan := unmarshalScanProto(res.Scan)

	return &scan, nil
}

func (r *scannerRepo) GetHistory() ([]Scan, error) {
	res, err := r.client.GetHistory(r.ctx, &scannerpb.GetHistoryRequest{})
	if err != nil {
		return nil, err
	}

	var scans []Scan
	for _, s := range res.Scans {
		scan := unmarshalScanProto(s)
		scans = append(scans, scan)
	}

	return scans, nil
}
