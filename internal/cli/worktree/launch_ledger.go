package worktree

// Launch-ledger reclamation at worktree disposal (card t297).
//
// The launch ledger's projects map (internal/profile, launch.yaml) used to
// grow one row per worktree a `-p` launch was recorded from, and nothing
// collected the rows whose worktree later died. The write side now folds
// subtree launches into the registered project root (REQ-009); this file is
// the reclamation half: every moai worktree disposal path prunes the rows
// whose directory is gone — the row a pre-normalization binary wrote for the
// disposed tree, and any other row whose directory is already gone.

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// pruneLaunchLedgerFn is the seam the disposal paths reclaim the launch
// ledger through. It exists so tests can inject a fake: the production target
// writes ~/.moai/claude-profiles/launch.yaml, which a test must never touch.
var pruneLaunchLedgerFn = profile.PruneStaleProjectEntries

// pruneLaunchLedgerAfterDisposal reclaims the launch-ledger rows whose
// project directory died with the worktree just removed. Best-effort by
// contract: a failure warns on warn and never fails the disposal that already
// succeeded; nothing pruned produces no output.
//
// out == nil is the quiet form (--auto suppresses success output): the
// reclamation still happens, only the report line is dropped.
func pruneLaunchLedgerAfterDisposal(out, warn io.Writer) {
	pruned, err := pruneLaunchLedgerFn()
	if err != nil {
		_, _ = fmt.Fprintf(warn, "Warning: could not prune the launch ledger: %v\n", err)
		return
	}
	if out != nil && len(pruned) > 0 {
		_, _ = fmt.Fprintf(out, "Pruned %d stale launch-ledger project row(s).\n", len(pruned))
	}
}
