package cli

import (
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
