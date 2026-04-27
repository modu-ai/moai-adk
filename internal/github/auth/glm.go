package auth

import (
	"context"
	"fmt"
	"strings"
)

// GLMAuthHandler는 GLM 인증을 처리합니다.
type GLMAuthHandler struct {
	secrets SecretSetter
}

// NewGLMAuthHandler는 새로운 GLMAuthHandler를 생성합니다.
func NewGLMAuthHandler(secrets SecretSetter) *GLMAuthHandler {
	return &GLMAuthHandler{
		secrets: secrets,
	}
}

// Setup은 GLM 인증 토큰을 GitHub 시크릿으로 저장하고
// SPEC-GLM-001 환경 변수 메타데이터를 주입합니다.
func (h *GLMAuthHandler) Setup(ctx context.Context, repo, token string) error {
	// 토큰 검증
	if err := validateGLMToken(token); err != nil {
		return fmt.Errorf("glm setup: %w", err)
	}

	// 1. GLM_API_KEY 시크릿 설정
	if err := h.secrets.SetSecret(ctx, repo, "GLM_API_KEY", token); err != nil {
		return fmt.Errorf("glm setup: %w", err)
	}

	// 2. SPEC-GLM-001 환경 변수 메타데이터 주입
	envVars := map[string]string{
		"DISABLE_BETAS":            "true",
		"DISABLE_PROMPT_CACHING":   "true",
		"CLAUDE_CODE_USE_bedrock":  "0",
		"CLAUDE_CODE_USE_vertex":   "0",
	}

	for name, value := range envVars {
		secretName := fmt.Sprintf("GLM_ENV_%s", name)
		if err := h.secrets.SetSecret(ctx, repo, secretName, value); err != nil {
			return fmt.Errorf("glm setup (env %s): %w", name, err)
		}
	}

	fmt.Println("GLM 인증이 완료되었습니다.")
	fmt.Println("SPEC-GLM-001 환경 변수 메타데이터가 주입되었습니다:")
	fmt.Println("  - DISABLE_BETAS=true")
	fmt.Println("  - DISABLE_PROMPT_CACHING=true")
	fmt.Println("  - CLAUDE_CODE_USE_bedrock=0")
	fmt.Println("  - CLAUDE_CODE_USE_vertex=0")

	return nil
}

// validateGLMToken은 GLM 토큰을 검증합니다.
func validateGLMToken(token string) error {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return fmt.Errorf("GLM token is empty")
	}
	return nil
}
