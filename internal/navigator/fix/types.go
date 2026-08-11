// Package fix implements the M3 Fix layer of the BAS Falconer loop
// (SPEC-NAVIGATOR-SYNC-005): the deterministic diff-scope engine that
// identifies stale doc subtrees and emits a draft-request manifest for the
// orchestrator-mediated AI-draft delegation.
//
// The package CONSUMES the M0 graph types from internal/navigator/sync
// (REQ-NS5-005 bridge-not-absorb): it imports sync.Graph / sync.Edge /
// sync.Node read-only and NEVER mutates the M0/M1/M2/M4 producer surfaces.
// The AI-draft step (layer 2) is orchestrator-mediated via a manager-develop
// delegation — the Go engine contains ZERO LLM-client imports (REQ-NS5-007).
package fix

import (
	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// Provenance stamps a draft-request to a git baseline (REQ-NS5-004).
// FixCommitSHA = `git rev-parse HEAD`; BaselineCommitSHA = the resolved
// baseline commit; CapturedAt = the committer date of FixCommitSHA
// (`git log -1 --format=%cI`). No wall-clock timestamp is used, so two runs
// on the same HEAD + baseline produce byte-identical output (idempotence,
// REQ-NS5-004b). Carries forward M0's Provenance model
// (internal/navigator/sync/schema.go:63) with fix-specific field names.
type Provenance struct {
	FixCommitSHA      string `json:"fix_commit_sha"`
	BaselineCommitSHA string `json:"baseline_commit_sha"`
	CapturedAt        string `json:"captured_at"`
}

// WorkItemRef is a compact reference to the M2 work-item that seeded a stale
// subtree, preserving M2 provenance (REQ-NS5-004 work_item_refs[]). Each ref
// carries the M2 work-item's source_kind + owner_path + action so the AI
// draft delegation inherits the per-subtree fix-strategy hint.
type WorkItemRef struct {
	SourceKind string `json:"source_kind"`
	OwnerPath  string `json:"owner_path"`
	Action     string `json:"action"`
}

// DiffScopeEntry is one (doc_surface, subtree_id, stale_reason, work_item_ref)
// triple — the manifest the AI draft consumes (REQ-NS5-003). The diff_scope[]
// array is deduplicated by (doc_surface, subtree_id) and sorted for
// byte-stable output.
type DiffScopeEntry struct {
	DocSurface  string       `json:"doc_surface"`
	SubtreeID   string       `json:"subtree_id"`
	StaleReason string       `json:"stale_reason"`
	WorkItemRef *WorkItemRef `json:"work_item_ref,omitempty"`
}

// DraftRequest is the layer-1 handoff artifact (request.json schema,
// plan.md §C.3). Produced by the deterministic Go engine; consumed by the
// orchestrator's manager-develop AI-draft delegation. The artifact is
// idempotent (same HEAD + baseline + inputs → byte-identical output) and
// provenance-attributed (no wall-clock).
type DraftRequest struct {
	Provenance        Provenance        `json:"provenance"`
	DiffScope         []DiffScopeEntry  `json:"diff_scope"`
	WorkItemRefs      []WorkItemRef     `json:"work_item_refs"`
	DraftInstructions DraftInstructions `json:"draft_instructions"`
}

// DraftInstructions carries per-subtree strategy hints derived from the M2
// action field (REQ-NS5-004 draft_instructions). Each strategy maps a
// subtree_id to a one-line fix directive the AI draft should apply.
type DraftInstructions struct {
	PerSubtree []SubtreeStrategy `json:"per_subtree"`
}

// SubtreeStrategy is one per-subtree strategy hint inside DraftInstructions.
type SubtreeStrategy struct {
	SubtreeID string `json:"subtree_id"`
	Strategy  string `json:"strategy"`
}

// AppliedLedger records the apply-on-approval outcome (layer 3, REQ-NS5-008c).
// Written at fix-drafts/<draft-id>/applied.json AFTER a human approval. The
// type lands in M3.1 per plan.md §D.1; the write logic is owned by M3.5
// apply.go. ApprovalTimestamp is git-committer-date (NOT wall-clock).
type AppliedLedger struct {
	Approver            string   `json:"approver"`
	ApprovalTimestamp   string   `json:"approval_timestamp"`
	AppliedSubtreeIDs   []string `json:"applied_subtree_ids"`
	ResultingLiveDocSHA string   `json:"resulting_live_doc_sha"`
}

// docSurfaceFor maps a graph node's entity_type to the doc surface file whose
// subtree the node's documentation lives in (REQ-NS5-003 subtree
// identification). This is the M3.1 contract mapping; M3.2+ may refine it.
//
//   - symbol   → capability-symbols.json (003 enrich chain output)
//   - spec     → audit-report.json (002 audit chain output)
//   - decision → capability-map.md (001 regen chain output)
func docSurfaceFor(et navsync.EntityType) string {
	switch et {
	case navsync.EntitySymbol:
		return "capability-symbols.json"
	case navsync.EntitySpec:
		return "audit-report.json"
	case navsync.EntityDecision:
		return "capability-map.md"
	default:
		return "capability-map.md"
	}
}
