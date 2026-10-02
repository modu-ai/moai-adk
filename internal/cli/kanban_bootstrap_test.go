package cli

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestEnterFactoryLeaderModeSetsRunID asserts the factory leader mints a run id
// and publishes the factory markers, and that the restore returns every
// variable to its PRIOR PRESENCE — an unset variable is unset again, not set to
// "". Re-pinned by SPEC-LAUNCHER-ENTRY-FLAGS-001 M5a from the removed kanban
// leader entry, whose run-id and restore contract it shares.
//
// Non-parallel by construction: os.Setenv mutates process-global state.
func TestEnterFactoryLeaderModeSetsRunID(t *testing.T) {
	clearFactoryTestEnv(t)
	keys := []string{config.EnvMoaiKanbanID, config.EnvMoaiKanbanLeadAddr, config.EnvMoaiFactoryWorkers}

	restore := enterFactoryLeaderMode(1, "")

	runID := os.Getenv(config.EnvMoaiKanbanID)
	if runID == "" {
		t.Fatalf("%s not set by enterFactoryLeaderMode", config.EnvMoaiKanbanID)
	}
	if _, ok := kanban.SplitLeaderLabel(kanban.RoleLeader + "-" + runID); !ok {
		t.Errorf("run id %q does not produce a parseable leader label", runID)
	}
	if os.Getenv(config.EnvMoaiFactoryWorkers) != "1" {
		t.Errorf("%s = %q, want 1", config.EnvMoaiFactoryWorkers, os.Getenv(config.EnvMoaiFactoryWorkers))
	}

	restore()
	for _, key := range keys {
		if _, present := os.LookupEnv(key); present {
			t.Errorf("%s still present after restore (prior presence not restored)", key)
		}
	}
}
