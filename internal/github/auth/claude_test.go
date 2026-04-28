package auth

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

// MockSecretSetter is a test mock for the SecretSetter interface.
type MockSecretSetter struct {
	SetSecretFunc func(ctx context.Context, repo, name, value string) error
}

func (m *MockSecretSetter) SetSecret(ctx context.Context, repo, name, value string) error {
	if m.SetSecretFunc != nil {
		return m.SetSecretFunc(ctx, repo, name, value)
	}
	return nil
}

func TestClaudeAuthHandler_Check(t *testing.T) {
	t.Run("claude CLI installed", func(t *testing.T) {
		if _, err := exec.LookPath("claude"); err != nil {
			t.Skip("skipping: claude CLI not available")
		}

		handler := NewClaudeAuthHandler(&MockSecretSetter{})
		status, err := handler.Check(context.Background())

		if err != nil {
			t.Errorf("Check() error = %v, want nil", err)
		}
		if !status.Installed {
			t.Error("Check().Installed = false, want true")
		}
	})

	t.Run("token exists", func(t *testing.T) {
		t.Skip("requires actual claude CLI")
	})
}

func TestClaudeAuthHandler_Setup(t *testing.T) {
	t.Run("secret set success", func(t *testing.T) {
		ctx := context.Background()
		setSecretCalled := false

		mockSetter := &MockSecretSetter{
			SetSecretFunc: func(ctx context.Context, repo, name, value string) error {
				setSecretCalled = true
				if name != "CLAUDE_CODE_OAUTH_TOKEN" {
					t.Errorf("secret name = %s, want CLAUDE_CODE_OAUTH_TOKEN", name)
				}
				if repo != "owner/repo" {
					t.Errorf("repo = %s, want owner/repo", repo)
				}
				return nil
			},
		}

		handler := NewClaudeAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", "test-token-value")

		if err != nil {
			t.Errorf("Setup() error = %v, want nil", err)
		}
		if !setSecretCalled {
			t.Error("SetSecret was not called")
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

		handler := NewClaudeAuthHandler(mockSetter)
		err := handler.Setup(ctx, "owner/repo", "test-token")

		if err == nil {
			t.Error("Setup() error = nil, want error")
		}
		if !errors.Is(err, expectedErr) {
			t.Errorf("Setup() error = %v, want %v", err, expectedErr)
		}
	})
}
