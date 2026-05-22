package grpcserver

import (
	"fmt"
	"log"
	"net"

	scannerpb "github.com/mrtdeh/scanners-management/internal/scanner/pb"

	"google.golang.org/grpc"
)

func RunGRPCServer(h *ScannerHandler, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	scannerpb.RegisterScannerServiceServer(grpcServer, h)

	log.Printf("gRPC server listening on %s", addr)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
