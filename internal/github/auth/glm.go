package auth

import (
	"context"
	"fmt"
	"strings"
)

// GLMAuthHandler handles GLM authentication.
type GLMAuthHandler struct {
	secrets SecretSetter
}

// NewGLMAuthHandler creates a new GLMAuthHandler.
func NewGLMAuthHandler(secrets SecretSetter) *GLMAuthHandler {
	return &GLMAuthHandler{
		secrets: secrets,
	}
}

// Setup stores the GLM auth token as a GitHub secret and injects
// SPEC-GLM-001 environment variable metadata.
func (h *GLMAuthHandler) Setup(ctx context.Context, repo, token string) error {
	if err := validateGLMToken(token); err != nil {
		return fmt.Errorf("glm setup: %w", err)
	}

	// 1. Set GLM_API_KEY secret
	if err := h.secrets.SetSecret(ctx, repo, "GLM_API_KEY", token); err != nil {
		return fmt.Errorf("glm setup: %w", err)
	}

	// 2. Inject SPEC-GLM-001 environment variable metadata
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

	fmt.Println("GLM authentication complete.")
	fmt.Println("SPEC-GLM-001 environment variable metadata injected:")
	fmt.Println("  - DISABLE_BETAS=true")
	fmt.Println("  - DISABLE_PROMPT_CACHING=true")
	fmt.Println("  - CLAUDE_CODE_USE_bedrock=0")
	fmt.Println("  - CLAUDE_CODE_USE_vertex=0")

	return nil
}

// validateGLMToken validates a GLM token.
func validateGLMToken(token string) error {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return fmt.Errorf("GLM token is empty")
	}
	return nil
}
