package ports

import (
	"context"

	"github.com/fuzztobread/job-scheduler/internal/core/domain"
)

// JobRepository defines the interface for storing and retrieving job data
type JobRepository interface {
	SaveJobCollection(ctx context.Context, jobs domain.JobCollection) error
	GetLatestJobCollection(ctx context.Context, url string) (domain.JobCollection, error)
}
