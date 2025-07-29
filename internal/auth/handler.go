package auth

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service *UserService
}

func NewAuthHandler(service *UserService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Request"})
	}

	err := h.service.RegisterUser(context.Background(), req)
	if err != nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var cred Credential
	if err := c.Bind(&cred); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	token, err := h.service.AuthenticateUser(context.Background(), cred)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	var req VerifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Request"})
	}
	err := h.service.VerifyEmail(context.Background(), req.Token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Email Verified successfully"})
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	err := h.service.ForgotPassword(context.Background(), req.Email)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset Email sent"})
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	err := h.service.ResetPassword(context.Background(), req.Token, req.NewPassword)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Password successfully reset"})
}

func (h *AuthHandler) Profile(c echo.Context) error {
	user := c.Get("user")
	log.Printf("[Profile] user context: %+v", user)
	claims, ok := user.(*JWTClaims)
	if !ok || claims == nil {
		log.Printf("[Profile] Invalid or missing JWT claims: %+v", user)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing token"})
	}
	ctx := c.Request().Context()
	log.Printf("[Profile] Looking up user by email: %s", claims.Email)
	foundUser, err := h.service.repo.FindByEmail(ctx, claims.Email)
	if err != nil {
		log.Printf("[Profile] Error in FindByEmail: %v", err)
	}
	if foundUser == nil {
		log.Printf("[Profile] User not found for email: %s", claims.Email)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}
	log.Printf("[Profile] Found user: %+v", foundUser)
	var userID, cmsID string
	if foundUser.ID.IsZero() {
		log.Printf("[Profile] foundUser.ID is zero value!")
		userID = ""
	} else {
		userID = foundUser.ID.Hex()
	}
	cmsID = foundUser.CMSID
	log.Printf("[Profile] Returning user profile: email=%s, role=%s, faculty=%s, name=%s, _id=%s, cms_id=%s", claims.Email, claims.Role, claims.Faculty, claims.Name, userID, cmsID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Authenticated User",
		"email":   claims.Email,
		"role":    claims.Role,
		"faculty": claims.Faculty,
		"name":    claims.Name,
		"_id":     userID,
		"cms_id":  cmsID,
	})
}
