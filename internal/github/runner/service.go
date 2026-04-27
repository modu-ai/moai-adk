// Package runner는 GitHub Actions runner 서비스 관리 기능을 제공합니다.
// Package runner provides service management for GitHub Actions runners.
package runner

import (
	"context"
	"fmt"
	"strings"
)

// CheckStatus는 서비스 상태를 나타냅니다.
// CheckStatus represents service status.
type CheckStatus string

const (
	// StatusRunning은 서비스 실행 중 상태입니다.
	StatusRunning CheckStatus = "running"
	// StatusStopped은 서비스 중지 상태입니다.
	StatusStopped CheckStatus = "stopped"
	// StatusUnknown은 알 수 없는 상태입니다.
	StatusUnknown CheckStatus = "unknown"
)

// ServiceManager는 runner 서비스 관리 인터페이스입니다.
// ServiceManager interface for runner service management.
type ServiceManager interface {
	Install(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status(ctx context.Context) (CheckStatus, error)
}

// ShellExecutor는 셸 명령 실행을 위한 인터페이스입니다 (테스트용).
// ShellExecutor executes shell commands (for mocking).
type ShellExecutor interface {
	RunCommand(ctx context.Context, dir, name string, args ...string) error
	RunCommandOutput(ctx context.Context, dir, name string, args ...string) (string, error)
}

// defaultShellExecutor는 실제 셸 명령을 실행하는 구현입니다.
// defaultShellExecutor implements ShellExecutor using real shell commands.
type defaultShellExecutor struct{}

// RunCommand는 셸 명령을 실행합니다.
func (d *defaultShellExecutor) RunCommand(ctx context.Context, dir, name string, args ...string) error {
	// TODO T-04: 실제 구현은 os/exec.Command 사용
	return nil
}

// RunCommandOutput은 셸 명령을 실행하고 출력을 반환합니다.
func (d *defaultShellExecutor) RunCommandOutput(ctx context.Context, dir, name string, args ...string) (string, error) {
	// TODO T-04: 실제 구현은 os/exec.Command 사용
	return "", nil
}

// LaunchdManager는 macOS에서 svc.sh를 통해 runner 서비스를 관리합니다.
// LaunchdManager manages runner service on macOS via svc.sh.
type LaunchdManager struct {
	ghRunnerDir string
	executor    ShellExecutor
}

// NewLaunchdManager는 새로운 LaunchdManager 인스턴스를 생성합니다.
// NewLaunchdManager creates a new LaunchdManager instance.
func NewLaunchdManager(ghRunnerDir string, executor ShellExecutor) *LaunchdManager {
	exec := executor
	if exec == nil {
		exec = &defaultShellExecutor{}
	}

	return &LaunchdManager{
		ghRunnerDir: ghRunnerDir,
		executor:    exec,
	}
}

// Install은 runner 서비스를 설치합니다.
func (m *LaunchdManager) Install(ctx context.Context) error {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	err := m.executor.RunCommand(ctx, runnerDir, "./svc.sh", "install")
	if err != nil {
		return fmt.Errorf("install service: %w", err)
	}
	return nil
}

// Start는 runner 서비스를 시작합니다.
func (m *LaunchdManager) Start(ctx context.Context) error {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	err := m.executor.RunCommand(ctx, runnerDir, "./svc.sh", "start")
	if err != nil {
		return fmt.Errorf("start service: %w", err)
	}
	return nil
}

// Stop은 runner 서비스를 중지합니다.
func (m *LaunchdManager) Stop(ctx context.Context) error {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	err := m.executor.RunCommand(ctx, runnerDir, "./svc.sh", "stop")
	if err != nil {
		return fmt.Errorf("stop service: %w", err)
	}
	return nil
}

// Status는 runner 서비스 상태를 확인합니다.
func (m *LaunchdManager) Status(ctx context.Context) (CheckStatus, error) {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	output, err := m.executor.RunCommandOutput(ctx, runnerDir, "./svc.sh", "status")
	if err != nil {
		return StatusUnknown, fmt.Errorf("check status: %w", err)
	}

	// 출력에서 상태 판별
	output = strings.ToLower(output)
	if strings.Contains(output, "running") {
		return StatusRunning, nil
	}
	if strings.Contains(output, "stopped") {
		return StatusStopped, nil
	}
	return StatusUnknown, nil
}

// SystemdStub은 Linux systemd 미지원 스텁 구현입니다 (v1.1).
// SystemdStub returns "not yet supported" for all operations (v1.1).
type SystemdStub struct{}

// Install은 미지원 에러를 반환합니다.
func (s *SystemdStub) Install(ctx context.Context) error {
	return fmt.Errorf("systemd service management not yet supported (v1.1)")
}

// Start는 미지원 에러를 반환합니다.
func (s *SystemdStub) Start(ctx context.Context) error {
	return fmt.Errorf("systemd service management not yet supported (v1.1)")
}

// Stop은 미지원 에러를 반환합니다.
func (s *SystemdStub) Stop(ctx context.Context) error {
	return fmt.Errorf("systemd service management not yet supported (v1.1)")
}

// Status는 미지원 에러를 반환합니다.
func (s *SystemdStub) Status(ctx context.Context) (CheckStatus, error) {
	return StatusUnknown, fmt.Errorf("systemd service management not yet supported (v1.1)")
}
