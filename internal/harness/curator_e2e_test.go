package harness

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// mustReadE2E reads a file, failing the test on error.
func mustReadE2E(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

// TestM2VerificationChain_EndToEnd (spec.md §7 M2 verification target,
// REQ-HEV2-003..024): exercises the full write-layer chain over the EXISTING
// implementation, end to end —
//
//	AddBullet → VerifyBudgetEnforced → CreateSurfaceSnapshot → DeleteBullet →
//	RestoreSnapshot (byte-identical) → LineageEntry audit trail.
//
// This is a harness-package integration test so it can drive both the curator
// CRUD/writer surface and the harness snapshot/rollback/lineage surface in one
// realistic flow.
func TestM2VerificationChain_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	seed := "# Project\n\n## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n" +
		"- rule A <!-- key: A -->\n<!-- moai:learned-end -->\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// (1) AddBullet — add bullet B to the digest block.
	if err := curator.AddBullet(path, curator.BlockTypeLearnedWorkflow,
		curator.Bullet{LedgerKey: "B", Text: "rule B"}); err != nil {
		t.Fatalf("AddBullet: %v", err)
	}
	afterAdd := string(mustReadE2E(t, path))
	if !strings.Contains(afterAdd, "rule B") {
		t.Fatalf("AddBullet did not persist bullet B")
	}

	// (2) VerifyBudgetEnforced — an over-budget digest write (16 bullets ×
	// 200 chars = 3200 > MaxDigestBlockChars 3000, and 16 < the 20-bullet cap so
	// the BUDGET gate fires, not the cap) is rejected and the file is NOT touched
	// (REQ-HEV2-008).
	beforeBudget := afterAdd
	overBudget := make([]curator.Bullet, 0, 16)
	for i := 0; i < 16; i++ {
		overBudget = append(overBudget, curator.Bullet{Text: strings.Repeat("x", 200)})
	}
	if err := curator.WriteManagedBlock(path, curator.BlockTypeLearnedWorkflow,
		curator.BlockContent{Bullets: overBudget}); !errors.Is(err, curator.ErrDigestBudgetExceeded) {
		t.Fatalf("over-budget write should return ErrDigestBudgetExceeded, got %v", err)
	}
	if string(mustReadE2E(t, path)) != beforeBudget {
		t.Errorf("file modified despite budget rejection")
	}

	// (3) CreateSurfaceSnapshot BEFORE the next mutation.
	snapBase := filepath.Join(dir, "snapshots")
	preMutate := mustReadE2E(t, path)
	snapshotDir, err := CreateSurfaceSnapshot(snapBase, "hev2-e2e", []SurfaceRestoreUnit{
		{LearnedSurface: "claude.md.learned-workflow", OriginalPath: path, BulletsAffected: []string{"B"}},
	})
	if err != nil {
		t.Fatalf("CreateSurfaceSnapshot: %v", err)
	}

	// (4) DeleteBullet — remove B.
	if err := curator.DeleteBullet(path, curator.BlockTypeLearnedWorkflow, "B"); err != nil {
		t.Fatalf("DeleteBullet: %v", err)
	}
	if strings.Contains(string(mustReadE2E(t, path)), "rule B") {
		t.Fatalf("DeleteBullet did not remove bullet B")
	}

	// (5) RestoreSnapshot — byte-identical rollback to the pre-mutation state.
	if err := RestoreSnapshot(snapshotDir); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}
	if string(mustReadE2E(t, path)) != string(preMutate) {
		t.Errorf("rollback not byte-identical to pre-mutation state")
	}

	// (6) LineageEntry audit trail — record the rollback transition with the M5
	// additive fields (LearnedSurface / BulletsChanged / SnapshotDir), then
	// verify it round-trips through the lineage loader.
	manifest := filepath.Join(dir, "manifest.jsonl")
	if err := WriteLineageEntry(manifest, LineageEntry{
		ProposalID:     "hev2-e2e",
		TargetPath:     path,
		Decision:       "rejected", // documented enum value for a not-kept transition
		Reason:         "regression rollback (e2e verification chain)",
		LearnedSurface: "claude.md.learned-workflow",
		BulletsChanged: []string{"B"},
		SnapshotDir:    snapshotDir,
	}); err != nil {
		t.Fatalf("WriteLineageEntry: %v", err)
	}
	entries, err := LoadManifest(manifest)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 lineage entry, got %d", len(entries))
	}
	e := entries[0]
	if e.LearnedSurface != "claude.md.learned-workflow" {
		t.Errorf("lineage LearnedSurface = %q, want claude.md.learned-workflow", e.LearnedSurface)
	}
	if len(e.BulletsChanged) != 1 || e.BulletsChanged[0] != "B" {
		t.Errorf("lineage BulletsChanged = %v, want [B]", e.BulletsChanged)
	}
	if e.SnapshotDir != snapshotDir {
		t.Errorf("lineage SnapshotDir = %q, want %q", e.SnapshotDir, snapshotDir)
	}
}
