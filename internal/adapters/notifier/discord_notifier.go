// internal/adapters/notifier/discord_notifier.go
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/fuzztobread/job-scheduler/internal/core/domain"
)

// DiscordNotifier implements the Notifier interface for Discord webhooks
type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

// DiscordEmbed represents a Discord embed object
type DiscordEmbed struct {
	Title       string                  `json:"title,omitempty"`
	Description string                  `json:"description,omitempty"`
	URL         string                  `json:"url,omitempty"`
	Color       int                     `json:"color,omitempty"`
	Fields      []DiscordEmbedField     `json:"fields,omitempty"`
	Author      *DiscordEmbedAuthor     `json:"author,omitempty"`
	Footer      *DiscordEmbedFooter     `json:"footer,omitempty"`
	Timestamp   string                  `json:"timestamp,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordEmbedAuthor represents the author of a Discord embed
type DiscordEmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordEmbedFooter represents the footer of a Discord embed
type DiscordEmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordWebhookPayload represents a Discord webhook payload
type DiscordWebhookPayload struct {
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Content   string         `json:"content,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

// NewDiscordNotifier creates a new DiscordNotifier instance
func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NotifyNewJobs sends a notification about new job listings to Discord
func (n *DiscordNotifier) NotifyNewJobs(ctx context.Context, diff domain.DiffResult) error {
	// Skip if there are no changes
	if len(diff.NewJobs) == 0 && len(diff.UpdatedJobs) == 0 && len(diff.RemovedJobs) == 0 {
		return nil
	}
	
	// Create the webhook payload
	payload := DiscordWebhookPayload{
		Username:  "Career Scraper",
		AvatarURL: "https://cdn-icons-png.flaticon.com/512/4365/4365271.png", // Job search icon
		Content:   fmt.Sprintf("Job updates for **%s**", diff.CompanyName),
		Embeds:    []DiscordEmbed{},
	}
	
	// Add source URL embed
	sourceEmbed := DiscordEmbed{
		Title:       "Career Page",
		URL:         diff.SourceURL,
		Description: "Click the title to visit the career page",
		Color:       3447003, // Blue color
		Footer: &DiscordEmbedFooter{
			Text: fmt.Sprintf("Last updated: %s", time.Now().Format(time.RFC1123)),
		},
	}
	payload.Embeds = append(payload.Embeds, sourceEmbed)
	
	// Add new jobs
	if len(diff.NewJobs) > 0 {
		newJobsEmbed := DiscordEmbed{
			Title:       fmt.Sprintf("New Jobs (%d)", len(diff.NewJobs)),
			Description: "The following jobs have been newly listed:",
			Color:       5763719, // Green color
			Fields:      []DiscordEmbedField{},
		}
		
		for _, job := range diff.NewJobs {
			// Create details string
			var details []string
			if job.Department != "" {
				details = append(details, fmt.Sprintf("Department: %s", job.Department))
			}
			if job.Location != "" {
				details = append(details, fmt.Sprintf("Location: %s", job.Location))
			}
			
			detailsStr := "No additional details"
			if len(details) > 0 {
				detailsStr = strings.Join(details, " | ")
			}
			
			// Add job field
			field := DiscordEmbedField{
				Name:   job.Title,
				Value:  fmt.Sprintf("[View Job](%s)\n%s", job.URL, detailsStr),
				Inline: false,
			}
			newJobsEmbed.Fields = append(newJobsEmbed.Fields, field)
			
			// Add description if available and not too long
			if job.Description != "" {
				desc := job.Description
				if len(desc) > 200 {
					desc = desc[:197] + "..."
				}
				
				descField := DiscordEmbedField{
					Name:   "Description",
					Value:  desc,
					Inline: false,
				}
				newJobsEmbed.Fields = append(newJobsEmbed.Fields, descField)
			}
		}
		
		payload.Embeds = append(payload.Embeds, newJobsEmbed)
	}
	
	// Add updated jobs
	if len(diff.UpdatedJobs) > 0 {
		updatedJobsEmbed := DiscordEmbed{
			Title:       fmt.Sprintf("Updated Jobs (%d)", len(diff.UpdatedJobs)),
			Description: "The following jobs have been updated:",
			Color:       16776960, // Yellow color
			Fields:      []DiscordEmbedField{},
		}
		
		for _, job := range diff.UpdatedJobs {
			field := DiscordEmbedField{
				Name:   job.Title,
				Value:  fmt.Sprintf("[View Job](%s)", job.URL),
				Inline: false,
			}
			updatedJobsEmbed.Fields = append(updatedJobsEmbed.Fields, field)
		}
		
		payload.Embeds = append(payload.Embeds, updatedJobsEmbed)
	}
	
	// Add removed jobs
	if len(diff.RemovedJobs) > 0 {
		removedJobsEmbed := DiscordEmbed{
			Title:       fmt.Sprintf("Removed Jobs (%d)", len(diff.RemovedJobs)),
			Description: "The following jobs are no longer listed:",
			Color:       15158332, // Red color
			Fields:      []DiscordEmbedField{},
		}
		
		for _, job := range diff.RemovedJobs {
			field := DiscordEmbedField{
				Name:   job.Title,
				Value:  job.Department + (func() string { if job.Location != "" { return " | " + job.Location }; return "" })(),
				Inline: false,
			}
			removedJobsEmbed.Fields = append(removedJobsEmbed.Fields, field)
		}
		
		payload.Embeds = append(payload.Embeds, removedJobsEmbed)
	}
	
	// Send the webhook
	return n.sendWebhook(ctx, payload)
}

// sendWebhook sends a payload to the Discord webhook
func (n *DiscordNotifier) sendWebhook(ctx context.Context, payload DiscordWebhookPayload) error {
	// Marshal payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord webhook payload: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Discord webhook request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Send request
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord webhook returned non-success status: %d", resp.StatusCode)
	}
	
	return nil
}
