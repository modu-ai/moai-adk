// SPEC-V3R3-HARNESS-001 / T-M4-02
// Integration tests for the archiveLegacySkills function.
// When archiveLegacySkills is invoked on a project containing the 16 legacy
// skills plus a user harness skill plus moai-meta-harness:
//   - the 16 legacy skills must be archived
//   - harness-* skills must not be touched
//   - the output format must be verified

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestArchiveLegacySkills_Integration verifies the legacy-skill archive flow
// as an integration test.
func TestArchiveLegacySkills_Integration(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create all 16 legacy skills
	for _, id := range legacySkillIDs {
		makeSkillDir(t, root, id, "# "+id+" content")
	}

	// Create a user-customized skill (must be preserved)
	makeSkillDir(t, root, "harness-test", "# user custom skill")

	// moai-meta-harness (core skill, must not be touched)
	makeSkillDir(t, root, "moai-meta-harness", "# meta harness")

	// force=false: SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 default contract.
	var out bytes.Buffer
	archived, err := archiveLegacySkills(root, &out, false)
	if err != nil {
		t.Fatalf("archiveLegacySkills: %v", err)
	}

	// Verify all 16 skills were archived
	if archived != len(legacySkillIDs) {
		t.Errorf("archived count = %d, want %d", archived, len(legacySkillIDs))
	}

	// Verify the archive directory exists for each legacy skill
	for _, id := range legacySkillIDs {
		archiveDir := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, id)
		if _, statErr := os.Stat(archiveDir); statErr != nil {
			t.Errorf("archive not created for %s: %v", id, statErr)
		}
	}

	// Verify the harness-test skill is NOT archived
	userArchive := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, "harness-test")
	if _, statErr := os.Stat(userArchive); statErr == nil {
		t.Error("harness-test should NOT be archived (user customization)")
	}

	// moai-meta-harness must not be archived either
	metaArchive := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, "moai-meta-harness")
	if _, statErr := os.Stat(metaArchive); statErr == nil {
		t.Error("moai-meta-harness should NOT be archived (not a legacy skill)")
	}

	// Verify the output format: "archive: <id> → ..." lines
	output := out.String()
	for _, id := range legacySkillIDs {
		expected := "archive: " + id
		if !strings.Contains(output, expected) {
			t.Errorf("output missing archive line for %s, output:\n%s", id, output)
		}
	}

	// Summary line: "total: N skills archived"
	if !strings.Contains(output, "total:") {
		t.Errorf("output missing summary line, got:\n%s", output)
	}
}

// TestArchiveLegacySkills_PartialPresent verifies correct behavior when only a
// subset of the legacy skills is present.
func TestArchiveLegacySkills_PartialPresent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create only the first 5 skills
	presentSkills := legacySkillIDs[:5]
	for _, id := range presentSkills {
		makeSkillDir(t, root, id, "# "+id)
	}

	// force=false: default contract for partial-present scenarios.
	var out bytes.Buffer
	archived, err := archiveLegacySkills(root, &out, false)
	if err != nil {
		t.Fatalf("archiveLegacySkills partial: %v", err)
	}

	if archived != len(presentSkills) {
		t.Errorf("archived = %d, want %d", archived, len(presentSkills))
	}
}
