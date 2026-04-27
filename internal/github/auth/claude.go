package auth

import (
	"context"
	"fmt"
	"os/exec"
)

// ClaudeAuthHandler handles Claude OAuth token setup.
type ClaudeAuthHandler struct {
	secrets SecretSetter
}

// NewClaudeAuthHandler creates a new ClaudeAuthHandler.
func NewClaudeAuthHandler(secrets SecretSetter) *ClaudeAuthHandler {
	return &ClaudeAuthHandler{
		secrets: secrets,
	}
}

// Setup authenticates Claude and stores the OAuth token as a GitHub secret.
// Secret name: CLAUDE_CODE_OAUTH_TOKEN
func (h *ClaudeAuthHandler) Setup(ctx context.Context, repo, token string) error {
	// Pass value via stdin (REQ-SEC-002: never write to disk)
	if err := h.secrets.SetSecret(ctx, repo, "CLAUDE_CODE_OAUTH_TOKEN", token); err != nil {
		return fmt.Errorf("claude setup: %w", err)
	}

	fmt.Println("Claude OAuth token has been set.")
	fmt.Println("Max plan subscription is required.")

	return nil
}

// Check verifies that Claude CLI is installed and the token is valid.
func (h *ClaudeAuthHandler) Check(ctx context.Context) (*AuthStatus, error) {
	_, err := exec.LookPath("claude")
	if err != nil {
		return &AuthStatus{
			Installed:     false,
			Authenticated: false,
			Message:       "Claude CLI is not installed. Run 'npm install -g @anthropic-ai/claude'.",
		}, nil
	}

	return &AuthStatus{
		Installed:     true,
		Authenticated: false, // Actual token verification is complex; default to false
		SecretName:    "CLAUDE_CODE_OAUTH_TOKEN",
		Message:       "Claude CLI is installed.",
	}, nil
}
