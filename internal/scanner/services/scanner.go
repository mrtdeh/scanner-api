package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/mrtdeh/scanners-management/internal/model"
	engines "github.com/mrtdeh/scanners-management/internal/scanner/engins"
	"github.com/mrtdeh/scanners-management/internal/scanner/repositories"
	jobmng "github.com/mrtdeh/scanners-management/pkg/job_manager"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ScannerServerService struct {
	jm      *jobmng.JobManager
	ReqRepo repositories.ScanRequestRepository
	ResRepo repositories.ScanResultRepository

	yeng engines.YaraScannerDockerEngine
	heng engines.HashGeneratorEngine
	reng engines.RandomSleeperEngine
}

func NewScannerServerService(
	jm *jobmng.JobManager,
	ReqRepo repositories.ScanRequestRepository,
	ResRepo repositories.ScanResultRepository,
	yeng engines.YaraScannerDockerEngine,
	heng engines.HashGeneratorEngine,
	reng engines.RandomSleeperEngine,

) *ScannerServerService {
	s := &ScannerServerService{jm, ReqRepo, ResRepo, yeng, heng, reng}
	return s
}

func (s *ScannerServerService) AddFileToScanQueue(scanRequestID, fileRequestID, filepath string) error {
	d, _ := json.Marshal(model.ScanJob{
		ScanID:   scanRequestID,
		FileID:   fileRequestID,
		FilePath: filepath,
	})

	return s.jm.AddJob(jobmng.Job{
		ID:   fileRequestID,
		Data: d,
	})
}

func (s *ScannerServerService) processScanRequest(j *jobmng.Job) error {
	// TODO : process the scan job...
	start := time.Now()
	var results []model.ScannerResult
	var wg sync.WaitGroup
	var sjob model.ScanJob
	var l sync.Mutex
	var updateResults = func(r model.ScannerResult) {
		l.Lock()
		defer l.Unlock()
		results = append(results, r)
	}

	if err := json.Unmarshal(j.Data, &sjob); err != nil {
		return err
	}

	wg.Add(3)

	go func() {
		defer wg.Done()
		start := time.Now()
		ctx := context.Background()
		yres, err := s.yeng.Scan(ctx, sjob.FilePath)
		if err != nil {
			updateResults(model.ScannerResult{
				StartedAt:   start,
				Engine:      "yara_scanner",
				Status:      "failed",
				Error:       err.Error(),
				CompletedAt: time.Now(),
			})

			return
		}
		updateResults(model.ScannerResult{
			Engine:      "yara_scanner",
			Output:      yres.RawOutput,
			Status:      "completed",
			StartedAt:   start,
			CompletedAt: time.Now(),
		})
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		hash := s.heng.GenerateHash(sjob.FilePath)
		updateResults(model.ScannerResult{
			Engine:      "hash_generator",
			Output:      hash,
			Status:      "completed",
			StartedAt:   start,
			CompletedAt: time.Now(),
		})
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		dur := s.reng.RandomSleep()
		updateResults(model.ScannerResult{
			Engine:      "random_sleeper",
			Output:      dur.String(),
			Status:      "completed",
			StartedAt:   start,
			CompletedAt: time.Now(),
		})
	}()

	wg.Wait()
	s.ResRepo.Create(model.ScanResult{
		ScanFinished: true,
		FileName:     sjob.FilePath,
		CreatedAt:    start,
		FinishedAt:   time.Now(),
		Results:      results,
	})
	return nil
}

func (s *ScannerServerService) CreateScan(scanId string, files []model.ScanRequestFile) error {
	return s.ReqRepo.Create(model.ScanRequest{
		ScanID:    scanId,
		Files:     files,
		StartedAt: time.Now(),
	})
}

func (s *ScannerServerService) IsScanRequestExists(scanId string) bool {
	m, _ := s.ReqRepo.GetByID(scanId)
	return m != nil
}

func (s *ScannerServerService) SetFileAsReceived(scanId string, fileRequestId string) error {
	m, err := s.ReqRepo.GetByID(scanId)
	if err != nil {
		return err
	}
	var found bool
	var f model.ScanRequestFile
	for _, f = range m.Files {
		if f.RequestID == fileRequestId {
			found = true
			break
		}
	}
	if !found {
		return errors.New("request file id not found in scan request : " + fileRequestId)
	}

	f.Received = true
	if err := s.ReqRepo.UpdateFile(scanId, f); err != nil {
		return err
	}
	return nil
}

func (s *ScannerServerService) GetHistory() ([]model.ScanResponse, error) {
	var res []model.ScanResponse
	results, err := s.ReqRepo.List(bson.M{})
	if err != nil {
		log.Println("error in get by file hash : ", err)
		return nil, err
	}

	var scanResults []model.ScanResult
	for _, r := range results {
		for _, f := range r.Files {
			sr, err := s.ResRepo.GetByFileHash(f.Hash)
			if err != nil {
				log.Println("error in get by file hash : ", err)
				continue
			}

			scanResults = append(scanResults, *sr)
		}
		res = append(res, model.ScanResponse{
			ScanID:      r.ScanID,
			Result:      scanResults,
			StartedAt:   r.StartedAt,
			CompletedAt: r.CompletedAt,
		})

	}

	return res, nil
}

func (s *ScannerServerService) GetResultByScanID(scanId string) (*model.ScanResponse, error) {
	result, err := s.ReqRepo.GetByID(scanId)
	if err != nil {
		log.Println("error in get scan : ", err)
		return nil, err
	}

	var scanResults []model.ScanResult

	for _, f := range result.Files {
		sr, err := s.ResRepo.GetByFileHash(f.Hash)
		if err != nil {
			log.Println("error in get by file hash : ", err)
			continue
		}

		scanResults = append(scanResults, *sr)

	}

	return &model.ScanResponse{
		ScanID:      result.ScanID,
		Result:      scanResults,
		StartedAt:   result.StartedAt,
		CompletedAt: result.CompletedAt,
	}, nil

}
