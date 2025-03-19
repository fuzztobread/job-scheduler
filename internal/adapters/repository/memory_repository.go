// internal/adapters/repository/memory_repository.go
package repository

import (
	"context"
	"sync"

	"github.com/fuzztobread/job-scheduler/internal/core/domain"
	"github.com/fuzztobread/job-scheduler/internal/core/ports"
)

// MemoryRepository implements the JobRepository interface using in-memory storage
type MemoryRepository struct {
	collections map[string]domain.JobCollection
	mu          sync.RWMutex
}

// NewMemoryRepository creates a new MemoryRepository instance
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		collections: make(map[string]domain.JobCollection),
	}
}

// SaveJobCollection saves a job collection to the repository
func (r *MemoryRepository) SaveJobCollection(
	ctx context.Context,
	collection domain.JobCollection,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.collections[collection.SourceURL] = collection
	return nil
}

// GetLatestJobCollection retrieves the latest job collection for a URL
func (r *MemoryRepository) GetLatestJobCollection(
	ctx context.Context,
	url string,
) (domain.JobCollection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collection, exists := r.collections[url]
	if !exists {
		return domain.JobCollection{}, nil
	}

	return collection, nil
}

var _ ports.JobRepository = (*MemoryRepository)(nil) // Ensure interface compliance
