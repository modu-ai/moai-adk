package cli

// t508 — the `enabled` key is REQUIRED by codex, optional to this parser.
//
// Measured in this run (codex-cli 0.153.4, isolated CODEX_HOME):
//
//	CODEX_HOME=<good> codex mcp list -> rc=0
//	CODEX_HOME=<bad>  codex mcp list -> rc=1
//	  Error: failed to load bootstrap configuration
//	  Caused by: missing field `enabled` in `skills.config`
//
// A single [[skills.config]] entry declaring no `enabled` key therefore kills
// codex startup outright — not a partial degradation. This file is the RED
// guard for that shape: it must FAIL against the current tree.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal is the mutant: a
// [[skills.config]] entry whose path EXISTS but declares no `enabled` key.
// Nothing is stale, so the stale-path finding stays silent — and silence is
// exactly the defect, because codex refuses to start on this config.
func TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t)}, // no EnabledKey: the fatal shape
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)

	if check.Status != uikit.CheckFail {
		t.Errorf("status = %v, want CheckFail — codex cannot start on this config: %+v", check.Status, check)
	}
	text := check.Message + " " + codexDetailText(check)
	if !strings.Contains(text, "enabled") {
		t.Errorf("finding never names the `enabled` key: %+v", check)
	}
}

// TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet is the control: the same
// live path WITH an `enabled` key is the shape codex accepts (rc=0 measured
// above), so it must produce no fatal finding. Without this row the guard
// above passes just as happily on a check that fails every config.
func TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t), EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)

	if check.Status == uikit.CheckFail {
		t.Errorf("a well-formed entry must not be fatal: %+v", check)
	}
}
