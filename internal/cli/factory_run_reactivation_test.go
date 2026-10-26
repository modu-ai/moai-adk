package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// REQ-FCR-024 (card t1963, ledger L12) at the leader-entry surface: starting or
// resuming a leader under the id of a run a consolidation retired is refused
// with the typed error, and the message names both the consolidated run and the
// run its cards went to. Both writers wrap with %w, so the typed error survives.
func TestLeaderEntryRefusesAConsolidatedRun(t *testing.T) {
	check := func(t *testing.T, err error) {
		t.Helper()
		if !errors.Is(err, homestate.ErrRunRetiredByConsolidation) {
			t.Fatalf("err = %v, want homestate.ErrRunRetiredByConsolidation", err)
		}
		if !strings.Contains(err.Error(), "run-old") || !strings.Contains(err.Error(), "run-new") {
			t.Fatalf("refusal %q must name the consolidated run (run-old) and its target (run-new)", err)
		}
	}
	t.Run("start", func(t *testing.T) {
		_, _ = fcFixture(t)
		root := factoryCardRoot()
		caaSeedMovedCard(t, root)
		t.Setenv(config.EnvFactoryRunID, "run-old")
		check(t, recordLeaderRunStart(root, "", BackendClaude, "", 0))
	})
	t.Run("resume", func(t *testing.T) {
		_, _ = fcFixture(t)
		root := factoryCardRoot()
		caaSeedMovedCard(t, root)
		check(t, resumeDiscoveredRun(root, discovery.VerifiedLeader{RunID: "run-old", PID: 4242, ProcessStart: "fp", Basis: "test"}))
	})
}
