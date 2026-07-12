// Package curator — L2 shadow-apply canary for the Curator write path
// (SPEC-HARNESS-EVOLVE-003 M4).
//
// REQ-HEV3-012: shadow-apply the proposed block to an in-memory evaluation
// (NOT the real file) and run regression checks.
// REQ-HEV3-013: regression gate on held-out signals (3K budget, 20-bullet cap,
// anti-fabrication, marker integrity).
//
// The existing safety.EvaluateCanary + safety.CanaryVeto machinery serves the
// harness-applier path (session-metric canary). This file is the Curator-path
// sibling: content-regression canary (design.md §C.3).
//
// This file does NOT import the harness package (harness imports curator via
// layer3.go — importing back would cycle). The result type ShadowCanaryResult
// is curator-local; the caller (dispatch / harness layer) converts it to
// harness.CanaryResult when wiring into the Pipeline.
package curator

import "fmt"

// ShadowCanaryResult is the result of a shadow-apply canary check.
// The caller converts this to harness.CanaryResult when wiring into the Pipeline.
type ShadowCanaryResult struct {
	// Rejected is true when the proposed write fails a held-out regression signal.
	Rejected bool

	// Reason explains why the write was rejected (empty when Rejected is false).
	Reason string
}

// ShadowApplyCanary evaluates a proposed Curator managed-block write against
// held-out regression signals WITHOUT touching the real file (REQ-HEV3-012/013).
//
// The block content is validated in-memory against the same enforcement checks
// the real writer would apply. When Rejected == true, the proposal MUST NOT
// reach the file.
//
// Regression-gate held-out signals (design.md §C.3):
//   - 3K digest budget (REQ-HEV2-008, MaxDigestBlockChars) — Tier-4 only
//   - 20-bullet cap (REQ-HEV2-009, MaxDigestBullets) — Tier-4 only
//   - anti-fabrication forbidden content (REQ-HEV2-011) — Tier-4 only
//
// Tier-3 (BlockTypeLearnedLocal) is an append-only surface with no budget
// enforcement; ShadowApplyCanary passes it without rejection.
func ShadowApplyCanary(blockType BlockType, content BlockContent) ShadowCanaryResult {
	if blockType == BlockTypeLearnedWorkflow {
		if err := validateDigestBlock(content); err != nil {
			return ShadowCanaryResult{
				Rejected: true,
				Reason:   fmt.Sprintf("L2 shadow-apply regression gate: %v", err),
			}
		}
	}
	return ShadowCanaryResult{}
}
