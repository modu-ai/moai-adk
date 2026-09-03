package config

import (
	"testing"
)

// gate_precommit_test.go — gate.pre_commit.enabled schema tests
// (SPEC-PRECOMMIT-GATE-SCOPE-001 M1).
//
// The pre-commit context's heavy gate is default-OFF: the git pre-commit hook
// invokes `moai gate` with the MOAI_PRECOMMIT=1 marker, and the runner honors
// gate.pre_commit.enabled only under that marker. The key defaults to false so
// a single pre-existing project-wide failure never blocks unrelated commits.
// A standalone `moai gate` run never reads this key (operator decision 2).

// TestNewDefaultGateConfigPreCommitDefaultOff pins the default-OFF decision:
// NewDefaultGateConfig seeds PreCommit.Enabled=false (axis (b), operator
// decision 2026-09-03).
func TestNewDefaultGateConfigPreCommitDefaultOff(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultGateConfig()
	if cfg.PreCommit.Enabled {
		t.Error("NewDefaultGateConfig().PreCommit.Enabled = true, want false (pre-commit heavy gate is default-OFF)")
	}
}

// TestLoadGateSectionPreCommitOverride verifies gate.yaml's pre_commit.enabled
// reaches cfg.Gate.PreCommit.Enabled through the shared loadGateSection loader
// (opt-in path).
func TestLoadGateSectionPreCommitOverride(t *testing.T) {
	t.Parallel()
	moaiDir := writeGateYAML(t, "gate:\n  enabled: true\n  pre_commit:\n    enabled: true\n")

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Gate.PreCommit.Enabled {
		t.Error("Gate.PreCommit.Enabled = false, want true when gate.yaml opts in")
	}
	if !loader.LoadedSections()["gate"] {
		t.Error("loadedSections[gate] = false, want true when gate.yaml present")
	}
}

// TestLoadGateSectionPreCommitAbsentKeepsDefault verifies the default-OFF
// posture survives a gate.yaml that omits the pre_commit sub-key entirely
// (partial-override contract: omitted keys keep seeded defaults).
func TestLoadGateSectionPreCommitAbsentKeepsDefault(t *testing.T) {
	t.Parallel()
	moaiDir := writeGateYAML(t, "gate:\n  skip_tests: true\n")

	loader := NewLoader()
	cfg, err := loader.Load(moaiDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Gate.PreCommit.Enabled {
		t.Error("Gate.PreCommit.Enabled = true, want false when gate.yaml omits pre_commit")
	}
}
