package scheduler	


import (
	"context"
	"sync"
	
	"github.com/robfig/cron/v3"
	"log"
	"github.com/fuzztobread/job-scheduler/internal/core/ports"
)

// CronScheduler implements the Scheduler interface using cron
type CronScheduler struct {
	cron   *cron.Cron
	jobs   map[cron.EntryID]context.CancelFunc
	mu     sync.Mutex
}

// NewCronScheduler creates a new CronScheduler instance
func NewCronScheduler() *CronScheduler {
	return &CronScheduler{
		cron: cron.New(cron.WithSeconds()),
		jobs: make(map[cron.EntryID]context.CancelFunc),
	}
}

// Schedule schedules a new job with the given cron specification
func (s *CronScheduler) Schedule(spec string, job ports.Job) error {
    _, err := s.cron.AddFunc(spec, func() {
        // Just run the job with a background context
        ctx := context.Background()
        if err := job(ctx); err != nil {
            // Log the error
            log.Printf("Job execution error: %v", err)
        }
    })
    
    return err
}

// Start starts the scheduler
func (s *CronScheduler) Start(ctx context.Context) error {
    s.cron.Start()
    
    // Wait for the context to be done
    <-ctx.Done()
    return ctx.Err()
}

// Stop stops the scheduler
func (s *CronScheduler) Stop() error {
    // This stops all jobs
    s.cron.Stop()
    return nil
}