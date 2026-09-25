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
//
// The three hook cases assert a *hookSink rather than io.Discard (AC-HDS-015).
// They previously expected io.Discard, and SPEC-HOOK-DIAG-SINK-001 REQ-HDS-001
// deliberately inverts that expectation: the hook path now writes to a file
// sink. This is a contract change, not a regression — the carve-out these cases
// exist to protect is that neither standard stream is reachable, and that is
// what they still assert.
//
// The three case NAMES still read "discards" and are now inaccurate. They are
// kept deliberately: AC-HDS-015 pins them, so that deleting a case to make the
// suite green is mechanically visible. Renaming them is a later decision, not a
// tidy-up to make here.
//
// A *hookSink cannot be compared by pointer identity: resolveLoggingDecision
// constructs a fresh one per call, by design (each invocation resolves its own
// root). The hook cases therefore assert the TYPE plus the three destinations
// that must never appear.
func TestLoggingHandlerSelection(t *testing.T) {
	cases := []struct {
		name string
		env  string
		args []string
		// wantDest is compared by identity and applies to non-hook cases only;
		// it is nil exactly when wantHookSink is set.
		wantDest     io.Writer
		wantHookSink bool
	}{
		{name: "hook_discards", args: []string{"hook", "pre-tool"}, wantHookSink: true},
		{name: "doctor_writes_stderr", args: []string{"doctor"}, wantDest: os.Stderr},
		{name: "astgrep_writes_stderr", args: []string{"ast-grep", "."}, wantDest: os.Stderr},
		{name: "update_writes_stderr", args: []string{"update"}, wantDest: os.Stderr},
		{name: "hook_behind_a_flag_discards", args: []string{"--verbose", "hook", "stop"}, wantHookSink: true},
		// D-3: MOAI_LOG_LEVEL does not re-open the hook path. The hook contract
		// is that stdout carries structured JSON and stderr belongs to the
		// Claude Code runtime, so the variable moves the LEVEL and never the
		// destination (REQ-HDS-002, REQ-HDS-005). That is the assertion this
		// case exists for, and it survives the destination change intact.
		{name: "hook_ignores_log_level_env", env: "debug", args: []string{"hook", "session-start"}, wantHookSink: true},
		{name: "bare_invocation_writes_stderr", args: []string{}, wantDest: os.Stderr},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(config.EnvLogLevel, tc.env)

			got := resolveLoggingDecision(tc.args)

			if !tc.wantHookSink {
				if got.dest != tc.wantDest {
					t.Errorf("resolveLoggingDecision(%q).dest = %T(%v), want %T(%v)",
						tc.args, got.dest, got.dest, tc.wantDest, tc.wantDest)
				}
				return
			}

			if _, ok := got.dest.(*hookSink); !ok {
				t.Errorf("resolveLoggingDecision(%q).dest = %T, want *cli.hookSink "+
					"(the hook path writes to the file sink — REQ-HDS-001)", tc.args, got.dest)
			}
			// The carve-out itself: MOAI_LOG_LEVEL=%q must not open a standard
			// stream, and the records must not be thrown away either.
			if got.dest == os.Stdout || got.dest == os.Stderr || got.dest == io.Discard {
				t.Errorf("resolveLoggingDecision(%q).dest with %s=%q resolved to %T(%v); "+
					"the hook path must reach neither standard stream nor io.Discard "+
					"(REQ-HDS-002)", tc.args, config.EnvLogLevel, tc.env, got.dest, got.dest)
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
