package codexwiring

import "testing"

// TestSD_AC022_EnvVarsAllowlistFrozen — REQ-SD-022 (AC-SD-022): the generated
// Codex MCP server table's env_vars allowlist is frozen at the run-start
// develop tree. The production value is compared byte-identically against the
// snapshot literal pinned below; a test cannot read git at CI time reliably,
// so the expected literal is embedded and its provenance named here.
//
// Baseline attribution (verification-claim-integrity §2): the pinned literal
// is the verbatim `mcpServerEnvVarsValue` read from
// `git show a7190891d:internal/codexwiring/configtoml.go` — the run-start
// develop snapshot (SPEC-FACTORY-SELF-DISPATCH-001 §E.2 C.1).
func TestSD_AC022_EnvVarsAllowlistFrozen(t *testing.T) {
	const snapshotValue = `["MOAI_HOME", "MOAI_KANBAN_ID", "MOAI_SESSION_PID", "MOAI_KANBAN_BACKEND", "MOAI_FACTORY_WORKER", "MOAI_FACTORY_WORKERS", "CLAUDE_PROJECT_DIR", "CLAUDE_CODE_SESSION_ID"]`
	if mcpServerEnvVarsValue == "" {
		t.Fatal("mcpServerEnvVarsValue is empty — the freeze comparison would pass vacuously")
	}
	if mcpServerEnvVarsValue != snapshotValue {
		t.Errorf("mcpServerEnvVarsValue drifted from the run-start snapshot a7190891d:\n got: %s\nwant: %s",
			mcpServerEnvVarsValue, snapshotValue)
	}
}
