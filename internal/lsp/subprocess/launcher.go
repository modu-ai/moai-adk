package subprocess

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// ErrBinaryNotFound is returned by Launcher.Launch when the language server
// binary cannot be located in the system PATH (REQ-LC-004 warn_and_skip).
//
// Callers SHOULD log the error and skip the server rather than panic.
//
// @MX:ANCHOR: [AUTO] ErrBinaryNotFound — sentinel consumed by Manager, CLI, and tests
// @MX:REASON: fan_in >= 3 — manager, CLI warn_and_skip handler, and test assertions all use errors.Is(err, ErrBinaryNotFound)
var ErrBinaryNotFound = errors.New("subprocess: binary not found in PATH")

// LaunchResult holds handles to the running language server subprocess.
// All three pipe fields are guaranteed non-nil on successful launch.
type LaunchResult struct {
	// Cmd is the running subprocess. Callers may call Cmd.Wait to detect exit.
	Cmd *exec.Cmd

	// Stdin is the write end of the subprocess's stdin pipe.
	// The caller sends JSON-RPC messages here.
	Stdin io.WriteCloser

	// Stdout is the read end of the subprocess's stdout pipe.
	// The caller reads JSON-RPC messages from here.
	Stdout io.ReadCloser

	// Stderr is the read end of the subprocess's stderr pipe.
	// The caller may drain this for logging purposes.
	Stderr io.ReadCloser
}

// Launcher spawns language server subprocesses with isolated stdio pipes.
//
// @MX:ANCHOR: [AUTO] Launcher — central subprocess factory consumed by core.Client and integration tests
// @MX:REASON: fan_in >= 3 — core.Client.Start, integration tests, and Manager.spawnServer all call Launch
type Launcher struct{}

// NewLauncher creates a new Launcher.
func NewLauncher() *Launcher {
	return &Launcher{}
}

// Launch starts the language server described by cfg in a new subprocess.
//
// Binary lookup uses exec.LookPath when cfg.Command is not an absolute path.
// If the binary is not found, Launch returns ErrBinaryNotFound (warn_and_skip
// pattern: caller logs and continues without the server — no panic).
//
// Each launched server gets its own isolated stdin, stdout, and stderr pipes
// to prevent I/O cross-contamination between servers (REQ-LC-005).
//
// The returned LaunchResult.Cmd is already started; callers MUST call
// Cmd.Wait after the process is expected to exit to avoid zombie processes.
func (l *Launcher) Launch(ctx context.Context, cfg config.ServerConfig) (*LaunchResult, error) {
	// 바이너리 경로 결정: 절대 경로면 직접 사용, 아니면 PATH 탐색
	binPath, err := resolveBinary(cfg.Command)
	if err != nil {
		return nil, fmt.Errorf("subprocess.Launch %q (language %q): %w",
			cfg.Command, cfg.Language, ErrBinaryNotFound)
	}

	args := make([]string, len(cfg.Args))
	copy(args, cfg.Args)

	cmd := exec.CommandContext(ctx, binPath, args...) //nolint:gosec // 경로는 위에서 검증됨

	// 각 파이프를 독립적으로 생성 (REQ-LC-005 isolation)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("subprocess.Launch %q: stdin pipe: %w", cfg.Command, err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("subprocess.Launch %q: stdout pipe: %w", cfg.Command, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("subprocess.Launch %q: stderr pipe: %w", cfg.Command, err)
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("subprocess.Launch %q: start: %w", cfg.Command, err)
	}

	return &LaunchResult{
		Cmd:    cmd,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}

// resolveBinary finds the absolute path to the binary.
// If cmd is already absolute and the file exists, it is returned as-is.
// Otherwise exec.LookPath is used to search PATH.
func resolveBinary(cmd string) (string, error) {
	if isAbsolutePath(cmd) {
		if _, err := os.Stat(cmd); err != nil {
			return "", ErrBinaryNotFound
		}
		return cmd, nil
	}
	return exec.LookPath(cmd)
}

// isAbsolutePath reports whether path is absolute on the current OS.
func isAbsolutePath(path string) bool {
	if len(path) == 0 {
		return false
	}
	// Unix: starts with '/'
	// Windows: starts with drive letter + ':' or UNC '\\'
	if path[0] == '/' || path[0] == '\\' {
		return true
	}
	if len(path) >= 2 && path[1] == ':' {
		return true
	}
	return false
}
