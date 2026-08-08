package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowBackwardCompat pins the REQ-WCR-003 preservation contract: removing
// the token_budget / auto_clear fields from the WEB EDIT SURFACE must not remove
// the yaml keys, the struct members, or the accessors. A workflow.yaml carrying
// both blocks must still load without error and the accessors must return the
// on-disk values.
func TestWorkflowBackwardCompat(t *testing.T) {
	root := t.TempDir()
	sections := filepath.Join(root, "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	yaml := `workflow:
    auto_clear:
        enabled: true
        after_plan: true
        after_run: false
        token_threshold: 123456
    token_budget:
        plan: 11111
        run: 22222
        sync: 33333
`
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(root)
	if err != nil {
		t.Fatalf("Load returned error for a config carrying token_budget/auto_clear: %v", err)
	}

	if got := cfg.WorkflowPlanTokens(); got != 11111 {
		t.Errorf("WorkflowPlanTokens() = %d, want 11111 (yaml key must still bind after the web-surface removal)", got)
	}
	if got := cfg.WorkflowRunTokens(); got != 22222 {
		t.Errorf("WorkflowRunTokens() = %d, want 22222", got)
	}
	if got := cfg.WorkflowSyncTokens(); got != 33333 {
		t.Errorf("WorkflowSyncTokens() = %d, want 33333", got)
	}
	if !cfg.WorkflowAutoClearEnabled() {
		t.Error("WorkflowAutoClearEnabled() = false, want true (auto_clear.enabled must still bind)")
	}
	if cfg.Workflow.AutoClear.TokenThreshold != 123456 {
		t.Errorf("Workflow.AutoClear.TokenThreshold = %d, want 123456", cfg.Workflow.AutoClear.TokenThreshold)
	}
}
