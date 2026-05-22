package main

import (
	"fmt"
	"log"

	apiconfig "github.com/mrtdeh/scanners-management/internal/api/config"
	grpc_client "github.com/mrtdeh/scanners-management/internal/api/grpc_client"

	httpserver "github.com/mrtdeh/scanners-management/internal/api/server"
	"github.com/mrtdeh/scanners-management/internal/api/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("Start API Service")

	cnf, err := apiconfig.LoadConfiguration()
	if err != nil {
		log.Fatal("error in parse configuration : ", err)
	}

	conn, err := grpc.NewClient(cnf.ScannerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	scannerClient, err := grpc_client.NewScannerServiceClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer scannerClient.Close()

	serv := services.NewScanService(scannerClient)
	h := httpserver.NewScanHandler(serv)

	log.Fatal(
		"error in running http api server : ",
		httpserver.Run(h, cnf.HttpHost, cnf.HttpPort),
	)
}
