package codexwiring

import (
	"strings"
	"testing"
)

// TestMCPServerEnvVarsKeepFactoryMarkers — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1
// (card t1399), AC-016: the Codex MCP server's inherited-environment allowlist
// still lists the factory markers it lists today. The names are the frozen wire
// values (the six factory-read MOAI_KANBAN_* values keep their spelling through
// the whole change); the four below are the ones the allowlist carries, written
// here as literals so a rename of the Go constants cannot move them. The
// byte-identity of the whole allowlist is TestSD_AC022_EnvVarsAllowlistFrozen's.
func TestMCPServerEnvVarsKeepFactoryMarkers(t *testing.T) {
	if mcpServerEnvVarsValue == "" {
		t.Fatal("mcpServerEnvVarsValue is empty — the membership check would pass vacuously")
	}
	for _, name := range []string{
		"MOAI_KANBAN_ID",
		"MOAI_KANBAN_BACKEND",
		"MOAI_FACTORY_WORKER",
		"MOAI_FACTORY_WORKERS",
	} {
		if !strings.Contains(mcpServerEnvVarsValue, `"`+name+`"`) {
			t.Errorf("the allowlist %s no longer lists %q", mcpServerEnvVarsValue, name)
		}
	}
}
