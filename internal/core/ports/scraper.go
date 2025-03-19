// internal/core/ports/scraper.go
package ports

import (
	"context"

	"github.com/fuzztobread/job-scheduler/internal/core/domain"
)

// Scraper defines the interface for scraping career pages
type Scraper interface {
	Scrape(ctx context.Context, url string) (domain.JobCollection, error)
}
