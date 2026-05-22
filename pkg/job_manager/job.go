package jobmng

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	cnf   Config
	tasks []*Task
	err   error

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

type Config struct {
	// MaxRetry is the maximum number of retries for each task.
	MaxRetry int
	// MaxTimeout is the maximum duration for each task to complete.
	Timeout time.Duration
	// Backoff is the duration to wait before retrying a failed task.
	Backoff time.Duration
}

// NewJob, return none nil new job pointer. note that no need to nil check for this value
func NewJob(cnf Config) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	return &Job{
		cnf:    cnf,
		ctx:    ctx,
		cancel: cancel,

		tasks: []*Task{},
		wg:    &sync.WaitGroup{},
	}
}

func (j *Job) AddTasks(tasksFn ...TaskFunc) {
	now := time.Now()
	for _, fn := range tasksFn {
		j.tasks = append(j.tasks, &Task{
			CreatedAt: now,
			fn:        fn,
			ctx:       j.ctx,
		})
	}
}

func (j *Job) WaitForCompleteTasks() {
	j.wg.Wait()
}

func (j *Job) Close() {
	j.cancel()
	j.WaitForCompleteTasks()
}

func (j *Job) Run(concurrent bool) error {
	j.wg.Add(len(j.tasks))
	for _, t := range j.tasks {
		select {
		case <-j.ctx.Done():
			return fmt.Errorf("job canclled by context : %v", j.ctx.Err())
		default:
			if concurrent {
				go func() {
					t.run(j.cnf.MaxRetry, j.cnf.Timeout, j.cnf.Backoff)
					j.wg.Done()
				}()
			} else {
				t.run(j.cnf.MaxRetry, j.cnf.Timeout, j.cnf.Backoff)
				j.wg.Done()
			}
		}
	}
	return nil
}

// ===========================================================
