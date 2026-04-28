// Package runner는 GitHub Actions runner 서비스 관리 기능을 테스트합니다.
// Package runner provides tests for GitHub Actions runner service management.
package runner

import (
	"context"
	"errors"
	"testing"
)

// MockShellExecutor는 테스트용 ShellExecutor 인터페이스 구현입니다.
// MockShellExecutor is a test implementation of ShellExecutor interface.
type MockShellExecutor struct {
	// RunCommandFunc는 RunCommand 호출 시 실행됩니다.
	RunCommandFunc func(ctx context.Context, dir, name string, args ...string) error

	// RunCommandOutputFunc는 RunCommandOutput 호출 시 실행됩니다.
	RunCommandOutputFunc func(ctx context.Context, dir, name string, args ...string) (string, error)
}

// RunCommand는 RunCommandFunc를 실행합니다.
func (m *MockShellExecutor) RunCommand(ctx context.Context, dir, name string, args ...string) error {
	if m.RunCommandFunc == nil {
		return nil
	}
	return m.RunCommandFunc(ctx, dir, name, args...)
}

// RunCommandOutput은 RunCommandOutputFunc를 실행합니다.
func (m *MockShellExecutor) RunCommandOutput(ctx context.Context, dir, name string, args ...string) (string, error) {
	if m.RunCommandOutputFunc == nil {
		return "", nil
	}
	return m.RunCommandOutputFunc(ctx, dir, name, args...)
}

// TestLaunchdManager_Install은 svc.sh install을 테스트합니다.
func TestLaunchdManager_Install(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{
		RunCommandFunc: func(ctx context.Context, dir, name string, args ...string) error {
			// svc.sh install 호출 확인
			if name != "./svc.sh" {
				t.Errorf("Expected ./svc.sh, got %s", name)
			}
			if len(args) != 1 || args[0] != "install" {
				t.Errorf("Expected install arg, got %v", args)
			}
			return nil
		},
	}

	mgr := NewLaunchdManager("/tmp/actions-runner", mockExec)

	err := mgr.Install(context.Background())
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}
}

// TestLaunchdManager_Start는 svc.sh start를 테스트합니다.
func TestLaunchdManager_Start(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{
		RunCommandFunc: func(ctx context.Context, dir, name string, args ...string) error {
			if args[0] != "start" {
				t.Errorf("Expected start arg, got %v", args)
			}
			return nil
		},
	}

	mgr := NewLaunchdManager("/tmp/actions-runner", mockExec)

	err := mgr.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

// TestLaunchdManager_Stop은 svc.sh stop을 테스트합니다.
func TestLaunchdManager_Stop(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{
		RunCommandFunc: func(ctx context.Context, dir, name string, args ...string) error {
			if args[0] != "stop" {
				t.Errorf("Expected stop arg, got %v", args)
			}
			return nil
		},
	}

	mgr := NewLaunchdManager("/tmp/actions-runner", mockExec)

	err := mgr.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

// TestLaunchdManager_Status_Running은 svc.sh status 실행 중을 테스트합니다.
func TestLaunchdManager_Status_Running(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{
		RunCommandOutputFunc: func(ctx context.Context, dir, name string, args ...string) (string, error) {
			return "Running service", nil
		},
	}

	mgr := NewLaunchdManager("/tmp/actions-runner", mockExec)

	status, err := mgr.Status(context.Background())
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if status != StatusRunning {
		t.Errorf("Expected StatusRunning, got %s", status)
	}
}

// TestLaunchdManager_Status_Stopped는 svc.sh status 중지를 테스트합니다.
func TestLaunchdManager_Status_Stopped(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{
		RunCommandOutputFunc: func(ctx context.Context, dir, name string, args ...string) (string, error) {
			return "Stopped service", nil
		},
	}

	mgr := NewLaunchdManager("/tmp/actions-runner", mockExec)

	status, err := mgr.Status(context.Background())
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if status != StatusStopped {
		t.Errorf("Expected StatusStopped, got %s", status)
	}
}

// TestLaunchdManager_Status_Unknown은 svc.sh status 알 수 없음을 테스트합니다.
func TestLaunchdManager_Status_Unknown(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{
		RunCommandOutputFunc: func(ctx context.Context, dir, name string, args ...string) (string, error) {
			return "Unknown state", nil
		},
	}

	mgr := NewLaunchdManager("/tmp/actions-runner", mockExec)

	status, err := mgr.Status(context.Background())
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if status != StatusUnknown {
		t.Errorf("Expected StatusUnknown, got %s", status)
	}
}

// TestLaunchdManager_CommandError는 명령 실행 실패를 테스트합니다.
func TestLaunchdManager_CommandError(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{
		RunCommandFunc: func(ctx context.Context, dir, name string, args ...string) error {
			return errors.New("command failed")
		},
	}

	mgr := NewLaunchdManager("/tmp/actions-runner", mockExec)

	err := mgr.Install(context.Background())
	if err == nil {
		t.Fatal("Expected error for command failure, got nil")
	}
}

// TestSystemdStub_NotSupported는 systemd 미지원을 테스트합니다.
func TestSystemdStub_NotSupported(t *testing.T) {
	t.Parallel()

	stub := &SystemdStub{}

	ctx := context.Background()

	// 모든 메서드가 "not yet supported" 에러를 반환해야 함
	err := stub.Install(ctx)
	if err == nil || !containsString(err.Error(), "not yet supported") {
		t.Errorf("Install: expected 'not yet supported' error, got %v", err)
	}

	err = stub.Start(ctx)
	if err == nil || !containsString(err.Error(), "not yet supported") {
		t.Errorf("Start: expected 'not yet supported' error, got %v", err)
	}

	err = stub.Stop(ctx)
	if err == nil || !containsString(err.Error(), "not yet supported") {
		t.Errorf("Stop: expected 'not yet supported' error, got %v", err)
	}

	_, err = stub.Status(ctx)
	if err == nil || !containsString(err.Error(), "not yet supported") {
		t.Errorf("Status: expected 'not yet supported' error, got %v", err)
	}
}

// TestNewLaunchdManager는 생성자를 테스트합니다.
func TestNewLaunchdManager(t *testing.T) {
	t.Parallel()

	mockExec := &MockShellExecutor{}
	mgr := NewLaunchdManager("/test/dir", mockExec)

	if mgr == nil {
		t.Fatal("NewLaunchdManager returned nil")
	}

	if mgr.ghRunnerDir != "/test/dir" {
		t.Errorf("Expected ghRunnerDir=/test/dir, got %s", mgr.ghRunnerDir)
	}

	if mgr.executor != mockExec {
		t.Error("Executor not set correctly")
	}
}

// containsString은 부분 문자열 검증 헬퍼 함수입니다.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
