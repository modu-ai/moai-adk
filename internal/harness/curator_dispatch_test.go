// Package harness — Curator production wiring tests (SPEC-HARNESS-EVOLVE-003 M5).
// REQ-HEV3-005..011, 026/027/035: the full Dispatch cycle + production callers.
package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// makeManagedBlockFile creates a temp CLAUDE.md with an existing managed block
// so WriteManagedBlock/TierGatedWrite can replace it.
func makeManagedBlockFile(t *testing.T, path string) {
	t.Helper()
	content := "# CLAUDE.md\n\n## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n- existing bullet\n<!-- moai:learned-end -->\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write managed block file: %v", err)
	}
}

// TestExecuteCuratorDecision_ApprovalPath verifies REQ-HEV3-008 (production caller
// for TierGatedWrite on the approval path) + AC-HEV3-007b (file IS written + lineage "applied").
func TestExecuteCuratorDecision_ApprovalPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeMd := filepath.Join(dir, "CLAUDE.md")
	makeManagedBlockFile(t, claudeMd)
	registryPath := filepath.Join(dir, "negative-evidence.jsonl")

	dispatch := NewCuratorDispatch(CuratorDispatchConfig{
		RegistryPath: registryPath,
		ModelClass:   "claude",
	})

	artifact := &CuratorProposalArtifact{
		TargetSurface: curator.SurfaceTarget{Path: claudeMd, BlockType: curator.BlockTypeLearnedWorkflow, Tier: 4},
		Content: curator.BlockContent{
			Tier:    4,
			Bullets: []curator.Bullet{{LedgerKey: "lk-1", Text: "workflow pattern A"}},
		},
		Observations: 10,
		PatternKey:   "feature+plan+autopilot+success",
	}

	before, _ := os.ReadFile(claudeMd)

	decision := curator.ApprovalDecision{Approved: true}
	var recordedOutcome string
	recorder := func(outcome, rationale string) error {
		recordedOutcome = outcome
		return nil
	}

	err := dispatch.ExecuteCuratorDecision(artifact, decision, recorder)
	if err != nil {
		t.Fatalf("ExecuteCuratorDecision (approval) failed: %v", err)
	}

	// The file MUST have changed (the write landed).
	after, _ := os.ReadFile(claudeMd)
	if string(before) == string(after) {
		t.Error("file unchanged after approved write — TierGatedWrite did not fire (REQ-HEV3-008)")
	}

	// The recorder MUST NOT have been called on the approval path (no rejection).
	if recordedOutcome != "" {
		t.Errorf("recorder called with outcome %q on approval path — must not fire", recordedOutcome)
	}
}

// TestExecuteCuratorDecision_RejectionPath verifies REQ-HEV3-009 (production caller
// for WriteManagedBlockGated on the rejection path) + AC-HEV3-007a (file NOT touched +
// lineage "rejected" + A7 registry gains an entry).
func TestExecuteCuratorDecision_RejectionPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeMd := filepath.Join(dir, "CLAUDE.md")
	makeManagedBlockFile(t, claudeMd)
	registryPath := filepath.Join(dir, "negative-evidence.jsonl")

	dispatch := NewCuratorDispatch(CuratorDispatchConfig{
		RegistryPath: registryPath,
		ModelClass:   "claude",
	})

	artifact := &CuratorProposalArtifact{
		TargetSurface: curator.SurfaceTarget{Path: claudeMd, BlockType: curator.BlockTypeLearnedWorkflow, Tier: 4},
		Content: curator.BlockContent{
			Tier:    4,
			Bullets: []curator.Bullet{{LedgerKey: "lk-1", Text: "workflow pattern A"}},
		},
		Observations: 10,
		PatternKey:   "feature+plan+autopilot+success",
	}

	before, _ := os.ReadFile(claudeMd)

	decision := curator.ApprovalDecision{Approved: false, Rationale: "user rejected at L5"}
	var recordedOutcome, recordedRationale string
	recorder := func(outcome, rationale string) error {
		recordedOutcome = outcome
		recordedRationale = rationale
		return nil
	}

	err := dispatch.ExecuteCuratorDecision(artifact, decision, recorder)
	if err == nil {
		t.Fatal("ExecuteCuratorDecision (rejection) = nil, want ErrApprovalRejected")
	}

	// AC-HEV3-007a: file MUST NOT be touched.
	after, _ := os.ReadFile(claudeMd)
	if string(before) != string(after) {
		t.Error("file changed after rejection — WriteManagedBlockGated wrote despite rejection (REQ-HEV3-007)")
	}

	// The recorder MUST have been called with "rejected".
	if recordedOutcome != "rejected" {
		t.Errorf("recorder outcome = %q, want \"rejected\"", recordedOutcome)
	}
	if recordedRationale != "user rejected at L5" {
		t.Errorf("recorder rationale = %q, want the L5 rationale", recordedRationale)
	}

	// AC-HEV3-019: A7 registry gains an entry with outcome "rejected".
	entries, _ := ReadNegativeEvidence(registryPath)
	if len(entries) != 1 {
		t.Fatalf("A7 registry len = %d, want 1 (register-on-reject REQ-HEV3-019)", len(entries))
	}
	if entries[0].Outcome != NegativeEvidenceRejected {
		t.Errorf("A7 entry Outcome = %q, want %q", entries[0].Outcome, NegativeEvidenceRejected)
	}
}

// TestCuratorDispatch_Prepare_A7EarlyBlock verifies AC-HEV3-021a / REQ-HEV3-035:
// a same-key re-proposal is blocked BEFORE reaching L2/L3/L5 (early-block).
func TestCuratorDispatch_Prepare_A7EarlyBlock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	registryPath := filepath.Join(dir, "negative-evidence.jsonl")

	// Pre-populate the registry with a suppressed entry for the pattern key.
	now := time.Now()
	_ = AppendNegativeEvidence(registryPath, NegativeEvidence{
		PatternKey:           "feature+plan+autopilot+success",
		Outcome:              NegativeEvidenceRejected,
		Timestamp:            now,
		CooldownUntil:        now.Add(48 * time.Hour), // not elapsed
		NewEvidenceSinceEvent: 0,                      // < N=3
		GateOrigin:           GateOriginL3,
	})

	dispatch := NewCuratorDispatch(CuratorDispatchConfig{RegistryPath: registryPath, ModelClass: "claude"})
	_, err := dispatch.Prepare(curator.CuratorProposalInput{
		TargetPath:   "CLAUDE.md",
		PatternKey:   "feature+plan+autopilot+success",
		Observations: 10,
		BlockType:    curator.BlockTypeLearnedWorkflow,
	})
	if err == nil {
		t.Fatal("Prepare = nil error, want ErrReProposalSuppressed (A7 early-block — AC-HEV3-021a)")
	}
}

// TestCuratorDispatch_Prepare_GLMObserveOnly verifies AC-HEV3-026 / REQ-HEV3-027:
// a GLM session's Tier-3+ proposal is recorded observe-only (no write).
func TestCuratorDispatch_Prepare_GLMObserveOnly(t *testing.T) {
	t.Parallel()

	dispatch := NewCuratorDispatch(CuratorDispatchConfig{
		RegistryPath: filepath.Join(t.TempDir(), "neg.jsonl"),
		ModelClass:   "glm",
	})

	artifact, err := dispatch.Prepare(curator.CuratorProposalInput{
		TargetPath:   "CLAUDE.md",
		PatternKey:   "feature+plan+autopilot+success",
		Observations: 10,
		BlockType:    curator.BlockTypeLearnedWorkflow,
	})
	if err != nil {
		t.Fatalf("Prepare error: %v", err)
	}
	if !artifact.IsGLMObserveOnly {
		t.Error("IsGLMObserveOnly = false, want true (GLM session — REQ-HEV3-026)")
	}

	// Even with Approved=true, the GLM path records observe-only and does NOT write.
	dir := t.TempDir()
	claudeMd := filepath.Join(dir, "CLAUDE.md")
	makeManagedBlockFile(t, claudeMd)
	artifact.TargetSurface.Path = claudeMd

	before, _ := os.ReadFile(claudeMd)
	var recordedOutcome string
	recorder := func(outcome, rationale string) error {
		recordedOutcome = outcome
		return nil
	}

	err = dispatch.ExecuteCuratorDecision(artifact, curator.ApprovalDecision{Approved: true}, recorder)
	if err != nil {
		t.Fatalf("ExecuteCuratorDecision error: %v", err)
	}
	after, _ := os.ReadFile(claudeMd)
	if string(before) != string(after) {
		t.Error("file changed under GLM observe-only — MUST NOT write (REQ-HEV3-026)")
	}
	if recordedOutcome != "observe-only-glm" {
		t.Errorf("recorder outcome = %q, want \"observe-only-glm\"", recordedOutcome)
	}
}

// TestIsGLMSession verifies the model-class gate helper.
func TestIsGLMSession(t *testing.T) {
	t.Parallel()

	cases := []struct {
		mc   string
		want bool
	}{
		{"glm", true},
		{"GLM", true},
		{" Glm ", true},
		{"claude", false},
		{"opus", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsGLMSession(tc.mc); got != tc.want {
			t.Errorf("IsGLMSession(%q) = %v, want %v", tc.mc, got, tc.want)
		}
	}
}
