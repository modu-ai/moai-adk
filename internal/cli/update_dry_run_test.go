// SPEC-V3R3-HARNESS-001 / T-M4-04
// --dry-run flag tests: the planned operations must be printed without mutating the filesystem.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDryRunArchive verifies that dryRunArchiveLegacySkills emits output prefixed
// with [dry-run] and does not modify the filesystem.
func TestDryRunArchive(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create only 3 of the 16 skills
	presentSkills := []string{
		"moai-domain-backend",
		"moai-domain-frontend",
		"moai-domain-database",
	}
	for _, id := range presentSkills {
		makeSkillDir(t, root, id, "# "+id)
	}

	// Capture an mtime snapshot
	skillsDir := filepath.Join(root, ".claude", "skills")
	preMtimes := snapshotMtimes(t, skillsDir)

	// Run dry-run
	var out bytes.Buffer
	err := dryRunArchiveLegacySkills(root, &out)
	if err != nil {
		t.Fatalf("dryRunArchiveLegacySkills returned error: %v", err)
	}

	// Verify the filesystem is unchanged (mtime comparison)
	postMtimes := snapshotMtimes(t, skillsDir)
	for path, preMtime := range preMtimes {
		postMtime, ok := postMtimes[path]
		if !ok {
			t.Errorf("file disappeared during dry-run: %s", path)
			continue
		}
		if !preMtime.Equal(postMtime) {
			t.Errorf("mtime changed during dry-run for %s: pre=%v post=%v",
				path, preMtime, postMtime)
		}
	}

	// Verify the archive directory was not created
	archiveBase := filepath.Join(root, ".moai", "archive")
	if _, err := os.Stat(archiveBase); err == nil {
		t.Error("archive directory was created during dry-run (should not mutate filesystem)")
	}

	// Output verification: [dry-run] prefix + 3 skill IDs + summary line
	output := out.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("output missing [dry-run] prefix, got:\n%s", output)
	}
	for _, id := range presentSkills {
		if !strings.Contains(output, id) {
			t.Errorf("output missing skill ID %s, got:\n%s", id, output)
		}
	}

	// Verify the summary line
	if !strings.Contains(output, "total:") {
		t.Errorf("output missing summary line containing 'total:', got:\n%s", output)
	}
}

// TestDryRunArchive_NoSkills verifies that, when no legacy skills are present,
// dry-run emits an empty plan and exits without error.
func TestDryRunArchive_NoSkills(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	var out bytes.Buffer
	err := dryRunArchiveLegacySkills(root, &out)
	if err != nil {
		t.Fatalf("dryRunArchiveLegacySkills on empty project: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "total:") {
		t.Errorf("output missing summary line, got:\n%s", output)
	}
	// 0 skills archived
	if !strings.Contains(output, "0 skills archived") {
		t.Errorf("expected '0 skills archived' in output, got:\n%s", output)
	}
}

// snapshotMtimes returns a path→mtime map for every file under dir.
func snapshotMtimes(t *testing.T, dir string) map[string]time.Time {
	t.Helper()
	result := make(map[string]time.Time)
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		result[path] = info.ModTime()
		return nil
	})
	return result
}
