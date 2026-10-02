package cli

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The tier seed is the difference between a factory that advances on its own and
// one that stops at every commit. These tests are deliberately NOT parallel:
// they read and write a process-global variable, and t.Setenv forbids it anyway.
//
// They drive the two factory entry helpers that call seedAutonomyTier — the
// leader and the lane — because those are the only launch paths that seed it
// (SPEC-LAUNCHER-ENTRY-FLAGS-001 M5a re-pinned them from the removed kanban
// entries).

// factoryTierEntries are the two launch helpers that seed the tier.
var factoryTierEntries = []struct {
	name  string
	enter func() func()
}{
	{"leader", func() func() { return enterFactoryLeaderMode(1, "") }},
	{"lane", func() func() { return enterFactoryLaneMode("lane-1", 0, "", config.FactoryDispatchAuto) }},
}

// TestFactoryModeSeedsFullyAutonomousTier is the load-bearing property of the
// factory entry. Without the seed the variable stays unset, config.AutonomyTier
// fails safe to semi-auto, and every commit pays the synchronous vet+lint+test
// gate — the most-interrupted tier, reached by launching the mode built to
// avoid it.
func TestFactoryModeSeedsFullyAutonomousTier(t *testing.T) {
	for _, c := range factoryTierEntries {
		t.Run(c.name, func(t *testing.T) {
			clearFactoryTestEnv(t)
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

// TestFactoryModePreservesExplicitTier: an operator who names a tier has made a
// choice, and the factory entry is not entitled to overrule it. Someone running
// the factory at semi-auto on purpose — to watch a risky card go through — must
// get semi-auto.
func TestFactoryModePreservesExplicitTier(t *testing.T) {
	for _, c := range factoryTierEntries {
		for _, tier := range []string{config.AutonomyTierSemiAuto, config.AutonomyTierAutomatic} {
			t.Run(c.name+"_"+tier, func(t *testing.T) {
				clearFactoryTestEnv(t)
				t.Setenv(config.EnvAutonomyTier, tier)

				restore := c.enter()
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
}

// TestFactoryModeTreatsBlankTierAsUnset: a wrapper that exports the name with no
// value has not made a choice, and config.AutonomyTier already reads blank as
// semi-auto. Honoring the blank would hand the factory the most-interrupted tier
// through an empty string nobody typed on purpose.
func TestFactoryModeTreatsBlankTierAsUnset(t *testing.T) {
	for _, c := range factoryTierEntries {
		for _, blank := range []string{"", "   "} {
			t.Run(c.name+"_blank="+blank, func(t *testing.T) {
				clearFactoryTestEnv(t)
				t.Setenv(config.EnvAutonomyTier, blank)

				restore := c.enter()
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
}
