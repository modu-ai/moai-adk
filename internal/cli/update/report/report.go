// Package report contains the user-facing advisory / status emission
// functions extracted from the root update command during M3d-B2
// decomposition (SPEC-CLI-TUX-V3-003). Behavior is byte-identical to the
// pre-decomposition implementation; only the package location changed.
package report

import (
	"fmt"
	"io"
)

// EmitHooksReviewGuidance writes the advisory /hooks review guidance line.
//
// Claude Code snapshots hook configuration at session startup; a template
// re-render leaves running sessions stale until /hooks is reviewed or Claude
// Code is restarted. This is emitted only on the "Template sync complete"
// branch of runUpdate (no re-render → no stale-snapshot risk → no guidance).
//
// @MX:NOTE: [AUTO] Advisory /hooks review guidance (SPEC-HOOK-CONFIG-SAFETY-001).
// @MX:SPEC: SPEC-HOOK-CONFIG-SAFETY-001
func EmitHooksReviewGuidance(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Hooks: running Claude Code sessions need a /hooks review (or a Claude Code restart) for new or changed hooks to take effect.")
}
