// closer_test.go — TDD coverage for atomic close orchestrator.
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 AC-LSG-001, AC-LSG-006, AC-LSG-014, AC-LSG-018, AC-LSG-022.
package spec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// closeFixtureSpec mirrors auditFixtureSpec but adds acceptance.md content.
type closeFixtureSpec struct {
	id           string
	specMD       string
	progressMD   string
	acceptanceMD string
}

func buildCloseFixture(t *testing.T, fixture closeFixtureSpec) string {
	t.Helper()
	tempDir := t.TempDir()
	dir := filepath.Join(tempDir, ".moai", "specs", fixture.id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(fixture.specMD), 0644); err != nil {
		t.Fatal(err)
	}
	if fixture.progressMD != "" {
		if err := os.WriteFile(filepath.Join(dir, "progress.md"), []byte(fixture.progressMD), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.acceptanceMD != "" {
		if err := os.WriteFile(filepath.Join(dir, "acceptance.md"), []byte(fixture.acceptanceMD), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return tempDir
}

// AC-LSG-006 — Precondition matrix validation: missing §E.5 Mx section.
func TestClose_PreconditionMissingMx(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-MX-001",
		specMD: makeSpecMD("SPEC-TEST-MX-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234
`,
		// no §E.5 Mx section
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close("SPEC-TEST-MX-001", CloseOptions{BaseDir: baseDir})
	if err == nil {
		t.Fatal("expected ErrPreconditionMissing, got nil")
	}
	if !errors.Is(err, ErrPreconditionMissing) {
		t.Errorf("error = %v, want ErrPreconditionMissing", err)
	}
	if result == nil {
		t.Fatal("result should not be nil even on precondition failure")
	}
	if len(result.PreconditionsFailed) == 0 {
		t.Error("PreconditionsFailed should list the missing precondition")
	}
	found := false
	for _, p := range result.PreconditionsFailed {
		if strings.Contains(strings.ToLower(p), "§e.5") || strings.Contains(strings.ToLower(p), "mx") {
			found = true
		}
	}
	if !found {
		t.Errorf("PreconditionsFailed should mention §E.5 Mx; got %v", result.PreconditionsFailed)
	}
	if result.Result != "failure" {
		t.Errorf("Result = %q, want failure", result.Result)
	}
}

// AC-LSG-006 — Precondition missing §E.2 sync section.
func TestClose_PreconditionMissingSync(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-SYNC-001",
		specMD: makeSpecMD("SPEC-TEST-SYNC-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.5 Mx-phase
mx_commit_sha: abc1234
`,
		// no §E.2 sync section
	}
	baseDir := buildCloseFixture(t, fixture)

	_, err := Close("SPEC-TEST-SYNC-001", CloseOptions{BaseDir: baseDir})
	if !errors.Is(err, ErrPreconditionMissing) {
		t.Errorf("expected ErrPreconditionMissing, got %v", err)
	}
}

// AC-LSG-014 — Precondition abort atomicity: no staging on precondition failure.
// In M1 this is automatic because the commit phase is not yet implemented.
// The test verifies the result.Result == "failure" + PreconditionsFailed populated
// + no CommitSHA produced.
func TestClose_PreconditionAbortAtomicity(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-ABORT-001",
		specMD: makeSpecMD("SPEC-TEST-ABORT-001", "implemented", "V3R6", "2026-05-25"),
		// no progress.md at all
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close("SPEC-TEST-ABORT-001", CloseOptions{BaseDir: baseDir})
	if !errors.Is(err, ErrPreconditionMissing) {
		t.Errorf("expected ErrPreconditionMissing, got %v", err)
	}
	if result.CommitSHA != "" {
		t.Errorf("CommitSHA should be empty on precondition failure, got %q", result.CommitSHA)
	}
	if result.Result != "failure" {
		t.Errorf("Result = %q, want failure", result.Result)
	}
	// At least 2 preconditions should fail (missing sync + missing mx)
	if len(result.PreconditionsFailed) < 2 {
		t.Errorf("expected ≥ 2 preconditions failed for missing progress.md; got %v",
			result.PreconditionsFailed)
	}
}

// AC-LSG-006 — All preconditions met allows close to proceed (M1 stub: no commit yet).
func TestClose_AllPreconditionsMet(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-OK-001",
		specMD: makeSpecMD("SPEC-TEST-OK-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha: def5678
`,
		acceptanceMD: `# Acceptance
| AC | Status |
|----|--------|
| AC-TEST-001 | **PASS** |
`,
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close("SPEC-TEST-OK-001", CloseOptions{BaseDir: baseDir})
	if err != nil {
		t.Fatalf("Close() error = %v; want nil", err)
	}
	if result.Result != "success" {
		t.Errorf("Result = %q, want success", result.Result)
	}
}

// AC-LSG-022 — Backfill-only mode handles already-completed SPECs as no-op.
// This is the v0.1.2 reframe fixture state `fully-completed-noop`.
func TestClose_BackfillOnly_FullyCompletedNoOp(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-NOOP-001",
		specMD: makeSpecMD("SPEC-TEST-NOOP-001", "completed", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha: def5678
`,
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close("SPEC-TEST-NOOP-001", CloseOptions{
		BaseDir:      baseDir,
		BackfillOnly: true,
	})
	if err != nil {
		t.Fatalf("Close() should succeed as no-op; got error = %v", err)
	}
	if !result.NoOp {
		t.Error("NoOp should be true for already-completed SPEC + backfill-only")
	}
	if result.Result != "noop" {
		t.Errorf("Result = %q, want noop", result.Result)
	}
	if result.Mode != "backfill-only" {
		t.Errorf("Mode = %q, want backfill-only", result.Mode)
	}
	if len(result.Transitions) != 0 {
		t.Errorf("Transitions should be empty for no-op; got %v", result.Transitions)
	}
}

// AC-LSG-022 — Backfill-only on Y_Y_Y_Y_StatusDrift: status implemented → completed.
func TestClose_BackfillOnly_Y4StatusDrift(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-DRIFT-001",
		specMD: makeSpecMD("SPEC-TEST-DRIFT-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha: def5678
`,
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close("SPEC-TEST-DRIFT-001", CloseOptions{
		BaseDir:      baseDir,
		BackfillOnly: true,
	})
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if result.NoOp {
		t.Error("NoOp should be false (status drift requires transition)")
	}
	// Must propose spec.md status transition to completed
	if v, ok := result.Transitions["spec.md:frontmatter.status"]; !ok || v != "completed" {
		t.Errorf("expected spec.md status transition to completed; got %v", result.Transitions)
	}
}

// AC-LSG-022 — DryRun returns ErrDryRun with computed transitions but no staging.
func TestClose_DryRun(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-DRYRUN-001",
		specMD: makeSpecMD("SPEC-TEST-DRYRUN-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha: def5678
`,
		acceptanceMD: "| AC-T-001 | **PASS** |\n",
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close("SPEC-TEST-DRYRUN-001", CloseOptions{
		BaseDir: baseDir,
		DryRun:  true,
	})
	if !errors.Is(err, ErrDryRun) {
		t.Errorf("expected ErrDryRun, got %v", err)
	}
	if result.Result != "success" {
		t.Errorf("DryRun Result = %q, want success", result.Result)
	}
	// Transitions should be populated even in dry-run
	if v, ok := result.Transitions["spec.md:frontmatter.status"]; !ok || v != "completed" {
		t.Errorf("expected dry-run to compute transitions; got %v", result.Transitions)
	}
	if result.CommitSHA != "" {
		t.Errorf("DryRun must NOT produce CommitSHA; got %q", result.CommitSHA)
	}
}

// AC-LSG-006 — PASS-WITH-DEBT presence blocks close unless --force.
func TestClose_PassWithDebtBlocksClose(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-DEBT-001",
		specMD: makeSpecMD("SPEC-TEST-DEBT-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n`,
		acceptanceMD: `| AC-T-001 | **PASS-WITH-DEBT** |
`,
	}
	baseDir := buildCloseFixture(t, fixture)

	_, err := Close("SPEC-TEST-DEBT-001", CloseOptions{BaseDir: baseDir})
	if !errors.Is(err, ErrPreconditionMissing) {
		t.Errorf("expected ErrPreconditionMissing on PASS-WITH-DEBT; got %v", err)
	}

	// Force overrides
	result, err := Close("SPEC-TEST-DEBT-001", CloseOptions{BaseDir: baseDir, Force: true})
	if err != nil {
		t.Errorf("Force should override PASS-WITH-DEBT; got error %v", err)
	}
	if result.Result != "success" {
		t.Errorf("Result with Force = %q, want success", result.Result)
	}
}

// AC-LSG-001 (precondition for full implementation in M3) — empty specID rejected.
func TestClose_EmptySpecIDRejected(t *testing.T) {
	t.Parallel()
	_, err := Close("", CloseOptions{})
	if err == nil {
		t.Fatal("expected error for empty specID, got nil")
	}
}

// AC-LSG-001 — missing SPEC directory rejected with informative error.
func TestClose_MissingSpecDir(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	_, err := Close("SPEC-DOES-NOT-EXIST-001", CloseOptions{BaseDir: tempDir})
	if err == nil {
		t.Fatal("expected error for missing spec dir, got nil")
	}
	if !strings.Contains(err.Error(), "spec directory not found") {
		t.Errorf("error should mention missing spec dir; got %v", err)
	}
}

// AC-LSG-010 — Close acquires + releases lock; subsequent Close on same SPEC succeeds.
func TestClose_LockReleasedOnReturn(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-LOCK-001",
		specMD: makeSpecMD("SPEC-TEST-LOCK-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n`,
	}
	baseDir := buildCloseFixture(t, fixture)

	// First close (succeeds)
	if _, err := Close("SPEC-TEST-LOCK-001", CloseOptions{BaseDir: baseDir}); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	// Second close — lock must have been released, so this should also succeed
	// (not fail with ErrSpecCloseLockHeld).
	if _, err := Close("SPEC-TEST-LOCK-001", CloseOptions{BaseDir: baseDir}); err != nil {
		t.Errorf("second Close() error = %v; lock should have been released", err)
	}
}

// Verify result.AuditedAt populated + DurationMs set.
func TestClose_ResultMetadata(t *testing.T) {
	t.Parallel()

	fixture := closeFixtureSpec{
		id:     "SPEC-TEST-META-001",
		specMD: makeSpecMD("SPEC-TEST-META-001", "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n`,
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close("SPEC-TEST-META-001", CloseOptions{BaseDir: baseDir})
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if result.AuditedAt.IsZero() {
		t.Error("AuditedAt should be populated")
	}
	if result.DurationMs < 0 {
		t.Error("DurationMs should be non-negative")
	}
	if result.Mode != "full-close" {
		t.Errorf("Mode = %q, want full-close", result.Mode)
	}
}
