package cli

// factory_dispatch_test.go holds the SPEC-FACTORY-BOOTSTRAP-001 dispatch-truth-
// table tests. The four-row table at spec.md §A.2 is the semantic core of this
// SPEC; these tests pin each row against the resolveFactoryBranch decision
// function the cc.go / glm.go dispatch sites call.

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// clearAllFactoryEnv unsets EVERY factory variable so each dispatch case starts
// from a known-absent state. t.Setenv registers the restore.
func clearAllFactoryEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiFactory,
		config.EnvMoaiFactoryID,
		config.EnvMoaiFactorySpec,
		config.EnvMoaiFactoryLabel,
		config.EnvMoaiFactorySettingsInjected,
	} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

// TestResolveFactoryBranchTruthTable is AC-FB-001..004 + AC-FB-027: the four-row
// truth table at spec.md §A.2, including the breaking-change 5th case
// (companion-shape --name alone → no-op).
func TestResolveFactoryBranchTruthTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		factoryEnabled bool
		isCompanion    bool
		want           factoryBranch
	}{
		{"row1: -f alone → lead", true, false, factoryBranchLead},
		{"row2: -f --name <companion> → companion", true, true, factoryBranchCompanion},
		{"row3: --name <non-companion> alone → no-op", false, false, factoryBranchNone},
		{"row4: -f --name <non-companion> → lead", true, false, factoryBranchLead},
		{"row5 (BREAKING): --name <companion-shape> alone → no-op", false, true, factoryBranchNone},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := resolveFactoryBranch(c.factoryEnabled, c.isCompanion)
			if got != c.want {
				t.Errorf("resolveFactoryBranch(factoryEnabled=%v, isCompanion=%v) = %v, want %v",
					c.factoryEnabled, c.isCompanion, got, c.want)
			}
		})
	}
}

// TestDispatchOutcome_LeadEnvState is AC-FB-001 / AC-FB-006 / AC-FB-007: driving
// the lead branch through the real enterFactoryMode produces the correct env
// state (MOAI_FACTORY set, MOAI_FACTORY_ID set, MOAI_FACTORY_LABEL unset).
func TestDispatchOutcome_LeadEnvState(t *testing.T) {
	clearAllFactoryEnv(t)
	if resolveFactoryBranch(true, false) != factoryBranchLead {
		t.Fatal("expected lead branch")
	}
	restore := enterFactoryMode("SPEC-FOO-001")
	defer restore()

	if os.Getenv(config.EnvMoaiFactory) == "" {
		t.Errorf("MOAI_FACTORY not set (lead must carry the chain seed)")
	}
	if os.Getenv(config.EnvMoaiFactoryID) == "" {
		t.Errorf("MOAI_FACTORY_ID not set")
	}
	if os.Getenv(config.EnvMoaiFactorySpec) != "SPEC-FOO-001" {
		t.Errorf("MOAI_FACTORY_SPEC = %q, want SPEC-FOO-001", os.Getenv(config.EnvMoaiFactorySpec))
	}
	if _, present := os.LookupEnv(config.EnvMoaiFactoryLabel); present {
		t.Errorf("MOAI_FACTORY_LABEL must NOT be set on a lead")
	}
}

// TestDispatchOutcome_CompanionEnvState is AC-FB-002 / AC-FB-006: driving the
// companion branch through enterFactoryCompanionMode produces the correct env
// state (MOAI_FACTORY_LABEL set, MOAI_FACTORY unset, MOAI_FACTORY_ID unset).
func TestDispatchOutcome_CompanionEnvState(t *testing.T) {
	clearAllFactoryEnv(t)
	if resolveFactoryBranch(true, true) != factoryBranchCompanion {
		t.Fatal("expected companion branch")
	}
	restore := enterFactoryCompanionMode("run-abc123")
	defer restore()

	if os.Getenv(config.EnvMoaiFactoryLabel) != "run-abc123" {
		t.Errorf("MOAI_FACTORY_LABEL = %q, want \"run-abc123\"", os.Getenv(config.EnvMoaiFactoryLabel))
	}
	if _, present := os.LookupEnv(config.EnvMoaiFactory); present {
		t.Errorf("MOAI_FACTORY must NOT be set on a companion")
	}
}

// TestDispatchOutcome_NoOpEnvState is AC-FB-003 / AC-FB-027: the no-op branch
// leaves every factory env var unset — including the breaking-change case where
// --name carries the companion shape but -f is absent.
func TestDispatchOutcome_NoOpEnvState(t *testing.T) {
	clearAllFactoryEnv(t)
	if resolveFactoryBranch(false, true) != factoryBranchNone {
		t.Fatal("expected no-op branch for companion-shape --name alone (AC-FB-027)")
	}
	for _, key := range []string{config.EnvMoaiFactory, config.EnvMoaiFactoryID, config.EnvMoaiFactoryLabel, config.EnvMoaiFactorySpec} {
		if _, present := os.LookupEnv(key); present {
			t.Errorf("%s set on a no-op dispatch", key)
		}
	}
}
