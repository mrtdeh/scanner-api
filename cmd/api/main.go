package main

import (
	"fmt"
	"log"

	apiconfig "github.com/mrtdeh/scanners-management/internal/api/config"
	httpserver "github.com/mrtdeh/scanners-management/internal/api/server"
)

func main() {
	fmt.Println("Start API Service")

	cnf, err := apiconfig.LoadConfiguration()
	if err != nil {
		log.Fatal("error in parse configuration : ", err)
	}

	log.Fatal(
		"error in running http api server : ",
		httpserver.Run(cnf.HttpHost, cnf.HttpPort),
	)
}
