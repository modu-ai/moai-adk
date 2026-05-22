// SPEC-V3R3-HARNESS-001 / T-M4-05
// Idempotency tests for archiveLegacySkills:
// verifies that the archive state does not change when invoked twice.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestArchiveIdempotency verifies that the archive state (file list + hashes)
// does not change after calling archiveLegacySkills twice.
func TestArchiveIdempotency(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create all 16 legacy skills
	for _, id := range legacySkillIDs {
		makeSkillDir(t, root, id, "# "+id+" idempotency test content")
	}

	// First run (force=false: SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 REQ-UAC-006 — default idempotency contract)
	var out1 bytes.Buffer
	archived1, err := archiveLegacySkills(root, &out1, false)
	if err != nil {
		t.Fatalf("first archiveLegacySkills: %v", err)
	}

	// Snapshot the archive state
	archiveRoot := filepath.Join(root, ".moai", "archive", "skills", archiveVersion)
	snap1 := takeArchiveSnapshot(t, archiveRoot)

	// Second run (idempotent, force=false)
	var out2 bytes.Buffer
	archived2, err := archiveLegacySkills(root, &out2, false)
	if err != nil {
		t.Fatalf("second archiveLegacySkills: %v", err)
	}

	// Take another archive-state snapshot
	snap2 := takeArchiveSnapshot(t, archiveRoot)

	// File count must be identical
	if len(snap1) != len(snap2) {
		t.Errorf("archive file count changed: first=%d second=%d",
			len(snap1), len(snap2))
	}

	// Hashes must be identical (no duplication / no tampering)
	for path, hash1 := range snap1 {
		hash2, ok := snap2[path]
		if !ok {
			t.Errorf("file missing in second snapshot: %s", path)
			continue
		}
		if hash1 != hash2 {
			t.Errorf("archive content changed for %s: first=%s second=%s",
				path, hash1, hash2)
		}
	}

	// Verify archived counts: no new files should be archived on the second run
	// (with the source unchanged, no drift error is raised and the return is 0
	// or matches the first run.)
	// archived1 == len(legacySkillIDs), archived2 == 0 (already archived) or the same.
	if archived1 != len(legacySkillIDs) {
		t.Errorf("first run archived=%d, want %d", archived1, len(legacySkillIDs))
	}
	// Second run: the source is unchanged, so everything is skipped without drift → 0 or identical
	if archived2 != 0 {
		t.Errorf("second run should archive 0 (already done), got %d", archived2)
	}
}

// TestArchiveIdempotency_NoNewFiles verifies that no new files are added to the
// archive after a second run.
func TestArchiveIdempotency_NoNewFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create only a subset of the skills
	for _, id := range legacySkillIDs[:8] {
		makeSkillDir(t, root, id, "content for "+id)
	}

	var out bytes.Buffer
	if _, err := archiveLegacySkills(root, &out, false); err != nil {
		t.Fatalf("first run: %v", err)
	}

	archiveRoot := filepath.Join(root, ".moai", "archive", "skills", archiveVersion)
	before := countDirFiles(t, archiveRoot)

	if _, err := archiveLegacySkills(root, &out, false); err != nil {
		t.Fatalf("second run: %v", err)
	}

	after := countDirFiles(t, archiveRoot)
	if before != after {
		t.Errorf("archive file count changed: before=%d after=%d (duplicate files?)",
			before, after)
	}
}

// takeArchiveSnapshot returns a path → hash map for every file under archiveRoot.
func takeArchiveSnapshot(t *testing.T, archiveRoot string) map[string]string {
	t.Helper()
	result := make(map[string]string)
	_ = filepath.WalkDir(archiveRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(archiveRoot, path)
		if err != nil {
			return err
		}
		rawHash, err := hashFile(path)
		if err != nil {
			return err
		}
		result[rel] = string(rawHash)
		return nil
	})
	return result
}

// countDirFiles returns the number of files under dir.
func countDirFiles(t *testing.T, dir string) int {
	t.Helper()
	var count int
	_ = filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		count++
		return nil
	})
	return count
}
