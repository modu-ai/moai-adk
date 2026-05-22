package template

import (
	"io/fs"
	"path"
	"testing"
)

// TestDevOnlySkillLeak ensures dev-only skills moai-workflow-github and moai-workflow-release
// are NOT registered in the user-facing template tree. Source: SPEC-V3R2-WF-002 REQ-WF002-014
func TestDevOnlySkillLeak(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Set of dev-only skills that MUST NOT appear in the template tree
	devOnlySkills := map[string]bool{
		"moai-workflow-github":  true,
		"moai-workflow-release": true,
	}

	// Walk every path under .claude/skills/ in the EmbeddedTemplates() tree
	walkErr := fs.WalkDir(fsys, ".claude/skills", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip when the .claude/skills/ directory itself does not exist (normal case)
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		// Check whether the last path segment matches a dev-only skill name
		lastName := path.Base(filePath)
		if devOnlySkills[lastName] {
			t.Errorf(
				"DEV_ONLY_SKILL_LEAK: skill %q found at %q. This skill is dev-only (REQ-WF002-014, SPEC-V3R2-WF-002).",
				lastName, filePath,
			)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir error: %v", walkErr)
	}
}
