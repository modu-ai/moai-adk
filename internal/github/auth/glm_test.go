package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGLMAuthHandler_Setup(t *testing.T) {
	t.Run("GLM token set success", func(t *testing.T) {
		ctx := context.Background()
		setSecretCalls := []string{}
		testToken := "test-glm-token-12345"

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				setSecretCalls = append(setSecretCalls, name)

				if name == "GLM_API_KEY" && value != testToken {
					t.Errorf("GLM_API_KEY value mismatch: got %s, want %s", value, testToken)
				}
				return nil
			},
		}

		handler := NewGLMAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", testToken)

		if err != nil {
			t.Errorf("Setup() error = %v, want nil", err)
		}

		// 5 secrets should be set (GLM_API_KEY + 4 env vars)
		expectedCount := 5
		if len(setSecretCalls) != expectedCount {
			t.Errorf("SetSecret call count = %d, want %d", len(setSecretCalls), expectedCount)
		}

		if setSecretCalls[0] != "GLM_API_KEY" {
			t.Errorf("first secret = %s, want GLM_API_KEY", setSecretCalls[0])
		}

		// SPEC-GLM-001 env vars should be set
		expectedEnvVars := []string{
			"DISABLE_BETAS",
			"DISABLE_PROMPT_CACHING",
			"CLAUDE_CODE_USE_bedrock",
			"CLAUDE_CODE_USE_vertex",
		}

		for _, envVar := range expectedEnvVars {
			found := false
			for _, call := range setSecretCalls {
				if strings.Contains(call, envVar) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SPEC-GLM-001 env var %s was not set", envVar)
			}
		}
	})

	t.Run("empty token returns error", func(t *testing.T) {
		ctx := context.Background()
		mockSetter := &MockSecretSetter{}

		handler := NewGLMAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", "")

		if err == nil {
			t.Error("Setup() error = nil, want error (empty token)")
		}
	})

	t.Run("secret set failure returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("secret set failed")

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				return expectedErr
			},
		}

		handler := NewGLMAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", "test-token")

		if err == nil {
			t.Error("Setup() error = nil, want error")
		}
	})
}

func TestValidateGLMToken(t *testing.T) {
	t.Run("valid tokens", func(t *testing.T) {
		validTokens := []string{
			"test-token-12345",
			"sk-valid-key",
			"a",
		}

		for _, token := range validTokens {
			err := validateGLMToken(token)
			if err != nil {
				t.Errorf("validateGLMToken(%s) error = %v, want nil", token, err)
			}
		}
	})

	t.Run("empty token returns error", func(t *testing.T) {
		err := validateGLMToken("")
		if err == nil {
			t.Error("validateGLMToken() error = nil, want error (empty token)")
		}
	})

	t.Run("whitespace-only token returns error", func(t *testing.T) {
		err := validateGLMToken("   ")
		if err == nil {
			t.Error("validateGLMToken() error = nil, want error (whitespace only)")
		}
	})
}
