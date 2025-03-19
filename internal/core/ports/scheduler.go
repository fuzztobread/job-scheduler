// internal/core/ports/scheduler.go
package ports

import (
	"context"
)

// Job represents a scheduled job to be executed
type Job func(ctx context.Context) error

// Scheduler defines the interface for scheduling jobs
type Scheduler interface {
	Schedule(spec string, job Job) error
	Start(ctx context.Context) error
	Stop() error
}
