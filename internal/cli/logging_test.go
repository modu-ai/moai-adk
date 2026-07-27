package cli

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestResolveLogLevel covers the MOAI_LOG_LEVEL resolution table
// (REQ-FAG-004, 005, 006, 007).
//
// Not parallel, and no subtest declares t.Parallel: every case drives the real
// os.Getenv read through t.Setenv, which panics when the calling test is
// parallel. Reading the variable for real (rather than passing the value as a
// parameter) is deliberate — it is what proves the production path reads
// config.EnvLogLevel and not some other name.
func TestResolveLogLevel(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want slog.Level
	}{
		// t.Setenv with "" leaves os.Getenv returning "", exactly as it does for
		// a genuinely unset variable, and clears any ambient value the developer
		// happens to have exported.
		{"unset_defaults_to_warn", "", slog.LevelWarn},
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"uppercase_is_accepted", "ERROR", slog.LevelError},
		{"surrounding_space_is_tolerated", "  debug  ", slog.LevelDebug},
		{"unrecognized_falls_back_to_warn", "nonsense", slog.LevelWarn},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(config.EnvLogLevel, tc.raw)

			if got := resolveLogLevel(); got != tc.want {
				t.Errorf("resolveLogLevel() with %s=%q = %v, want %v",
					config.EnvLogLevel, tc.raw, got, tc.want)
			}
		})
	}
}

// TestLoggingHandlerSelection covers the hook / non-hook destination split
// (REQ-FAG-002, 003) and the unconditional nature of the hook carve-out (D-3).
//
// The destination is asserted on the decision value rather than on the installed
// global logger: a slog.Handler does not expose its writer, so routing the
// assertion through slog.SetDefault would only prove that *a* handler was
// installed, not *which* one. Not parallel — subtests call t.Setenv.
func TestLoggingHandlerSelection(t *testing.T) {
	cases := []struct {
		name     string
		env      string
		args     []string
		wantDest io.Writer
	}{
		{"hook_discards", "", []string{"hook", "pre-tool"}, io.Discard},
		{"doctor_writes_stderr", "", []string{"doctor"}, os.Stderr},
		{"astgrep_writes_stderr", "", []string{"ast-grep", "."}, os.Stderr},
		{"update_writes_stderr", "", []string{"update"}, os.Stderr},
		{"hook_behind_a_flag_discards", "", []string{"--verbose", "hook", "stop"}, io.Discard},
		// D-3: MOAI_LOG_LEVEL does not re-open the hook path. The hook contract
		// is that stdout carries structured JSON and stderr belongs to the
		// Claude Code runtime, so log records stay discarded regardless.
		{"hook_ignores_log_level_env", "debug", []string{"hook", "session-start"}, io.Discard},
		{"bare_invocation_writes_stderr", "", []string{}, os.Stderr},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(config.EnvLogLevel, tc.env)

			got := resolveLoggingDecision(tc.args)
			if got.dest != tc.wantDest {
				t.Errorf("resolveLoggingDecision(%q).dest = %T(%v), want %T(%v)",
					tc.args, got.dest, got.dest, tc.wantDest, tc.wantDest)
			}
		})
	}
}

// TestLoggingNeverTargetsStdout covers REQ-FAG-008: stdout carries the CLI's
// machine-readable output (--format=json / sarif payloads), so no handler this
// package installs may be constructed over it.
//
// The check is deliberately STATIC — it reads logging.go and asserts the token
// is absent — because the runtime equivalent is not decidable here. Under
// `go test ./...` this repo's test binary receives ONE *os.File for both
// streams: a comparison of the chosen destination against os.Stdout was measured
// returning dest == os.Stdout == os.Stderr == the same pointer, so a runtime
// identity check reports a false failure while the production selection
// (os.Stderr) is correct. A source-level guard is falsifiable (reference
// os.Stdout in logging.go and this fails) and immune to that aliasing.
//
// The behavioral half of REQ-FAG-008 is covered elsewhere: TestLoggingHandlerSelection
// asserts which writer is chosen, and the SPEC's AC-FAG-005 asserts against the
// real binary that stdout carries no slog record.
func TestLoggingNeverTargetsStdout(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("logging.go")
	if err != nil {
		t.Fatalf("read logging.go: %v", err)
	}
	if bytes.Contains(src, []byte("os.Stdout")) {
		t.Error("logging.go references os.Stdout; the log handler must never be " +
			"constructed over stdout (REQ-FAG-008) — stdout carries --format=json/sarif payloads")
	}
}
