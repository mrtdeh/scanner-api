package services

import (
	"github.com/google/uuid"
	"github.com/mrtdeh/scanners-management/internal/api/repositories"
	"github.com/mrtdeh/scanners-management/internal/model"
)

type ScannerClientService struct {
	repo repositories.ScannerRepository
}

func NewScannerClientService(sr repositories.ScannerRepository) *ScannerClientService {
	return &ScannerClientService{sr}
}

type FormFile struct {
	Name     string
	TempPath string
	Size     int64
}
type ScanRequest struct {
	ScanID    string
	FormFiles []FormFile
}

func (ss *ScannerClientService) CreateScan(sr ScanRequest) error {
	var files []model.ScanRequestFile
	var startUpload chan struct{}
	defer close(startUpload)

	for _, sf := range sr.FormFiles {
		fileReq := uuid.NewString()
		files = append(files, model.ScanRequestFile{
			Name:      sf.Name,
			Size:      sf.Size,
			RequestID: fileReq,
		})

		go func() {
			<-startUpload
			ss.repo.SendFile(sr.ScanID, fileReq, sf.TempPath)
		}()
	}

	_, err := ss.repo.CreateScan(model.ScanRequest{
		ScanID: sr.ScanID,
		Files:  files,
	})
	if err != nil {
		return err
	}

	return nil
}
