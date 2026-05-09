// Package pr provides the `moai pr watch` CLI command and its report emitters.
// HARD: This package MUST NOT call AskUserQuestion or prompt the user directly.
// The orchestrator (skill body) is responsible for presenting the emitted
// report via AskUserQuestion. See .claude/rules/moai/core/askuser-protocol.md.
package pr

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
)

// EmitReadyToMergeReport writes a markdown-formatted report to w that the
// orchestrator can parse and present via AskUserQuestion.
//
// Format contract (Wave 2 forward-stable):
//   - Starts with a ## header identifying the PR
//   - Includes check summary (N/N pass)
//   - Lists recommended action option labeled "(권장)" (FIRST option)
//   - Lists alternative options (do not merge, investigate)
//   - Ends with a separator line
//
// The orchestrator reads this report from stdout and constructs an
// AskUserQuestion call. This function MUST NOT call AskUserQuestion.
func EmitReadyToMergeReport(w io.Writer, state ciwatch.CIState) error {
	totalReq := state.RequiredPassed + len(state.RequiredFailed) + len(state.RequiredPending)
	auxFail := len(state.AuxiliaryFailed)

	// Build advisory note.
	advisoryNote := ""
	if auxFail > 0 {
		advisoryNote = fmt.Sprintf("\n> **Advisory**: %d auxiliary check(s) failed (non-blocking — see advisory list).\n", auxFail)
	}

	// Build auxiliary check names for advisory note.
	auxNames := ""
	for _, af := range state.AuxiliaryFailed {
		auxNames += "- `" + af.Name + "` (advisory, non-blocking)\n"
	}

	report := fmt.Sprintf(`## CI Watch: PR #%d Ready-to-Merge

**Branch**: %s
**Required checks**: %d/%d pass
%s
### Required Checks Summary

All required CI checks have completed successfully.

%s
### Recommended Action

The following options are available:

1. **Merge PR #%d (권장)** — All required checks pass. Proceed with merge.
2. **Hold** — Keep PR open and monitor for additional changes.
3. **Investigate auxiliary failures** — Review advisory check failures before merging.

---
`,
		state.PRNumber,
		state.Branch,
		state.RequiredPassed, totalReq,
		advisoryNote,
		auxNames,
		state.PRNumber,
	)

	_, err := fmt.Fprint(w, report)
	return err
}

// EmitFailureHandoff writes a structured JSON handoff to w for consumption by
// the Wave 3 expert-debug agent. The JSON shape is defined by ciwatch.Handoff.
func EmitFailureHandoff(w io.Writer, state ciwatch.CIState) error {
	h := ciwatch.NewHandoff(state)

	// Collect all parts first, then write atomically to propagate errors.
	parts := make([]string, 0, len(h.FailedChecks)*2+3)
	parts = append(parts, fmt.Sprintf(`{"prNumber":%d,"branch":%q,"failedChecks":[`, h.PRNumber, h.Branch))
	for i, fc := range h.FailedChecks {
		if i > 0 {
			parts = append(parts, ",")
		}
		parts = append(parts, fmt.Sprintf(`{"name":%q,"runId":%q,"logUrl":%q,"conclusionDetail":%q}`,
			fc.Name, fc.RunID, fc.LogURL, fc.ConclusionDetail))
	}
	parts = append(parts, fmt.Sprintf(`],"auxiliaryFailCount":%d,"totalRequired":%d}`, h.AuxiliaryFailCount, h.TotalRequired))
	parts = append(parts, "\n")

	for _, p := range parts {
		if _, err := fmt.Fprint(w, p); err != nil {
			return err
		}
	}
	return nil
}
