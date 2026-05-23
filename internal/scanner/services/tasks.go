package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
	jobmng "github.com/mrtdeh/scanners-management/pkg/job_manager"
)

func (s *ScannerServerService) YaraScannerTask(finfo *FileInfo) jobmng.TaskFunc {
	return func(t *jobmng.Task) error {
		now := time.Now()
		r := domains.ScannerResult{
			Engine:    "yara_scanner",
			Status:    "processing",
			Output:    "",
			StartedAt: now,
		}

		// Store the initial scan result with "processing" status
		err := s.resRepo.PutScanResultByID(finfo.ResultID, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		ctx := t.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
		defer cancel()

		// Run the main task function
		res, err := s.yeng.Scan(ctx, finfo.FilePath)
		if err != nil {
			return fmt.Errorf("failed to run yara scanner: %w", err)
		}

		r.Output = res.RawOutput
		r.Status = "completed"
		r.CompletedAt = time.Now()

		// Update the scan result with the final output and status
		err = s.resRepo.PutScanResultByID(finfo.ResultID, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		fmt.Println("complete yara scanner task : output =", res.RawOutput)
		return nil
	}
}

func (s *ScannerServerService) HashGeneratorTask(finfo *FileInfo) jobmng.TaskFunc {
	return func(t *jobmng.Task) error {
		now := time.Now()
		hash := s.heng.GenerateHash(finfo.FilePath)
		err := s.resRepo.PutScanResultByID(finfo.ResultID, domains.ScannerResult{
			Engine:      "hash_generator",
			Status:      "completed",
			Output:      hash,
			StartedAt:   now,
			CompletedAt: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}
		fmt.Println("complete hash generator task : hash =", hash)
		return nil
	}
}

func (s *ScannerServerService) RandomSleeperTask(finfo *FileInfo) jobmng.TaskFunc {
	return func(t *jobmng.Task) error {
		now := time.Now()
		r := domains.ScannerResult{
			Engine:    "random_sleeper",
			Status:    "processing",
			Output:    "",
			StartedAt: now,
		}

		// Store the initial scan result with "processing" status
		err := s.resRepo.PutScanResultByID(finfo.ResultID, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		// Run the main task function
		dur := s.reng.RandomSleep()

		r.Output = dur.String()
		r.Status = "completed"
		r.CompletedAt = time.Now()

		// Update the scan result with the final output and status
		err = s.resRepo.PutScanResultByID(finfo.ResultID, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		fmt.Println("complete random sleeper task : duration =", dur)
		return nil
	}
}

func (s *ScannerServerService) OnFileScanStarted(job *jobmng.Job, finfo *FileInfo) jobmng.OnJobStartedFunc {
	return func() {
		log.Printf("Started processing file: %s", finfo.ID)

		err := s.reqRepo.UpdateFileStatus(finfo.ScanID, finfo.ID, "processing")
		if err != nil {
			log.Printf("Failed to update file status: %v", err)
		}
	}
}

func (s *ScannerServerService) OnFileScanFinished(job *jobmng.Job, finfo *FileInfo) jobmng.OnJobFinishedFunc {
	return func() {
		scanId := finfo.ScanID
		status := "completed"
		// Check job tasks error and update file status to failed if any task failed
		// Else update file status to completed
		if len(job.TasksError()) > 0 {
			status = "failed"
		}
		err := s.reqRepo.UpdateFileStatus(scanId, finfo.ID, status)
		if err != nil {
			log.Printf("Failed to update file status: %v", err)
		}

		// Check if all files in scan request are completed or failed,
		// then update scan request status to completed
		scanReq, err := s.reqRepo.GetByID(scanId)
		if err != nil {
			log.Printf("Failed to get scan request: %v", err)
			return
		}

		allDone := true
		for _, f := range scanReq.Files {
			if f.Status != "completed" && f.Status != "failed" {
				allDone = false
				break
			}
		}

		if allDone {
			scanReq.Status = "completed"
			scanReq.CompletedAt = time.Now()
			err := s.reqRepo.Update(*scanReq)
			if err != nil {
				log.Printf("Failed to update scan request status: %v", err)
			}
		}

		log.Printf("Finished processing file: %s", finfo.ID)
	}
}
