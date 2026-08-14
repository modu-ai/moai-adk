package cli

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The tier seed is the difference between a board that advances on its own and
// one that stops at every commit. These tests are deliberately NOT parallel:
// they read and write a process-global variable, and t.Setenv forbids it anyway.

// TestKanbanModeSeedsFullyAutonomousTier is the load-bearing property of -k.
// Without the seed the variable stays unset, config.AutonomyTier fails safe to
// semi-auto, and every commit pays the synchronous vet+lint+test gate — the
// most-interrupted tier, reached by launching the mode built to avoid it.
func TestKanbanModeSeedsFullyAutonomousTier(t *testing.T) {
	cases := []struct {
		name  string
		enter func() func()
	}{
		{"lead", func() func() { return enterKanbanMode("SPEC-KANBAN-001") }},
		{"companion", func() func() { return enterKanbanCompanionMode("plan-tjlgt1") }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(config.EnvAutonomyTier, "placeholder")
			if err := os.Unsetenv(config.EnvAutonomyTier); err != nil {
				t.Fatalf("unsetenv: %v", err)
			}

			restore := c.enter()

			if got := os.Getenv(config.EnvAutonomyTier); got != config.AutonomyTierFullyAutonomous {
				t.Errorf("%s tier = %q, want %q", c.name, got, config.AutonomyTierFullyAutonomous)
			}
			if got := config.AutonomyTier(); got != config.AutonomyTierFullyAutonomous {
				t.Errorf("config.AutonomyTier() = %q, want %q — the seed must be readable through the canonical reader, not just the raw variable", got, config.AutonomyTierFullyAutonomous)
			}

			restore()

			if _, present := os.LookupEnv(config.EnvAutonomyTier); present {
				t.Errorf("%s left %s set after restore; prior presence was absent", c.name, config.EnvAutonomyTier)
			}
		})
	}
}

// TestKanbanModePreservesExplicitTier: an operator who names a tier has made a
// choice, and -k is not entitled to overrule it. Someone running the board at
// semi-auto on purpose — to watch a risky card go through — must get semi-auto.
func TestKanbanModePreservesExplicitTier(t *testing.T) {
	for _, tier := range []string{config.AutonomyTierSemiAuto, config.AutonomyTierAutomatic} {
		t.Run(tier, func(t *testing.T) {
			t.Setenv(config.EnvAutonomyTier, tier)

			restore := enterKanbanMode("")
			if got := os.Getenv(config.EnvAutonomyTier); got != tier {
				t.Errorf("explicit tier %q was overwritten with %q", tier, got)
			}
			restore()

			if got := os.Getenv(config.EnvAutonomyTier); got != tier {
				t.Errorf("after restore tier = %q, want the operator's %q", got, tier)
			}
		})
	}
}

// TestKanbanModeTreatsBlankTierAsUnset: a wrapper that exports the name with no
// value has not made a choice, and config.AutonomyTier already reads blank as
// semi-auto. Honoring the blank would hand -k the most-interrupted tier through
// an empty string nobody typed on purpose.
func TestKanbanModeTreatsBlankTierAsUnset(t *testing.T) {
	for _, blank := range []string{"", "   "} {
		t.Run("blank="+blank, func(t *testing.T) {
			t.Setenv(config.EnvAutonomyTier, blank)

			restore := enterKanbanMode("")
			if got := os.Getenv(config.EnvAutonomyTier); got != config.AutonomyTierFullyAutonomous {
				t.Errorf("blank tier %q left as %q, want it filled in with %q", blank, got, config.AutonomyTierFullyAutonomous)
			}
			restore()

			if got := os.Getenv(config.EnvAutonomyTier); got != blank {
				t.Errorf("after restore tier = %q, want the prior blank %q", got, blank)
			}
		})
	}
}
