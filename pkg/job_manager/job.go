package jobmng

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID         string
	tasks      []*Task
	MaxRetry   int
	MaxTimeout time.Duration
	CreatedAt  time.Time
	Params     map[string]any
	Err        error

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

// NewJob, return none nil new job pointer. note that no need to nil check for this value
func NewJob(maxRetries int, maxTimeout time.Duration) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	return &Job{
		MaxRetry:   maxRetries,
		MaxTimeout: maxTimeout,
		CreatedAt:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,

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
					t.run(j.MaxRetry, j.MaxTimeout, time.Second)
					j.wg.Done()
				}()
			} else {
				t.run(j.MaxRetry, j.MaxTimeout, time.Second)
				j.wg.Done()
			}
		}
	}
	return nil
}

// ===========================================================
