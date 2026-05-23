// Sensitive env injection tests for SPEC-V3R5-SECURITY-CRIT-001 P0-2
// (CWE-214: information exposure through process argv).
//
// The DefaultSessionManager's pre-existing InjectEnv path passes values
// directly via `tmux set-environment <key> <value>`, which makes the value
// visible to any local process that reads /proc/<pid>/cmdline (or `ps -ef`
// on macOS). For credentials such as ANTHROPIC_AUTH_TOKEN this is a P0
// information leak. InjectSensitiveEnv must route through a temp file +
// `tmux source-file` so the value never appears in argv.

//go:build !windows
// +build !windows

package tmux

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

// secretToken is a recognizable sentinel used across the sensitive tests so
// that any argv leak shows up obviously in assertion output.
const secretToken = "sk-glm-SENSITIVE-VALUE-1234567890abcdef"

// TestInjectSensitiveEnvNoArgvLeak covers AC-SEC-005:
// InjectSensitiveEnv must NOT pass the value as a positional argument to
// tmux. The mock command recorder collects every argv slice; the secret
// token must never appear in any of them.
func TestInjectSensitiveEnvNoArgvLeak(t *testing.T) {
	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))

	err := mgr.InjectSensitiveEnv(context.Background(), "ANTHROPIC_AUTH_TOKEN", secretToken)
	if err != nil {
		t.Fatalf("InjectSensitiveEnv unexpected error: %v", err)
	}

	// The mock runner recorded every call. Walk all argv slices and assert
	// the secret value is never present as a positional argument.
	for _, call := range calls {
		for _, arg := range call {
			if strings.Contains(arg, secretToken) {
				t.Errorf("AC-SEC-005 regression: secret value leaked into tmux argv\n  call: %v", call)
			}
		}
	}

	// Sanity: there must be at least one tmux source-file invocation, because
	// that is the prescribed channel for sensitive values.
	sawSourceFile := false
	for _, call := range calls {
		if len(call) >= 2 && call[0] == "tmux" && call[1] == "source-file" {
			sawSourceFile = true
			break
		}
	}
	if !sawSourceFile {
		t.Errorf("InjectSensitiveEnv did not invoke 'tmux source-file'; calls=%v", calls)
	}
}

// TestInjectSensitiveEnvTempFileCleaned covers DoD-1 (no residual temp file).
// After InjectSensitiveEnv returns, the temp script must be unlinked even on
// the success path so the secret does not linger on disk.
func TestInjectSensitiveEnvTempFileCleaned(t *testing.T) {
	var sourcedPaths []string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		if name == "tmux" && len(args) >= 2 && args[0] == "source-file" {
			sourcedPaths = append(sourcedPaths, args[1])
		}
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))

	if err := mgr.InjectSensitiveEnv(context.Background(), "ANTHROPIC_AUTH_TOKEN", secretToken); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sourcedPaths) == 0 {
		t.Fatal("expected at least one source-file invocation")
	}

	for _, p := range sourcedPaths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("temp script %s was not cleaned up (err=%v)", p, err)
		}
	}
}

// TestInjectSensitiveEnvFailureNoArgvFallback covers AC-SEC-007:
// When the temp-file path fails, InjectSensitiveEnv MUST return
// ErrTmuxSensitiveInjectFailed — it must NEVER fall back to the
// argv-exposing `set-environment` path.
func TestInjectSensitiveEnvFailureNoArgvFallback(t *testing.T) {
	// Force source-file to fail so the helper enters the error branch.
	var calls [][]string
	runner := func(_ context.Context, name string, args ...string) (string, error) {
		calls = append(calls, append([]string{name}, args...))
		if name == "tmux" && len(args) >= 1 && args[0] == "source-file" {
			return "", errors.New("simulated tmux source-file failure")
		}
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))

	err := mgr.InjectSensitiveEnv(context.Background(), "ANTHROPIC_AUTH_TOKEN", secretToken)
	if err == nil {
		t.Fatal("expected error from failed source-file, got nil")
	}
	if !errors.Is(err, ErrTmuxSensitiveInjectFailed) {
		t.Errorf("error should wrap ErrTmuxSensitiveInjectFailed; got %v", err)
	}

	// Critical: no `set-environment <key> <value>` call must have been made
	// after the failure. Any such call would re-introduce the argv leak.
	for _, call := range calls {
		if len(call) < 4 {
			continue
		}
		if call[0] != "tmux" || call[1] != "set-environment" {
			continue
		}
		// Allow `set-environment -u <key>` (unset) — that has no value arg.
		if slices.Contains(call, "-u") {
			continue
		}
		// Detect `tmux set-environment <key> <value>` with our secret as value.
		if slices.Contains(call, secretToken) {
			t.Errorf("AC-SEC-007 regression: fallback to argv after temp-file failure: %v", call)
		}
	}
}

// TestInjectSensitiveEnvTempFilePermissionRestrictive covers DoD-2:
// the temp script created in ~/.moai/run/ MUST have mode 0o600 (or stricter)
// during its brief lifetime. Anything 0o644 leaks the secret to local users.
func TestInjectSensitiveEnvTempFilePermissionRestrictive(t *testing.T) {
	// Capture the temp script path by hooking into `source-file` argv and
	// asserting on the file BEFORE the helper unlinks it. We do this by
	// inspecting the script during the runner callback itself.
	var capturedMode os.FileMode
	var captureErr error

	runner := func(_ context.Context, name string, args ...string) (string, error) {
		if name == "tmux" && len(args) >= 2 && args[0] == "source-file" {
			info, err := os.Stat(args[1])
			if err != nil {
				captureErr = err
				return "", nil
			}
			capturedMode = info.Mode().Perm()
		}
		return "", nil
	}

	mgr := NewSessionManager(WithSessionRunFunc(runner))
	if err := mgr.InjectSensitiveEnv(context.Background(), "ANTHROPIC_AUTH_TOKEN", secretToken); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captureErr != nil {
		t.Fatalf("stat temp script during source-file: %v", captureErr)
	}
	if capturedMode == 0 {
		t.Fatal("source-file was never invoked; temp script mode could not be captured")
	}
	const want os.FileMode = 0o600
	if capturedMode&0o077 != 0 {
		t.Errorf("temp script mode = %#o, want %#o (group/world bits must be zero for sensitive temp scripts)",
			capturedMode, want)
	}
}
