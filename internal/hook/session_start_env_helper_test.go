package hook

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// clearFactoryEnv unsets every launch marker the retained factory notices and
// the session-record reader read, so each case starts from a known-absent
// state. t.Setenv registers the restore, so the process env is returned to its
// prior value when the test ends.
//
// Moved from session_start_kanban_test.go (SPEC-LAUNCHER-ENTRY-FLAGS-001 M5b
// first step): the retained factory SessionStart tests call it. The three
// retired markers (the leader marker, the companion label, the SPEC marker) left
// its list with their constants.
func clearFactoryEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvFactoryRunID,
		config.EnvFactorySettingsInjected,
		config.EnvFactoryLeadAddr,
		config.EnvFactoryBackend,
		config.EnvMoaiFactoryWorkers,
		config.EnvMoaiFactoryWorker,
	} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}
