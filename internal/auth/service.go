package auth

import (
	"ExamSeatPlanner/internal/config"
	"context"
	"errors"
	"fmt"
	"log"
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
func validateInstituteEmail(email string) error {
	// Check if email contains .edu (required for institute emails)
	if !strings.Contains(email, ".edu") {
		return errors.New("staff and admin emails must be from an institute domain (.edu)")
	}

	// Optional: Check for .pk domain (Pakistan institutes)
	if strings.HasSuffix(email, ".pk") && !strings.Contains(email, ".edu.pk") {
		return errors.New("Pakistan institute emails must follow .edu.pk format")
	}

	return nil
}

func (s *UserService) RegisterUser(ctx context.Context, req RegisterRequest) error {
	// Validate email format for staff/admin (must be institute email)
	if req.Role == "admin" || req.Role == "staff" {
		if err := validateInstituteEmail(req.Email); err != nil {
			return err
		}
	}

	// Check if user already exists by email
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("Email already registered")
	}

	// For students, also check CMS ID
	if req.Role == "student" {
		if req.CMSID == "" {
			return errors.New("Student ID is required for student registration")
		}
		existingStudent, err := s.repo.FindByCMS(ctx, req.CMSID)
		if err != nil {
			return err
		}
		if existingStudent != nil {
			return errors.New("Student ID already registered")
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
	token, _ := GenerateJWT(user.Email, user.CMSID, user.Role, user.Faculty, user.Department, user.Batch, time.Hour*24) // Include email and CMS ID for JWT
	err = s.authService.SendVerificationEmail(user.Email, token)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, cred Credential) (string, error) {
	var user *User
	var err error

	// Determine if identifier is CMS ID (for students) or email (for staff/admin)
	// Try to find by CMS ID first (for students)
	if strings.Contains(cred.Identifier, "@") {
		// It's an email, find by email (for staff/admin)
		user, err = s.repo.FindByEmail(ctx, cred.Identifier)
	} else {
		// It's a CMS ID, find by CMS ID (for students)
		user, err = s.repo.FindByCMS(ctx, cred.Identifier)
	}

	if err != nil || !CheckPasswordHash(cred.Password, user.PasswordHash) {
		return "", errors.New("Invalid Credentials")
	}

	if !user.Verified {
		return "", errors.New(("Email not verified"))
	}

	token, err := GenerateJWT(user.Email, user.CMSID, user.Role, user.Faculty, user.Department, user.Batch, time.Minute*10) // Include email and CMS ID for JWT
	if err != nil {
		return "", errors.New("Token not generated")
	}
	return token, nil
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	email, err := ValidateJWT(token)
	if err != nil {
		return errors.New("Invalid token")
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
	resetToken, _ := GenerateJWT(user.Email, user.CMSID, user.Role, user.Faculty, user.Department, user.Batch, time.Minute*15) // Include email and CMS ID for JWT
	user.ResetToken = resetToken
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	user.ResetToken = resetToken

	if err := s.authService.SendResetPasswordEmail(email, resetToken); err != nil {
		log.Println("Email sending error:", err)
		return errors.New("Failed to send reset password email")
	}
	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	email, err := ValidateJWT(token)
	if err != nil {
		return errors.New("Invalid Token")
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
	body := fmt.Sprintf("Click the link to verify your email: https://yourdomain.com/verify-email?token=%s", token)

	return a.EmailService.SendEmail(email, subject, body)
}

func (a *AuthService) SendResetPasswordEmail(email, token string) error {
	subject := "Password Reset"
	body := fmt.Sprintf("Click the link to reset your password: https://yourdomain.com/reset-password?token=%s", token)

	return a.EmailService.SendEmail(email, subject, body)
}
