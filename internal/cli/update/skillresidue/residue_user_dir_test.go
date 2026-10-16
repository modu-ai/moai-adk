package skillresidue

import (
	"os"
	"path/filepath"
	"testing"
)

// A user-owned directory named skill.md is not a lowercase skill file. Clean must
// leave it, with its contents, where it is. Renaming it to SKILL.md would move the
// user's files under a new name.
func TestClean_KeepsUserDirectoryNamedSkillMd(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude", "skills", "moai-ref-owasp-checklist")
	mustWriteFile(t, filepath.Join(dir, "skill.md", "notes.txt"), "user notes")
	shipped := func(rel string) bool { return rel == ".claude/skills/moai-ref-owasp-checklist/SKILL.md" }

	if _, err := Clean(root, Options{Shipped: shipped}); err != nil {
		t.Fatalf("Clean: %v", err)
	}

	info, err := os.Lstat(filepath.Join(dir, "skill.md"))
	if err != nil || !info.IsDir() {
		t.Fatalf("user directory skill.md/ was moved or removed (lstat err=%v)", err)
	}
	// Compare the listed names, not a stat: on a case-insensitive filesystem a stat of
	// SKILL.md resolves to the skill.md directory and would hide the rename.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() == "SKILL.md" {
			t.Errorf("user directory skill.md/ was renamed to SKILL.md (entries = %v)", entries)
		}
	}
	if got := readOrFail(t, filepath.Join(dir, "skill.md", "notes.txt")); got != "user notes" {
		t.Errorf("user file inside skill.md/ = %q, want it untouched", got)
	}
}
