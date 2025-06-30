package notification

import (
	"ExamSeatPlanner/internal/auth"
	"ExamSeatPlanner/internal/config"
	"context"
	"log"
	"os"
	"time"
)

// NotificationService handles scheduling and sending notifications.
type NotificationService struct {
	repo         *NotificationRepository
	emailService *config.EmailService
	userRepo     *auth.UserRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(repo *NotificationRepository, emailService *config.EmailService, userRepo *auth.UserRepository) *NotificationService {
	return &NotificationService{repo: repo, emailService: emailService, userRepo: userRepo}
}

// ScheduleNotification saves a new notification to the DB.
func (s *NotificationService) ScheduleNotification(ctx context.Context, n *Notification) error {
	n.Status = "scheduled"
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	return s.repo.CreateNotification(ctx, n)
}

// SendDueNotifications finds and sends all notifications that are due.
func (s *NotificationService) SendDueNotifications(ctx context.Context) {
	notifications, err := s.repo.GetPendingNotifications(ctx)
	if err != nil {
		log.Println("Failed to fetch pending notifications:", err)
		return
	}
	for _, n := range notifications {
		sentTo, err := s.sendNotification(ctx, n)
		status := "sent"
		if err != nil {
			log.Println("Failed to send notification:", err)
			status = "failed"
		}
		s.repo.UpdateNotificationStatus(ctx, n.ID, status, sentTo)
	}
}

// sendNotification sends the notification email to all matching users.
func (s *NotificationService) sendNotification(ctx context.Context, n *Notification) ([]string, error) {
	users, err := s.userRepo.FindByRolesAndFaculties(ctx, n.Roles, n.Faculties)
	if err != nil {
		return nil, err
	}

	// Use environment variable for email subject, with fallback
	subject := os.Getenv("NOTIFICATION_EMAIL_SUBJECT")
	if subject == "" {
		subject = "Notification" // Default subject if not specified
	}

	var sentTo []string
	for _, user := range users {
		err := s.emailService.SendEmail(user.Email, subject, n.Message)
		if err == nil {
			sentTo = append(sentTo, user.Email)
		}
	}
	return sentTo, nil
}

// Why: This service coordinates notification scheduling, user filtering, and email delivery. Scheduling is handled by periodically calling SendDueNotifications (e.g., from a goroutine or cron job).
