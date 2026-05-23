package services

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	responses "github.com/mrtdeh/scanners-management/internal/models"
	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
	engines "github.com/mrtdeh/scanners-management/internal/scanner/engins"
	"github.com/mrtdeh/scanners-management/internal/scanner/repositories"

	jobmng "github.com/mrtdeh/scanners-management/pkg/job_manager"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ScannerServerService struct {
	jm      *jobmng.JobManager
	reqRepo repositories.ScanRequestRepository
	resRepo repositories.ScanResultRepository

	yeng engines.YaraScannerDockerEngine
	heng engines.HashGeneratorEngine
	reng engines.RandomSleeperEngine
}

func NewScannerServerService(
	jm *jobmng.JobManager,
	reqRepo repositories.ScanRequestRepository,
	resRepo repositories.ScanResultRepository,
	yeng engines.YaraScannerDockerEngine,
	heng engines.HashGeneratorEngine,
	reng engines.RandomSleeperEngine,

) *ScannerServerService {
	s := &ScannerServerService{jm, reqRepo, resRepo, yeng, heng, reng}
	return s
}

func (s *ScannerServerService) AddFileToScanQueue(finfo *FileInfo) error {

	job := jobmng.NewJob(jobmng.Config{
		MaxRetry: 3,
		Timeout:  time.Second * 10,
		Backoff:  time.Second * 2,
	})
	job.AddTasks(
		s.YaraScannerTask(finfo),
		s.HashGeneratorTask(finfo),
		s.RandomSleeperTask(finfo),
	)

	return s.jm.AddJob(job)
}

func (s *ScannerServerService) CreateScan(req CreateScanRequest) (*responses.CreateScanResponse, error) {
	var files []domains.ScanRequestFile
	var states []responses.FileStateResponse
	now := time.Now()

	for _, rf := range req.Files {
		// Initlize variables
		var scanResultID string
		var state = responses.FileStateResponse{FileID: rf.ID}
		f := domains.ScanRequestFile{
			ID:       rf.ID,
			Name:     rf.Name,
			Size:     rf.Size,
			Hash:     rf.Hash,
			Status:   "pending",
			ResultID: scanResultID,
			FilePath: fmt.Sprintf("/tmp/%s_%s", rf.ID, rf.Name),
		}

		// Check scan result for file is exist or not
		latest, _ := s.resRepo.GetLatestResultByFileHash(rf.Hash)
		isExpired := latest != nil && latest.IsExpired()
		if latest == nil || isExpired {
			// If scan result is not exist or expired, then create new scan result for this file
			scanResultID = uuid.NewString()
			err := s.resRepo.Create(domains.ScanResult{
				ID:        scanResultID,
				FileHash:  rf.Hash,
				Results:   []domains.ScannerResult{},
				CreatedAt: now,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create scan result: %w", err)
			}
		} else {
			// If scan result is exist and not expired, then use this result for response and skip scan for this file
			state.Cached = true
			f.Status = "cached"
			scanResultID = latest.ID
		}

		f.ResultID = scanResultID

		states = append(states, state)
		files = append(files, f)

	}
	// Convert requested scan info to business model
	m := domains.ScanRequest{
		ScanID:    req.ScanID,
		Files:     files,
		StartedAt: now,
		CreatedAt: now,
		Status:    "created",
	}

	// Store model to db
	err := s.reqRepo.Create(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan request: %w", err)
	}

	// Response to controller
	return &responses.CreateScanResponse{
		FilesStates: states,
	}, nil
}

func (s *ScannerServerService) IsScanRequestExists(scanId string) bool {
	m, _ := s.reqRepo.GetByID(scanId)
	return m != nil
}

func (s *ScannerServerService) GetHistory() ([]responses.ScanRequestResultResponse, error) {
	scans, err := s.reqRepo.List(bson.M{})
	if err != nil {
		log.Println("error in get by file hash : ", err)
		return nil, err
	}

	var results []responses.ScanRequestResultResponse

	for _, sr := range scans {
		res := s.convertRequestScanToResponse(sr)
		results = append(results, res)
	}

	return results, nil
}

func (s *ScannerServerService) GetResultByScanID(scanId string) (*responses.ScanRequestResultResponse, error) {
	result, err := s.reqRepo.GetByID(scanId)
	if err != nil {
		log.Println("error in get scan : ", err)
		return nil, err
	}

	res := s.convertRequestScanToResponse(*result)

	return &res, nil
}

func (s *ScannerServerService) GetFileInfo(scanId, fileId string) (*FileInfo, error) {
	result, err := s.reqRepo.GetByID(scanId)
	if err != nil {
		log.Println("error in get scan : ", err)
		return nil, err
	}

	for _, f := range result.Files {
		if f.ID == fileId {
			return &FileInfo{
				ID:       f.ID,
				ScanID:   scanId,
				Status:   f.Status,
				Hash:     f.Hash,
				Name:     f.Name,
				Size:     f.Size,
				ResultID: f.ResultID,
				FilePath: f.FilePath,
			}, nil
		}
	}

	return nil, fmt.Errorf("scan request or file not found")
}
