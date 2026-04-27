package auth

import (
	"context"
	"errors"
	"testing"
)

// MockSecretSetter는 테스트용 SecretSetter 모의 구현입니다.
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
	t.Run("claude CLI 설치됨", func(t *testing.T) {
		handler := NewClaudeAuthHandler(&MockSecretSetter{})
		status, err := handler.Check(context.Background())

		if err != nil {
			t.Errorf("Check() error = %v, want nil", err)
		}
		if !status.Installed {
			t.Error("Check().Installed = false, want true")
		}
	})

	t.Run("토큰이 존재함", func(t *testing.T) {
		// 이 테스트는 실제 환경에서만 작동하므로 skip
		t.Skip("실제 claude CLI가 필요함")
	})
}

func TestClaudeAuthHandler_Setup(t *testing.T) {
	t.Run("secret 설정 성공", func(t *testing.T) {
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
			t.Error("SetSecret이 호출되지 않음")
		}
	})

	t.Run("secret 설정 실패 시 에러 반환", func(t *testing.T) {
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
