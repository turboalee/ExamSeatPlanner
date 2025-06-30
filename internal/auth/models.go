package auth

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	CMSID        string             `bson:"cms_id"` // CMS ID for students (required), empty for staff/admin
	Name         string             `bson:"name"`
	Email        string             `bson:"email"` // Email for notifications (personal or institute for students, institute only for staff/admin)
	PasswordHash string             `bson:"password_hash"`
	Verified     bool               `bson:"verified"`
	ResetToken   string             `bson:"reset_token,omitempty"`
	Role         string             `bson:"role"`       // Role is required for RBAC (admin, staff, student)
	Faculty      string             `bson:"faculty"`    // Faculty is needed for notification targeting and grouping
	Department   string             `bson:"department"` // Department is needed for seating algorithms and grouping
	Batch        string             `bson:"batch"`      // Batch is needed for seating algorithms and grouping
}

type RegisterRequest struct {
	CMSID      string `json:"cms_id,omitempty"` // CMS ID required for students, omit for staff/admin
	Name       string `json:"name"`
	Email      string `json:"email"` // Email for notifications (any email for students, institute only for staff/admin)
	Password   string `json:"password"`
	Role       string `json:"role"`       // Role is required at registration to assign permissions
	Faculty    string `json:"faculty"`    // Faculty is required to group users for notifications and seating
	Department string `json:"department"` // Department is required for seating and grouping
	Batch      string `json:"batch"`      // Batch is required for seating and grouping
}

type Credential struct {
	Identifier string `json:"identifier"` // CMS ID for students, email for staff/admin
	Password   string `json:"password"`
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
