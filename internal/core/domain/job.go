// internal/core/domain/job.go
package domain

import "time"

// Job represents a job listing from a career page
type Job struct {
	ID          string
	Title       string
	Description string
	Location    string
	Department  string
	URL         string
	PostedDate  time.Time
	ScrapedAt   time.Time
}

// JobCollection represents a collection of jobs from a career page
type JobCollection struct {
	CompanyName string
	SourceURL   string
	ScrapedAt   time.Time
	Jobs        []Job
	RawContent  string // Raw HTML content for debugging
}

// DiffResult represents the difference between two job collections
type DiffResult struct {
	CompanyName string
	SourceURL   string
	NewJobs     []Job
	RemovedJobs []Job
	UpdatedJobs []Job
}