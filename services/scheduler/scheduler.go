package scheduler

import (
	"context"
	"sync"
	"time"
)

type Scheduler struct {
	sync.Mutex
	ticker *time.Ticker
	jobs   []*Job
}
type Job struct {
	D       time.Duration
	Apply   func()
	lastRun time.Time
}

func NewScheduler(tickDuration time.Duration) *Scheduler {
	return &Scheduler{ticker: time.NewTicker(tickDuration)}
}

func (s *Scheduler) AddJob(job *Job) {
	job.lastRun = time.Now()
	s.Lock()
	defer s.Unlock()
	s.jobs = append(s.jobs, job)
}
func (s *Scheduler) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-s.ticker.C:
				for _, job := range s.jobs {
					if job.lastRun.Add(job.D).Before(time.Now()) {
						go func() {
							defer func() { job.lastRun = time.Now() }()
							job.Apply()
						}()
					}
				}
			case <-ctx.Done():
				break
			}
		}
	}()
}
