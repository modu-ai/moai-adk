package config

// SPEC-SESSION-ANCHOR-ATTR-001 REQ-SAA-005 — the anchor relocation guard
// ships default-OFF like its BranchGuard / AgentStopGuard siblings: the
// ownership judgment and audit row always run, and only the refusal of a
// flagged relocation is opt-in.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAnchorRelocationGuardDefaultFalse pins the shipped default: the
// refusal layer is OFF — the ownership judgment and its audit row always
// run, a maintainer opts into refusal via local config. Template neutrality:
// no `enabled: true` anywhere under internal/template/templates/.
func TestAnchorRelocationGuardDefaultFalse(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultConfig()
	if cfg.Workflow.AnchorRelocationGuard.Enabled {
		t.Errorf("Workflow.AnchorRelocationGuard.Enabled: got true, want false (shipped default)")
	}
}

// TestAnchorRelocationGuardLoaderRoundTrip pins the workflow.yaml surface:
// `anchor_relocation_guard.enabled: true` in a section file overrides the
// default and enables the refusal layer.
func TestAnchorRelocationGuardLoaderRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := "workflow:\n  anchor_relocation_guard:\n    enabled: true\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Workflow.AnchorRelocationGuard.Enabled {
		t.Errorf("anchor_relocation_guard.enabled: got false, want true after section load")
	}
}
