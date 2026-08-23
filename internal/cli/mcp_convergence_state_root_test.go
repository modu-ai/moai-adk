package cli

// Card t182 — the convergence state file follows the audited tree.
//
// The writer resolved its directory once at package load, from the process's
// own project dir, while the reader (loadConvergenceResult, multi_review_gate.go)
// takes the caller's projectDir per call. A run inside a worktree therefore
// wrote its verdict into the primary checkout, where that worktree's gate never
// looks — and where every other worktree's verdict landed too.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// The named project root wins over the process-wide state dir.
func TestPersistConvergenceResult_HonorsProjectRoot(t *testing.T) {
	primary := t.TempDir()
	worktree := t.TempDir()

	orig := convergenceStateDir
	convergenceStateDir = filepath.Join(primary, ".moai", "state")
	t.Cleanup(func() { convergenceStateDir = orig })

	r := ConvergenceResult{OverallVerdict: "pass"}
	if err := persistConvergenceResult(r, "sess-wt", worktree); err != nil {
		t.Fatalf("persistConvergenceResult: %v", err)
	}

	want := filepath.Join(worktree, ".moai", "state", "audit-multi", "sess-wt.json")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("state file not written under the named project root: %v", err)
	}
	// The reader resolves exactly this path, so the writer must not have left a
	// copy in the primary checkout instead.
	if _, ok := loadConvergenceResult(worktree, "sess-wt"); !ok {
		t.Error("loadConvergenceResult(worktree) could not read the file the writer just wrote")
	}
	stray := filepath.Join(convergenceStateDir, "audit-multi", "sess-wt.json")
	if _, err := os.Stat(stray); err == nil {
		t.Errorf("verdict also landed in the primary checkout at %s", stray)
	}
}

// An empty project root keeps the pre-existing behavior: the package-level
// convergenceStateDir, which the older tests override by assignment.
func TestPersistConvergenceResult_EmptyProjectRootKeepsStateDir(t *testing.T) {
	tmp := t.TempDir()
	orig := convergenceStateDir
	convergenceStateDir = tmp
	t.Cleanup(func() { convergenceStateDir = orig })

	if err := persistConvergenceResult(ConvergenceResult{OverallVerdict: "pass"}, "sess-default", ""); err != nil {
		t.Fatalf("persistConvergenceResult: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "audit-multi", "sess-default.json")); err != nil {
		t.Fatalf("state file not written to convergenceStateDir: %v", err)
	}
}

// End to end through the fan-out: the config's ProjectRoot decides where the
// run's verdict is recorded, not the process's own checkout.
func TestRunMultiAudit_RecordsVerdictUnderTheConfiguredRoot(t *testing.T) {
	primary := t.TempDir()
	worktree := t.TempDir()

	origDir := convergenceStateDir
	convergenceStateDir = filepath.Join(primary, ".moai", "state")
	t.Cleanup(func() { convergenceStateDir = origDir })

	rc := &recordingCaller{}
	origCall := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = origCall })

	cfg := MultiAuditConfig{SessionID: "sess-fanout", ProjectRoot: worktree}
	runMultiAudit(context.Background(), claudeReview("pass"), "uncommittedChanges", "", cfg, nil)

	if _, ok := loadConvergenceResult(worktree, "sess-fanout"); !ok {
		t.Error("no convergence state readable under the configured project root")
	}
	if _, ok := loadConvergenceResult(primary, "sess-fanout"); ok {
		t.Error("convergence state was recorded in the primary checkout instead")
	}
}
