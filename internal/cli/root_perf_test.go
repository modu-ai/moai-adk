package cli

import (
	"bytes"
	"log/slog"
	"os"
	"testing"
)

// TestIsTrivialCommand (AC-PERF-003a) verifies that trivial subcommands are
// correctly identified and would skip InitDependencies.
func TestIsTrivialCommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"version_long", []string{"--version"}, true},
		{"version_short", []string{"-v"}, true},
		{"version_subcommand", []string{"version"}, true},
		{"help_long", []string{"--help"}, true},
		{"help_short", []string{"-h"}, true},
		{"help_subcommand", []string{"help"}, true},
		{"completion", []string{"completion"}, true},
		{"hook_command", []string{"hook", "session-start"}, false},
		{"spec_lint", []string{"spec", "lint"}, false},
		{"update", []string{"update"}, false},
		{"init", []string{"init", "myproject"}, false},
		{"doctor", []string{"doctor"}, false},
		{"empty_args", []string{}, false},
		{"flag_only", []string{"--json"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isTrivialCommand(tc.args)
			if got != tc.want {
				t.Errorf("isTrivialCommand(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

// TestTrivialPathSkipsFullInit (REQ-FAG-009) asserts that Execute() actually
// TAKES the trivial fast path, not merely that isTrivialCommand classifies
// correctly — TestIsTrivialCommand above already covers the predicate, and a
// predicate test alone would still pass if the branch were deleted from
// Execute().
//
// The witness is the package-level `deps` global. It is written by exactly two
// non-test sites, InitDependencies and the SetDeps test helper (deps.go), so
// after an explicit reset to nil its nil-ness is an exact, binary,
// host-independent record of whether full initialization ran.
//
// Log volume is NOT usable as the witness: every warn-or-above site reachable on
// the full-init path is conditional on a failure state (gopls config load
// failure, bridge init failure, absent config sections dir), none of which fires
// on a healthy host — so a log-based assertion would pass whether or not the
// branch was taken.
//
// Not parallel: it mutates os.Args, the `deps` global, the shared rootCmd, and
// the process default logger that Execute installs via configureLogging.
func TestTrivialPathSkipsFullInit(t *testing.T) {
	origArgs := os.Args
	origDeps := deps
	origLogger := slog.Default()
	t.Cleanup(func() {
		os.Args = origArgs
		deps = origDeps
		slog.SetDefault(origLogger)
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	// SetDeps is a second writer of `deps`; reset explicitly so a value left
	// behind by an earlier test cannot be mistaken for full initialization here.
	deps = nil

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	// Execute() branches on os.Args; rootCmd.SetArgs keeps cobra's own parse in
	// step so the two cannot disagree if a sibling test left args behind.
	os.Args = []string{"moai", "--version"}
	rootCmd.SetArgs([]string{"--version"})

	if err := Execute(); err != nil {
		t.Fatalf("Execute() on the --version path returned error: %v", err)
	}

	if deps != nil {
		t.Error("Execute() ran full initialization on the trivial --version path: " +
			"`deps` is non-nil, so the isTrivialCommand branch in Execute() was not " +
			"taken (REQ-FAG-009)")
	}
}

// BenchmarkExecuteVersion measures the startup cost of a trivial command path
// (AC-PERF-003c). The lazy-init path should be significantly cheaper than the
// full InitDependencies path.
func BenchmarkIsTrivialCommand(b *testing.B) {
	args := []string{"--version"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isTrivialCommand(args)
	}
}
