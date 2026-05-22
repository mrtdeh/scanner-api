package services

import (
	"fmt"
	"log"
	"time"

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

func (s *ScannerServerService) checkFileResultIsExist(fileHash string) bool {
	// Implementation for checking if file is cached
	_, err := s.resRepo.GetByFileHash(fileHash)
	if err == nil {
		return true
	}
	return false
}

func (s *ScannerServerService) AddFileToScanQueue(scanRequestID, fileHash, filepath string) error {

	job := jobmng.NewJob(3, time.Second*5)
	job.AddTasks(
		s.HashGeneratorTask(fileHash, filepath),
		s.RandomSleeperTask(fileHash),
	)

	return s.jm.AddJob(job)
}

func (s *ScannerServerService) CreateScan(req CreateScanRequest) (*CreateScanResponse, error) {
	var files []domains.ScanRequestFile
	var states []FileState

	// Convert DTO to business model
	for _, rf := range req.Files {
		f := domains.ScanRequestFile{
			FileID: rf.UID,
			Name:   rf.Name,
			Size:   rf.Size,
			Hash:   rf.Hash,
			Status: "pending",
		}

		var state = FileState{FileID: rf.UID}
		isExist := s.checkFileResultIsExist(rf.Hash)
		if isExist {
			state.Cached = true
			f.Status = "cached"
		}
		states = append(states, state)

		files = append(files, f)
	}
	m := domains.ScanRequest{
		ScanID:    req.ScanID,
		Files:     files,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		Status:    "created",
	}

	// Store model to db
	err := s.reqRepo.Create(m)
	if err != nil {
		return nil, err
	}

	// Response to controller
	return &CreateScanResponse{
		FilesStates: states,
	}, nil
}

func (s *ScannerServerService) IsScanRequestExists(scanId string) bool {
	m, _ := s.reqRepo.GetByID(scanId)
	return m != nil
}

func (s *ScannerServerService) GetHistory() ([]ScanRequestResult, error) {
	scans, err := s.reqRepo.List(bson.M{})
	if err != nil {
		log.Println("error in get by file hash : ", err)
		return nil, err
	}

	var results []ScanRequestResult

	for _, sr := range scans {
		res := s.convertRequestScanToResponse(sr)
		results = append(results, res)
	}

	return results, nil
}

func (s *ScannerServerService) GetResultByScanID(scanId string) (*ScanRequestResult, error) {
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
		if f.FileID == fileId {
			return &FileInfo{
				UID:    f.FileID,
				Status: f.Status,
				Hash:   f.Hash,
				Name:   f.Name,
				Size:   f.Size,
			}, nil
		}
	}

	return nil, fmt.Errorf("scan request or file not found")
}
