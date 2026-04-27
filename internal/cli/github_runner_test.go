// Package cli는 GitHub Actions runner CLI 명령을 테스트합니다.
// Package cli provides tests for GitHub Actions runner CLI commands.
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/github/runner"
)

// TestNewRunnerInstallCmd는 install 서브커맨드를 테스트합니다.
func TestNewRunnerInstallCmd(t *testing.T) {
	t.Parallel()

	cmd := newRunnerInstallCmd()

	if cmd == nil {
		t.Fatal("newRunnerInstallCmd returned nil")
	}

	if cmd.Use != "install" {
		t.Errorf("Expected Use=install, got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Args 검증 (NoArgs는 nil 인터페이스)
	if cmd.Args != nil {
		// cobra.NoArgs는 nil이 아닌 함수이므로 실행 가능한지 확인
		err := cmd.Args(cmd, []string{"arg1"})
		if err == nil {
			t.Error("Expected NoArgs to reject arguments")
		}
	}
}

// TestNewRunnerRegisterCmd는 register 서브커맨드를 테스트합니다.
func TestNewRunnerRegisterCmd(t *testing.T) {
	t.Parallel()

	cmd := newRunnerRegisterCmd()

	if cmd == nil {
		t.Fatal("newRunnerRegisterCmd returned nil")
	}

	// Use는 "register <repo>" 형식
	expectedUse := "register <repo>"
	if cmd.Use != expectedUse {
		t.Errorf("Expected Use=%s, got %s", expectedUse, cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

// TestNewRunnerStartCmd는 start 서브커맨드를 테스트합니다.
func TestNewRunnerStartCmd(t *testing.T) {
	t.Parallel()

	cmd := newRunnerStartCmd()

	if cmd == nil {
		t.Fatal("newRunnerStartCmd returned nil")
	}

	if cmd.Use != "start" {
		t.Errorf("Expected Use=start, got %s", cmd.Use)
	}
}

// TestNewRunnerStopCmd는 stop 서브커맨드를 테스트합니다.
func TestNewRunnerStopCmd(t *testing.T) {
	t.Parallel()

	cmd := newRunnerStopCmd()

	if cmd == nil {
		t.Fatal("newRunnerStopCmd returned nil")
	}

	if cmd.Use != "stop" {
		t.Errorf("Expected Use=stop, got %s", cmd.Use)
	}
}

// TestNewRunnerStatusCmd는 status 서브커맨드를 테스트합니다.
func TestNewRunnerStatusCmd(t *testing.T) {
	t.Parallel()

	cmd := newRunnerStatusCmd()

	if cmd == nil {
		t.Fatal("newRunnerStatusCmd returned nil")
	}

	if cmd.Use != "status" {
		t.Errorf("Expected Use=status, got %s", cmd.Use)
	}
}

// TestNewRunnerUpgradeCmd는 upgrade 서브커맨드를 테스트합니다.
func TestNewRunnerUpgradeCmd(t *testing.T) {
	t.Parallel()

	cmd := newRunnerUpgradeCmd()

	if cmd == nil {
		t.Fatal("newRunnerUpgradeCmd returned nil")
	}

	// Use는 "upgrade <repo>" 형식
	expectedUse := "upgrade <repo>"
	if cmd.Use != expectedUse {
		t.Errorf("Expected Use=%s, got %s", expectedUse, cmd.Use)
	}
}

// --- Mock Factory Functions for Dependency Injection ---

// mockInstallerFactory는 테스트용 Installer 생성자입니다.
func mockInstallerFactory() *runner.Installer {
	return runner.NewInstaller("/tmp", nil)
}

// mockRegistrarFactory는 테스트용 Registrar 생성자입니다.
func mockRegistrarFactory() *runner.Registrar {
	return runner.NewRegistrar("/tmp", nil)
}

// mockServiceManagerFactory는 테스트용 ServiceManager 생성자입니다.
func mockServiceManagerFactory() runner.ServiceManager {
	return runner.NewLaunchdManager("/tmp", nil)
}

// mockVersionCheckerFactory는 테스트용 VersionChecker 생성자입니다.
func mockVersionCheckerFactory() *runner.VersionChecker {
	return runner.NewVersionChecker("/tmp", nil)
}
