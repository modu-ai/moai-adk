// m2_ownership_classification_test.go — SPEC-USERASSET-DEPLOY-GUARD-001,
// gate round 18 #2: a flag-less journal entry is "not written THIS run",
// not "not owned". An ALREADY-TRACKED target whose bytes a user edited is
// REQ-023 divergence (backup + preserve + report) — the pre-fix form
// classified it as a collision and lost the shipped backup and the correct
// guidance.
package userassets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJournalTrackedMismatchIsDivergenceNotCollision(t *testing.T) {
	f := newFixture(t)

	// Run 1: a clean install — the manifest now TRACKS manager-x.
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("run 1: %v", err)
	}
	const trackedKey = "claude-agents/manager-x.md"
	trackedAbs := filepath.Join(f.home, ".claude", "agents", "manager-x.md")

	// The failed-save up-to-date-run state: a staged journal whose entry
	// for the TRACKED file carries WriteCompleted=false (nothing was
	// written this run) and the OLD (installed) hash.
	seedRecoveredJournal(t, f.home, trackedKey, shaHex(readBytes(t, trackedAbs)), false)

	// The user edits the tracked file — its bytes now mismatch both the
	// journal record and the shipped copy.
	if err := os.WriteFile(trackedAbs, []byte("---\nname: manager-x\n---\nthe user's own edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if res.CollisionSkipped != 0 {
		t.Fatalf("RED-regression: the tracked, user-edited file was classified as a collision (%v) — 'not written this run' was read as 'not owned' and the divergence backup was lost", res.Collisions)
	}
	if res.DivergencePreserved != 1 {
		t.Fatalf("the tracked mismatch was not classified as divergence: %+v", res)
	}
	// The divergence contract's backup arm ran: the shipped bytes are under
	// the backup home.
	backup := filepath.Join(f.home, ".moai", "backups", "claude-agents", "manager-x.md")
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("the divergence backup is missing: %v", err)
	}
	// The user's bytes stay on disk (REQ-010/REQ-023 inviolability).
	if got := readBytes(t, trackedAbs); string(got) != "---\nname: manager-x\n---\nthe user's own edit\n" {
		t.Fatalf("the user's edit was clobbered: %q", got)
	}
}
