package cli

// codex_debug_flag_test.go — SPEC-CODEX-DEBUG-MODE-001 M1: the codex
// launcher's debug flag surface (AC-001..AC-004). The debug tokens are
// launcher-owned on this launcher — the codex CLI rejects them (verified
// 0.159.3) — so a pre--- token is stripped before the verb lookup (REQ-001),
// a post--- token is codex's own verbatim argument with no activation
// (REQ-002), and a readout verb refuses the token with the named diagnostic
// (REQ-003), mirroring the --spawn readout-refusal discipline.
//
// The "debug trace is active" half of AC-001/002 is observable only once the
// M2 trace engine exists; the codex_debug_trace_test.go cells assert it. This
// file pins the flag surface: what the child receives, what refuses, and
// where the tokens may appear.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCodexDebugTokenStrippedFromChildArgs(t *testing.T) {
	root := t.TempDir()
	// The -w row resolves a real fixture tree; the writer check is pinned
	// open by the capture harness.
	if err := os.MkdirAll(filepath.Join(root, ".moai", "worktrees", "wt-a"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "short spelling on the cli verb (AC-001)", args: []string{"cli", "-d"}},
		{name: "bare form with the long spelling", args: []string{"-d"}},
		{name: "long spelling beside -w (AC-002)", args: []string{"-w", "wt-a", "--debug"}},
		{name: "repeated tokens are idempotent", args: []string{"-d", "-d"}},
		{name: "both spellings together", args: []string{"cli", "-d", "--debug"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			_, _, err := runCodexCmd(t, tc.args...)
			if err != nil {
				t.Fatalf("runCodex(%v): %v", tc.args, err)
			}
			if cap.count() != 1 {
				t.Fatalf("launches = %d, want 1", cap.count())
			}
			argv := cap.records[0].Argv
			for _, token := range []string{"-d", "--debug"} {
				if slices.Contains(argv, token) {
					t.Errorf("child argv carries the debug token %q — the codex CLI rejects it: %v", token, argv)
				}
			}
			if tc.args[0] == "-w" && !strings.HasSuffix(filepath.Clean(cap.records[0].Dir), filepath.Join(".moai", "worktrees", "wt-a")) {
				t.Errorf("launch dir = %q, want the wt-a fixture tree", cap.records[0].Dir)
			}
		})
	}
}

// The -w value position is never consumed by the debug scan: `-w -d` leaves
// -w without a value, which is the worktree diagnostic — not a silent launch
// with -d forwarded (acceptance §D.1, token-shape parity with stripSpawnFlag).
func TestCodexDebugScanNeverConsumesWorktreeValue(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	_, stderr, err := runCodexCmd(t, "-w", "-d")
	if err == nil {
		t.Fatalf("runCodex(-w -d) = nil error, want the worktree-value diagnostic")
	}
	if !strings.Contains(stderr, codexWorktreeValueDiag) {
		t.Errorf("stderr = %q, want it to name %q", stderr, codexWorktreeValueDiag)
	}
	if cap.count() != 0 {
		t.Errorf("failed -w parse started %d launches, want 0", cap.count())
	}
}

func TestCodexDebugTokenAfterDashDashForwarded(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "post--- short token on the cli verb (AC-003)", args: []string{"cli", "--", "-d"}},
		{name: "empty head, only -- and the token", args: []string{"--", "-d"}},
		{name: "post--- long token", args: []string{"--", "--debug"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			_, _, err := runCodexCmd(t, tc.args...)
			if err != nil {
				t.Fatalf("runCodex(%v): %v", tc.args, err)
			}
			if cap.count() != 1 {
				t.Fatalf("launches = %d, want 1", cap.count())
			}
			argv := cap.records[0].Argv
			if !slices.Contains(argv, tc.args[len(tc.args)-1]) {
				t.Errorf("child argv dropped the post--- token %q: %v", tc.args[len(tc.args)-1], argv)
			}
		})
	}
}

// A readout starts nothing there is nothing to trace: the debug token on
// `moai codex status` refuses with the named diagnostic and exit code 1,
// mirroring the --spawn readout refusal (AC-004).
func TestCodexDebugTokenRefusedOnReadoutVerb(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	_, stderr, err := runCodexCmd(t, "status", "-d")
	var ecErr *exitCodeError
	if !errors.As(err, &ecErr) || ecErr.ExitCode() != 1 {
		t.Fatalf("runCodex(status -d) err = %v, want exitCodeError(1)", err)
	}
	const wantDiag = "-d/--debug applies to the launch verbs only (moai codex cli -d / moai codex app -d)"
	if stderr != wantDiag+"\n" {
		t.Errorf("stderr = %q, want the named refusal %q", stderr, wantDiag)
	}
	if cap.count() != 0 {
		t.Errorf("refused readout started %d launches, want 0", cap.count())
	}
}
