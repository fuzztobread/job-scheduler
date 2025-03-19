// internal/adapters/scraper/gorod_scraper.go
package scraper

import (
	"context"
	"fmt"
	"time"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	
	"github.com/go-rod/rod"
	"log"
	"github.com/PuerkitoBio/goquery"
	
	"github.com/fuzztobread/job-scheduler/internal/core/domain"
)

// GoRodScraper implements the Scraper interface using go-rod
type GoRodScraper struct {
	timeout time.Duration
}

// NewGoRodScraper creates a new GoRodScraper instance
func NewGoRodScraper(timeout time.Duration) *GoRodScraper {
	return &GoRodScraper{
		timeout: timeout,
	}
}

// Scrape scrapes a career page and returns the job listings
func (s *GoRodScraper) Scrape(ctx context.Context, url string) (domain.JobCollection, error) {
	log.Printf("Starting to scrape URL: %s", url)
	
	result := domain.JobCollection{
		SourceURL: url,
		ScrapedAt: time.Now(),
	}
	
	// Extract company name from URL
	result.CompanyName = extractCompanyName(url)
	log.Printf("Extracted company name: %s", result.CompanyName)
	
	// Launch a new browser
	log.Printf("Launching browser...")
	browser := rod.New().Timeout(s.timeout)
	defer browser.Close()
	
	// Connect to the browser
	log.Printf("Connecting to browser...")
	if err := browser.Connect(); err != nil {
		return result, fmt.Errorf("failed to connect to browser: %w", err)
	}
	
	// Create a new page
	log.Printf("Creating new page...")
	page := browser.MustPage()
	defer page.Close()
	
	// Navigate to the career page
	log.Printf("Navigating to %s...", url)
	if err := page.Navigate(url); err != nil {
		return result, fmt.Errorf("failed to navigate to career page: %w", err)
	}
	
	// Wait for the page to load
	log.Printf("Waiting for page to stabilize...")
	if err := page.WaitStable(2 * time.Second); err != nil {
		return result, fmt.Errorf("failed to wait for page to stabilize: %w", err)
	}
	
	// Get the HTML content
	log.Printf("Getting HTML content...")
	html, err := page.HTML()
	if err != nil {
		return result, fmt.Errorf("failed to get HTML content: %w", err)
	}
	
	result.RawContent = html
	log.Printf("Retrieved HTML content (%d bytes)", len(html))
	
	// Parse the HTML
	log.Printf("Parsing jobs from HTML...")
	jobs, err := s.parseJobs(html, url)
	if err != nil {
		return result, fmt.Errorf("failed to parse jobs: %w", err)
	}
	
	result.Jobs = jobs
	log.Printf("Found %d jobs on page", len(jobs))
	
	return result, nil
}

// parseJobs parses job listings from HTML content
func (s *GoRodScraper) parseJobs(html, sourceURL string) ([]domain.Job, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	var jobs []domain.Job
	
	// This is a generic selector - you'll need to customize it for each site
	// Common job listing patterns to look for
	jobSelectors := []string{
		".job-listing", 
		".careers-listing", 
		".job-post", 
		".job-card",
		"[data-job-id]",
		"article.job",
		// F1soft career site specific selector
		".features-job",
		// Google careers specific XPath
		"/html/body/div[2]/section[2]/div/div[2]/div[1]/div",
	}
	
	// Try each selector until we find something
	for _, selector := range jobSelectors {
		// Handle XPath selectors (starting with /)
		var selection *goquery.Selection
		if strings.HasPrefix(selector, "/") {
			log.Printf("Trying XPath selector: %s", selector)
			// For XPath selectors, we need a different approach
			// Convert the XPath to a CSS selector if possible, or handle it specially
			
			// Special case for the Google careers XPath
			if selector == "/html/body/div[2]/section[2]/div/div[2]/div[1]/div" {
				selection = doc.Find("div > section:nth-child(2) > div > div:nth-child(2) > div:nth-child(1) > div")
			} else {
				log.Printf("Skipping unsupported XPath selector: %s", selector)
				continue
			}
		} else {
			log.Printf("Trying CSS selector: %s", selector)
			selection = doc.Find(selector)
		}
		
		selection.Each(func(i int, s *goquery.Selection) {
			job := domain.Job{
				ScrapedAt: time.Now(),
			}
			
			// Try to extract job ID
			jobID, exists := s.Attr("data-job-id")
			if !exists {
				jobID, exists = s.Attr("id")
			}
			
			// If we still don't have an ID, generate one from the content
			if !exists || jobID == "" {
				// Create a hash from the job content
				hash := sha256.Sum256([]byte(s.Text()))
				jobID = hex.EncodeToString(hash[:])
			}
			
			job.ID = jobID
			
			// Special handling for F1soft career site structure
			if s.HasClass("features-job") {
				log.Printf("Processing F1soft career site job listing")
				
				// Job title is in h3 > a
				job.Title = s.Find("h3 a").Text()
				job.Title = strings.TrimSpace(job.Title)
				
				// Job URL
				jobURL, exists := s.Find("h3 a").Attr("href")
				if exists {
					if strings.HasPrefix(jobURL, "/") {
						urlParts := strings.Split(sourceURL, "/")
						baseURL := strings.Join(urlParts[:3], "/")
						jobURL = baseURL + jobURL
					}
					job.URL = jobURL
				}
				
				// Company name
				job.Department = s.Find(".box-content a.fw-600").Text()
				job.Department = strings.TrimSpace(job.Department)
				
				// Location
				job.Location = s.Find(".icon-map-pin + span").Text()
				job.Location = strings.TrimSpace(job.Location)
				
				// Additional info in description
				var descParts []string
				
				// Job type
				jobType := s.Find(".job-tag li:nth-child(1) a").Text()
				if jobType != "" {
					descParts = append(descParts, "Type: "+jobType)
				}
				
				// Job level
				jobLevel := s.Find(".job-tag li:nth-child(2) a").Text()
				if jobLevel != "" {
					descParts = append(descParts, "Level: "+jobLevel)
				}
				
				// Category
				category := s.Find(".job-tag li:nth-child(3) a").Text()
				if category != "" {
					descParts = append(descParts, "Category: "+category)
				}
				
				// Deadline
				deadline := s.Find("p.days").Text()
				deadline = strings.TrimSpace(deadline)
				if deadline != "" {
					descParts = append(descParts, deadline)
				}
				
				job.Description = strings.Join(descParts, " | ")
			} else {
				// Try different selectors for job title
				job.Title = s.Find(".job-title, h2, h3").First().Text()
				job.Title = strings.TrimSpace(job.Title)
				
				// Try different selectors for job description
				job.Description = s.Find(".job-description, .description, p").Text()
				job.Description = strings.TrimSpace(job.Description)
				
				// Try different selectors for job location
				job.Location = s.Find(".job-location, .location").Text()
				job.Location = strings.TrimSpace(job.Location)
				
				// Try different selectors for job department
				job.Department = s.Find(".job-department, .department, .category").Text()
				job.Department = strings.TrimSpace(job.Department)
				
				// Try to extract job URL
				jobURL, exists := s.Find("a").First().Attr("href")
				if exists {
					// If it's a relative URL, make it absolute
					if strings.HasPrefix(jobURL, "/") {
						urlParts := strings.Split(sourceURL, "/")
						baseURL := strings.Join(urlParts[:3], "/")
						jobURL = baseURL + jobURL
					}
					job.URL = jobURL
				}
			}
			
			// Only add jobs with at least a title
			if job.Title != "" {
				jobs = append(jobs, job)
				log.Printf("Found job: %s", job.Title)
			}
		})
		
		// If we found jobs with this selector, stop trying others
		if len(jobs) > 0 {
			log.Printf("Successfully found %d jobs using selector: %s", len(jobs), selector)
			break
		}
	}
	
	// If we still haven't found any jobs, try a more aggressive approach for Google careers
	if len(jobs) == 0 && strings.Contains(sourceURL, "google.com") {
		log.Printf("Trying aggressive approach for Google careers page")
		
		// Look for any divs that might contain job information
		doc.Find("div.career-item, div.job-item, div.position-item, div.opening").Each(func(i int, s *goquery.Selection) {
			job := domain.Job{
				ScrapedAt: time.Now(),
				ID:        fmt.Sprintf("google-job-%d", i),
			}
			
			// Look for title in various elements
			job.Title = s.Find("h3, h4, .title, .position-title").First().Text()
			job.Title = strings.TrimSpace(job.Title)
			
			// Look for other job details
			job.Description = s.Find("p, .description").Text()
			job.Description = strings.TrimSpace(job.Description)
			
			job.Location = s.Find(".location").Text()
			job.Location = strings.TrimSpace(job.Location)
			
			// Extract URL if available
			jobURL, exists := s.Find("a").First().Attr("href")
			if exists {
				if strings.HasPrefix(jobURL, "/") {
					urlParts := strings.Split(sourceURL, "/")
					baseURL := strings.Join(urlParts[:3], "/")
					jobURL = baseURL + jobURL
				}
				job.URL = jobURL
			}
			
			if job.Title != "" {
				jobs = append(jobs, job)
				log.Printf("Found Google job: %s", job.Title)
			}
		})
		
		// If that didn't work, try to extract any text that looks like job titles
		if len(jobs) == 0 {
			log.Printf("Trying to extract any potential job titles from the page")
			doc.Find("h1, h2, h3, h4, h5, strong").Each(func(i int, s *goquery.Selection) {
				text := strings.TrimSpace(s.Text())
				if len(text) > 0 && len(text) < 100 { // Job titles are usually not too long
					job := domain.Job{
						ScrapedAt:   time.Now(),
						ID:          fmt.Sprintf("google-text-%d", i),
						Title:       text,
						Description: "Extracted from Google careers page",
					}
					jobs = append(jobs, job)
					log.Printf("Extracted potential job title: %s", text)
				}
			})
		}
	}
	
	return jobs, nil
}

// extractCompanyName extracts the company name from a URL
func extractCompanyName(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	
	// Extract domain
	urlParts := strings.Split(url, "/")
	if len(urlParts) > 0 {
		domainParts := strings.Split(urlParts[0], ".")
		if len(domainParts) > 1 {
			return strings.Title(domainParts[len(domainParts)-2])
		}
		return strings.Title(domainParts[0])
	}
	
	return "Unknown Company"
}