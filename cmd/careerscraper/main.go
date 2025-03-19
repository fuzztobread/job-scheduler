// cmd/careerscraper/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/fuzztobread/job-scheduler/internal/adapters/notifier"
	"github.com/fuzztobread/job-scheduler/internal/adapters/repository"
	"github.com/fuzztobread/job-scheduler/internal/adapters/scheduler"
	"github.com/fuzztobread/job-scheduler/internal/adapters/scraper"
	"github.com/fuzztobread/job-scheduler/internal/config"
	"github.com/fuzztobread/job-scheduler/internal/core/ports"
	"github.com/fuzztobread/job-scheduler/internal/core/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Create scraper
	scraper := scraper.NewGoRodScraper(30 * time.Second)
	
	// Create repository
	repo := repository.NewMemoryRepository()
	
	// Create notifier
	var notifierInstance ports.Notifier
	switch cfg.NotifierType {
	case "discord":
		if cfg.DiscordWebhookURL == "" {
			log.Fatalf("Discord webhook URL is required for Discord notifier")
		}
		notifierInstance = notifier.NewDiscordNotifier(cfg.DiscordWebhookURL)
	
	default:
		log.Fatalf("Unknown notifier type: %s", cfg.NotifierType)
	}
	
	// Create service
	service := services.NewCareerScraperService(scraper, notifierInstance, repo, cfg.URLs)
	
	// Create scheduler
	scheduler := scheduler.NewCronScheduler()
	
	// For testing - run the job immediately once
	log.Println("Running initial scrape job...")
	if err := service.ScrapeAndNotify(context.Background()); err != nil {
		log.Printf("Initial scrape job failed: %v", err)
	}
	
	// Schedule the scraping job
	log.Printf("Scheduling job with cron expression: %s", cfg.ScrapeInterval)
	if err := scheduler.Schedule(cfg.ScrapeInterval, service.ScrapeAndNotify); err != nil {
		log.Fatalf("Failed to schedule job: %v", err)
	}
	
	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start the scheduler
	go func() {
		if err := scheduler.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Scheduler stopped with error: %v", err)
		}
	}()
	
	log.Printf("Career scraper started, monitoring %d URLs every %s", len(cfg.URLs), cfg.ScrapeInterval)
	
	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for termination signal
	<-sigCh
	log.Println("Shutting down...")
	
	// Stop the scheduler
	cancel()
	if err := scheduler.Stop(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}
	
	log.Println("Shutdown complete")
}