// SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 / M3
// TestArchiveForce covers the four force-path scenarios introduced by
// REQ-UAC-001 / REQ-UAC-003 / REQ-UAC-006 / REQ-UAC-007:
//
//   - force=true + drift  → existing archive moved to v2.16-drift-<ts>/<id>/
//     and re-archived from the live source.
//   - force=true + no drift → idempotent; no drift-backup directory created.
//   - force=false + drift → preserves the BC-V3R3-007 ARCHIVE_DRIFT error.
//   - force=true + backup parent denied → original archive is preserved and
//     the call returns a wrapped error (drift overwrite is best-effort).
//
// All subtests use t.TempDir() per CLAUDE.local.md §6 (test isolation rule);
// the dev project's .claude/ and .moai/archive/ are never touched.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestArchiveForce groups the four force-path subtests.
func TestArchiveForce(t *testing.T) {
	t.Parallel()

	t.Run("force_with_drift_creates_backup_and_overwrites", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		// Pick a single legacy skill ID for the drift scenario.
		id := legacySkillIDs[0]

		// Live source content (post-drift version).
		srcContent := "# live source after edit\nupdated content\n"
		makeSkillDir(t, root, id, srcContent)

		// Pre-populate the archive with old content (drift baseline).
		archiveDir := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, id)
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			t.Fatalf("seed archive dir: %v", err)
		}
		oldContent := "# old archive content (pre-drift)\n"
		if err := os.WriteFile(filepath.Join(archiveDir, "SKILL.md"), []byte(oldContent), 0o644); err != nil {
			t.Fatalf("seed archive file: %v", err)
		}

		// force=true should detect drift, back up the archive, and re-archive.
		var out bytes.Buffer
		_, err := archiveLegacySkills(root, &out, true)
		if err != nil {
			t.Fatalf("archiveLegacySkills force=true: %v", err)
		}

		// Verify exactly one drift-backup directory exists matching the glob.
		// Glob pattern resolves R3 timestamp-flakiness risk per spec.md §E.1.
		matches, err := filepath.Glob(filepath.Join(root, ".moai", "archive", "skills",
			archiveVersion+"-drift-*"))
		if err != nil {
			t.Fatalf("glob drift backup: %v", err)
		}
		if len(matches) != 1 {
			t.Fatalf("expected 1 drift backup directory, got %d: %v", len(matches), matches)
		}

		// Backup directory must contain the pre-drift archive content.
		backupSkillFile := filepath.Join(matches[0], id, "SKILL.md")
		backupBytes, err := os.ReadFile(backupSkillFile)
		if err != nil {
			t.Fatalf("read backup file: %v", err)
		}
		if string(backupBytes) != oldContent {
			t.Errorf("backup content mismatch: got %q, want %q", string(backupBytes), oldContent)
		}

		// The live archive must now match the live source.
		liveArchiveFile := filepath.Join(archiveDir, "SKILL.md")
		liveBytes, err := os.ReadFile(liveArchiveFile)
		if err != nil {
			t.Fatalf("read live archive: %v", err)
		}
		if string(liveBytes) != srcContent {
			t.Errorf("live archive content mismatch: got %q, want %q", string(liveBytes), srcContent)
		}

		// Output should mention the drift backup line for traceability.
		if !strings.Contains(out.String(), "archive drift backup: "+id) {
			t.Errorf("expected drift backup line for %s in output, got:\n%s", id, out.String())
		}
	})

	t.Run("force_without_drift_is_idempotent", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		id := legacySkillIDs[1]
		content := "# clean content\n"
		makeSkillDir(t, root, id, content)

		// Pre-populate the archive with matching content (no drift).
		archiveDir := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, id)
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			t.Fatalf("seed archive dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(archiveDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("seed archive file: %v", err)
		}

		var out bytes.Buffer
		_, err := archiveLegacySkills(root, &out, true)
		if err != nil {
			t.Fatalf("archiveLegacySkills force=true (no drift): %v", err)
		}

		// No drift-backup directory should be created when content matches.
		matches, err := filepath.Glob(filepath.Join(root, ".moai", "archive", "skills",
			archiveVersion+"-drift-*"))
		if err != nil {
			t.Fatalf("glob drift backup: %v", err)
		}
		if len(matches) != 0 {
			t.Errorf("expected no drift backup directories, got %d: %v", len(matches), matches)
		}
	})

	t.Run("force_false_preserves_drift_error", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		id := legacySkillIDs[2]
		makeSkillDir(t, root, id, "# live content\n")

		// Seed mismatched archive content (drift).
		archiveDir := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, id)
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			t.Fatalf("seed archive dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(archiveDir, "SKILL.md"), []byte("# stale\n"), 0o644); err != nil {
			t.Fatalf("seed archive file: %v", err)
		}

		var out bytes.Buffer
		_, err := archiveLegacySkills(root, &out, false)
		if err == nil {
			t.Fatalf("force=false with drift: expected ARCHIVE_DRIFT error, got nil")
		}
		// BC-V3R3-007 contract: error must be ARCHIVE_DRIFT typed.
		if !strings.Contains(err.Error(), "ARCHIVE_DRIFT") &&
			!strings.Contains(err.Error(), "archive already exists but content differs") {
			t.Errorf("force=false drift error should reference ARCHIVE_DRIFT, got: %v", err)
		}

		// No drift-backup directory should be created on force=false.
		matches, _ := filepath.Glob(filepath.Join(root, ".moai", "archive", "skills",
			archiveVersion+"-drift-*"))
		if len(matches) != 0 {
			t.Errorf("force=false must not create backup directories, got: %v", matches)
		}
	})

	t.Run("force_with_drift_backup_failure_preserves_original", func(t *testing.T) {
		// On Windows, chmod-based permission denial is unreliable. The
		// underlying invariant — "if backup creation fails, the original
		// archive is preserved and the call returns an error" — is identical
		// across platforms, so we skip on Windows rather than introduce a
		// platform-specific failure injection.
		if runtime.GOOS == "windows" {
			t.Skip("chmod-based permission denial is unreliable on Windows")
		}
		t.Parallel()
		root := t.TempDir()
		id := legacySkillIDs[3]
		srcContent := "# live\n"
		makeSkillDir(t, root, id, srcContent)

		// Seed mismatched archive content.
		archiveDir := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, id)
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			t.Fatalf("seed archive dir: %v", err)
		}
		stale := "# stale archive\n"
		if err := os.WriteFile(filepath.Join(archiveDir, "SKILL.md"), []byte(stale), 0o644); err != nil {
			t.Fatalf("seed archive file: %v", err)
		}

		// Make the .moai/archive/skills parent read-only so the drift-backup
		// parent directory cannot be created. This forces the backup branch
		// to fail and verifies the original archive is preserved.
		archiveSkillsParent := filepath.Join(root, ".moai", "archive", "skills")
		if err := os.Chmod(archiveSkillsParent, 0o555); err != nil {
			t.Fatalf("chmod parent read-only: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(archiveSkillsParent, 0o755)
		})

		var out bytes.Buffer
		_, err := archiveLegacySkills(root, &out, true)
		if err == nil {
			t.Fatalf("expected backup failure error, got nil")
		}
		if !strings.Contains(err.Error(), "drift backup") {
			t.Errorf("expected 'drift backup' in error, got: %v", err)
		}

		// Restore write permission so we can inspect the archive.
		if err := os.Chmod(archiveSkillsParent, 0o755); err != nil {
			t.Fatalf("restore chmod: %v", err)
		}
		// Original archive must still hold the stale content (preserved).
		liveBytes, err := os.ReadFile(filepath.Join(archiveDir, "SKILL.md"))
		if err != nil {
			t.Fatalf("read live archive: %v", err)
		}
		if string(liveBytes) != stale {
			t.Errorf("original archive should be preserved on backup failure: got %q, want %q",
				string(liveBytes), stale)
		}
	})
}
