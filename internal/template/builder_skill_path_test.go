package template

import (
	"io/fs"
	"strings"
	"testing"
)

// TestBuilderSkillPathStructure verifies that the builder-skill agent instructions
// explicitly enforce the correct skill file path structure.
//
// Regression test for: https://github.com/modu-ai/moai-adk/issues/443
//
// The builder-skill agent was creating files at wrong nested paths like:
//
//	.claude/skills/moai/library/pykrx.md  (WRONG)
//
// instead of the correct flat structure:
//
//	.claude/skills/moai-library-pykrx/SKILL.md  (CORRECT)
func TestBuilderSkillPathStructure(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const agentPath = ".claude/agents/moai/builder-skill.md"

	data, err := fs.ReadFile(fsys, agentPath)
	if err != nil {
		t.Fatalf("read %s: %v", agentPath, err)
	}

	content := string(data)

	// The instructions must explicitly state that the skill file is always
	// named SKILL.md (never the skill name with a .md extension).
	if !strings.Contains(content, "SKILL.md") {
		t.Error("builder-skill.md must mention 'SKILL.md' as the required filename inside the skill directory")
	}

	// The instructions must prohibit creating nested subdirectories under
	// .claude/skills/. The full skill name (e.g. moai-library-pykrx) must
	// become a single directory, never be split into subdirectories.
	hasNestedProhibition := strings.Contains(content, "NEVER create nested") ||
		strings.Contains(content, "never create nested") ||
		strings.Contains(content, "Do NOT create subdirectories") ||
		strings.Contains(content, "do not create subdirectories") ||
		strings.Contains(content, "not split into subdirectories") ||
		strings.Contains(content, "single directory")
	if !hasNestedProhibition {
		t.Error("builder-skill.md must explicitly prohibit creating nested subdirectories inside .claude/skills/ (e.g., 'NEVER create nested subdirectories' or 'single directory')")
	}

	// The instructions must give a concrete example showing the correct path
	// pattern for a skill whose name contains category segments separated by
	// hyphens (e.g., moai-library-pykrx → .claude/skills/moai-library-pykrx/SKILL.md).
	// At minimum the docs should show the full skill-name is used as-is for the directory.
	hasConcreteExample := strings.Contains(content, ".claude/skills/{skill-name}/SKILL.md") ||
		strings.Contains(content, ".claude/skills/<skill-name>/SKILL.md") ||
		strings.Contains(content, ".claude/skills/moai-library-") ||
		strings.Contains(content, "{full-skill-name}/SKILL.md")
	if !hasConcreteExample {
		t.Error("builder-skill.md must include a concrete example demonstrating that the full skill name (including category segments) maps to a single directory, e.g., '.claude/skills/{skill-name}/SKILL.md'")
	}
}
