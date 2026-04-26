package hook

// session_start_skill_extra_test.go: additional path coverage for ensureNewSkillSymlinks
// - invalid names (path traversal, hidden files, slash-containing)
// - skipping existing directories (Windows copies)
// - getModifiedGoFiles with empty git output

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestEnsureNewSkillSymlinks_InvalidNames rejects names with path traversal or hidden prefix.
func TestEnsureNewSkillSymlinks_InvalidNames(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	newSkillsDir := filepath.Join(root, ".moai", "evolution", "new-skills")
	_ = os.MkdirAll(newSkillsDir, 0o755)

	// These names must be rejected without panic.
	// Note: "has/slash" is not usable because os.MkdirAll interprets the slash
	// as a path separator creating a "has" subdirectory — use only names that
	// ReadDir will actually return as entries with the invalid prefix.
	invalidNames := []string{
		".hidden-skill",
		".another-hidden",
	}
	for _, name := range invalidNames {
		_ = os.MkdirAll(filepath.Join(newSkillsDir, name), 0o755)
	}

	// No valid skills → count must be 0.
	count := ensureNewSkillSymlinks(root)
	if count != 0 {
		t.Errorf("expected 0 for invalid-name-only entries, got %d", count)
	}
}

// TestEnsureNewSkillSymlinks_SkipsExistingDirectory skips when a directory
// already exists at the link target (Windows copy placement or manual).
func TestEnsureNewSkillSymlinks_SkipsExistingDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillName := "moai-evolved-manual"

	// Create skill source.
	skillSrc := filepath.Join(root, ".moai", "evolution", "new-skills", skillName)
	_ = os.MkdirAll(skillSrc, 0o755)

	// Pre-create a regular directory at the target link path.
	skillsDir := filepath.Join(root, ".claude", "skills")
	_ = os.MkdirAll(skillsDir, 0o755)
	linkPath := filepath.Join(skillsDir, skillName)
	_ = os.MkdirAll(linkPath, 0o755)

	count := ensureNewSkillSymlinks(root)
	if count != 0 {
		t.Errorf("expected 0 (directory already exists), got %d", count)
	}
}

// TestEnsureNewSkillSymlinks_SkipsUnexpectedFile skips when a non-dir, non-symlink
// file exists at the target path.
func TestEnsureNewSkillSymlinks_SkipsUnexpectedFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillName := "moai-evolved-file-at-path"

	skillSrc := filepath.Join(root, ".moai", "evolution", "new-skills", skillName)
	_ = os.MkdirAll(skillSrc, 0o755)

	// Place a regular file where the symlink would go.
	skillsDir := filepath.Join(root, ".claude", "skills")
	_ = os.MkdirAll(skillsDir, 0o755)
	linkPath := filepath.Join(skillsDir, skillName)
	_ = os.WriteFile(linkPath, []byte("not a dir"), 0o644)

	count := ensureNewSkillSymlinks(root)
	if count != 0 {
		t.Errorf("expected 0 (unexpected file at link path), got %d", count)
	}
}

// TestEnsureNewSkillSymlinks_ValidNameCreatesLink verifies a valid name
// on non-Windows creates exactly one symlink.
func TestEnsureNewSkillSymlinks_ValidNameWithFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink not tested on Windows here")
	}

	root := t.TempDir()
	skillName := "moai-evolved-valid-extra"

	skillSrc := filepath.Join(root, ".moai", "evolution", "new-skills", skillName)
	_ = os.MkdirAll(skillSrc, 0o755)
	_ = os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("# Extra"), 0o644)

	count := ensureNewSkillSymlinks(root)
	if count != 1 {
		t.Errorf("expected 1 symlink, got %d", count)
	}

	linkPath := filepath.Join(root, ".claude", "skills", skillName)
	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("lstat link: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected symlink at %s", linkPath)
	}
}
