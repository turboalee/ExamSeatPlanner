package auth

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	CMSID        string             `bson:"cms_id"`
	Name         string             `bson:"name"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"password_hash"`
	Verified     bool               `bson:"verified"`
	ResetToken   string             `bson:"reset_token,omitempty"`
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

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}
