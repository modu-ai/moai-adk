package curator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// learnedBlockABC is a CLAUDE.md fixture with a MOAI:LEARNED-WORKFLOW block
// containing bullets [A, B, C], each carrying a distinct ledger_key.
const learnedBlockABC = "# Project\n\n" +
	"## MOAI:LEARNED-WORKFLOW\n" +
	"<!-- moai:learned-start -->\n" +
	"- rule A <!-- key: A -->\n" +
	"- rule B <!-- key: B -->\n" +
	"- rule C <!-- key: C -->\n" +
	"<!-- moai:learned-end -->\n"

// TestRollbackTrigger_MechanicalOnly_NoModelSelfReport (AC-HEV2-042,
// REQ-HEV2-033 + REQ-HEV2-022): the rollback is driven purely by the mechanical
// snapshot artifact (a snapshot directory), never by model self-report. The
// full chain AddBullet(D) → CreateSurfaceSnapshot → DeleteBullet(B) →
// RestoreSnapshot restores the file byte-identically to the pre-write state,
// leaves the managed-block markers intact (not orphaned), and is idempotent
// (a deterministic function of the snapshot, not a model decision).
func TestRollbackTrigger_MechanicalOnly_NoModelSelfReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(path, []byte(learnedBlockABC), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	preWrite, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pre-write: %v", err)
	}

	// Mechanical snapshot BEFORE mutation. The ONLY rollback trigger input is
	// this snapshot dir — a filesystem artifact, not a model decision.
	snapBase := filepath.Join(dir, "snapshots")
	snapshotDir, err := harness.CreateSurfaceSnapshot(snapBase, "hev2-042", []harness.SurfaceRestoreUnit{
		{LearnedSurface: "claude.md.learned-workflow", OriginalPath: path, BulletsAffected: []string{"D", "B"}},
	})
	if err != nil {
		t.Fatalf("CreateSurfaceSnapshot: %v", err)
	}

	// Mutate the digest via CRUD: add D, delete B.
	if err := curator.AddBullet(path, curator.BlockTypeLearnedWorkflow,
		curator.Bullet{LedgerKey: "D", Text: "rule D"}); err != nil {
		t.Fatalf("AddBullet: %v", err)
	}
	if err := curator.DeleteBullet(path, curator.BlockTypeLearnedWorkflow, "B"); err != nil {
		t.Fatalf("DeleteBullet: %v", err)
	}
	mutated, _ := os.ReadFile(path)
	if string(mutated) == string(preWrite) {
		t.Fatalf("mutation had no effect — the test cannot prove rollback")
	}

	// Mechanical rollback: RestoreSnapshot's ONLY input is the snapshot dir.
	if err := harness.RestoreSnapshot(snapshotDir); err != nil {
		t.Fatalf("RestoreSnapshot: %v", err)
	}
	restored, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read restored: %v", err)
	}
	if string(restored) != string(preWrite) {
		t.Errorf("rollback not byte-identical:\n pre-write=%q\n restored =%q", preWrite, restored)
	}
	// Markers intact (not orphaned).
	rs := string(restored)
	if !strings.Contains(rs, "moai:learned-start") || !strings.Contains(rs, "moai:learned-end") {
		t.Errorf("managed-block markers orphaned after rollback")
	}

	// Idempotent: restoring again from the same mechanical artifact yields the
	// same bytes (the trigger is a deterministic function of the snapshot).
	if err := harness.RestoreSnapshot(snapshotDir); err != nil {
		t.Fatalf("second RestoreSnapshot: %v", err)
	}
	again, _ := os.ReadFile(path)
	if string(again) != string(preWrite) {
		t.Errorf("rollback not idempotent — second restore diverged")
	}
}
