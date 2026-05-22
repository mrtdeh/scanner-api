package grpc_client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	scannerpb "github.com/mrtdeh/scanners-management/internal/scanner/pb"
	"google.golang.org/grpc"
)

// type Client interface {
// 	CreateScan(*scannerpb.CreateScanRequest) (*scannerpb.CreateScanResponse, error)
// 	SendFile(scanID, fileReqID string, filePath string) error
// 	GetScanResultByID(scanID string) (*scannerpb.ScanResponse, error)
// 	GetHistory() ([]*scannerpb.ScanResponse, error)
// }

type Client struct {
	client scannerpb.ScannerServiceClient
	ctx    context.Context
	cancel context.CancelFunc
}

func NewScannerServiceClient(conn *grpc.ClientConn) (*Client, error) {
	client := scannerpb.NewScannerServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (r *Client) CreateScan(ctx context.Context, s *scannerpb.CreateScanRequest) (*scannerpb.CreateScanResponse, error) {
	ctx, cancel := r.newCombinedContext(ctx)
	defer cancel()

	res, err := r.client.CreateScan(ctx, s)
	if err != nil {
		return nil, err
	}

	if !res.Success {
		return nil, errors.New(res.Message)
	}
	return res, nil
}

func (r *Client) SendFile(ctx context.Context, scanID string, fileReqID string, filepath string) error {
	ctx, cancel := r.newCombinedContext(ctx)
	defer cancel()

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
	s, err := r.client.SendFile(ctx)
	if err != nil {
		fmt.Printf("http api -> Error Sending Info: %v", err)
	}

	// send file info to server
	s.Send(&scannerpb.SendFileRequest{
		Data: &scannerpb.SendFileRequest_Info{
			Info: &scannerpb.FileHeader{
				ScanId: scanID,
				FileId: fileReqID,
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

	_, err = s.CloseAndRecv()
	if err != nil {
		return err
	}

	return nil
}

func (r *Client) GetScanResultByID(ctx context.Context, scanID string) (*scannerpb.ScanResponse, error) {
	ctx, cancel := r.newCombinedContext(ctx)
	defer cancel()

	res, err := r.client.GetScanResultByID(ctx, &scannerpb.GetScanResultByIDRequest{
		ScanId: scanID,
	})
	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, errors.New("scan not found : " + scanID)
	}
	return res.Scan, nil
}

func (r *Client) GetHistory(ctx context.Context) ([]*scannerpb.ScanResponse, error) {
	ctx, cancel := r.newCombinedContext(ctx)
	defer cancel()

	res, err := r.client.GetHistory(ctx, &scannerpb.GetHistoryRequest{})
	if err != nil {
		return nil, err
	}
	return res.Scans, nil
}

func (r *Client) Close() error {
	select {
	case <-r.ctx.Done():
		return errors.New("client already closed")
	default:
		r.cancel()
	}
	return nil
}
