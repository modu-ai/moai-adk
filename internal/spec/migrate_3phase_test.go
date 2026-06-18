// migrate_3phase_test.go — SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-007 / M3 coverage.
package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateProgressMD_FoldsE5IntoE4(t *testing.T) {
	t.Parallel()
	input := `# Progress

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: "abc123"

---

## §E.5 Mx-phase Audit-Ready Signal

` + "```yaml" + `
mx_commit_sha: "def456"
mx_status: ready
` + "```" + `

---

Version: 0.1.0
`
	out, mxSHA, folded := MigrateProgressMD(input)
	if !folded {
		t.Fatal("folded=false; expected §E.5 to be folded into §E.4")
	}
	if mxSHA != "def456" {
		t.Errorf("mx_commit_sha = %q, want %q", mxSHA, "def456")
	}
	if strings.Contains(out, "## §E.5") {
		t.Error("§E.5 heading still present after migration; should be removed")
	}
	if !strings.Contains(out, "### (Migrated from §E.5)") {
		t.Error("migrated sub-heading not inserted under §E.4")
	}
	if !strings.Contains(out, "mx_commit_sha: \"def456\"") {
		t.Error("§E.5 content (mx_commit_sha) not folded into §E.4")
	}
	// §E.4 must still carry its original sync_commit_sha.
	if !strings.Contains(out, "sync_commit_sha: \"abc123\"") {
		t.Error("§E.4 original content lost during fold")
	}
}

func TestMigrateProgressMD_Idempotent(t *testing.T) {
	t.Parallel()
	// No §E.5 section → no-op.
	input := `## §E.4 Sync-phase Audit-Ready Signal
sync_commit_sha: "abc"
`
	out, mxSHA, folded := MigrateProgressMD(input)
	if folded {
		t.Error("folded=true for input without §E.5; expected no-op")
	}
	if mxSHA != "" {
		t.Errorf("mx_commit_sha = %q, want empty", mxSHA)
	}
	if out != input {
		t.Errorf("content changed on no-op migration")
	}
}

func TestRunMigration_SkipsGrandfathered(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	specsDir := filepath.Join(baseDir, ".moai", "specs")

	// V3R6 modern SPEC with legacy §E.5 → should be migrated.
	mustWrite := func(path, content string) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(filepath.Join(specsDir, "SPEC-V3R6-LEGACY-001", "spec.md"),
		"---\nid: SPEC-V3R6-LEGACY-001\ntitle: \"t\"\nversion: \"0.1.0\"\nstatus: completed\ncreated: 2026-05-01\nupdated: 2026-05-02\nauthor: t\npriority: P1\nphase: \"v3.0.0\"\nmodule: \"m\"\nlifecycle: spec-anchored\ntags: \"t\"\nera: \"V3R6\"\n---\n# Body\n")
	mustWrite(filepath.Join(specsDir, "SPEC-V3R6-LEGACY-001", "progress.md"),
		"## §E.2 Run\n## §E.4 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n")

	// Grandfather-protected V2.x SPEC (no progress.md) → must be SKIPPED (N4).
	mustWrite(filepath.Join(specsDir, "SPEC-V2X-001", "spec.md"),
		"---\nid: SPEC-V2X-001\ntitle: \"t\"\nversion: \"0.1.0\"\nstatus: completed\ncreated: 2026-01-01\nupdated: 2026-01-02\nauthor: t\npriority: P1\nphase: \"v2.0\"\nmodule: \"m\"\nlifecycle: spec-anchored\ntags: \"t\"\nera: \"V2.x\"\n---\n# Body\n")

	logPath := filepath.Join(baseDir, ".moai", "state", "lifecycle-redesign-migration.json")
	log, err := RunMigration(baseDir, logPath, false)
	if err != nil {
		t.Fatalf("RunMigration error: %v", err)
	}
	if len(log.Entries) != 1 {
		t.Fatalf("expected 1 migrated entry, got %d", len(log.Entries))
	}
	if log.Entries[0].SpecID != "SPEC-V3R6-LEGACY-001" {
		t.Errorf("migrated entry SpecID = %q", log.Entries[0].SpecID)
	}
	if log.Entries[0].MxCommitSha != "def" {
		t.Errorf("MxCommitSha = %q, want %q", log.Entries[0].MxCommitSha, "def")
	}
	if log.Skipped != 1 {
		t.Errorf("skipped grandfathered = %d, want 1", log.Skipped)
	}
	// Verify the migration log file was written.
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("migration log not written: %v", err)
	}
	// Verify §E.5 was actually removed from the migrated progress.md.
	migrated, _ := os.ReadFile(filepath.Join(specsDir, "SPEC-V3R6-LEGACY-001", "progress.md"))
	if strings.Contains(string(migrated), "## §E.5") {
		t.Error("§E.5 heading still present after migration on disk")
	}
}

func TestRunMigration_DryRun(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	specsDir := filepath.Join(baseDir, ".moai", "specs")
	progressPath := filepath.Join(specsDir, "SPEC-V3R6-DRY-001", "progress.md")
	os.MkdirAll(filepath.Dir(progressPath), 0755)
	os.WriteFile(filepath.Join(specsDir, "SPEC-V3R6-DRY-001", "spec.md"),
		[]byte("---\nid: SPEC-V3R6-DRY-001\ntitle: \"t\"\nversion: \"0.1.0\"\nstatus: completed\ncreated: 2026-05-01\nupdated: 2026-05-02\nauthor: t\npriority: P1\nphase: \"v3.0.0\"\nmodule: \"m\"\nlifecycle: spec-anchored\ntags: \"t\"\nera: \"V3R6\"\n---\n# Body\n"), 0644)
	original := "## §E.2 Run\n## §E.4 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n"
	os.WriteFile(progressPath, []byte(original), 0644)

	log, err := RunMigration(baseDir, "", true) // dryRun, no log path
	if err != nil {
		t.Fatalf("RunMigration dryRun error: %v", err)
	}
	if len(log.Entries) != 1 {
		t.Fatalf("dryRun should still populate 1 entry, got %d", len(log.Entries))
	}
	// On-disk progress.md must be UNCHANGED (dry run).
	onDisk, _ := os.ReadFile(progressPath)
	if string(onDisk) != original {
		t.Error("dryRun mutated progress.md on disk; must be a no-op")
	}
}
