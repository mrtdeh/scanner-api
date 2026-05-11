package scanner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mrtdeh/scanners-management/internal/scanner/config"
	"github.com/mrtdeh/scanners-management/internal/scanner/controller"
	engines "github.com/mrtdeh/scanners-management/internal/scanner/engins"
	"github.com/mrtdeh/scanners-management/internal/scanner/repositories"
	grpcserver "github.com/mrtdeh/scanners-management/internal/scanner/server"
	"github.com/mrtdeh/scanners-management/internal/scanner/services"
	jobmng "github.com/mrtdeh/scanners-management/pkg/job_manager"
	mongodb "github.com/mrtdeh/scanners-management/pkg/mongo"
)

func main() {
	fmt.Println("Start Scanner Service")

	cnf, err := config.LoadConfiguration()
	if err != nil {
		log.Fatal("error in parse configuration : ", err)
	}

	db, err := mongodb.ConnectMongoDB(mongodb.Config{
		Host:     cnf.MongoHost,
		Port:     cnf.MongoPort,
		DBName:   cnf.MongoDatabase,
		Username: cnf.MongoUsername,
		Password: cnf.MongoPassword,
	})

	ctx := context.Background()

	jm, err := jobmng.NewJobManager(ctx, jobmng.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	scanRequestRepo := repositories.NewScanResultRepository(db)
	scanResultRepo := repositories.NewScanRequestRepository(db)
	yeng, err := engines.NewYaraScannerDockerEngine(&engines.DefaultCommandExcutor{}, cnf.YaraImage, cnf.YaraRulesPath)
	if err != nil {
		log.Fatal(err)
	}

	heng := engines.NewHashGeneratorEngine("aaa")
	reng := engines.NewRandomSleeperEngine(time.Second, time.Second*10)
	scanService := services.NewScannerServerService(
		jm, scanResultRepo, scanRequestRepo,
		yeng, heng, reng)
	ctr := controller.NewScannerServerController(scanService)

	log.Fatal(
		"error in run gRPC server :",
		grpcserver.RunGRPCServer(ctr, fmt.Sprintf("%s:%d", cnf.GRPCHost, cnf.GRPCPort)),
	)
}
