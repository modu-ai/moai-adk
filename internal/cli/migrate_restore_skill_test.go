// SPEC-V3R3-HARNESS-001 / T-M4-03
// Tests for the moai migrate restore-skill subcommand.
// Round-trip: verify byte-identical results after archiveSkill → restoreSkill.

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRestoreSkill_RoundTrip verifies that the original file and the restored
// file are byte-identical after an archive → restore round-trip.
func TestRestoreSkill_RoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	skillID := "moai-domain-backend"
	content := "# moai-domain-backend\noriginal content line 1\nline 2"
	makeSkillDir(t, root, skillID, content)

	// Record the original hashes
	srcDir := filepath.Join(root, ".claude", "skills", skillID)
	origHashes := hashDir(t, srcDir)

	// Step 1: archive
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("archiveSkill: %v", err)
	}

	// Remove the source directory (simulate the clean state before restore)
	if err := os.RemoveAll(srcDir); err != nil {
		t.Fatalf("RemoveAll source: %v", err)
	}

	// Step 2: restore
	if err := restoreSkill(root, skillID, false); err != nil {
		t.Fatalf("restoreSkill: %v", err)
	}

	// Compare the restored hashes against the original hashes
	restoredHashes := hashDir(t, srcDir)
	if len(origHashes) != len(restoredHashes) {
		t.Errorf("file count: original=%d restored=%d", len(origHashes), len(restoredHashes))
	}
	for rel, origHash := range origHashes {
		restoredHash, ok := restoredHashes[rel]
		if !ok {
			t.Errorf("file missing after restore: %s", rel)
			continue
		}
		if origHash != restoredHash {
			t.Errorf("content mismatch for %s: original=%s restored=%s",
				rel, origHash, restoredHash)
		}
	}
}

// TestRestoreSkill_ArchiveMissing verifies that an error is returned when the archive is missing.
func TestRestoreSkill_ArchiveMissing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	err := restoreSkill(root, "moai-domain-frontend", false)
	if err == nil {
		t.Fatal("expected error when archive is missing, got nil")
	}
}

// TestRestoreSkill_TargetExists verifies that an error is returned without
// --force when the restore target already exists.
func TestRestoreSkill_TargetExists(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-database"

	makeSkillDir(t, root, skillID, "original")
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("archiveSkill: %v", err)
	}
	// The source still exists

	// Attempt restore without --force → expect an error
	err := restoreSkill(root, skillID, false)
	if err == nil {
		t.Fatal("expected error when target exists without --force, got nil")
	}
}

// TestRestoreSkill_TargetExistsWithForce verifies that an existing target is
// overwritten when --force is supplied.
func TestRestoreSkill_TargetExistsWithForce(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skillID := "moai-domain-db-docs"

	makeSkillDir(t, root, skillID, "original content")
	if err := archiveSkill(root, skillID); err != nil {
		t.Fatalf("archiveSkill: %v", err)
	}

	// Modify the source content (now diverges from the archive)
	skillFile := filepath.Join(root, ".claude", "skills", skillID, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("write modified: %v", err)
	}

	// Restore with --force=true → expect success
	if err := restoreSkill(root, skillID, true); err != nil {
		t.Fatalf("restoreSkill with force: %v", err)
	}

	// After restore, the content must match the archived (original) content
	restored, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(restored) != "original content" {
		t.Errorf("restored content = %q, want %q", string(restored), "original content")
	}
}

// TestRestoreSkill_EmptySkillID verifies that calling restoreSkill with an
// empty skillID returns the RESTORE_INVALID_SKILL_ID error.
func TestRestoreSkill_EmptySkillID(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	err := restoreSkill(root, "", false)
	if err == nil {
		t.Fatal("expected error for empty skillID, got nil")
	}
	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T: %v", err, err)
	}
	if me.Code != "RESTORE_INVALID_SKILL_ID" {
		t.Errorf("code = %q, want RESTORE_INVALID_SKILL_ID", me.Code)
	}
}

// TestRestoreSkill_InvalidPathTraversal verifies that the
// RESTORE_INVALID_SKILL_ID error is returned when skillID contains "..".
func TestRestoreSkill_InvalidPathTraversal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	err := restoreSkill(root, "../evil", false)
	if err == nil {
		t.Fatal("expected error for path traversal skillID, got nil")
	}
}

// TestRestoreSkill_All16RoundTrip verifies the archive → restore round-trip for
// all 16 skills.
func TestRestoreSkill_All16RoundTrip(t *testing.T) {
	t.Parallel()
	for _, skillID := range legacySkillIDs {
		skillID := skillID
		t.Run(skillID, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			content := "# " + skillID + " test content\nsome data"
			makeSkillDir(t, root, skillID, content)

			srcDir := filepath.Join(root, ".claude", "skills", skillID)
			origHashes := hashDir(t, srcDir)

			if err := archiveSkill(root, skillID); err != nil {
				t.Fatalf("archiveSkill: %v", err)
			}
			if err := os.RemoveAll(srcDir); err != nil {
				t.Fatalf("RemoveAll: %v", err)
			}
			if err := restoreSkill(root, skillID, false); err != nil {
				t.Fatalf("restoreSkill: %v", err)
			}

			restoredHashes := hashDir(t, srcDir)
			for rel, origHash := range origHashes {
				if restoredHashes[rel] != origHash {
					t.Errorf("hash mismatch for %s in %s", rel, skillID)
				}
			}
		})
	}
}
