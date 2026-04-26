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

// InstallHintError wraps ErrBinaryNotFound and carries the user-facing install hint
// from lsp.yaml (REQ-LM-004). Callers may use errors.As to extract and log the hint.
//
// @MX:NOTE: [AUTO] InstallHintError — wraps ErrBinaryNotFound with install hint for user-facing messages
type InstallHintError struct {
	// Hint is the human-readable install command from lsp.yaml install_hint field.
	Hint string
	// Err is the underlying error (always wraps ErrBinaryNotFound).
	Err error
}

// Error implements the error interface.
func (e *InstallHintError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%v (install hint: %s)", e.Err, e.Hint)
	}
	return e.Err.Error()
}

// Unwrap allows errors.Is(err, ErrBinaryNotFound) to work through InstallHintError.
func (e *InstallHintError) Unwrap() error {
	return e.Err
}

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
// If the primary binary is not found, fallback_binaries from cfg are tried in
// order (REQ-LM-008). If all binaries are missing, Launch returns
// ErrBinaryNotFound wrapped in an InstallHintError when cfg.InstallHint is set
// (REQ-LM-004).
//
// Each launched server gets its own isolated stdin, stdout, and stderr pipes
// to prevent I/O cross-contamination between servers (REQ-LC-005).
//
// The returned LaunchResult.Cmd is already started; callers MUST call
// Cmd.Wait after the process is expected to exit to avoid zombie processes.
func (l *Launcher) Launch(ctx context.Context, cfg config.ServerConfig) (*LaunchResult, error) {
	// Try the primary command, then fallback_binaries in order (REQ-LM-008).
	candidates := append([]string{cfg.Command}, cfg.FallbackBinaries...)
	var binPath string
	var resolveErr error
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		p, err := resolveBinary(candidate)
		if err == nil {
			binPath = p
			break
		}
		resolveErr = err
	}

	if binPath == "" {
		// All candidates failed. Wrap with InstallHintError when hint is available (REQ-LM-004).
		notFound := fmt.Errorf("subprocess.Launch %q (language %q): %w",
			cfg.Command, cfg.Language, ErrBinaryNotFound)
		if cfg.InstallHint != "" {
			return nil, &InstallHintError{Hint: cfg.InstallHint, Err: notFound}
		}
		_ = resolveErr
		return nil, notFound
	}

	args := make([]string, len(cfg.Args))
	copy(args, cfg.Args)

	cmd := exec.CommandContext(ctx, binPath, args...) //nolint:gosec // path is validated above

	// Create each pipe independently (REQ-LC-005 isolation)
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
