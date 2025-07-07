package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

type JWTClaims struct {
	Name       string `json:"name"`
	Email      string `json:"email"`            // Primary identifier for staff/admin, secondary for students
	CMSID      string `json:"cms_id,omitempty"` // CMS ID only for students, omit for staff/admin
	Role       string `json:"role"`             // Role is needed for RBAC in protected endpoints
	Faculty    string `json:"faculty"`          // Faculty is needed for notification and grouping
	Department string `json:"department"`       // Department is needed for seating/grouping in protected endpoints
	Batch      string `json:"batch"`            // Batch is needed for seating/grouping in protected endpoints
	jwt.RegisteredClaims
}

func GenerateJWT(name, email, cmsID, role, faculty, department, batch string, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		Name:       name,
		Email:      email,
		CMSID:      cmsID,
		Role:       role,
		Faculty:    faculty,
		Department: department,
		Batch:      batch,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return "", errors.New("token expired")
	}
	return claims.Email, nil
}

func GetJWTKey() []byte {
	return jwtKey
}

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
