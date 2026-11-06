package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// A lowercase skill.md must not be renamed by the housekeeping step, which runs before
// the user confirms the template sync. A cancelled update would otherwise leave the
// rename behind with no way back. The rename runs only after the update is confirmed.
func TestHousekeeping_DoesNotRenameSkillFileBeforeConfirmation(t *testing.T) {
	projectDir := initIdentityProject(t)
	dir := filepath.Join(projectDir, ".claude", "skills", "moai-domain-backend")
	upper := filepath.Join(dir, "SKILL.md")
	if err := os.Remove(upper); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove shipped SKILL.md to set up the lowercase case: %v", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	lower := filepath.Join(dir, "skill.md")
	if err := os.WriteFile(lower, []byte("lower body\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runUpdateHousekeepingDeferRenames(projectDir)

	if _, err := os.Lstat(lower); err != nil {
		t.Errorf("skill.md was renamed before the update was confirmed (lstat err=%v)", err)
	}
	// Compare listed names: on a case-insensitive filesystem a stat of SKILL.md resolves
	// to skill.md and would report a rename that did not happen.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() == "SKILL.md" {
			t.Errorf("SKILL.md exists before the update was confirmed (entries = %v)", entries)
		}
	}

	// Once the update is confirmed, the held-back rename runs.
	applyDeferredSkillRenames(projectDir)
	entries, err = os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var sawUpper bool
	for _, e := range entries {
		if e.Name() == "SKILL.md" {
			sawUpper = true
		}
	}
	if !sawUpper {
		t.Errorf("confirmed update did not apply the held-back rename (entries = %v)", entries)
	}
}
