package hook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodexReviewGateExecutableDirectoryDoesNotShadowBinary(t *testing.T) {
	requireBash(t)
	base := t.TempDir()
	project := filepath.Join(base, "workspace", "project")
	config := filepath.Join(project, ".moai", "config", "sections")
	for _, dir := range []string{config, filepath.Join(base, "moai")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(config, "workflow.yaml"), []byte(reviewGateWorkflowYAML("codex", true)), 0644); err != nil {
		t.Fatal(err)
	}
	if got := runReviewGateWrapper(t, "handle-codex-review-gate.sh", project); got != 1 {
		t.Fatalf("executable directory shadowed PATH binary: calls=%d want=1", got)
	}
}
