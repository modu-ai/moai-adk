// Package harness — Curator production wiring (SPEC-HARNESS-EVOLVE-003 M5).
//
// This file is the PRODUCTION CALLER that drives the EVOLVE-002 write-layer API
// (TierGatedWrite + WriteManagedBlockGated) out of INERT status. The baseline of
// 0 production callers (research.md §B.2) rises to ≥1 here.
//
// REQ-HEV3-005..011 + REQ-HEV3-026/027/035:
//   - L5 ApprovalDecision + RejectionRecorder injection (REQ-HEV3-005/007)
//   - A7 registry early-block consult (REQ-HEV3-035)
//   - GLM observe-only gate (REQ-HEV3-026/027 — in the dispatch layer, NOT the writer)
//   - Tier→surface dispatch + content.Tier read (REQ-HEV3-003/010)
//   - No autonomous write path (REQ-HEV3-006 — no direct WriteManagedBlock bypass)
//   - DISJOINT branches: approval → TierGatedWrite (sole write); rejection →
//     WriteManagedBlockGated (RejectionRecorder). Never both (D3 double-write-avoidance).
//   - No inert wiring / no feature-flag-off-by-default (REQ-HEV3-011)
//
// This file lives in the harness package (NOT curator) so it CAN import both
// harness types (NegativeEvidence, A7 registry) and curator types (TierGatedWrite,
// WriteManagedBlockGated, Dispatch). harness already imports curator via layer3.go.
package harness

import (
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// CuratorDispatchConfig configures the Curator production dispatch.
type CuratorDispatchConfig struct {
	// RegistryPath is the .moai/state/negative-evidence.jsonl path (A7 consult).
	RegistryPath string

	// ModelClass is the session model class ("claude", "glm", etc.) for the
	// observe-only gate (REQ-HEV3-026). Empty means "unknown — admit".
	ModelClass string
}

// CuratorDispatch is the production wiring that drives the EVOLVE-002 write-layer
// API (H-4 resolved: a separate Dispatch type; Applier is PRESERVE for the
// existing harness-applier path). The orchestrator drives a two-phase cycle:
//
//  1. Prepare — A7 early-block + GLM observe-only gate + tier→surface dispatch.
//     Returns a ProposalArtifact for the orchestrator to surface via L5 AskUserQuestion.
//  2. ExecuteCuratorDecision — the orchestrator re-enters with the ApprovalDecision.
//     Approval → curator.TierGatedWrite (the SOLE write). Rejection →
//     curator.WriteManagedBlockGated (RejectionRecorder + A7 register).
//
// The Dispatch itself does NOT call AskUserQuestion (subagent boundary, REQ-HEV3-031).
type CuratorDispatch struct {
	cfg CuratorDispatchConfig
}

// NewCuratorDispatch constructs a production Dispatch.
func NewCuratorDispatch(cfg CuratorDispatchConfig) *CuratorDispatch {
	return &CuratorDispatch{cfg: cfg}
}

// CuratorProposalArtifact is the output of Prepare — the proposal + resolved
// surface + gate verdicts, ready for the orchestrator's L5 AskUserQuestion round.
type CuratorProposalArtifact struct {
	// TargetSurface is the resolved write target (path + BlockType + tier).
	TargetSurface curator.SurfaceTarget

	// Content is the block payload with the explicit Tier field set (REQ-HEV3-010).
	Content curator.BlockContent

	// Observations is the pattern's evidence count (machine signal).
	Observations int

	// PatternKey is the structural pattern token (for A7 + lineage recording).
	PatternKey string

	// IsGLMObserveOnly is true when the session is GLM and the proposal is recorded
	// observe-only (no write — REQ-HEV3-026). When true, ExecuteCuratorDecision
	// records in the ledger but does NOT write any Learned surface.
	IsGLMObserveOnly bool
}

// Prepare evaluates a Curator proposal through A7 + GLM + tier→surface dispatch
// (REQ-HEV3-035/026/003). Returns a ProposalArtifact for the L5 round, or an
// early-block error (ErrReProposalSuppressed).
//
// This is PHASE 1 of the two-phase dispatch cycle. The orchestrator receives
// the artifact, runs the L5 AskUserQuestion round, and re-enters via
// ExecuteCuratorDecision with the ApprovalDecision.
func (d *CuratorDispatch) Prepare(input curator.CuratorProposalInput) (*CuratorProposalArtifact, error) {
	// ── 0. A7 registry early-block (REQ-HEV3-035) ──────────────────────────
	// The A7 consult fires BEFORE any gate (L2/L3/L5) — same-key re-proposal
	// is blocked early per design.md §A.1 step 0.
	if d.cfg.RegistryPath != "" {
		entries, err := ReadNegativeEvidence(d.cfg.RegistryPath)
		if err == nil { // Edge-2: read failure is non-fatal (empty registry)
			check := IsReProposalSuppressed(entries, input.PatternKey, time.Now())
			if check.Suppressed {
				return nil, fmt.Errorf("%w: pattern key %q suppressed (%s)",
					ErrReProposalSuppressed, input.PatternKey, check.Reason)
			}
		}
	}

	// ── 1. GLM observe-only gate (REQ-HEV3-026/027) ────────────────────────
	// The gate lives in the dispatch layer (upstream of the writer), NOT in the
	// writer itself (REQ-HEV3-027 — the writer is a mechanical primitive).
	isGLM := IsGLMSession(d.cfg.ModelClass)

	// ── 2. Tier→surface dispatch (REQ-HEV3-003/004/010) ────────────────────
	surface, content, err := curator.PrepareTierDispatch(input)
	if err != nil {
		return nil, err
	}

	return &CuratorProposalArtifact{
		TargetSurface:    surface,
		Content:          content,
		Observations:     input.Observations,
		PatternKey:       input.PatternKey,
		IsGLMObserveOnly: isGLM,
	}, nil
}

// IsGLMSession returns true when the model class indicates a GLM session
// (REQ-HEV3-026 — GLM sessions observe only; Tier 3+ promotion is observe-only).
// The check is case-insensitive on the model class string.
func IsGLMSession(modelClass string) bool {
	switch normalizeModelClass(modelClass) {
	case "glm":
		return true
	default:
		return false
	}
}

// normalizeModelClass lowercases + trims the model class for comparison.
func normalizeModelClass(mc string) string {
	mc = strings.ToLower(strings.TrimSpace(mc))
	return mc
}

// ExecuteCuratorDecision is PHASE 2 of the dispatch cycle. The orchestrator
// re-enters with the L5 ApprovalDecision. The two branches are DISJOINT
// (D3 double-write-avoidance invariant):
//
//   - Approval  → curator.TierGatedWrite (the SOLE write entry point on the
//     approval path). WriteManagedBlockGated is NOT called here — its internal
//     WriteManagedBlock delegate would double-write (REQ-HEV3-008).
//   - Rejection → curator.WriteManagedBlockGated (RejectionRecorder + no write).
//     TierGatedWrite is NOT called here (REQ-HEV3-009).
//
// REQ-HEV3-006: there is NO autonomous write path. Both branches require a live
// ApprovalDecision; a zero-value decision{Approved:false} rejects.
//
// recorder is the RejectionRecorder injected by the caller (bound to the lineage
// writer). It MAY be nil (rejection is still enforced — the file is not written).
//
// This function is the PRODUCTION CALLER for TierGatedWrite (REQ-HEV3-008) and
// WriteManagedBlockGated (REQ-HEV3-009) — the baseline-0 → ≥1 reachability breakthrough.
func (d *CuratorDispatch) ExecuteCuratorDecision(
	artifact *CuratorProposalArtifact,
	decision curator.ApprovalDecision,
	recorder curator.RejectionRecorder,
) error {
	// GLM observe-only: record in ledger (if recorder provided), return without writing.
	if artifact.IsGLMObserveOnly {
		if recorder != nil {
			_ = recorder("observe-only-glm", "GLM session records observe-only — no Learned surface write")
		}
		return nil
	}

	// ── DISJOINT branch point (D3) ─────────────────────────────────────────
	if decision.Approved {
		// 7a. Approval path: curator.TierGatedWrite is the SOLE write (REQ-HEV3-008).
		// WriteManagedBlockGated is NOT called on this branch (double-write-avoidance).
		return curator.TierGatedWrite(
			artifact.TargetSurface.Path,
			artifact.TargetSurface.BlockType,
			artifact.Observations,
			artifact.Content,
		)
	}

	// 7b. Rejection path: curator.WriteManagedBlockGated (REQ-HEV3-009).
	// TierGatedWrite is NOT called on this branch. The gate invokes the
	// RejectionRecorder and returns ErrApprovalRejected WITHOUT touching the file.
	err := curator.WriteManagedBlockGated(
		artifact.TargetSurface.Path,
		artifact.TargetSurface.BlockType,
		artifact.Content,
		decision,
		recorder,
	)

	// REQ-HEV3-019: register the rejection in the A7 registry (register-on-reject).
	if err != nil && d.cfg.RegistryPath != "" {
		_ = AppendNegativeEvidence(d.cfg.RegistryPath, NegativeEvidence{
			PatternKey:           artifact.PatternKey,
			Outcome:              NegativeEvidenceRejected,
			Timestamp:            time.Now(),
			EvidenceCountAtEvent: artifact.Observations,
			CooldownUntil:        time.Now().Add(NegativeEvidenceCooldown),
			GateOrigin:           GateOriginL5,
		})
	}

	return err
}
