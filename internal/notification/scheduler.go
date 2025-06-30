package notification

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/fx"
)

// NotificationScheduler handles periodic checking and sending of due notifications.
type NotificationScheduler struct {
	service *NotificationService
}

// NewNotificationScheduler creates a new scheduler for notifications.
func NewNotificationScheduler(service *NotificationService) *NotificationScheduler {
	return &NotificationScheduler{service: service}
}

// StartScheduler starts the background goroutine to periodically check and send due notifications.
func (s *NotificationScheduler) StartScheduler(lc fx.Lifecycle) {
	// Get scheduler interval from environment, default to 1 minute
	intervalStr := os.Getenv("NOTIFICATION_SCHEDULER_INTERVAL_MINUTES")
	interval := 1 // Default to 1 minute
	if intervalStr != "" {
		if parsed, err := strconv.Atoi(intervalStr); err == nil && parsed > 0 {
			interval = parsed
		}
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	done := make(chan bool)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Starting notification scheduler (checking every %d minutes)...", interval)
			go func() {
				// Create a new context for the goroutine that won't be cancelled
				schedulerCtx := context.Background()
				for {
					select {
					case <-ticker.C:
						s.service.SendDueNotifications(schedulerCtx)
					case <-done:
						return
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping notification scheduler...")
			ticker.Stop()
			done <- true
			return nil
		},
	})
}

// Why: This scheduler runs in the background to automatically send notifications when they are due, without requiring manual intervention.
