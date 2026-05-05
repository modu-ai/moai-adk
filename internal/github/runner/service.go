// Package runner provides service management for GitHub Actions runners.
package runner

import (
	"context"
	"fmt"
	"strings"
)

// CheckStatus represents service status.
type CheckStatus string

const (
	// StatusRunning indicates service is running.
	StatusRunning CheckStatus = "running"
	// StatusStopped indicates service is stopped.
	StatusStopped CheckStatus = "stopped"
	StatusUnknown CheckStatus = "unknown"
)

// ServiceManager interface for runner service management.
type ServiceManager interface {
	Install(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status(ctx context.Context) (CheckStatus, error)
}

// ShellExecutor executes shell commands (for mocking).
type ShellExecutor interface {
	RunCommand(ctx context.Context, dir, name string, args ...string) error
	RunCommandOutput(ctx context.Context, dir, name string, args ...string) (string, error)
}

// defaultShellExecutor implements ShellExecutor using real shell commands.
type defaultShellExecutor struct{}

func (d *defaultShellExecutor) RunCommand(ctx context.Context, dir, name string, args ...string) error {
	// TODO T-04: Actual implementation uses os/exec.Command
	return nil
}

// RunCommandOutput executes shell commands and returns output.
func (d *defaultShellExecutor) RunCommandOutput(ctx context.Context, dir, name string, args ...string) (string, error) {
	// TODO T-04: Actual implementation uses os/exec.Command
	return "", nil
}

// LaunchdManager manages runner service on macOS via svc.sh.
type LaunchdManager struct {
	ghRunnerDir string
	executor    ShellExecutor
}

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

// Install installs the runner service.
func (m *LaunchdManager) Install(ctx context.Context) error {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	err := m.executor.RunCommand(ctx, runnerDir, "./svc.sh", "install")
	if err != nil {
		return fmt.Errorf("install service: %w", err)
	}
	return nil
}

// Start starts the runner service.
func (m *LaunchdManager) Start(ctx context.Context) error {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	err := m.executor.RunCommand(ctx, runnerDir, "./svc.sh", "start")
	if err != nil {
		return fmt.Errorf("start service: %w", err)
	}
	return nil
}

// Stop stops the runner service.
func (m *LaunchdManager) Stop(ctx context.Context) error {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	err := m.executor.RunCommand(ctx, runnerDir, "./svc.sh", "stop")
	if err != nil {
		return fmt.Errorf("stop service: %w", err)
	}
	return nil
}

// Status verifies the runner service status.
func (m *LaunchdManager) Status(ctx context.Context) (CheckStatus, error) {
	runnerDir := m.ghRunnerDir + "/actions-runner"
	output, err := m.executor.RunCommandOutput(ctx, runnerDir, "./svc.sh", "status")
	if err != nil {
		return StatusUnknown, fmt.Errorf("check status: %w", err)
	}

	output = strings.ToLower(output)
	if strings.Contains(output, "running") {
		return StatusRunning, nil
	}
	if strings.Contains(output, "stopped") {
		return StatusStopped, nil
	}
	return StatusUnknown, nil
}

// SystemdStub returns "not yet supported" for all operations (v1.1).
type SystemdStub struct{}

// Install returns "not yet supported" error.
func (s *SystemdStub) Install(ctx context.Context) error {
	return fmt.Errorf("systemd service management not yet supported (v1.1)")
}

// Start returns "not yet supported" error.
func (s *SystemdStub) Start(ctx context.Context) error {
	return fmt.Errorf("systemd service management not yet supported (v1.1)")
}

// Stop returns "not yet supported" error.
func (s *SystemdStub) Stop(ctx context.Context) error {
	return fmt.Errorf("systemd service management not yet supported (v1.1)")
}

// Status returns "not yet supported" error.
func (s *SystemdStub) Status(ctx context.Context) (CheckStatus, error) {
	return StatusUnknown, fmt.Errorf("systemd service management not yet supported (v1.1)")
}
