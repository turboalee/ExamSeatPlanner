package auth

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
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
	}

	return s.repo.CreateUser(ctx, user)
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
