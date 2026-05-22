package jobmng

import (
	"context"
	"fmt"
	"sync"
)

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

	if err := job.Run(true); err != nil {
		job.err = err
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

func (jm *JobManager) AddJob(job *Job) error {
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
	case jm.jobs <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}
