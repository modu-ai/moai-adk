package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowCandidateCIDefaultFalse pins the shipped default
// (SPEC-CANDIDATE-CI-001 REQ-CCI-006): the candidate-CI path ships OFF —
// no candidate verb operates and the merge step's landing check stays the
// absent no-op seam — while guard_bundle_required ships TRUE, preserving
// today's gating where the drift-prone guard families ride the ordinary
// required suite (design.md D6).
func TestWorkflowCandidateCIDefaultFalse(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultConfig()
	if cfg.Workflow.CandidateCI.Enabled {
		t.Errorf("Workflow.CandidateCI.Enabled: got true, want false (shipped default)")
	}
	if !cfg.Workflow.CandidateCI.GuardBundleRequired {
		t.Errorf("Workflow.CandidateCI.GuardBundleRequired: got false, want true (default preserves today's guard gating)")
	}
}

// TestWorkflowCandidateCILoaderRoundTrip pins the workflow.yaml surface:
// a section file carrying candidate_ci overrides both defaults.
func TestWorkflowCandidateCILoaderRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := "workflow:\n  candidate_ci:\n    enabled: true\n    guard_bundle_required: false\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Workflow.CandidateCI.Enabled {
		t.Errorf("candidate_ci.enabled: got false, want true after section load")
	}
	if cfg.Workflow.CandidateCI.GuardBundleRequired {
		t.Errorf("candidate_ci.guard_bundle_required: got true, want false after section load")
	}
}
