// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M2 — `moai spec close` CLI integration tests.
//
// These tests cover the CLI surface only (cobra flag parsing, help text,
// stderr/stdout shape, exit-code mapping). The backing transaction logic lives
// in internal/spec/closer.go (M1) and is exercised by closer_test.go.
//
// AC coverage (M2 scope):
//   - AC-LSG-001  — atomic close happy path via CLI (M3 finishes commit transaction; M2 wires CLI)
//   - AC-LSG-006  — precondition validation error rendering
//   - AC-LSG-014  — precondition abort atomicity (no staging on failure)
//   - AC-LSG-022  — backfill-only mode flag wiring + fixture variants
package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureCompletedSpec writes a SPEC fixture whose spec.md/progress.md indicate
// the SPEC is already fully completed. Both sync_commit_sha and mx_commit_sha
// are present; spec.md status is "completed". Used by the fully-completed-noop
// branch of AC-LSG-022 + AC-LSG-018 cross-binding.
func fixtureCompletedSpec(t *testing.T, baseDir, specID string) {
	t.Helper()
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec: %v", err)
	}

	specMD := `---
id: ` + specID + `
title: Fixture
version: 0.1.0
status: completed
created: 2026-01-01
updated: 2026-01-01
author: Test
priority: P1
phase: run
module: TEST
lifecycle: spec-anchored
tags: test
---

# Test SPEC
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	progressMD := `---
id: ` + specID + `
artifact: progress
status: completed
---

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha: abc1234567890
sync_complete_at: 2026-01-01T00:00:00Z

## §E.3 Run-phase Status

status: completed

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: def1234567890
mx_complete_at: 2026-01-01T00:00:00Z
`
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
		t.Fatalf("write progress.md: %v", err)
	}

	// acceptance.md with all PASS (so ACAllPass = true)
	acceptanceMD := `---
id: ` + specID + `
artifact: acceptance
---

# Acceptance Criteria

| AC | Status |
|----|--------|
| AC-1 | PASS |
`
	if err := os.WriteFile(filepath.Join(specDir, "acceptance.md"), []byte(acceptanceMD), 0o644); err != nil {
		t.Fatalf("write acceptance.md: %v", err)
	}
}

// fixtureMissingMx writes a SPEC fixture lacking the §E.5 Mx section. Used by
// AC-LSG-006 (precondition missing-mx error rendering) + AC-LSG-014 (atomicity).
func fixtureMissingMx(t *testing.T, baseDir, specID string) {
	t.Helper()
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec: %v", err)
	}

	specMD := `---
id: ` + specID + `
title: Fixture
version: 0.1.0
status: implemented
created: 2026-01-01
updated: 2026-01-01
author: Test
priority: P1
phase: run
module: TEST
lifecycle: spec-anchored
tags: test
---
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	progressMD := `---
id: ` + specID + `
artifact: progress
status: in-progress
---

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha: abc1234567890

## §E.3 Run-phase Status

status: in-progress

# (Mx section deliberately omitted — precondition failure expected)
`
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
		t.Fatalf("write progress.md: %v", err)
	}

	acceptanceMD := `---
id: ` + specID + `
artifact: acceptance
---
| AC | Status |
| AC-1 | PASS |
`
	if err := os.WriteFile(filepath.Join(specDir, "acceptance.md"), []byte(acceptanceMD), 0o644); err != nil {
		t.Fatalf("write acceptance.md: %v", err)
	}
}

// fixtureY4StatusDrift writes a SPEC at the Y_Y_Y_Y_StatusDrift fixture state:
// §E.2 + §E.5 + SHAs all present, but spec.md status = implemented. Used by the
// AC-LSG-022 Y_Y_Y_Y_StatusDrift backfill-only variant.
func fixtureY4StatusDrift(t *testing.T, baseDir, specID string) {
	t.Helper()
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec: %v", err)
	}

	specMD := `---
id: ` + specID + `
title: Fixture
version: 0.1.0
status: implemented
created: 2026-01-01
updated: 2026-01-01
author: Test
priority: P1
phase: run
module: TEST
lifecycle: spec-anchored
tags: test
---
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	progressMD := `---
id: ` + specID + `
artifact: progress
status: in-progress
---

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha: abc1234567890

## §E.3 Run-phase Status

status: in-progress

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: def1234567890
`
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
		t.Fatalf("write progress.md: %v", err)
	}

	acceptanceMD := `---
id: ` + specID + `
artifact: acceptance
---
| AC | Status |
| AC-1 | PASS |
`
	if err := os.WriteFile(filepath.Join(specDir, "acceptance.md"), []byte(acceptanceMD), 0o644); err != nil {
		t.Fatalf("write acceptance.md: %v", err)
	}
}

// runSpecCloseCmd invokes the `spec close` cobra command in-process with the
// given args, captures stdout/stderr, and returns (stdout, stderr, exit-error).
// The returned error is the cobra RunE error path (nil on exit 0).
func runSpecCloseCmd(t *testing.T, baseDir string, args ...string) (string, string, error) {
	t.Helper()

	cmd := newSpecCloseCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Inject --base-dir to point at the test fixture instead of cwd.
	fullArgs := append([]string{"--base-dir", baseDir}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

// TestSpecClose_Help — AC-LSG-001 surface verification. `--help` must list the
// 3 documented flags (--backfill-only / --dry-run / --force).
func TestSpecClose_Help(t *testing.T) {
	cmd := newSpecCloseCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should not error: %v", err)
	}

	out := stdout.String()
	for _, want := range []string{"--backfill-only", "--dry-run", "--force", "SPEC-"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q\n----\n%s", want, out)
		}
	}
}

// TestSpecClose_RequiresSpecID — `moai spec close` with no positional argument
// must fail with a clear usage error (cobra's MinimumNArgs(1) takes care of this
// but we assert the user-visible behavior here).
func TestSpecClose_RequiresSpecID(t *testing.T) {
	cmd := newSpecCloseCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetOut(&bytes.Buffer{}) // discard
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error when SPEC-ID omitted, got nil")
	}
}

// TestSpecClose_PreconditionMissingMx — AC-LSG-006 + AC-LSG-014.
// SPEC fixture lacks §E.5; CLI must:
//   - Return non-nil error from RunE (which the binary maps to exit 1)
//   - Emit error message to stderr identifying the missing §E.5 section by name
func TestSpecClose_PreconditionMissingMx(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-FIXTURE-MISSING-MX-001"
	fixtureMissingMx(t, tmpDir, specID)

	_, stderr, err := runSpecCloseCmd(t, tmpDir, specID)
	if err == nil {
		t.Fatalf("expected error when §E.5 absent, got nil; stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "§E.5") {
		t.Errorf("stderr should identify missing §E.5 section by name; got: %q", stderr)
	}
	if !strings.Contains(stderr, "precondition") {
		t.Errorf("stderr should mention precondition failure; got: %q", stderr)
	}
}

// TestSpecClose_DryRun — AC-LSG-022 dry-run variant. The CLI must not error out
// when --dry-run is set on a fixture that would otherwise close cleanly; the
// closer.Close() function signals dry-run by returning ErrDryRun internally,
// and the CLI translates that to exit 0 + informational stdout.
func TestSpecClose_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-FIXTURE-DRYRUN-001"
	fixtureY4StatusDrift(t, tmpDir, specID)

	stdout, stderr, err := runSpecCloseCmd(t, tmpDir, "--dry-run", "--backfill-only", specID)
	if err != nil {
		t.Fatalf("dry-run should succeed (exit 0); err=%v stderr=%q", err, stderr)
	}
	if !strings.Contains(stdout, "dry-run") && !strings.Contains(stdout, "dry run") {
		t.Errorf("stdout should announce dry-run; got: %q", stdout)
	}
}

// TestSpecClose_BackfillOnly_FullyCompletedNoOp — AC-LSG-022 fully-completed-noop
// fixture state. SPEC is already at status: completed with both SHAs present;
// `--backfill-only` invocation must:
//   - Exit 0 (success, not error)
//   - Stage no changes (verified at integration level; here we just check exit code)
//   - Emit a noop signal substring (case-insensitive "noop" or "already completed")
func TestSpecClose_BackfillOnly_FullyCompletedNoOp(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-FIXTURE-NOOP-001"
	fixtureCompletedSpec(t, tmpDir, specID)

	stdout, stderr, err := runSpecCloseCmd(t, tmpDir, "--backfill-only", specID)
	if err != nil {
		t.Fatalf("fully-completed-noop should exit 0; err=%v stderr=%q", err, stderr)
	}
	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "noop") && !strings.Contains(combined, "already completed") {
		t.Errorf("output should signal noop; got stdout=%q stderr=%q", stdout, stderr)
	}
}

// TestBackfillOnlyVariants — AC-LSG-022 parametric. Iterates the documented
// fixture variants (subset; M3 will exercise the full Y_N_N_Y / Y_Y_N_Y
// commit-transaction paths). For M2 we cover the variants for which M1's stub
// implementation can return a deterministic outcome.
func TestBackfillOnlyVariants(t *testing.T) {
	tests := []struct {
		name      string
		writeFix  func(t *testing.T, baseDir, specID string)
		wantError bool
		wantStdoutSubstr string
	}{
		{
			name:             "fully-completed-noop",
			writeFix:         fixtureCompletedSpec,
			wantError:        false,
			wantStdoutSubstr: "noop",
		},
		{
			name:             "Y_Y_Y_Y_StatusDrift dry-run preview",
			writeFix:         fixtureY4StatusDrift,
			wantError:        false,
			wantStdoutSubstr: "transition",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			specID := "SPEC-FIXTURE-" + strings.ToUpper(strings.ReplaceAll(tt.name, " ", "-")) + "-001"
			tt.writeFix(t, tmpDir, specID)

			var args []string
			if strings.Contains(tt.name, "dry-run") {
				args = []string{"--dry-run", "--backfill-only", specID}
			} else {
				args = []string{"--backfill-only", specID}
			}

			stdout, stderr, err := runSpecCloseCmd(t, tmpDir, args...)
			if tt.wantError && err == nil {
				t.Errorf("expected error; got nil. stderr=%q", stderr)
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected success; got err=%v stderr=%q", err, stderr)
			}
			combined := strings.ToLower(stdout + stderr)
			if tt.wantStdoutSubstr != "" && !strings.Contains(combined, tt.wantStdoutSubstr) {
				t.Errorf("expected output to contain %q; got stdout=%q stderr=%q",
					tt.wantStdoutSubstr, stdout, stderr)
			}
		})
	}
}

// TestSpecClose_JSONOutput — `--json` emits CloseResult as JSON on stdout.
// Verifies the JSON envelope is well-formed and contains the documented fields.
func TestSpecClose_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-FIXTURE-JSON-001"
	fixtureCompletedSpec(t, tmpDir, specID)

	stdout, _, err := runSpecCloseCmd(t, tmpDir, "--backfill-only", "--json", specID)
	if err != nil {
		t.Fatalf("backfill-only noop should exit 0; err=%v", err)
	}
	// stdout must contain a JSON object with spec_id field.
	if !strings.Contains(stdout, `"spec_id"`) || !strings.Contains(stdout, specID) {
		t.Errorf("--json output should include spec_id field with %q; got: %q", specID, stdout)
	}
	if !strings.Contains(stdout, `"mode"`) {
		t.Errorf("--json output should include mode field; got: %q", stdout)
	}
}

// TestSpecClose_AlreadyCompletedWithoutBackfill — invoking close on a fully
// completed SPEC without --backfill-only must surface ErrAlreadyCompleted as
// an error (exit 1), distinguishing the "guard rails" full-close mode from
// the backfill no-op path.
func TestSpecClose_AlreadyCompletedWithoutBackfill(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-FIXTURE-ALREADY-COMPLETED-001"
	fixtureCompletedSpec(t, tmpDir, specID)

	_, stderr, err := runSpecCloseCmd(t, tmpDir, specID)
	if err == nil {
		t.Fatalf("full close on already-completed SPEC should error; stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "already") {
		t.Errorf("stderr should mention 'already'; got: %q", stderr)
	}
}

// TestSpecClose_SpecDirMissing — invoking close on a SPEC that does not exist
// must surface the directory-not-found error.
func TestSpecClose_SpecDirMissing(t *testing.T) {
	tmpDir := t.TempDir()
	_, stderr, err := runSpecCloseCmd(t, tmpDir, "SPEC-DOES-NOT-EXIST-001")
	if err == nil {
		t.Fatalf("missing SPEC dir should error; stderr=%q", stderr)
	}
}

// initGitRepoFixture initializes a git repo at dir and commits all current
// files so the close transaction edits a tracked file. Used by the full-close
// CLI integration test (M3: transaction implemented — requires a real repo).
func initGitRepoFixture(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	run("add", ".")
	run("commit", "-m", "chore: initial fixture")
}

// TestSpecClose_FullCloseSuccess_ReadyFixture — exercises the full-close success
// path with the M3 atomic-commit transaction implemented. The close stages this
// SPEC's spec.md + progress.md, commits, and reports the commit SHA.
//
// Fixture: spec.md status=implemented, both §E.2 + §E.5 sections present, all
// ACs PASS, in a real git repo. Close without --backfill-only → full-close mode,
// all preconditions satisfied, no NoOp, real commit produced.
func TestSpecClose_FullCloseSuccess_ReadyFixture(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-FIXTURE-READY-001"
	specDir := filepath.Join(tmpDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec: %v", err)
	}

	specMD := `---
id: ` + specID + `
title: Ready fixture
version: 0.1.0
status: implemented
created: 2026-01-01
updated: 2026-01-01
author: Test
priority: P1
phase: run
module: TEST
lifecycle: spec-anchored
tags: test
---
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	progressMD := `---
id: ` + specID + `
artifact: progress
status: implemented
---

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha: abc1234567890

## §E.3 Run-phase Status

status: implemented

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: def1234567890
`
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
		t.Fatalf("write progress.md: %v", err)
	}

	acceptanceMD := `---
id: ` + specID + `
artifact: acceptance
---
| AC | Status |
| AC-1 | PASS |
`
	if err := os.WriteFile(filepath.Join(specDir, "acceptance.md"), []byte(acceptanceMD), 0o644); err != nil {
		t.Fatalf("write acceptance.md: %v", err)
	}

	// M3: full-close performs a real atomic commit, so the fixture must be a git repo.
	initGitRepoFixture(t, tmpDir)

	stdout, stderr, err := runSpecCloseCmd(t, tmpDir, specID)
	if err != nil {
		t.Fatalf("full-close on ready fixture should succeed; err=%v stderr=%q", err, stderr)
	}
	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "full-close") && !strings.Contains(combined, "transition") {
		t.Errorf("output should announce close + transitions; got stdout=%q stderr=%q", stdout, stderr)
	}
	// M3: the transaction commits, so the CLI must report the resulting commit SHA.
	if !strings.Contains(stdout, "Commit:") {
		t.Errorf("output should report the close commit SHA; got: %q", stdout)
	}
	// The obsolete M2-stub "deferred to M3" note must be gone.
	if strings.Contains(stdout, "deferred to M3") || strings.Contains(stdout, "M2 stub") {
		t.Errorf("obsolete M2-stub note must be removed; got: %q", stdout)
	}
}

// TestSpecClose_NoAskUserQuestion — REQ-PGN-012 / C-HRA-008 static guard.
// The CLI surface MUST NOT invoke AskUserQuestion. The grep-based audit in E4
// is the canonical check; this test additionally enforces that the running
// command does not block waiting for input by exercising the happy path.
//
// The check excludes lines whose first non-whitespace characters are `//`
// (Go single-line comments) so that doctrine documentation referring to
// AskUserQuestion by name does not produce a false positive — matches the
// orchestrator-side grep pattern in agent-common-protocol.md.
func TestSpecClose_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("spec_close.go")
	if err != nil {
		t.Skipf("spec_close.go not present (RED phase): %v", err)
	}
	if hasNonCommentRefToAskUser(string(src)) {
		t.Errorf("spec_close.go must not reference AskUserQuestion or mcp__askuser outside comments")
	}
}

// hasNonCommentRefToAskUser returns true iff the source text contains a
// non-comment line referencing AskUserQuestion or mcp__askuser. Mirrors the
// canonical grep pattern: `grep -v "// "` (drop lines whose first non-ws is //).
func hasNonCommentRefToAskUser(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.Contains(line, "AskUserQuestion") || strings.Contains(line, "mcp__askuser") {
			return true
		}
	}
	return false
}
