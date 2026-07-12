// Package harness — full Curator cycle integration test (SPEC-HARNESS-EVOLVE-003 M6).
// Exercises the complete chain: Prepare -> rejection -> A7 register -> re-Prepare -> early-block.
package harness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// TestIntegration_FullCuratorCycle_RejectThenBlock verifies the end-to-end Curator
// lifecycle: a proposal is prepared, rejected at L5, registered in A7, and a
// subsequent same-key re-proposal is early-blocked (REQ-HEV3-019/021/035).
func TestIntegration_FullCuratorCycle_RejectThenBlock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeMd := filepath.Join(dir, "CLAUDE.md")
	registryPath := filepath.Join(dir, "negative-evidence.jsonl")

	// Seed the managed block file.
	seed := "# CLAUDE.md\n\n## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n- existing\n<!-- moai:learned-end -->\n"
	if err := os.WriteFile(claudeMd, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	dispatch := NewCuratorDispatch(CuratorDispatchConfig{
		RegistryPath: registryPath,
		ModelClass:   "claude",
	})

	input := curator.CuratorProposalInput{
		TargetPath:   claudeMd,
		PatternKey:   "feature+plan+autopilot+success",
		Observations: 10,
		BlockType:    curator.BlockTypeLearnedWorkflow,
		Content: curator.BlockContent{
			Bullets: []curator.Bullet{{LedgerKey: "lk-1", Text: "workflow pattern A"}},
		},
	}

	// ── Phase 1: Prepare (passes — no prior A7 entry) ──────────────────────
	artifact, err := dispatch.Prepare(input)
	if err != nil {
		t.Fatalf("Prepare (first) failed: %v", err)
	}
	artifact.TargetSurface.Path = claudeMd

	// ── Phase 2: orchestrator returns rejection ─────────────────────────────
	decision := curator.ApprovalDecision{Approved: false, Rationale: "integration test rejection"}
	rejectionErr := dispatch.ExecuteCuratorDecision(artifact, decision, nil)
	if rejectionErr == nil {
		t.Fatal("ExecuteCuratorDecision (rejection) = nil, want ErrApprovalRejected")
	}

	// ── Verify: file untouched + A7 registry has entry ──────────────────────
	after, _ := os.ReadFile(claudeMd)
	if string(after) != seed {
		t.Error("file changed after rejection — must be untouched")
	}
	entries, _ := ReadNegativeEvidence(registryPath)
	if len(entries) != 1 {
		t.Fatalf("A7 registry len = %d, want 1 (register-on-reject)", len(entries))
	}
	if entries[0].Outcome != NegativeEvidenceRejected {
		t.Errorf("A7 Outcome = %q, want rejected", entries[0].Outcome)
	}

	// ── Phase 3: re-Prepare same key → early-blocked (REQ-HEV3-021a/035) ────
	_, err = dispatch.Prepare(input)
	if err == nil {
		t.Fatal("re-Prepare = nil, want ErrReProposalSuppressed (A7 early-block — AC-HEV3-021a)")
	}
}
