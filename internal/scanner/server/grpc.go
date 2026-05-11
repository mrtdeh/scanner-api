package grpcserver

import (
	"fmt"
	"log"
	"net"

	"github.com/mrtdeh/scanners-management/internal/scanner/controller"
	"github.com/mrtdeh/scanners-management/proto/scannerpb"
	"google.golang.org/grpc"
)

type Controller = controller.ScannerServerController

func RunGRPCServer(ctr *Controller, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	scannerpb.RegisterScannerServiceServer(grpcServer, ctr)

	log.Printf("gRPC server listening on %s", addr)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
