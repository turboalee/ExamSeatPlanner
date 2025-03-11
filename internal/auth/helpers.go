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
	CMSID string `json:"cms_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(cmsID string, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		CMSID: cmsID,
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
		return "", errors.New("Invalid token")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return "", errors.New("Token expired")
	}
	return claims.CMSID, nil
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
