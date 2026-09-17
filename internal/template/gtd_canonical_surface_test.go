package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGTDCanonicalSurfaceGolden(t *testing.T) {
	paths := []string{
		"templates/.claude/commands/moai/gtd.md",
		"templates/.claude/skills/moai/workflows/gtd.md",
		"templates/.agents/skills/moai-gtd/SKILL.md",
		"templates/.claude/commands/moai/todo.md",
		"templates/.agents/skills/moai-todo/SKILL.md",
	}
	for _, path := range paths {
		data, err := os.ReadFile(filepath.FromSlash(path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(path, "todo") && !strings.Contains(string(data), "arguments: `gtd` $ARGUMENTS") {
			t.Fatalf("%s is not a thin gtd compatibility path", path)
		}
	}
	if len(publishedSkillNames) != 17 {
		t.Fatalf("published skill count = %d, want 17", len(publishedSkillNames))
	}
}
