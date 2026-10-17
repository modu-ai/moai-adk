package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// writeCandidateCIConfig seeds root with a workflow.yaml section carrying
// the given candidate_ci values. It is the config-side fixture for the
// landing-gate wiring tests (SPEC-CANDIDATE-CI-001).
func writeCandidateCIConfig(t *testing.T, root, enabled, guardRequired string) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := "workflow:\n  candidate_ci:\n    enabled: " + enabled + "\n    guard_bundle_required: " + guardRequired + "\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// TestCandidateCIEnabled pins the real gate read (SPEC-CANDIDATE-CI-001
// REQ-CCI-006): the config key workflow.candidate_ci.enabled decides, an
// absent key reads false, and a load failure reads false — the gate never
// enables itself on an unreadable config.
func TestCandidateCIEnabled(t *testing.T) {
	t.Parallel()

	t.Run("absent key reads false", func(t *testing.T) {
		t.Parallel()
		if candidateCIEnabled(t.TempDir()) {
			t.Errorf("candidateCIEnabled: got true, want false for a root with no workflow.yaml")
		}
	})

	t.Run("explicit false reads false", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeCandidateCIConfig(t, root, "false", "true")
		if candidateCIEnabled(root) {
			t.Errorf("candidateCIEnabled: got true, want false with enabled: false")
		}
	})

	t.Run("explicit true reads true", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeCandidateCIConfig(t, root, "true", "true")
		if !candidateCIEnabled(root) {
			t.Errorf("candidateCIEnabled: got false, want true with enabled: true")
		}
	})
}
