package template

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSkillEntrypointUppercase guards the cross-platform skill-loading contract:
// every template skill MUST use an uppercase SKILL.md entrypoint, never a
// lowercase skill.md. Claude Code looks up SKILL.md (case-sensitively on
// Linux/CI/Docker/Vercel), so a lowercase entrypoint silently fails to load on
// case-sensitive filesystems even though it works on case-insensitive macOS.
//
// On case-insensitive macOS the os.Stat-based check would report a false
// positive (matching SKILL.md when asked for skill.md), so this test reads the
// directory entries directly — ReadDir returns the on-disk, case-preserved
// filenames — and asserts the exact-case name set.
func TestSkillEntrypointUppercase(t *testing.T) {
	skillsDir := filepath.Join("templates", ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read skills dir %q: %v", skillsDir, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(skillsDir, e.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read skill dir %q: %v", e.Name(), err)
		}
		names := make(map[string]bool, len(files))
		for _, f := range files {
			names[f.Name()] = true
		}
		if names["skill.md"] {
			t.Errorf("skill %q has a lowercase entrypoint skill.md — rename to SKILL.md (invisible on case-sensitive Linux/CI/Docker/Vercel per the Claude Code skill spec)", e.Name())
		}
		if !names["SKILL.md"] {
			t.Errorf("skill %q is missing the required SKILL.md entrypoint", e.Name())
		}
	}
}
