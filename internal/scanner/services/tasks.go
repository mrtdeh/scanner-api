package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mrtdeh/scanners-management/internal/scanner/domains"
	jobmng "github.com/mrtdeh/scanners-management/pkg/job_manager"
)

func (s *ScannerServerService) YaraScannerTask(fileHash, filePath string) jobmng.TaskFunc {
	return func(t *jobmng.Task) error {
		now := time.Now()
		r := domains.ScannerResult{
			Engine:    "yara_scanner",
			Status:    "processing",
			Output:    "",
			StartedAt: now,
		}

		// Store the initial scan result with "processing" status
		err := s.resRepo.PutScanResultByFileHash(fileHash, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		ctx := t.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
		defer cancel()

		// Run the main task function
		res, err := s.yeng.Scan(ctx, filePath)
		if err != nil {
			return fmt.Errorf("failed to run yara scanner: %w", err)
		}

		r.Output = res.RawOutput
		r.Status = "completed"
		r.CompletedAt = time.Now()

		// Update the scan result with the final output and status
		err = s.resRepo.PutScanResultByFileHash(fileHash, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		fmt.Println("complete yara scanner task : output =", res.RawOutput)
		return nil
	}
}

func (s *ScannerServerService) HashGeneratorTask(fileHash, filePath string) jobmng.TaskFunc {
	return func(t *jobmng.Task) error {
		now := time.Now()
		hash := s.heng.GenerateHash(filePath)
		err := s.resRepo.PutScanResultByFileHash(fileHash, domains.ScannerResult{
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

func (s *ScannerServerService) RandomSleeperTask(fileHash string) jobmng.TaskFunc {
	return func(t *jobmng.Task) error {
		now := time.Now()
		r := domains.ScannerResult{
			Engine:    "random_sleeper",
			Status:    "processing",
			Output:    "",
			StartedAt: now,
		}

		// Store the initial scan result with "processing" status
		err := s.resRepo.PutScanResultByFileHash(fileHash, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		// Run the main task function
		dur := s.reng.RandomSleep()

		r.Output = dur.String()
		r.Status = "completed"
		r.CompletedAt = time.Now()

		// Update the scan result with the final output and status
		err = s.resRepo.PutScanResultByFileHash(fileHash, r)
		if err != nil {
			return fmt.Errorf("failed to put scan result: %w", err)
		}

		fmt.Println("complete random sleeper task : duration =", dur)
		return nil
	}
}
