package notification

import (
	"context"
	"net/http"
	"time"

	"ExamSeatPlanner/internal/auth"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// ListNotifications handles GET /api/notifications
func (h *NotificationHandler) ListNotifications(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.JWTClaims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	faculty := claims.Faculty
	role := claims.Role
	// Optionally, allow admin to see all, or filter by faculty/role
	notifications, err := h.service.ListNotifications(c.Request().Context(), faculty, role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch notifications"})
	}
	return c.JSON(http.StatusOK, notifications)
}

// DeleteNotification handles DELETE /api/notifications/:id
func (h *NotificationHandler) DeleteNotification(c echo.Context) error {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid notification id"})
	}
	err = h.service.DeleteNotification(c.Request().Context(), objID)
	if err != nil {
		if err.Error() == "not found" {
			return c.JSON(404, map[string]string{"error": "Notification not found"})
		}
		return c.JSON(500, map[string]string{"error": "Failed to delete notification: " + err.Error()})
	}
	return c.JSON(200, map[string]string{"message": "Notification deleted successfully"})
}

// Why: This handler provides the HTTP interface for scheduling notifications, with validation to ensure send times are in the future.
