package curator

import (
	"errors"
	"fmt"
)

// ErrApprovalRejected is returned by WriteManagedBlockGated when the L5 approval
// decision is a rejection. The file is NOT touched — no autonomous write path
// exists (REQ-HEV2-032, AP-HEV2-003).
var ErrApprovalRejected = errors.New("curator: managed-block write rejected (L5 approval)")

// ApprovalDecision is the L5 approval decision routed from the orchestrator
// (REQ-HEV2-032). The Curator writer is a subagent-side function and MUST NOT
// call AskUserQuestion (subagent boundary, REQ-HEV2-035 / C-HRA-008); the
// orchestrator runs the AskUserQuestion round and passes the decision in here.
type ApprovalDecision struct {
	// Approved is true when the orchestrator's L5 round returned an approval
	// token, false on rejection.
	Approved bool
	// Rationale is the human-supplied rejection rationale recorded to the audit
	// trail (lineage) when Approved is false. It is ignored when Approved.
	Rationale string
}

// RejectionRecorder records a rejected proposal's audit trail. The curator
// package cannot import internal/harness (harness imports curator — an import
// cycle), so the applier/orchestrator injects the recorder. It is invoked with
// the rejection outcome ("rejected") and rationale when a write is rejected, so
// the caller can append a LineageEntry (REQ-HEV2-032 "records the rejection in
// lineage"). A nil recorder is permitted (the rejection is still enforced — the
// file is not written — but no lineage entry is appended).
type RejectionRecorder func(outcome, rationale string) error

// WriteManagedBlockGated gates a managed-block write on an L5 approval decision
// (REQ-HEV2-032). When the decision is approved it delegates to WriteManagedBlock.
// When rejected it invokes recorder (if non-nil) with ("rejected", rationale)
// and returns ErrApprovalRejected WITHOUT touching the file. This is the
// "requires an approval token to execute" contract: no approval token → no
// file write.
func WriteManagedBlockGated(path string, blockType BlockType, content BlockContent, decision ApprovalDecision, recorder RejectionRecorder) error {
	if !decision.Approved {
		if recorder != nil {
			if err := recorder("rejected", decision.Rationale); err != nil {
				return fmt.Errorf("curator: rejection lineage record failed: %w", err)
			}
		}
		return ErrApprovalRejected
	}
	return WriteManagedBlock(path, blockType, content)
}
