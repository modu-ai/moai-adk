package hook

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// clearKanbanEnv unsets every kanban variable so each case starts from a
// known-absent state. t.Setenv registers the restore, so the process env is
// returned to its prior value when the test ends.
//
// Moved verbatim from session_start_kanban_test.go (SPEC-LAUNCHER-ENTRY-FLAGS-001
// M5b first step): the retained factory SessionStart tests call it.
func clearKanbanEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanID,
		config.EnvMoaiKanbanSpec,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanSettingsInjected,
		config.EnvMoaiKanbanLeadAddr,
		config.EnvMoaiKanbanBackend,
		config.EnvMoaiFactoryWorkers,
		config.EnvMoaiFactoryWorker,
	} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}
