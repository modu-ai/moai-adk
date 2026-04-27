package auth

import (
	"context"
	"fmt"
	"regexp"
)

// GeminiAuthHandler는 Gemini API key를 처리합니다.
type GeminiAuthHandler struct {
	secrets SecretSetter
}

// NewGeminiAuthHandler는 새로운 GeminiAuthHandler를 생성합니다.
func NewGeminiAuthHandler(secrets SecretSetter) *GeminiAuthHandler {
	return &GeminiAuthHandler{
		secrets: secrets,
	}
}

// Setup은 Gemini API key를 GitHub 시크릿으로 저장합니다.
// 키 형식: alphanumeric + dash/underscore, ~39자 (REQ-CI-010.1).
func (h *GeminiAuthHandler) Setup(ctx context.Context, repo, apiKey string) error {
	// API key 형식 검증
	if err := validateGeminiAPIKey(apiKey); err != nil {
		return fmt.Errorf("gemini setup: %w", err)
	}

	// gh secret set GEMINI_API_KEY -R REPO
	if err := h.secrets.SetSecret(ctx, repo, "GEMINI_API_KEY", apiKey); err != nil {
		return fmt.Errorf("gemini setup: %w", err)
	}

	maskedKey := maskGeminiKey(apiKey)
	fmt.Printf("Gemini API key가 설정되었습니다: %s\n", maskedKey)
	fmt.Println("Free tier 제한에 주의하세요 (REQ-CI-010.2).")

	return nil
}

// validateGeminiAPIKey는 Gemini API key 형식을 검증합니다.
// REQ-CI-010.1: alphanumeric + dash/underscore, ~39자
func validateGeminiAPIKey(key string) error {
	if key == "" {
		return fmt.Errorf("API key is empty")
	}

	// Gemini API key는 보통 "AIza"로 시작하고 ~39자입니다
	// 형식: alphanumeric + dash + underscore
	pattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !pattern.MatchString(key) {
		return fmt.Errorf("API key contains invalid characters (only alphanumeric, dash, underscore allowed)")
	}

	// 최소 길이 검증
	if len(key) < 8 {
		return fmt.Errorf("API key too short (minimum 8 characters)")
	}

	return nil
}

// maskGeminiKey는 API key를 출력용으로 마스킹합니다.
// 첫 문자 + 마지막 4자만 표시.
func maskGeminiKey(key string) string {
	if len(key) <= 4 {
		return "***"
	}
	// 첫 문자 + 마지막 4자만 표시
	return key[:1] + "..." + key[len(key)-4:]
}
