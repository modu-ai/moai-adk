package github

import (
	"context"
	"testing"
)

// mockGHError는 테스트용 모의 에러입니다.
type mockGHError struct {
	msg string
}

func (e *mockGHError) Error() string {
	return e.msg
}

// MockGHSecretExecutor는 테스트용 GHSecretExecutor 모의 구현입니다.
type MockGHSecretExecutor struct {
	RunGHFunc         func(ctx context.Context, args ...string) error
	RunGHOutputFunc   func(ctx context.Context, args ...string) (string, error)
	RunGHWithStdinFunc func(ctx context.Context, stdin string, args ...string) error
}

func (m *MockGHSecretExecutor) RunGH(ctx context.Context, args ...string) error {
	if m.RunGHFunc != nil {
		return m.RunGHFunc(ctx, args...)
	}
	return nil
}

func (m *MockGHSecretExecutor) RunGHOutput(ctx context.Context, args ...string) (string, error) {
	if m.RunGHOutputFunc != nil {
		return m.RunGHOutputFunc(ctx, args...)
	}
	return "", nil
}

func (m *MockGHSecretExecutor) RunGHWithStdin(ctx context.Context, stdin string, args ...string) error {
	if m.RunGHWithStdinFunc != nil {
		return m.RunGHWithStdinFunc(ctx, stdin, args...)
	}
	// Fallback to RunGH if no stdin-specific function is set
	return m.RunGH(ctx, args...)
}

func TestGHSecretManager_SetSecret(t *testing.T) {
	t.Run("secret을 stdin 통해 성공적으로 설정", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := &MockGHSecretExecutor{
			RunGHFunc: func(ctx context.Context, args ...string) error {
				// gh secret set NAME -R REPO 형식 검증
				if len(args) < 4 {
					t.Errorf("예상보다 적은 인자: got %d, want at least 4", len(args))
				}
				if args[0] != "secret" || args[1] != "set" || args[2] != "TEST_SECRET" {
					t.Errorf("잘못된 인자: %v", args)
				}
				return nil
			},
		}

		manager := NewGHSecretManager(mockExecutor)
		err := manager.SetSecret(ctx, "owner/repo", "TEST_SECRET", "secret-value")

		if err != nil {
			t.Errorf("SetSecret() error = %v, want nil", err)
		}
	})

	t.Run("gh CLI 실행 실패 시 에러 반환", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := &MockGHSecretExecutor{
			RunGHFunc: func(ctx context.Context, args ...string) error {
				return &mockGHError{"gh command failed"}
			},
		}

		manager := NewGHSecretManager(mockExecutor)
		err := manager.SetSecret(ctx, "owner/repo", "TEST_SECRET", "secret-value")

		if err == nil {
			t.Error("SetSecret() error = nil, want error")
		}
	})
}

func TestGHSecretManager_ListSecrets(t *testing.T) {
	t.Run("secret 목록을 성공적으로 가져옴", func(t *testing.T) {
		ctx := context.Background()
		mockOutput := "SECRET1\tUpdated at 2024-01-01\nSECRET2\tUpdated at 2024-01-02\n"
		mockExecutor := &MockGHSecretExecutor{
			RunGHOutputFunc: func(ctx context.Context, args ...string) (string, error) {
				if len(args) < 4 {
					t.Errorf("예상보다 적은 인자: got %d, want at least 4", len(args))
				}
				if args[0] != "secret" || args[1] != "list" {
					t.Errorf("잘못된 인자: %v", args)
				}
				return mockOutput, nil
			},
		}

		manager := NewGHSecretManager(mockExecutor)
		secrets, err := manager.ListSecrets(ctx, "owner/repo")

		if err != nil {
			t.Errorf("ListSecrets() error = %v, want nil", err)
		}
		if len(secrets) != 2 {
			t.Errorf("ListSecrets() length = %d, want 2", len(secrets))
		}
		if secrets[0] != "SECRET1" {
			t.Errorf("ListSecrets()[0] = %s, want SECRET1", secrets[0])
		}
	})

	t.Run("빈 목록 처리", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := &MockGHSecretExecutor{
			RunGHOutputFunc: func(ctx context.Context, args ...string) (string, error) {
				return "", nil
			},
		}

		manager := NewGHSecretManager(mockExecutor)
		secrets, err := manager.ListSecrets(ctx, "owner/repo")

		if err != nil {
			t.Errorf("ListSecrets() error = %v, want nil", err)
		}
		if len(secrets) != 0 {
			t.Errorf("ListSecrets() length = %d, want 0", len(secrets))
		}
	})

	t.Run("gh CLI 실행 실패 시 에러 반환", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := &MockGHSecretExecutor{
			RunGHOutputFunc: func(ctx context.Context, args ...string) (string, error) {
				return "", &mockGHError{"gh command failed"}
			},
		}

		manager := NewGHSecretManager(mockExecutor)
		_, err := manager.ListSecrets(ctx, "owner/repo")

		if err == nil {
			t.Error("ListSecrets() error = nil, want error")
		}
	})
}

func TestGHSecretManager_DeleteSecret(t *testing.T) {
	t.Run("secret을 성공적으로 삭제", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := &MockGHSecretExecutor{
			RunGHFunc: func(ctx context.Context, args ...string) error {
				if len(args) < 4 {
					t.Errorf("예상보다 적은 인자: got %d, want at least 4", len(args))
				}
				if args[0] != "secret" || args[1] != "delete" || args[2] != "TEST_SECRET" {
					t.Errorf("잘못된 인자: %v", args)
				}
				return nil
			},
		}

		manager := NewGHSecretManager(mockExecutor)
		err := manager.DeleteSecret(ctx, "owner/repo", "TEST_SECRET")

		if err != nil {
			t.Errorf("DeleteSecret() error = %v, want nil", err)
		}
	})
}
