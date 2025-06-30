package auth

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{collection: db.Collection("users")}
}

func (r *UserRepository) FindByCMS(ctx context.Context, cmsID string) (*User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"cms_id": cmsID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("User not found")
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("CMS ID already exists")
		}
		return err
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("User not found")
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// FindByRolesAndFaculties finds users matching any of the given roles and faculties.
func (r *UserRepository) FindByRolesAndFaculties(ctx context.Context, roles, faculties []string) ([]*User, error) {
	filter := bson.M{"role": bson.M{"$in": roles}, "faculty": bson.M{"$in": faculties}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var users []*User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
