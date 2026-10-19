package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Active review gates must use head-bound Codex evidence; retired providers
// cannot hold sync or merge. Historical reports and runtime test fixtures are excluded.
func TestReviewRetirementActivePolicies(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	files := []string{"AGENTS.md.tmpl", ".claude/agents/moai/manager-lead.md"}
	for _, n := range []string{"factory-dispatch.md", "factory-dispatch-core.md", "factory-dispatch-detail.md", "factory-dispatch-gates.md", "factory-dispatch-mechanics.md", "factory-dispatch-cards.md"} {
		files = append(files, ".claude/rules/moai/workflow/"+n)
	}
	for _, rel := range files {
		raw, err := os.ReadFile(filepath.Join(root, "internal/template/templates", rel))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "CodeRabbit") {
			t.Errorf("retired review policy remains: %s", rel)
		}
	}
}
