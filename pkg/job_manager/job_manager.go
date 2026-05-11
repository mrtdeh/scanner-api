package jobmng

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID         string
	qu         chan *Task
	MaxRetry   int
	MaxTimeout time.Duration
	CreatedAt  time.Time
	Data       []byte
	Err        error

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

type Task struct {
	Err         error
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time

	ctx context.Context
	fn  TaskFunc
}

func (t *Task) run(maxRetries int, timeout, backoff time.Duration) error {

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context donw before process
		select {
		case <-t.ctx.Done():
			return t.ctx.Err()
		default:
		}

		timeoutCtx, cancel := context.WithTimeout(t.ctx, timeout)
		defer cancel()

		// Run process
		err := t.fn(timeoutCtx)

		// If has error, then set failed status and return
		if err == nil {
			// t.Status = TaskStatusCompleted
			now := time.Now()
			t.CompletedAt = &now
			return nil
		}

		// else, check for retry
		if attempt == maxRetries {
			// t.Status = TaskStatusFailed
			now := time.Now()
			t.CompletedAt = &now
			return fmt.Errorf("max retry for this task is exsited : %d", attempt)
		}

		// Wait for a second and if context is not cancelled, continue retry
		select {
		case <-time.After(backoff):
			continue
		case <-t.ctx.Done():
			// t.Status = TaskStatusCancelled
			// t.Error = t.ctx.Err()
			return fmt.Errorf("task timeout")
		}
	}

	return nil
}

func NewJob(maxRetries int, maxTimeout time.Duration, data any) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	d, _ := json.Marshal(data)
	return &Job{
		Data:       d,
		MaxRetry:   maxRetries,
		MaxTimeout: maxTimeout,
		CreatedAt:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,

		qu: make(chan *Task),
		wg: &sync.WaitGroup{},
	}
}

func (j *Job) AddTask(fn TaskFunc) *Job {

	j.qu <- &Task{
		CreatedAt: time.Now(),
		fn:        fn,
		ctx:       j.ctx,
	}
	return j
}

func (j *Job) WaitForCompleteTasks() {
	j.wg.Wait()
}

func (j *Job) Close() {
	j.cancel()
	j.WaitForCompleteTasks()
	close(j.qu)
}

func (j *Job) RunTasks(concurrent bool) error {
	for {
		select {
		case t := <-j.qu:
			j.wg.Add(1)
			if concurrent {
				go func() {
					err := t.run(j.MaxRetry, j.MaxTimeout, time.Second)
					if err != nil {
						t.Err = err
					}
					j.wg.Done()
				}()
			} else {
				err := t.run(j.MaxRetry, j.MaxTimeout, time.Second)
				if err != nil {
					t.Err = err
				}
				j.wg.Done()
			}
		case <-j.ctx.Done():
			return fmt.Errorf("job canclled by context : %v", j.ctx.Err())
		default:
			return nil

		}

	}
}

// ===========================================================

type TaskFunc func(ctx context.Context) error

type JobManager struct {
	mu       sync.RWMutex
	wg       sync.WaitGroup
	jobs     chan *Job
	wrkLimit chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
	isClosed bool
}

type JobManagerConfig struct {
	WorkerLimit int
	QueueSize   int
}

func DefaultConfig() JobManagerConfig {
	return JobManagerConfig{
		WorkerLimit: 10,
		QueueSize:   100,
	}
}

func NewJobManager(ctx context.Context, config JobManagerConfig) (*JobManager, error) {
	if config.WorkerLimit <= 0 {
		config.WorkerLimit = 10
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 100
	}

	ctx, cancel := context.WithCancel(ctx)

	jm := &JobManager{
		jobs:     make(chan *Job, config.QueueSize),
		wrkLimit: make(chan struct{}, config.WorkerLimit),
		ctx:      ctx,
		cancel:   cancel,
	}

	for i := 0; i < config.WorkerLimit; i++ {
		jm.wg.Add(1)
		go jm.worker(i)
	}

	return jm, nil
}

func (jm *JobManager) worker(workerID int) {
	defer jm.wg.Done()

	for {
		select {
		case <-jm.ctx.Done():
			fmt.Printf("Worker %d stopping\n", workerID)
			return

		case job, ok := <-jm.jobs:
			if !ok {
				return
			}

			jm.process(job)
		}
	}
}

func (jm *JobManager) process(job *Job) {
	select {
	case jm.wrkLimit <- struct{}{}:
		defer func() { <-jm.wrkLimit }()
	case <-jm.ctx.Done():
		return
	}

	if err := job.RunTasks(true); err != nil {
		job.Err = err
		return
	}

	job.WaitForCompleteTasks()
	job.Close()

}

func (jm *JobManager) Close() error {
	jm.mu.Lock()
	if jm.isClosed {
		jm.mu.Unlock()
		return fmt.Errorf("already closed")
	}

	jm.isClosed = true
	jm.mu.Unlock()

	jm.cancel()

	jm.wg.Wait()

	close(jm.jobs)
	close(jm.wrkLimit)
	return nil
}

func (jm *JobManager) AddJob(job Job) error {
	jm.mu.RLock()
	isClosed := jm.isClosed
	jm.mu.RUnlock()

	if isClosed {
		return fmt.Errorf("job manager is closed")
	}
	ctx, cancel := context.WithCancel(jm.ctx)
	job.ctx = ctx
	job.cancel = cancel
	select {
	case jm.jobs <- &job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}
