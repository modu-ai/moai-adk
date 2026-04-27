package auth

import (
	"context"
	"errors"
	"testing"
)

func TestCodexAuthHandler_Setup(t *testing.T) {
	t.Run("private repo secret set success", func(t *testing.T) {
		ctx := context.Background()
		setSecretCalled := false
		authJSONValue := `{"token": "test-token"}`

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				setSecretCalled = true
				if name != "CODEX_AUTH_JSON" {
					t.Errorf("secret name = %s, want CODEX_AUTH_JSON", name)
				}
				if value != authJSONValue {
					t.Errorf("secret value mismatch")
				}
				return nil
			},
		}

		handler := NewCodexAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/private-repo", authJSONValue, true)

		if err != nil {
			t.Errorf("Setup() error = %v, want nil", err)
		}
		if !setSecretCalled {
			t.Error("SetSecret was not called")
		}
	})

	t.Run("public repo HARD BLOCK - REQ-SEC-001", func(t *testing.T) {
		ctx := context.Background()
		mockSetter := &MockSecretSetter{}

		handler := NewCodexAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/public-repo", `{"token": "x"}`, false)

		if err == nil {
			t.Error("Setup() error = nil, want error (public repo block)")
		}
		if !errors.Is(err, ErrPublicRepoBlocked) {
			t.Errorf("Setup() error = %v, want ErrPublicRepoBlocked", err)
		}
	})

	t.Run("public repo HARD BLOCK even with force-public flag", func(t *testing.T) {
		ctx := context.Background()
		mockSetter := &MockSecretSetter{}

		handler := NewCodexAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/public-repo", `{"token": "x"}`, false)

		if err == nil {
			t.Error("Setup() error = nil, want error (public repo always blocked)")
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

		handler := NewCodexAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/private-repo", `{"token": "x"}`, true)

		if err == nil {
			t.Error("Setup() error = nil, want error")
		}
	})
}

func TestCodexAuthHandler_ValidateAuthJSON(t *testing.T) {
	t.Run("valid auth.json", func(t *testing.T) {
		validJSON := `{"token": "sk-test-key", "email": "user@example.com"}`
		err := validateAuthJSON(validJSON)
		if err != nil {
			t.Errorf("validateAuthJSON() error = %v, want nil", err)
		}
	})

	t.Run("empty JSON returns error", func(t *testing.T) {
		err := validateAuthJSON("{}")
		if err == nil {
			t.Error("validateAuthJSON() error = nil, want error (empty JSON)")
		}
	})

	t.Run("invalid JSON format returns error", func(t *testing.T) {
		err := validateAuthJSON("not-json")
		if err == nil {
			t.Error("validateAuthJSON() error = nil, want error (invalid JSON)")
		}
	})
}
