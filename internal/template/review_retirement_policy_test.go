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
	files := []string{"CLAUDE.md.tmpl", ".claude/agents/moai/manager-lead.md"}
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
	runtimePolicy, err := os.ReadFile(filepath.Join(root, "internal/hook/session_start_factory_i18n.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, retired := range []string{"CodeRabbit", "Merge Risk", "commit-status API"} {
		if strings.Contains(string(runtimePolicy), retired) {
			t.Errorf("retired review wait remains in active factory runtime guidance: %q", retired)
		}
	}
	if got := strings.Count(string(runtimePolicy), "codex-review-receipt:v1"); got != 4 {
		t.Errorf("native Codex receipt guidance appears in %d locales, want 4", got)
	}
}
