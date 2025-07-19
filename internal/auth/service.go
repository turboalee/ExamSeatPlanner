package auth

import (
	"ExamSeatPlanner/internal/config"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	EmailService *config.EmailService
}
type UserService struct {
	repo        *UserRepository
	authService *AuthService
}

func NewUserService(repo *UserRepository, authService *AuthService) *UserService {
	return &UserService{repo: repo, authService: authService}
}

func NewAuthService(emailService *config.EmailService) *AuthService {
	return &AuthService{EmailService: emailService}
}

// validateInstituteEmail validates that the email is from an institute domain (only for staff/admin)
// func validateInstituteEmail(email string) error {
// 	// Check if email contains .edu (required for institute emails)
// 	if !strings.Contains(email, ".edu") {
// 		return errors.New("staff and admin emails must be from an institute domain (.edu)")
// 	}
//
// 	// Optional: Check for .pk domain (Pakistan institutes)
// 	if strings.HasSuffix(email, ".pk") && !strings.Contains(email, ".edu.pk") {
// 		return errors.New("Pakistan institute emails must follow .edu.pk format")
// 	}
//
// 	return nil
// }

func (s *UserService) RegisterUser(ctx context.Context, req RegisterRequest) error {
	// Validate email format for staff/admin (must be institute email)
	// if req.Role == "admin" || req.Role == "staff" {
	// 	if err := validateInstituteEmail(req.Email); err != nil {
	// 		return err
	// 	}
	// }

	// Check if user already exists by email
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email already registered")
	}

	// For students, also check CMS ID
	if req.Role == "student" {
		if req.CMSID == "" {
			return errors.New("student ID is required for student registration")
		}
		existingStudent, err := s.repo.FindByCMS(ctx, req.CMSID)
		if err != nil {
			return err
		}
		if existingStudent != nil {
			return errors.New("student ID already registered")
		}
	}

	if req.Role == "student" {
		if req.Batch == "" {
			return errors.New("batch is required for students")
		}
	}

	hashPassword, err := HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &User{
		ID:           primitive.NewObjectID(),
		CMSID:        req.CMSID, // Required for students, empty for staff/admin
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashPassword,
		Verified:     false,
		Role:         req.Role,       // Set role from registration for RBAC
		Faculty:      req.Faculty,    // Set faculty from registration for grouping/notifications
		Department:   req.Department, // Set department for seating/grouping
		Batch:        req.Batch,      // Set batch for seating/grouping
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return err
	}
	token, _ := GenerateJWT(user.Name, user.Email, user.CMSID, user.Role, user.Faculty, user.Department, user.Batch, time.Hour*24) // Include name, email and CMS ID for JWT
	err = s.authService.SendVerificationEmail(user.Email, token)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, cred Credential) (string, error) {
	var user *User
	var err error

	log.Printf("Login attempt: identifier=%s", cred.Identifier)

	// Determine if identifier is CMS ID (for students) or email (for staff/admin)
	if strings.Contains(cred.Identifier, "@") {
		// It's an email, find by email (for staff/admin)
		user, err = s.repo.FindByEmail(ctx, cred.Identifier)
		log.Printf("FindByEmail: %v", user)
	} else {
		// It's a CMS ID, find by CMS ID (for students)
		user, err = s.repo.FindByCMS(ctx, cred.Identifier)
		log.Printf("FindByCMS: %v", user)
	}

	if user != nil {
		log.Printf("User found: email=%s, cms_id=%s, role=%s", user.Email, user.CMSID, user.Role)
	} else {
		log.Printf("No user found for identifier: %s", cred.Identifier)
	}

	if err != nil || user == nil || !CheckPasswordHash(cred.Password, user.PasswordHash) {
		log.Printf("Invalid credentials for identifier: %s", cred.Identifier)
		return "", errors.New("invalid Credentials")
	}

	if !user.Verified {
		log.Printf("Email not verified for user: %s", user.Email)
		return "", errors.New(("email not verified"))
	}

	token, err := GenerateJWT(user.Name, user.Email, user.CMSID, user.Role, user.Faculty, user.Department, user.Batch, time.Hour*24) // Include name, email and CMS ID for JWT
	if err != nil {
		log.Printf("Token not generated for user: %s", user.Email)
		return "", errors.New("token not generated")
	}
	log.Printf("JWT generated for user: %s, role: %s", user.Email, user.Role)
	return token, nil
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	email, err := ValidateJWT(token)
	if err != nil {
		return errors.New("invalid token")
	}
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("User not found")
	}
	user.Verified = true
	return s.repo.UpdateUser(ctx, user)
}

func (s *UserService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("User not found")
	}
	resetToken, _ := GenerateJWT(user.Name, user.Email, user.CMSID, user.Role, user.Faculty, user.Department, user.Batch, time.Minute*15) // Include name, email and CMS ID for JWT
	user.ResetToken = resetToken
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	user.ResetToken = resetToken

	if err := s.authService.SendResetPasswordEmail(email, resetToken); err != nil {
		log.Println("Email sending error:", err)
		return errors.New("failed to send reset password email")
	}
	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	email, err := ValidateJWT(token)
	if err != nil {
		return errors.New("invalid Token")
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("User not found")
	}
	hashPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hashPassword
	user.ResetToken = ""
	return s.repo.UpdateUser(ctx, user)
}

func (a *AuthService) SendVerificationEmail(email, token string) error {
	subject := "Email Verification"
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // fallback for dev
	}
	body := fmt.Sprintf("Click the link to verify your email: %s/verify-email?token=%s", frontendURL, token)
	return a.EmailService.SendEmail(email, subject, body)
}

func (a *AuthService) SendResetPasswordEmail(email, token string) error {
	subject := "Password Reset"
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // fallback for dev
	}
	body := fmt.Sprintf("Click the link to reset your password: %s/reset-password?token=%s", frontendURL, token)
	return a.EmailService.SendEmail(email, subject, body)
}
