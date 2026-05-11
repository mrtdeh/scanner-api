package main

import (
	"fmt"
	"log"

	apiconfig "github.com/mrtdeh/scanners-management/internal/api/config"
	"github.com/mrtdeh/scanners-management/internal/api/controller"
	"github.com/mrtdeh/scanners-management/internal/api/repositories"
	httpserver "github.com/mrtdeh/scanners-management/internal/api/server"
	"github.com/mrtdeh/scanners-management/internal/api/services"
)

func main() {
	fmt.Println("Start API Service")

	cnf, err := apiconfig.LoadConfiguration()
	if err != nil {
		log.Fatal("error in parse configuration : ", err)
	}

	repo, err := repositories.NewScannerRepository(cnf.ScannerAddress)
	if err != nil {
		log.Fatal(err)
	}
	serv := services.NewScannerClientService(repo)
	ctr := controller.NewScanController(serv)

	log.Fatal(
		"error in running http api server : ",
		httpserver.Run(ctr, cnf.HttpHost, cnf.HttpPort),
	)
}
