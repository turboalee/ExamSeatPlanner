package auth

import (
	"ExamSeatPlanner/internal/config"
	"context"
	"errors"
	"fmt"
	"log"

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

func (s *UserService) RegisterUser(ctx context.Context, req RegisterRequest) error {
	existingUser, err := s.repo.FindByCMS(ctx, req.CMSID)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("CMS ID already registered")
	}

	hashPassword, err := HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &User{
		ID:           primitive.NewObjectID(),
		CMSID:        req.CMSID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashPassword,
		Verified:     false,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return err
	}
	token, _ := GenerateJWT(user.CMSID)
	err = s.authService.SendVerificationEmail(user.Email, token)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, cred Credential) (string, error) {
	user, err := s.repo.FindByCMS(ctx, cred.CMSID)

	if err != nil {
		return "", errors.New("Invalid Credentials")
	}

	if !CheckPasswordHash(cred.Password, user.PasswordHash) {
		return "", errors.New("Invalid Credentials")
	}

	token, err := GenerateJWT(user.CMSID)
	if err != nil {
		return "", errors.New("Token not generated")
	}
	return token, nil
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	cmsID, err := ValidateJWT(token)
	if err != nil {
		return errors.New("Invalid token")
	}
	user, err := s.repo.FindByCMS(ctx, cmsID)
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
	resetToken, _ := GenerateJWT(user.CMSID)
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
	cmsID, err := ValidateJWT(token)
	if err != nil {
		return errors.New("Invalid Token")
	}

	user, err := s.repo.FindByCMS(ctx, cmsID)
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
