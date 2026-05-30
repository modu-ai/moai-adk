// closer_test.go — TDD coverage for atomic close orchestrator.
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 AC-LSG-001, AC-LSG-006, AC-LSG-014, AC-LSG-018, AC-LSG-020, AC-LSG-022.
package spec

import (
	"bufio"
	"encoding/json"
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

// ----------------------------------------------------------------------------
// M1/M2 remediation — Defect 1 (no-op detection relaxation) RED tests.
//
// The 5 already-discharged target SPECs were retroactively closed to
// `status: completed` by orchestrator-direct Mx chores that left progress.md
// §E.5 mx_commit_sha in non-extractable forms. Under --backfill-only these
// MUST be no-op (exit 0, 0 commits) per AC-LSG-018. The pre-remediation triple-
// AND gate (status == completed && SyncCommitSHA != "" && MxCommitSHA != "")
// fails for all 5 production states, so they fall through to the precondition
// matrix or compute-transitions instead of the no-op branch.
//
// AC-LSG-022 truth-table guard: the relaxed predicate keys ONLY on
// `spec.md status == "completed"`, which is the sole field distinguishing the
// no-op fixtures from the transition fixtures (Y_N_N_Y / Y_Y_N_Y /
// Y_Y_Y_Y_StatusDrift all carry status: implemented).
// ----------------------------------------------------------------------------

// completedProgressVariant enumerates the production §E.5 mx_commit_sha states
// observed across the 5 already-discharged target SPECs. All 5 SPECs are at
// spec.md status: completed; only the mx_commit_sha serialization differs.
type completedProgressVariant struct {
	name       string
	progressMD string
}

var completedNoOpVariants = []completedProgressVariant{
	{
		// ARR-001 / FCG-001 / TMD-001 — §E.5 section present, mx_commit_sha
		// field entirely ABSENT (orchestrator Mx chore did not write it).
		name: "mx-field-absent",
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: 11abb9a30

## §E.5 Mx-phase Audit-Ready Signal
mx_complete_at: 2026-05-25T00:00:00Z
`,
	},
	{
		// HCW-001 — mx_commit_sha present but literal value `null`.
		name: "mx-literal-null",
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: 2d9871208

## §E.5 Mx-phase Audit-Ready Signal
mx_commit_sha: null
`,
	},
	{
		// TMC-001 — mx_commit_sha present but `(this commit)` placeholder
		// (markdown-list style).
		name: "mx-this-commit-placeholder",
		progressMD: `## §E.2 Sync-phase
- sync_commit_sha: 1f42eecb1

## §E.5 Mx-phase Audit-Ready Signal
- mx_commit_sha: (this commit)
`,
	},
	{
		// Empty-value variant (`mx_commit_sha:` with nothing after).
		name: "mx-empty-value",
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: deadbeef0

## §E.5 Mx-phase Audit-Ready Signal
mx_commit_sha:
`,
	},
}

// TestClose_BackfillOnly_CompletedProductionVariantsAreNoOp — Defect 1 RED.
// Each of the 5 production §E.5 states (status: completed + non-extractable
// mx_commit_sha) MUST return no-op success under --backfill-only.
//
// Pre-remediation behavior: only the literal both-SHA-present fixture hits the
// no-op branch; these 4 variants fall through and either fail preconditions or
// compute transitions (NoOp=false). This test FAILS before the fix.
func TestClose_BackfillOnly_CompletedProductionVariantsAreNoOp(t *testing.T) {
	t.Parallel()

	for _, v := range completedNoOpVariants {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			specID := "SPEC-TEST-COMPLETED-" + strings.ToUpper(v.name) + "-001"
			fixture := closeFixtureSpec{
				id:         specID,
				specMD:     makeSpecMD(specID, "completed", "V3R6", "2026-05-25"),
				progressMD: v.progressMD,
			}
			baseDir := buildCloseFixture(t, fixture)

			result, err := Close(specID, CloseOptions{
				BaseDir:      baseDir,
				BackfillOnly: true,
			})
			if err != nil {
				t.Fatalf("completed SPEC + backfill-only should be no-op success; got err = %v", err)
			}
			if !result.NoOp {
				t.Errorf("NoOp should be true for already-completed SPEC; got result.Result=%q transitions=%v",
					result.Result, result.Transitions)
			}
			if result.Result != "noop" {
				t.Errorf("Result = %q, want noop", result.Result)
			}
			if len(result.Transitions) != 0 {
				t.Errorf("no-op must produce 0 transitions (0 commits); got %v", result.Transitions)
			}
			if result.CommitSHA != "" {
				t.Errorf("no-op must NOT produce a commit; got CommitSHA=%q", result.CommitSHA)
			}
		})
	}
}

// TestClose_BackfillOnly_StatusDriftNotNoOp — Defect 1 truth-table guard.
// A SPEC at status: implemented with BOTH SHAs present (Y_Y_Y_Y_StatusDrift)
// MUST NOT be treated as no-op by the relaxed predicate — it requires a status
// transition. This guards against over-relaxation that keys on SHA presence.
func TestClose_BackfillOnly_StatusDriftNotNoOp(t *testing.T) {
	t.Parallel()

	specID := "SPEC-TEST-DRIFT-GUARD-001"
	fixture := closeFixtureSpec{
		id:     specID,
		specMD: makeSpecMD(specID, "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha: def5678
`,
	}
	baseDir := buildCloseFixture(t, fixture)

	result, err := Close(specID, CloseOptions{BaseDir: baseDir, BackfillOnly: true})
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if result.NoOp {
		t.Error("status: implemented MUST NOT be no-op (status drift requires transition)")
	}
	if v, ok := result.Transitions["spec.md:frontmatter.status"]; !ok || v != "completed" {
		t.Errorf("expected status transition to completed; got %v", result.Transitions)
	}
}

// ----------------------------------------------------------------------------
// M1/M2 remediation — Defect 2 (lifecycle-close.log emission) RED tests.
//
// NFR-LSG-004 / AC-LSG-020: every Close() invocation MUST append one JSON line
// to .moai/logs/lifecycle-close.log with the 7-field schema. The log `result`
// field is `success` for a no-op (no-op IS a success); AC-LSG-018's jq filter
// expects `result == "success" && transitions == {}` for the 5 no-op closes.
// ----------------------------------------------------------------------------

// lifecycleLogEntry mirrors the on-disk NFR-LSG-004 log schema for assertions.
type lifecycleLogEntry struct {
	Timestamp  string            `json:"timestamp"`
	SpecID     string            `json:"spec_id"`
	Mode       string            `json:"mode"`
	Transitions map[string]string `json:"transitions"`
	CommitSHA  string            `json:"commit_sha"`
	Result     string            `json:"result"`
	DurationMs int64             `json:"duration_ms"`
}

func readLifecycleLog(t *testing.T, path string) []lifecycleLogEntry {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("lifecycle-close.log not written at %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	var entries []lifecycleLogEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e lifecycleLogEntry
		if jErr := json.Unmarshal([]byte(line), &e); jErr != nil {
			t.Fatalf("log line is not valid JSON: %q (%v)", line, jErr)
		}
		entries = append(entries, e)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan log: %v", err)
	}
	return entries
}

// TestClose_LogEmission_NoOpEntry — Defect 2 RED. A no-op close MUST write a
// single log line with mode=backfill-only, result=success (NOT "noop"),
// transitions={} (empty object), and the 7 schema fields populated.
func TestClose_LogEmission_NoOpEntry(t *testing.T) {
	t.Parallel()

	specID := "SPEC-TEST-LOG-NOOP-001"
	fixture := closeFixtureSpec{
		id:     specID,
		specMD: makeSpecMD(specID, "completed", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha:
`,
	}
	baseDir := buildCloseFixture(t, fixture)
	logPath := filepath.Join(t.TempDir(), "lifecycle-close.log")

	result, err := Close(specID, CloseOptions{
		BaseDir:      baseDir,
		BackfillOnly: true,
		LogPath:      logPath,
	})
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !result.NoOp {
		t.Fatalf("precondition: expected NoOp result")
	}

	entries := readLifecycleLog(t, logPath)
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log entry, got %d", len(entries))
	}
	e := entries[0]
	if e.SpecID != specID {
		t.Errorf("log spec_id = %q, want %q", e.SpecID, specID)
	}
	if e.Mode != "backfill-only" {
		t.Errorf("log mode = %q, want backfill-only", e.Mode)
	}
	// Reconciliation point: in-memory Result is "noop" but log result is "success".
	if e.Result != "success" {
		t.Errorf("log result = %q, want success (no-op IS a success per AC-LSG-020)", e.Result)
	}
	if len(e.Transitions) != 0 {
		t.Errorf("log transitions must be empty object for no-op; got %v", e.Transitions)
	}
	if e.Timestamp == "" {
		t.Error("log timestamp must be populated (RFC3339)")
	}
	if e.DurationMs < 0 {
		t.Errorf("log duration_ms must be non-negative; got %d", e.DurationMs)
	}
}

// TestClose_LogEmission_FailureEntry — Defect 2: failure paths also log, with
// result=failure. Verifies the log fires on every invocation including the
// precondition-failure path.
func TestClose_LogEmission_FailureEntry(t *testing.T) {
	t.Parallel()

	specID := "SPEC-TEST-LOG-FAIL-001"
	fixture := closeFixtureSpec{
		id:     specID,
		specMD: makeSpecMD(specID, "implemented", "V3R6", "2026-05-25"),
		// no progress.md → missing both sync + mx sections → precondition failure
	}
	baseDir := buildCloseFixture(t, fixture)
	logPath := filepath.Join(t.TempDir(), "lifecycle-close.log")

	_, err := Close(specID, CloseOptions{BaseDir: baseDir, LogPath: logPath})
	if !errors.Is(err, ErrPreconditionMissing) {
		t.Fatalf("expected ErrPreconditionMissing, got %v", err)
	}

	entries := readLifecycleLog(t, logPath)
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log entry on failure path, got %d", len(entries))
	}
	if entries[0].Result != "failure" {
		t.Errorf("log result = %q, want failure", entries[0].Result)
	}
}

// TestClose_LogEmission_FullCloseSuccessEntry — Defect 2: full-close success
// path logs result=success with mode=full-close.
func TestClose_LogEmission_FullCloseSuccessEntry(t *testing.T) {
	t.Parallel()

	specID := "SPEC-TEST-LOG-FULL-001"
	fixture := closeFixtureSpec{
		id:     specID,
		specMD: makeSpecMD(specID, "implemented", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha: def5678
`,
	}
	baseDir := buildCloseFixture(t, fixture)
	logPath := filepath.Join(t.TempDir(), "lifecycle-close.log")

	result, err := Close(specID, CloseOptions{BaseDir: baseDir, LogPath: logPath})
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if result.Result != "success" {
		t.Fatalf("precondition: expected success result, got %q", result.Result)
	}

	entries := readLifecycleLog(t, logPath)
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	if entries[0].Mode != "full-close" {
		t.Errorf("log mode = %q, want full-close", entries[0].Mode)
	}
	if entries[0].Result != "success" {
		t.Errorf("log result = %q, want success", entries[0].Result)
	}
}

// TestClose_LogEmission_Appends — Defect 2: multiple Close() invocations
// accumulate log entries (append, not truncate). Mirrors the M6 dogfood where
// 5 sequential closes accumulate ≥5 lines in the real log.
func TestClose_LogEmission_Appends(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "lifecycle-close.log")

	for i := 0; i < 5; i++ {
		specID := "SPEC-TEST-LOG-APPEND-" + strings.ToUpper(string(rune('A'+i))) + "-001"
		fixture := closeFixtureSpec{
			id:     specID,
			specMD: makeSpecMD(specID, "completed", "V3R6", "2026-05-25"),
			progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha:
`,
		}
		baseDir := buildCloseFixture(t, fixture)
		if _, err := Close(specID, CloseOptions{
			BaseDir:      baseDir,
			BackfillOnly: true,
			LogPath:      logPath,
		}); err != nil {
			t.Fatalf("close %d error = %v", i, err)
		}
	}

	entries := readLifecycleLog(t, logPath)
	if len(entries) != 5 {
		t.Fatalf("expected 5 accumulated log entries, got %d", len(entries))
	}
	// Replicate AC-LSG-018 jq filter: mode==backfill-only && result==success && transitions=={}
	noopSuccessCount := 0
	for _, e := range entries {
		if e.Mode == "backfill-only" && e.Result == "success" && len(e.Transitions) == 0 {
			noopSuccessCount++
		}
	}
	if noopSuccessCount < 5 {
		t.Errorf("AC-LSG-018 jq filter equivalent: want ≥5 noop-success entries, got %d", noopSuccessCount)
	}
}

// TestClose_LogEmission_ParentDirCreated — Defect 2: the log path's parent
// directory is created if absent (production path .moai/logs/ may not exist).
func TestClose_LogEmission_ParentDirCreated(t *testing.T) {
	t.Parallel()

	specID := "SPEC-TEST-LOG-MKDIR-001"
	fixture := closeFixtureSpec{
		id:     specID,
		specMD: makeSpecMD(specID, "completed", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha:
`,
	}
	baseDir := buildCloseFixture(t, fixture)
	// Nested non-existent parent dir.
	logPath := filepath.Join(t.TempDir(), "nested", "logs", "lifecycle-close.log")

	if _, err := Close(specID, CloseOptions{
		BaseDir:      baseDir,
		BackfillOnly: true,
		LogPath:      logPath,
	}); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("log file should be created with parent dir; stat err = %v", err)
	}
}

// TestClose_LogEmission_DefaultPath — Defect 2: when LogPath is empty, the log
// defaults to <baseDir>/.moai/logs/lifecycle-close.log.
func TestClose_LogEmission_DefaultPath(t *testing.T) {
	t.Parallel()

	specID := "SPEC-TEST-LOG-DEFAULT-001"
	fixture := closeFixtureSpec{
		id:     specID,
		specMD: makeSpecMD(specID, "completed", "V3R6", "2026-05-25"),
		progressMD: `## §E.2 Sync-phase
sync_commit_sha: abc1234

## §E.5 Mx-phase
mx_commit_sha:
`,
	}
	baseDir := buildCloseFixture(t, fixture)

	if _, err := Close(specID, CloseOptions{
		BaseDir:      baseDir,
		BackfillOnly: true,
		// LogPath omitted → default location
	}); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	defaultPath := filepath.Join(baseDir, ".moai", "logs", "lifecycle-close.log")
	if _, err := os.Stat(defaultPath); err != nil {
		t.Errorf("log should default to .moai/logs/lifecycle-close.log; stat err = %v", err)
	}
}
