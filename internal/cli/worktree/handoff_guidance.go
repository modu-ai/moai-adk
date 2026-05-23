// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 handoff guidance (P4 dispatch).
//
// Implements REQ-WTL-004: when --team is absent (or pane spawn fails as P4
// fallback), print paste-ready manual launch instructions to stdout instead
// of attempting any in-process exec or tmux pane spawn.
//
// The output format is intentionally simple: a single `cd <path> && moai <llm>`
// line that the user can copy verbatim into a fresh terminal. AC-WTL-004
// requires both `cd ` and ` && moai` literals to appear in stdout.

package worktree

import (
	"fmt"
	"io"
)

// printHandoff writes paste-ready manual launch instructions for the worktree
// to `out`. Used by:
//
//   - PatternP4Handoff (no --team flag) — primary path.
//   - Pattern P1/P2 fallback (tmux pane spawn failure, AC-WTL-007) — via
//     printHandoffWithError.
//
// The output contains:
//
//  1. A blank line for visual separation.
//  2. A summary line stating the worktree is ready.
//  3. A paste-ready command line: `cd <worktree-path> && moai <llm>`.
//  4. A trailing blank line.
//
// The function never returns an error; io.Writer failures are silently
// dropped (consistent with fmt.Fprintln's no-error-propagation convention in
// the rest of the worktree CLI).
func printHandoff(out io.Writer, cfg TeamLaunchConfig) {
	llm := cfg.LLM
	if llm == "" {
		llm = "cc"
	}
	_, _ = fmt.Fprintln(out, "")
	_, _ = fmt.Fprintln(out, "Worktree ready. To start a Claude session inside it, run:")
	_, _ = fmt.Fprintln(out, "")
	// AC-WTL-004: stdout must contain `cd ` AND ` && moai` literals.
	_, _ = fmt.Fprintf(out, "  cd %s && moai %s\n", cfg.WorktreePath, llm)
	_, _ = fmt.Fprintln(out, "")
}

// printHandoffWithError is the fallback variant used when an earlier dispatch
// step (tmux pane spawn, in-process exec setup, etc.) failed and we want the
// user to know WHY before showing the manual instructions.
//
// It emits a `warning:` line on `errOut` carrying the supplied reason, then
// delegates the normal handoff body to `printHandoff` on `out`. This matches
// AC-WTL-007 (P4 fallback with stderr notice) and the Windows fallback path.
//
// Both writers may be the same underlying stream (e.g., os.Stderr); callers
// SHOULD separate them so that scripts piping stdout for parsing still
// receive a clean paste-ready command.
func printHandoffWithError(out, errOut io.Writer, cfg TeamLaunchConfig, reason string) {
	_, _ = fmt.Fprintf(errOut, "warning: %s — falling back to manual handoff guidance\n", reason)
	printHandoff(out, cfg)
}
