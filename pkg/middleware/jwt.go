package middleware

import (
	"ExamSeatPlanner/internal/auth"
	"log"
	"net/http"
	"strings"

	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var (
	enforcer     *casbin.Enforcer
	enforcerOnce sync.Once
)

// InitCasbinEnforcer initializes the Casbin enforcer singleton.
func InitCasbinEnforcer() (*casbin.Enforcer, error) {
	var err error
	enforcerOnce.Do(func() {
		m, errM := model.NewModelFromFile("rbac_model.conf")
		if errM != nil {
			err = errM
			return
		}
		enforcer, err = casbin.NewEnforcer(m, "rbac_policy.csv")
	})
	return enforcer, err
}

// CasbinMiddleware enforces RBAC using Casbin for each request.
func CasbinMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*auth.JWTClaims)
		if !ok || claims == nil {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized: missing user claims"})
		}
		enf, err := InitCasbinEnforcer()
		if err != nil {
			log.Println("Casbin enforcer error:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "RBAC system error"})
		}
		role := claims.Role
		obj := c.Path() // Use route path for object
		act := c.Request().Method
		allowed, err := enf.Enforce(role, obj, act)
		if err != nil {
			log.Println("Casbin enforce error:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "RBAC system error"})
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: insufficient permissions"})
		}
		return next(c)
	}
}

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

		claims := &auth.JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.GetJWTKey(), nil
		})
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid Token"})
		}
		c.Set("user", claims)
		return next(c)
	}
}
