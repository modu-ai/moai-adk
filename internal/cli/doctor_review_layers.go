// doctor_review_layers.go — card t1683 (SPEC-REVIEW-GATE-REPAIR-001
// follow-up F4): the doctor warning for the two review layers being active
// together.
//
// The turn-end codex review gate (workflow.codex.review_gate.enabled) and the
// card-review stage of a factory lane both review the card diff, so a project
// running both reviews it twice. The card-review stage is a mandatory lane
// step with no config switch, so its presence is read from the factory
// session records (a persisted lane record) — the same store checkFactoryRun
// reads. Advisory and read-only: every failure path reads as OK and the check
// never returns CheckFail.
package cli

import (
	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// reviewLayerOverlapCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const reviewLayerOverlapCheckName = "Review Layer Overlap"

// checkReviewLayerOverlap warns when the turn-end review gate is enabled and a
// factory lane session record exists. The gate is read through the hook's own
// reader from the hook's own config root (a config-orphaned linked worktree
// reads its primary checkout; unreadable config reads as off), so doctor and
// the hook cannot disagree on "enabled"; an unreadable record store reads as
// no lane.
func checkReviewLayerOverlap(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: reviewLayerOverlapCheckName, Status: uikit.CheckOK}

	if !readCodexReviewGateEnabled(reviewGateConfigRoot(projectRoot)) {
		check.Message = "turn-end review gate is off; no overlapping review layers"
		return check
	}

	lane, readable := laneRecordPresent(projectRoot)
	switch {
	case lane:
		check.Status = uikit.CheckWarn
		check.Message = "a factory lane session record is present while the turn-end review gate is enabled; the card-review stage may review the same card diff twice"
		check.Detail = "The record may be stale: the record store carries no liveness signal, so a record left by a finished run also triggers this. " +
			"The card-review stage has no switch of its own; set workflow.codex.review_gate.enabled: false to keep a single review layer. " +
			"Scope: the Claude Stop path (the Codex Stop chain resolves its enabled flag from a different root)."
	case !readable:
		check.Message = "turn-end review gate is on; factory records unreadable, overlap not judged"
	default:
		check.Message = "turn-end review gate is on; no factory lane session record"
	}
	return check
}

// laneRecordPresent looks for a lane session record under projectRoot and under
// the primary checkout of its repository: the session-start writer keys the
// record by the session's project directory, which is the primary checkout for
// a session working in a linked worktree. readable is false when no lane was
// found and any candidate root failed to read: an absence is only reported for
// roots that were actually observed.
func laneRecordPresent(projectRoot string) (lane, readable bool) {
	roots := []string{projectRoot}
	if primary, _, err := auditreceipt.IdentifyPrimaryCheckout(projectRoot); err == nil && primary != projectRoot {
		roots = append(roots, primary)
	}
	readable = true
	for _, root := range roots {
		records, err := factory.ReadAll(root)
		if err != nil {
			readable = false
			continue
		}
		for i := range records {
			if records[i].Role == "lane" {
				return true, true
			}
		}
	}
	return false, readable
}
