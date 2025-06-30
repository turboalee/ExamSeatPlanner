package notification

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// NotificationHandler handles HTTP requests for notifications.
type NotificationHandler struct {
	service *NotificationService
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(service *NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// ScheduleNotificationRequest represents the request to schedule a notification.
type ScheduleNotificationRequest struct {
	Message   string    `json:"message"`   // The email message to send
	SendTime  time.Time `json:"send_time"` // When to send the email
	Roles     []string  `json:"roles"`     // Target user roles
	Faculties []string  `json:"faculties"` // Target faculties
}

// ScheduleNotification allows admins to schedule a new email notification.
func (h *NotificationHandler) ScheduleNotification(c echo.Context) error {
	var req ScheduleNotificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate that send time is in the future
	if req.SendTime.Before(time.Now()) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Send time must be in the future"})
	}

	notification := &Notification{
		Message:   req.Message,
		SendTime:  req.SendTime,
		Roles:     req.Roles,
		Faculties: req.Faculties,
	}

	err := h.service.ScheduleNotification(context.Background(), notification)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to schedule notification"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Notification scheduled successfully"})
}

// Why: This handler provides the HTTP interface for scheduling notifications, with validation to ensure send times are in the future. 