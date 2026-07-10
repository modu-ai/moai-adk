package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
)

// TestFangExitCoderCharacterization pins the root-execution seam contract that
// SPEC-CLI-TUX-V3-001 M1c must preserve across the charm.land/fang/v2 swap
// (AC-CTX-020). It is written BEFORE the fang wiring against the raw-cobra
// baseline and MUST continue to pass unmodified once runFang routes through
// fang.Execute. It characterizes three invariants:
//
//  1. Exit-code chain — an error implementing ExitCode() int (the
//     `moai worktree verify` 0/1/2/3 ExitCoder) propagates out of the seam so
//     cmd/moai/main.go's errors.As unwrapping surfaces the custom code; a plain
//     cobra error maps to exit 1; --help returns nil (exit 0).
//  2. Non-TTY + NO_COLOR — --help emits no ANSI escape on either channel.
//  3. Trivial fast-path — the lazy-init command set (root.go trivialCommands)
//     is unchanged, so version/help/completion still skip InitDependencies.
//
// The synthetic command trees are isolated (never the global rootCmd), so the
// fang mutations fang.Execute applies (SetHelpFunc / SilenceErrors / Version)
// cannot leak into the ~20 sibling cli test files.
func TestFangExitCoderCharacterization(t *testing.T) {
	t.Run("ExitCoderChain", func(t *testing.T) {
		cases := []struct {
			name     string
			args     []string
			wantCode int
		}{
			{"success_nil", []string{"ok"}, 0},
			{"local_exit_1", []string{"code1"}, 1},
			{"worktree_verify_2", []string{"code2"}, 2},
			{"worktree_verify_3", []string{"code3"}, 3},
			{"plain_error_maps_to_1", []string{"boom"}, 1},
			{"help_exit_0", []string{"--help"}, 0},
			{"unknown_command_maps_to_1", []string{"does-not-exist"}, 1},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				var out, errw bytes.Buffer
				root := newCharRoot(&out, &errw)
				root.SetArgs(tc.args)
				err := runFang(context.Background(), root)
				if got := exitCodeForError(err); got != tc.wantCode {
					t.Fatalf("args=%v: exit code = %d, want %d (err=%v)", tc.args, got, tc.wantCode, err)
				}
			})
		}
	})

	t.Run("NonTTYNoColorHelpHasNoANSI", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		var out, errw bytes.Buffer
		root := newCharRoot(&out, &errw)
		root.SetArgs([]string{"--help"})
		if err := runFang(context.Background(), root); err != nil {
			t.Fatalf("help exec: %v", err)
		}
		for name, buf := range map[string]*bytes.Buffer{"stdout": &out, "stderr": &errw} {
			if strings.Contains(buf.String(), "\x1b[") {
				t.Errorf("%s: --help under NO_COLOR/non-TTY contains an ANSI escape sequence", name)
			}
		}
	})

	t.Run("TrivialFastPathPreserved", func(t *testing.T) {
		for _, arg := range []string{"--version", "-v", "version", "help", "--help", "-h", "completion"} {
			if !isTrivialCommand([]string{arg}) {
				t.Errorf("isTrivialCommand(%q) = false, want true (lazy-init fast-path must persist under fang)", arg)
			}
		}
		for _, arg := range []string{"init", "update", "hook", "doctor"} {
			if isTrivialCommand([]string{arg}) {
				t.Errorf("isTrivialCommand(%q) = true, want false (non-trivial commands need full InitDependencies)", arg)
			}
		}
	})
}

// exitCodeForError mirrors cmd/moai/main.go's ExitCoder unwrapping verbatim:
// nil → 0, an error implementing ExitCode() int → that code, any other → 1.
func exitCodeForError(err error) int {
	if err == nil {
		return 0
	}
	var ec interface{ ExitCode() int }
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}

// localExitErr is a test double for a subcommand-supplied ExitCoder error that
// is not the worktree carrier (verifies the interface-based unwrapping, not a
// concrete-type match).
type localExitErr struct{ code int }

func (e localExitErr) Error() string { return "local exit-code carrier" }
func (e localExitErr) ExitCode() int { return e.code }

// newCharRoot builds an isolated root command whose subcommands exercise every
// exit-code path. It deliberately does NOT reuse the global rootCmd.
func newCharRoot(out, errw *bytes.Buffer) *cobra.Command {
	root := &cobra.Command{Use: "charroot", Short: "characterization root"}
	root.SetOut(out)
	root.SetErr(errw)

	root.AddCommand(&cobra.Command{
		Use:  "ok",
		RunE: func(*cobra.Command, []string) error { return nil },
	})
	root.AddCommand(&cobra.Command{
		Use: "code1",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			return localExitErr{code: 1}
		},
	})
	root.AddCommand(&cobra.Command{
		Use: "code2",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			return &worktree.ExitCodeError{Code: 2}
		},
	})
	root.AddCommand(&cobra.Command{
		Use: "code3",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			return &worktree.ExitCodeError{Code: 3}
		},
	})
	root.AddCommand(&cobra.Command{
		Use: "boom",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			return errors.New("plain non-exitcoder failure")
		},
	})
	return root
}
