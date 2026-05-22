package jobmng

import (
	"context"
	"fmt"
	"time"
)

type TaskFunc func(t *Task) error

type Task struct {
	job         *Job
	Err         error
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time

	ctx context.Context
	fn  TaskFunc
}

func (t *Task) Context() context.Context {
	return t.ctx
}

func (t *Task) run(maxRetries int, timeout, backoff time.Duration) {

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context donw before process
		select {
		case <-t.ctx.Done():
			t.Err = t.ctx.Err()
			return
		default:
		}

		timeoutCtx, cancel := context.WithTimeout(t.ctx, timeout)
		defer cancel()

		// Run process
		t.ctx = timeoutCtx
		err := t.fn(t)

		// If has error, then set failed status and return
		if err == nil {
			// t.Status = TaskStatusCompleted
			now := time.Now()
			t.CompletedAt = &now
			return
		}

		// else, check for retry
		if attempt == maxRetries {
			now := time.Now()
			t.CompletedAt = &now
			t.Err = err
			// return fmt.Errorf("max retry for this task is exsited : %d", attempt)
			return
		}

		// Wait for a second and if context is not cancelled, continue retry
		select {
		case <-time.After(backoff):
			continue
		case <-t.ctx.Done():
			// t.Status = TaskStatusCancelled
			// t.Error = t.ctx.Err()
			t.Err = fmt.Errorf("task timeout")
			return
		}
	}
}
