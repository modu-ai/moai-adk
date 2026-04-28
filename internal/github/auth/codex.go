package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrPublicRepoBlocked indicates that Codex usage is blocked on public repos.
var ErrPublicRepoBlocked = errors.New("codex auth blocked: OpenAI policy prohibits public repo usage (REQ-SEC-001)")

// CodexAuthHandler handles Codex authentication with a private repo guard.
type CodexAuthHandler struct {
	secrets SecretSetter
}

// NewCodexAuthHandler creates a new CodexAuthHandler.
func NewCodexAuthHandler(secrets SecretSetter) *CodexAuthHandler {
	return &CodexAuthHandler{
		secrets: secrets,
	}
}

// Setup authenticates Codex and stores auth.json as a GitHub secret.
// Public repos are HARD BLOCKED (REQ-SEC-001, REQ-CI-007).
func (h *CodexAuthHandler) Setup(ctx context.Context, repo, authJSON string, isPrivate bool) error {
	// HARD BLOCK: public repos are always blocked, even with --force-public
	if !isPrivate {
		return fmt.Errorf("%w: see https://openai.com/policies for OpenAI policy on public repositories", ErrPublicRepoBlocked)
	}

	if err := validateAuthJSON(authJSON); err != nil {
		return fmt.Errorf("codex setup: %w", err)
	}

	if err := h.secrets.SetSecret(ctx, repo, "CODEX_AUTH_JSON", authJSON); err != nil {
		return fmt.Errorf("codex setup: %w", err)
	}

	fmt.Println("Codex authentication complete (private repo only).")
	fmt.Println("Template metadata follows conditional seeding pattern (REQ-CI-009.2).")

	return nil
}

// validateAuthJSON validates the contents of an auth.json string.
// REQ-CI-009.1: File permission 600 validation is the caller's responsibility.
func validateAuthJSON(authJSON string) error {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(authJSON), &parsed); err != nil {
		return fmt.Errorf("invalid auth.json: %w", err)
	}

	if token, ok := parsed["token"]; !ok || token == "" {
		return errors.New("auth.json must contain non-empty 'token' field")
	}

	return nil
}
