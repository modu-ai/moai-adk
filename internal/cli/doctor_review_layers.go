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

	records, err := factory.ReadAll(projectRoot)
	if err != nil {
		check.Message = "turn-end review gate is on; factory records unreadable, overlap not judged"
		return check
	}
	for i := range records {
		if records[i].Role == "lane" {
			check.Status = uikit.CheckWarn
			check.Message = "turn-end review_gate and the lane card-review stage are both active; the same card diff is reviewed twice"
			check.Detail = "A factory lane session record exists while workflow.codex.review_gate.enabled is true. " +
				"The card-review stage has no switch of its own; set workflow.codex.review_gate.enabled: false to keep a single review layer."
			return check
		}
	}

	check.Message = "turn-end review gate is on; no factory lane session record"
	return check
}
