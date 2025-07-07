package middleware

import (
	"ExamSeatPlanner/internal/auth"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		log.Println("Received Authorization Header:", authHeader)
		if authHeader == "" {
			log.Println("No token provided")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing Token"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)
		log.Println("Token string length:", len(tokenString))

		claims := &auth.JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.GetJWTKey(), nil
		})
		if err != nil {
			log.Println("JWT parsing error:", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid Token"})
		}
		if !token.Valid {
			log.Println("Token is not valid")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid Token"})
		}
		log.Println("JWT claims set:", claims)
		c.Set("user", claims)
		return next(c)
	}
}
