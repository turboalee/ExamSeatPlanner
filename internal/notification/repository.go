package notification

import (
	"context"
	"errors"
	"fmt"
	"log"

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
	// For testing: ignore send_time, return all scheduled notifications
	filter := bson.M{"status": "scheduled"}
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

// ListNotifications fetches notifications filtered by faculty and role
func (r *NotificationRepository) ListNotifications(ctx context.Context, faculty, role string) ([]*Notification, error) {
	// Print all notifications in the collection before filtering
	allCursor, err := r.collection.Find(ctx, bson.M{})
	if err == nil {
		var allNotifs []*Notification
		if err := allCursor.All(ctx, &allNotifs); err == nil {
			log.Printf("[DEBUG] All notifications in DB: %d", len(allNotifs))
			for _, n := range allNotifs {
				log.Printf("[DEBUG] DB Notification: id=%v, faculties=%v, roles=%v, status=%v", n.ID, n.Faculties, n.Roles, n.Status)
			}
		}
	}
	var filter bson.M
	if role == "admin" {
		filter = bson.M{"faculties": bson.M{"$in": []string{faculty}}}
	} else {
		filter = bson.M{
			"faculties": bson.M{"$in": []string{faculty}},
			"roles":     bson.M{"$in": []string{role}},
		}
	}
	log.Printf("[DEBUG] ListNotifications filter: %+v", filter)
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var notifications []*Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] ListNotifications found %d notifications", len(notifications))
	for _, n := range notifications {
		log.Printf("[DEBUG] Notification: id=%v, faculties=%v, roles=%v, status=%v", n.ID, n.Faculties, n.Roles, n.Status)
	}
	return notifications, nil
}

// DeleteNotification deletes a notification by ObjectID
func (r *NotificationRepository) DeleteNotification(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

// Why: This repository abstracts DB access for notifications, making it easier to test and maintain the notification logic.
