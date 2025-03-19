// internal/core/services/careerscraper_service.go
package services

import (
	"context"
	"log"

	"fmt"
	"github.com/fuzztobread/job-scheduler/internal/core/domain"
	"github.com/fuzztobread/job-scheduler/internal/core/ports"
)

// CareerScraperService is responsible for orchestrating the scraping process
type CareerScraperService struct {
	scraper    ports.Scraper
	notifier   ports.Notifier
	repository ports.JobRepository
	urls       []string
}

// NewCareerScraperService creates a new instance of CareerScraperService
func NewCareerScraperService(
	scraper ports.Scraper,
	notifier ports.Notifier,
	repository ports.JobRepository,
	urls []string,
) *CareerScraperService {
	return &CareerScraperService{
		scraper:    scraper,
		notifier:   notifier,
		repository: repository,
		urls:       urls,
	}
}

// ScrapeAndNotify scrapes the specified URLs and sends notifications for changes
func (s *CareerScraperService) ScrapeAndNotify(ctx context.Context) error {
	log.Printf("Starting scrape job for %d URLs", len(s.urls))
	
	for _, url := range s.urls {
		log.Printf("Processing URL: %s", url)
		if err := s.processSingleURL(ctx, url); err != nil {
			log.Printf("Error processing URL %s: %v", url, err)
			// Continue with other URLs instead of failing entirely
			continue
		}
	}
	
	log.Printf("Completed scrape job for all URLs")
	return nil
}

// processSingleURL handles the scraping and notification for a single URL
func (s *CareerScraperService) processSingleURL(ctx context.Context, url string) error {
	log.Printf("Starting to scrape URL: %s", url)
	
	// Scrape the career page
	currentJobs, err := s.scraper.Scrape(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to scrape URL %s: %w", url, err)
	}
	
	log.Printf("Found %d jobs at %s", len(currentJobs.Jobs), url)
	
	// Get the previous job collection
	previousJobs, err := s.repository.GetLatestJobCollection(ctx, url)
	if err != nil {
		log.Printf("No previous job data found for %s: %v", url, err)
		// If it's the first time or there was an error, just save and don't notify
		return s.repository.SaveJobCollection(ctx, currentJobs)
	}
	
	log.Printf("Retrieved previous job collection with %d jobs", len(previousJobs.Jobs))
	
	// Compare and find differences
	diff := s.compareScrapeResults(previousJobs, currentJobs)
	
	// Log the diff results
	log.Printf("Diff results for %s: %d new, %d updated, %d removed", 
		url, len(diff.NewJobs), len(diff.UpdatedJobs), len(diff.RemovedJobs))
	
	// If there are changes, send notifications
	if len(diff.NewJobs) > 0 || len(diff.RemovedJobs) > 0 || len(diff.UpdatedJobs) > 0 {
		log.Printf("Sending notification for changes at %s", url)
		if err := s.notifier.NotifyNewJobs(ctx, diff); err != nil {
			log.Printf("Failed to send notification: %v", err)
			// Continue anyway and save the new results
		} else {
			log.Printf("Successfully sent notification")
		}
	} else {
		log.Printf("No changes detected for %s", url)
	}
	
	// Save the current results
	log.Printf("Saving current job collection for %s", url)
	if err := s.repository.SaveJobCollection(ctx, currentJobs); err != nil {
		return fmt.Errorf("failed to save job collection: %w", err)
	}
	
	log.Printf("Successfully processed URL: %s", url)
	return nil
}

// compareScrapeResults compares two job collections and returns the differences
func (s *CareerScraperService) compareScrapeResults(
	previous, current domain.JobCollection,
) domain.DiffResult {
	result := domain.DiffResult{
		CompanyName: current.CompanyName,
		SourceURL:   current.SourceURL,
	}
	
	// Create maps for easier comparison
	prevJobMap := make(map[string]domain.Job)
	currJobMap := make(map[string]domain.Job)
	
	for _, job := range previous.Jobs {
		prevJobMap[job.ID] = job
	}
	
	for _, job := range current.Jobs {
		currJobMap[job.ID] = job
		
		prevJob, exists := prevJobMap[job.ID]
		if !exists {
			// New job
			result.NewJobs = append(result.NewJobs, job)
		} else if job.Title != prevJob.Title || 
				 job.Description != prevJob.Description || 
				 job.Location != prevJob.Location || 
				 job.Department != prevJob.Department {
			// Updated job
			result.UpdatedJobs = append(result.UpdatedJobs, job)
		}
	}
	
	// Find removed jobs
	for _, prevJob := range previous.Jobs {
		if _, exists := currJobMap[prevJob.ID]; !exists {
			result.RemovedJobs = append(result.RemovedJobs, prevJob)
		}
	}
	
	return result
}