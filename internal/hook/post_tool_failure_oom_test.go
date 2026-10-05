package hook

import (
	"testing"
)

// t674: the OOM branch matched the bare substring "137". This repository names
// worktrees after card ids, so any failure text carrying a path like
// .claude/worktrees/t137 — a permission error, a missing file, anything —
// classified as OOMKilled once earlier branches declined to match. These tests
// pin the boundary: 137 is an OOM signal only in delimiter-bounded form
// ("exit status 137"), never as a fragment of a longer token or a path.

// TestOOMExitCode137Boundary pins the fix direction: a failure that merely
// carries "137" inside a longer token (a card-id worktree path) is NOT an OOM,
// while every real OOM exit-code phrasing still is.
func TestOOMExitCode137Boundary(t *testing.T) {
	t.Parallel()

	h := &postToolUseFailureHandler{}

	cases := []struct {
		name string
		err  string
		want ErrorCategory
	}{
		{
			// The defect shape: an unrelated failure mentioning the t137 path.
			name: "path fragment t137 is not OOM",
			err:  "bash: /proj/.claude/worktrees/t137/run.sh: no such file or directory",
			want: UnknownFailure,
		},
		{
			// Numeric-boundary direction: 1375 contains 137 but is not 137.
			name: "exit status 1375 is not OOM",
			err:  "command failed with exit status 1375",
			want: ExitError,
		},
		{
			// Real OOM phrasing survives the narrowing (and stays ahead of the
			// path contamination — the same text carries a t137 path).
			name: "exit status 137 still OOM even with t137 path",
			err:  "signal: killed (exit status 137) while running .claude/worktrees/t137/run.sh",
			want: OOMKilled,
		},
		{
			name: "exit code 137 still OOM",
			err:  "Command failed with exit code 137",
			want: OOMKilled,
		},
		{
			name: "bare delimited 137 still OOM",
			err:  "process terminated: 137",
			want: OOMKilled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := h.classifyError(&HookInput{
				Error:         tc.err,
				HookEventName: "PostToolUseFailure",
			})
			if got != tc.want {
				t.Errorf("classifyError(%q) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}
