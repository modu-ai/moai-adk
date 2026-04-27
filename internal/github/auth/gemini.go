package auth

import (
	"context"
	"fmt"
	"regexp"
)

// GeminiAuthHandler handles Gemini API key setup.
type GeminiAuthHandler struct {
	secrets SecretSetter
}

// NewGeminiAuthHandler creates a new GeminiAuthHandler.
func NewGeminiAuthHandler(secrets SecretSetter) *GeminiAuthHandler {
	return &GeminiAuthHandler{
		secrets: secrets,
	}
}

// Setup stores the Gemini API key as a GitHub secret.
// Key format: alphanumeric + dash/underscore, ~39 chars (REQ-CI-010.1).
func (h *GeminiAuthHandler) Setup(ctx context.Context, repo, apiKey string) error {
	if err := validateGeminiAPIKey(apiKey); err != nil {
		return fmt.Errorf("gemini setup: %w", err)
	}

	if err := h.secrets.SetSecret(ctx, repo, "GEMINI_API_KEY", apiKey); err != nil {
		return fmt.Errorf("gemini setup: %w", err)
	}

	maskedKey := maskGeminiKey(apiKey)
	fmt.Printf("Gemini API key has been set: %s\n", maskedKey)
	fmt.Println("Note: Free tier limits apply (REQ-CI-010.2).")

	return nil
}

// validateGeminiAPIKey validates the Gemini API key format.
// REQ-CI-010.1: alphanumeric + dash/underscore, ~39 chars.
func validateGeminiAPIKey(key string) error {
	if key == "" {
		return fmt.Errorf("API key is empty")
	}

	// Gemini API keys typically start with "AIza" and are ~39 characters long.
	pattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !pattern.MatchString(key) {
		return fmt.Errorf("API key contains invalid characters (only alphanumeric, dash, underscore allowed)")
	}

	if len(key) < 8 {
		return fmt.Errorf("API key too short (minimum 8 characters)")
	}

	return nil
}

// maskGeminiKey masks an API key for display output.
// Shows first character and last 4 characters only.
func maskGeminiKey(key string) string {
	if len(key) <= 4 {
		return "***"
	}
	return key[:1] + "..." + key[len(key)-4:]
}
