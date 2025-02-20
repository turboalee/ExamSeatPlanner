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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Request"})
	}

	token, err := h.service.AuthenticateUser(context.Background(), cred)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) Profile(c echo.Context) error {
	claims := c.Get("user").(*JWTClaims)
	log.Println("Profile requested for user:", claims)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Authenticated User",
		"cms_id":  claims.CMSID,
	})
}
