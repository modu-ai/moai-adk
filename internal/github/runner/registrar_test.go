// Package runner는 GitHub Actions runner 등록 기능을 테스트합니다.
// Package runner provides tests for GitHub Actions runner registration.
package runner

import (
	"context"
	"testing"
)

// MockGHExecutor는 테스트용 GHExecutor 인터페이스 구현입니다.
// MockGHExecutor is a test implementation of GHExecutor interface.
type MockGHExecutor struct {
	// RunGHFunc는 RunGH 호출 시 실행됩니다.
	RunGHFunc func(ctx context.Context, args ...string) error

	// RunGHOutputFunc는 RunGHOutput 호출 시 실행됩니다.
	RunGHOutputFunc func(ctx context.Context, args ...string) (string, error)
}

// RunGH는 RunGHFunc를 실행합니다. nil이면 에러를 반환합니다.
func (m *MockGHExecutor) RunGH(ctx context.Context, args ...string) error {
	if m.RunGHFunc == nil {
		return nil
	}
	return m.RunGHFunc(ctx, args...)
}

// RunGHOutput은 RunGHOutputFunc를 실행합니다. nil이면 빈 문자열과 nil을 반환합니다.
func (m *MockGHExecutor) RunGHOutput(ctx context.Context, args ...string) (string, error) {
	if m.RunGHOutputFunc == nil {
		return "", nil
	}
	return m.RunGHOutputFunc(ctx, args...)
}

// TestRegistrar_RegisterRunner_Success는 정상 등록을 테스트합니다.
func TestRegistrar_RegisterRunner_Success(t *testing.T) {
	t.Parallel()

	// GIVEN: 성공적인 token 응답과 config.sh 실행을 mock
	mockExec := &MockGHExecutor{
		RunGHOutputFunc: func(ctx context.Context, args ...string) (string, error) {
			if args[0] == "api" {
				// token 응답
				return `{"token": "AABBCCDD"}`, nil
			}
			return "", nil
		},
		RunGHFunc: func(ctx context.Context, args ...string) error {
			// config.sh 실행 성공
			return nil
		},
	}

	reg := NewRegistrar("/tmp/actions-runner", mockExec)

	// WHEN: runner 등록
	result, err := reg.RegisterRunner(context.Background(), "owner/repo", []string{"self-hosted", "macos"})

	// THEN: 성공 확인
	if err != nil {
		t.Fatalf("RegisterRunner failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected Success=true, got false")
	}

	if result.SettingsURL != "https://github.com/owner/repo/settings/actions/runners" {
		t.Errorf("Unexpected settings URL: %s", result.SettingsURL)
	}

	if len(result.Labels) != 2 || result.Labels[0] != "self-hosted" {
		t.Errorf("Unexpected labels: %v", result.Labels)
	}
}

// TestRegistrar_RegisterRunner_TokenError는 token 획득 실패를 테스트합니다.
func TestRegistrar_RegisterRunner_TokenError(t *testing.T) {
	t.Parallel()

	mockExec := &MockGHExecutor{
		RunGHOutputFunc: func(ctx context.Context, args ...string) (string, error) {
			return "", context.DeadlineExceeded
		},
	}

	reg := NewRegistrar("/tmp/actions-runner", mockExec)

	_, err := reg.RegisterRunner(context.Background(), "owner/repo", []string{})

	if err == nil {
		t.Fatal("Expected error for token failure, got nil")
	}
}

// TestRegistrar_RegisterRunner_ConfigError는 config.sh 실행 실패를 테스트합니다.
func TestRegistrar_RegisterRunner_ConfigError(t *testing.T) {
	t.Parallel()

	mockExec := &MockGHExecutor{
		RunGHOutputFunc: func(ctx context.Context, args ...string) (string, error) {
			return `{"token": "TESTTOKEN"}`, nil
		},
		RunGHFunc: func(ctx context.Context, args ...string) error {
			return context.Canceled
		},
	}

	reg := NewRegistrar("/tmp/actions-runner", mockExec)

	_, err := reg.RegisterRunner(context.Background(), "owner/repo", []string{})

	if err == nil {
		t.Fatal("Expected error for config.sh failure, got nil")
	}
}

// TestNewRegistrar는 Registrar 생성자를 테스트합니다.
func TestNewRegistrar(t *testing.T) {
	t.Parallel()

	reg := NewRegistrar("/test/dir", nil)

	if reg == nil {
		t.Fatal("NewRegistrar returned nil")
	}

	if reg.ghRunnerDir != "/test/dir" {
		t.Errorf("Expected ghRunnerDir=/test/dir, got %s", reg.ghRunnerDir)
	}

	if reg.executor == nil {
		// 기본 executor가 설정되어야 함
		t.Error("Expected default executor to be set")
	}
}
