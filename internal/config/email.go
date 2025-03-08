package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/fx"
)

type ResendConfig struct {
	APIKey string
	APIURL string
	From   string
}

func NewResendConfig() *ResendConfig {
	apiKey := os.Getenv("RESEND_API_KEY")
	apiURL := os.Getenv("RESEND_API_URL")
	fromEmail := os.Getenv("FROM_EMAIL")
	if apiKey == "" || apiURL == "" || fromEmail == "" {
		log.Fatal("Missing Environment variables")
	}
	return &ResendConfig{
		APIKey: apiKey,
		APIURL: apiURL,
		From:   fromEmail}
}

type EmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

type EmailService struct {
	Config *ResendConfig
}

func NewEmailService(lc fx.Lifecycle, config *ResendConfig) *EmailService {
	service := &EmailService{Config: config}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Email Service initialized")
			return nil
		},
	})
	return service
}

func (e *EmailService) SendEmail(to, subject, body string) error {
	payload := EmailRequest{
		From:    e.Config.From,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", e.Config.APIURL, bytes.NewBuffer(jsonData))

	if err != nil {
		return fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+e.Config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send Email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return fmt.Errorf("Failed to send email, status code: %d, error: %v", resp.StatusCode, errorResponse)
	}

	log.Println("Email sent successfully to ", to)
	return nil
}
