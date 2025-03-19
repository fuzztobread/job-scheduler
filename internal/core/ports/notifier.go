package ports

import (
	"context"
	
	"github.com/fuzztobread/job-scheduler/internal/core/domain"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	NotifyNewJobs(ctx context.Context, diff domain.DiffResult) error
}