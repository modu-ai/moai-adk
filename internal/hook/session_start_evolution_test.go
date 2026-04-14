package hook

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestEnsureNewSkillSymlinks_EmptyDirectory verifies that an empty new-skills
// directory results in no symlinks and no error.
func TestEnsureNewSkillSymlinks_EmptyDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create empty new-skills directory.
	newSkillsDir := filepath.Join(root, ".moai", "evolution", "new-skills")
	if err := os.MkdirAll(newSkillsDir, 0o755); err != nil {
		t.Fatalf("mkdir new-skills: %v", err)
	}

	count := ensureNewSkillSymlinks(root)
	if count != 0 {
		t.Errorf("expected 0 symlinks for empty directory, got %d", count)
	}
}

// TestEnsureNewSkillSymlinks_MissingDirectory verifies that a missing new-skills
// directory is handled gracefully (no panic, returns 0).
func TestEnsureNewSkillSymlinks_MissingDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Do NOT create .moai/evolution/new-skills/.

	count := ensureNewSkillSymlinks(root)
	if count != 0 {
		t.Errorf("expected 0 for missing new-skills dir, got %d", count)
	}
}

// TestEnsureNewSkillSymlinks_CreatesSymlink verifies that a subdirectory under
// new-skills results in a symlink under .claude/skills/.
// Skipped on Windows (symlink permissions vary).
func TestEnsureNewSkillSymlinks_CreatesSymlink(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink test skipped on Windows; copy-fallback covered by TestEnsureNewSkillSymlinks_WindowsCopyFallback")
	}

	root := t.TempDir()

	// Create a new evolved skill directory with a SKILL.md.
	skillName := "moai-evolved-auth"
	skillSrc := filepath.Join(root, ".moai", "evolution", "new-skills", skillName)
	if err := os.MkdirAll(skillSrc, 0o755); err != nil {
		t.Fatalf("mkdir new skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("# Evolved Auth Skill"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	count := ensureNewSkillSymlinks(root)
	if count != 1 {
		t.Errorf("expected 1 symlink created, got %d", count)
	}

	linkPath := filepath.Join(root, ".claude", "skills", skillName)
	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("lstat symlink: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink", linkPath)
	}

	// Verify the symlink resolves.
	if _, err := os.Stat(linkPath); err != nil {
		t.Errorf("symlink does not resolve: %v", err)
	}
}

// TestEnsureNewSkillSymlinks_SkipsExistingValidSymlink verifies that an existing
// valid symlink is not recreated.
func TestEnsureNewSkillSymlinks_SkipsExistingValidSymlink(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink test skipped on Windows")
	}

	root := t.TempDir()

	skillName := "moai-evolved-cache"
	skillSrc := filepath.Join(root, ".moai", "evolution", "new-skills", skillName)
	if err := os.MkdirAll(skillSrc, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Pre-create the symlink.
	skillsDir := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	linkPath := filepath.Join(skillsDir, skillName)
	relTarget := filepath.Join("..", "..", ".moai", "evolution", "new-skills", skillName)
	if err := os.Symlink(relTarget, linkPath); err != nil {
		t.Fatalf("create initial symlink: %v", err)
	}

	// Call again — should skip and return 0 (not 1).
	count := ensureNewSkillSymlinks(root)
	if count != 0 {
		t.Errorf("expected 0 newly created (existing valid symlink), got %d", count)
	}
}

// TestEnsureNewSkillSymlinks_CleansBrokenSymlink verifies that a broken symlink
// is removed and then recreated.
func TestEnsureNewSkillSymlinks_CleansBrokenSymlink(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink test skipped on Windows")
	}

	root := t.TempDir()

	skillName := "moai-evolved-broken"
	skillSrc := filepath.Join(root, ".moai", "evolution", "new-skills", skillName)
	if err := os.MkdirAll(skillSrc, 0o755); err != nil {
		t.Fatalf("mkdir skill src: %v", err)
	}

	// Pre-create a broken symlink (points to non-existent path).
	skillsDir := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	linkPath := filepath.Join(skillsDir, skillName)
	if err := os.Symlink("/non/existent/path", linkPath); err != nil {
		t.Fatalf("create broken symlink: %v", err)
	}

	count := ensureNewSkillSymlinks(root)
	if count != 1 {
		t.Errorf("expected 1 symlink recreated after cleaning broken one, got %d", count)
	}

	// Verify the new symlink is valid.
	if _, err := os.Stat(linkPath); err != nil {
		t.Errorf("new symlink does not resolve: %v", err)
	}
}

// TestEnsureNewSkillSymlinks_MultipleSkills verifies that multiple new-skill
// directories each get their own symlink.
func TestEnsureNewSkillSymlinks_MultipleSkills(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("symlink test skipped on Windows")
	}

	root := t.TempDir()
	skillNames := []string{"moai-evolved-alpha", "moai-evolved-beta", "moai-evolved-gamma"}

	for _, name := range skillNames {
		skillSrc := filepath.Join(root, ".moai", "evolution", "new-skills", name)
		if err := os.MkdirAll(skillSrc, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}

	count := ensureNewSkillSymlinks(root)
	if count != len(skillNames) {
		t.Errorf("expected %d symlinks, got %d", len(skillNames), count)
	}

	for _, name := range skillNames {
		linkPath := filepath.Join(root, ".claude", "skills", name)
		if _, err := os.Stat(linkPath); err != nil {
			t.Errorf("symlink for %s missing: %v", name, err)
		}
	}
}

// TestEnsureNewSkillSymlinks_IgnoresFiles verifies that plain files (not
// directories) inside new-skills are ignored.
func TestEnsureNewSkillSymlinks_IgnoresFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	newSkillsDir := filepath.Join(root, ".moai", "evolution", "new-skills")
	if err := os.MkdirAll(newSkillsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Place a plain file (e.g. .gitkeep) in new-skills.
	if err := os.WriteFile(filepath.Join(newSkillsDir, ".gitkeep"), []byte{}, 0o644); err != nil {
		t.Fatalf("write .gitkeep: %v", err)
	}

	count := ensureNewSkillSymlinks(root)
	if count != 0 {
		t.Errorf("expected 0 symlinks for file-only new-skills dir, got %d", count)
	}
}
