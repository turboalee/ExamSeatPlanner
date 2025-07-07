package notification

import (
	"ExamSeatPlanner/internal/auth"
	"ExamSeatPlanner/internal/config"
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	log.Printf("[DEBUG] Found %d pending notifications", len(notifications))
	for _, n := range notifications {
		log.Printf("[DEBUG] Processing notification: id=%v, message=%q, roles=%v, faculties=%v, send_time=%v", n.ID, n.Message, n.Roles, n.Faculties, n.SendTime)
		sentTo, err := s.sendNotification(ctx, n)
		if err != nil {
			log.Printf("[ERROR] Failed to send notification %v: %v", n.ID, err)
			continue
		}
		log.Printf("[DEBUG] Notification %v sent to: %v", n.ID, sentTo)
		s.repo.UpdateNotificationStatus(ctx, n.ID, "sent", sentTo)
	}
}

// sendNotification sends the notification email to all matching users.
func (s *NotificationService) sendNotification(ctx context.Context, n *Notification) ([]string, error) {
	log.Printf("[DEBUG] sendNotification: roles=%v, faculties=%v", n.Roles, n.Faculties)
	users, err := s.userRepo.FindByRolesAndFaculties(ctx, n.Roles, n.Faculties)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Found %d users to notify", len(users))

	// Use environment variable for email subject, with fallback
	subject := os.Getenv("NOTIFICATION_EMAIL_SUBJECT")
	if subject == "" {
		subject = "Notification" // Default subject if not specified
	}

	var sentTo []string
	for _, user := range users {
		log.Printf("[DEBUG] Sending email to: %s (%s)", user.Name, user.Email)
		err := s.emailService.SendEmail(user.Email, subject, n.Message)
		if err == nil {
			sentTo = append(sentTo, user.Email)
		}
	}
	return sentTo, nil
}

// ListNotifications fetches notifications filtered by faculty and role
func (s *NotificationService) ListNotifications(ctx context.Context, faculty, role string) ([]*Notification, error) {
	return s.repo.ListNotifications(ctx, faculty, role)
}

// DeleteNotification deletes a notification by ObjectID
func (s *NotificationService) DeleteNotification(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.DeleteNotification(ctx, id)
}

// Why: This service coordinates notification scheduling, user filtering, and email delivery. Scheduling is handled by periodically calling SendDueNotifications (e.g., from a goroutine or cron job).
