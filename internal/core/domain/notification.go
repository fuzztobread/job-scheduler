// internal/core/domain/notification.go
package domain

import "time"

// NotificationType defines the type of notification
type NotificationType string

const (
	// NotificationTypeNewJobs indicates new job listings were found
	NotificationTypeNewJobs NotificationType = "new_jobs"
	
	// NotificationTypeUpdatedJobs indicates job listings were updated
	NotificationTypeUpdatedJobs NotificationType = "updated_jobs"
	
	// NotificationTypeRemovedJobs indicates job listings were removed
	NotificationTypeRemovedJobs NotificationType = "removed_jobs"
	
	// NotificationTypeError indicates an error occurred during scraping
	NotificationTypeError NotificationType = "error"
)

// Notification represents a notification to be sent
type Notification struct {
	ID          string           `json:"id"`
	Type        NotificationType `json:"type"`
	CompanyName string           `json:"company_name"`
	SourceURL   string           `json:"source_url"`
	Title       string           `json:"title"`
	Message     string           `json:"message"`
	CreatedAt   time.Time        `json:"created_at"`
	Payload     interface{}      `json:"payload,omitempty"`
}

// NotificationHistory represents a record of sent notifications
type NotificationHistory struct {
	Notifications []Notification `json:"notifications"`
	LastSentAt    time.Time      `json:"last_sent_at"`
}

// CreateNewJobsNotification creates a notification for new jobs
func CreateNewJobsNotification(diff DiffResult) Notification {
	return Notification{
		Type:        NotificationTypeNewJobs,
		CompanyName: diff.CompanyName,
		SourceURL:   diff.SourceURL,
		Title:       "New Job Listings",
		Message:     createJobsMessage(diff.NewJobs, "new"),
		CreatedAt:   time.Now(),
		Payload:     diff.NewJobs,
	}
}

// CreateUpdatedJobsNotification creates a notification for updated jobs
func CreateUpdatedJobsNotification(diff DiffResult) Notification {
	return Notification{
		Type:        NotificationTypeUpdatedJobs,
		CompanyName: diff.CompanyName,
		SourceURL:   diff.SourceURL,
		Title:       "Updated Job Listings",
		Message:     createJobsMessage(diff.UpdatedJobs, "updated"),
		CreatedAt:   time.Now(),
		Payload:     diff.UpdatedJobs,
	}
}

// CreateRemovedJobsNotification creates a notification for removed jobs
func CreateRemovedJobsNotification(diff DiffResult) Notification {
	return Notification{
		Type:        NotificationTypeRemovedJobs,
		CompanyName: diff.CompanyName,
		SourceURL:   diff.SourceURL,
		Title:       "Removed Job Listings",
		Message:     createJobsMessage(diff.RemovedJobs, "removed"),
		CreatedAt:   time.Now(),
		Payload:     diff.RemovedJobs,
	}
}

// CreateErrorNotification creates a notification for scraping errors
func CreateErrorNotification(companyName, sourceURL, errMsg string) Notification {
	return Notification{
		Type:        NotificationTypeError,
		CompanyName: companyName,
		SourceURL:   sourceURL,
		Title:       "Scraping Error",
		Message:     errMsg,
		CreatedAt:   time.Now(),
	}
}

// createJobsMessage creates a human-readable message about job changes
func createJobsMessage(jobs []Job, changeType string) string {
	if len(jobs) == 0 {
		return "No " + changeType + " jobs found."
	}
	
	if len(jobs) == 1 {
		return "1 " + changeType + " job: " + jobs[0].Title
	}
	
	return string(len(jobs)) + " " + changeType + " jobs found."
}

// NotificationDeliveryStatus represents the delivery status of a notification
type NotificationDeliveryStatus string

const (
	// NotificationDeliveryStatusPending indicates the notification is pending delivery
	NotificationDeliveryStatusPending NotificationDeliveryStatus = "pending"
	
	// NotificationDeliveryStatusSent indicates the notification was sent successfully
	NotificationDeliveryStatusSent NotificationDeliveryStatus = "sent"
	
	// NotificationDeliveryStatusFailed indicates the notification failed to send
	NotificationDeliveryStatusFailed NotificationDeliveryStatus = "failed"
	
	// NotificationDeliveryStatusRetrying indicates the notification is being retried
	NotificationDeliveryStatusRetrying NotificationDeliveryStatus = "retrying"
)

// NotificationDelivery represents a delivery attempt for a notification
type NotificationDelivery struct {
	NotificationID string                   `json:"notification_id"`
	Status         NotificationDeliveryStatus `json:"status"`
	Attempts       int                      `json:"attempts"`
	LastAttemptAt  time.Time                `json:"last_attempt_at"`
	ErrorMessage   string                   `json:"error_message,omitempty"`
}