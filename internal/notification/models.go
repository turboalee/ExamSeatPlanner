package notification

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Notification represents a scheduled email notification.
type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // Unique identifier for the notification
	Message     string             `bson:"message"`      // The email message to be sent
	SendTime    time.Time          `bson:"send_time"`    // When the email should be sent (scheduled)
	Roles       []string           `bson:"roles"`        // Target user roles (admin, staff, student)
	Faculties   []string           `bson:"faculties"`    // Target faculties for filtering recipients
	Status      string             `bson:"status"`       // Status: scheduled, sent, failed, etc.
	CreatedAt   time.Time          `bson:"created_at"`   // When the notification was created
	UpdatedAt   time.Time          `bson:"updated_at"`   // When the notification was last updated
	SentTo      []string           `bson:"sent_to"`      // List of user emails the notification was sent to (for audit)
}

// Why: This model allows us to persist and track scheduled email notifications, including their target audience and delivery status.
