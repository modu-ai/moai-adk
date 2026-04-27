package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrPublicRepoBlocked는 public repo에서 Codex 사용이 차단됨을 나타냅니다.
var ErrPublicRepoBlocked = errors.New("Codex auth blocked: OpenAI policy prohibits public repo usage (REQ-SEC-001)")

// CodexAuthHandler는 Codex 인증을 처리합니다 (private repo guard).
type CodexAuthHandler struct {
	secrets SecretSetter
}

// NewCodexAuthHandler는 새로운 CodexAuthHandler를 생성합니다.
func NewCodexAuthHandler(secrets SecretSetter) *CodexAuthHandler {
	return &CodexAuthHandler{
		secrets: secrets,
	}
}

// Setup은 Codex를 인증하고 auth.json을 GitHub 시크릿으로 저장합니다.
// public repo에서는 HARD BLOCK됩니다 (REQ-SEC-001, REQ-CI-007).
func (h *CodexAuthHandler) Setup(ctx context.Context, repo, authJSON string, isPrivate bool) error {
	// HARD BLOCK: public repo는 항상 차단
	// --force-public 플래그가 있어도 block됨 (HARD requirement)
	if !isPrivate {
		return fmt.Errorf("%w: see https://openai.com/policies for OpenAI policy on public repositories", ErrPublicRepoBlocked)
	}

	// auth.json 검증
	if err := validateAuthJSON(authJSON); err != nil {
		return fmt.Errorf("codex setup: %w", err)
	}

	// gh secret set CODEX_AUTH_JSON -R REPO
	if err := h.secrets.SetSecret(ctx, repo, "CODEX_AUTH_JSON", authJSON); err != nil {
		return fmt.Errorf("codex setup: %w", err)
	}

	fmt.Println("Codex 인증이 완료되었습니다 (private repo only).")
	fmt.Println("template metadata는 조건부 시딩 패턴을 따릅니다 (REQ-CI-009.2).")

	return nil
}

// validateAuthJSON은 auth.json 내용을 검증합니다.
// REQ-CI-009.1: 파일 권한 600 검증은 호출자가 수행해야 합니다.
func validateAuthJSON(authJSON string) error {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(authJSON), &parsed); err != nil {
		return fmt.Errorf("invalid auth.json: %w", err)
	}

	// token 필드가 비어 있으면 에러
	if token, ok := parsed["token"]; !ok || token == "" {
		return errors.New("auth.json must contain non-empty 'token' field")
	}

	return nil
}
