package hook

// observability_master_cli_test.go — t62: the cwd_fallback WARN must not
// surface on operator CLI paths (where CLAUDE_PROJECT_DIR is legitimately
// unset and cwd is the right answer), while hook runtime keeps WARN.

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// captureSlogForT62 swaps the default slog handler for a text handler over a
// buffer at Debug level and restores the original on cleanup. Non-parallel
// only (mutates process-global logger state).
func captureSlogForT62(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(orig) })
	return &buf
}

// TestObservabilityCLIPath_NoWarnOnCwdFallback pins the operator-facing
// behavior from the t62 incident: loading the observability master toggle
// from a CLI process (no CLAUDE_PROJECT_DIR, cwd fallback legitimate) must
// not emit a WARN line — the fallback logs at Debug there. The Debug line
// itself stays visible in the buffer, proving the resolution still happened.
func TestObservabilityCLIPath_NoWarnOnCwdFallback(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	buf := captureSlogForT62(t)
	ResetObservabilityMasterForTesting()
	t.Cleanup(ResetObservabilityMasterForTesting)

	_ = IsObservabilityEnabledForCLI()

	out := buf.String()
	if strings.Contains(out, "level=WARN") {
		t.Fatalf("CLI-path cwd fallback must log at Debug, not WARN:\n%s", out)
	}
	if !strings.Contains(out, "cwd fallback used") {
		t.Fatalf("resolution must still happen (Debug line expected in buffer):\n%s", out)
	}
}

// TestObservabilityHookPath_KeepsWarnOnCwdFallback guards against
// over-fixing: the hook runtime keeps the WARN so a hook process that lost
// CLAUDE_PROJECT_DIR stays visible in diagnostics.
func TestObservabilityHookPath_KeepsWarnOnCwdFallback(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	buf := captureSlogForT62(t)
	ResetObservabilityMasterForTesting()
	t.Cleanup(ResetObservabilityMasterForTesting)

	_ = IsObservabilityEnabled()

	out := buf.String()
	if !strings.Contains(out, "level=WARN") || !strings.Contains(out, "cwd fallback used") {
		t.Fatalf("hook runtime must keep the WARN cwd_fallback line:\n%s", out)
	}
}
