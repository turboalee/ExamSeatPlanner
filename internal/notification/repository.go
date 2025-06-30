package notification

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NotificationRepository handles DB operations for notifications.
type NotificationRepository struct {
	collection *mongo.Collection
}

// NewNotificationRepository creates a new repository for notifications.
func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{collection: db.Collection("notifications")}
}

// CreateNotification inserts a new notification into the DB.
func (r *NotificationRepository) CreateNotification(ctx context.Context, n *Notification) error {
	_, err := r.collection.InsertOne(ctx, n)
	return err
}

// GetPendingNotifications fetches notifications scheduled to be sent (status = scheduled, send_time <= now).
func (r *NotificationRepository) GetPendingNotifications(ctx context.Context) ([]*Notification, error) {
	filter := bson.M{"status": "scheduled", "send_time": bson.M{"$lte": time.Now()}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var notifications []*Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

// UpdateNotificationStatus updates the status and sent_to fields of a notification.
func (r *NotificationRepository) UpdateNotificationStatus(ctx context.Context, id primitive.ObjectID, status string, sentTo []string) error {
	update := bson.M{"$set": bson.M{"status": status, "sent_to": sentTo}}
	res, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("notification not found")
	}
	return nil
}

// Why: This repository abstracts DB access for notifications, making it easier to test and maintain the notification logic.
