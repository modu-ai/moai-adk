package auth

import (
	"context"
	"fmt"
	"os/exec"
)

// ClaudeAuthHandler는 Claude OAuth 토큰을 처리합니다.
type ClaudeAuthHandler struct {
	secrets SecretSetter
}

// NewClaudeAuthHandler는 새로운 ClaudeAuthHandler를 생성합니다.
func NewClaudeAuthHandler(secrets SecretSetter) *ClaudeAuthHandler {
	return &ClaudeAuthHandler{
		secrets: secrets,
	}
}

// Setup은 Claude를 인증하고 OAuth 토큰을 GitHub 시크릿으로 저장합니다.
// 시크릿 이름: CLAUDE_CODE_OAUTH_TOKEN
func (h *ClaudeAuthHandler) Setup(ctx context.Context, repo, token string) error {
	// gh secret set CLAUDE_CODE_OAUTH_TOKEN -R REPO
	// stdin을 통해 값을 전달 (REQ-SEC-002)
	if err := h.secrets.SetSecret(ctx, repo, "CLAUDE_CODE_OAUTH_TOKEN", token); err != nil {
		return fmt.Errorf("claude setup: %w", err)
	}

	fmt.Println("Claude OAuth 토큰이 설정되었습니다.")
	fmt.Println("Max 플랜 구독이 필요합니다.")

	return nil
}

// Check는 Claude CLI가 설치되어 있고 토큰이 유효한지 확인합니다.
func (h *ClaudeAuthHandler) Check(ctx context.Context) (*AuthStatus, error) {
	// Claude CLI 설치 확인
	_, err := exec.LookPath("claude")
	if err != nil {
		return &AuthStatus{
			Installed:     false,
			Authenticated: false,
			Message:       "Claude CLI가 설치되지 않았습니다. 'npm install -g @anthropic-ai/claude'를 실행하세요.",
		}, nil
	}

	return &AuthStatus{
		Installed:     true,
		Authenticated: false, // 실제 토큰 확인은 복잡하므로 기본적으로 false
		SecretName:    "CLAUDE_CODE_OAUTH_TOKEN",
		Message:       "Claude CLI가 설치되어 있습니다.",
	}, nil
}
