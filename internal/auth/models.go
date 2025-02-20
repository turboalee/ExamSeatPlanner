package auth

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	CMSID        string             `bson:"cms_id"`
	Name         string             `bson:"name"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"password_hash"`
}

type RegisterRequest struct {
	CMSID    string `json:"cms_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Credential struct {
	CMSID    string `json:"cms_id"`
	Password string `json:"password"`
}
