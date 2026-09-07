package cli

// SPEC-CODEX-E2E-GUARD-001 M2 — the init→doctor end-to-end judgment
// (REQ-CEG-001 / REQ-CEG-002). doctor_codex_test.go wires through
// codexwiring.Wire directly, so the "Codex Wiring" check had never been
// judged behind the REAL init command path; these tests close that axis-1
// gap: run a full init (runInitForAutonomyAtHomeCapturingOut), then run the
// doctor judgment (checkCodexWiring) on the project init produced.
//
// Hermetic only: no real codex binary, no network — the doctor-side seams
// (stubMoaiLookup / stubCodexLookup / stubCodexHome) pin every PATH and home
// lookup, and HOME itself is pinned to a temp dir per the init-test
// convention. The production init flow and checkCodexWiring are OBSERVED,
// never modified (spec.md §G).

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// codexFindingPhrases are the directive/c finding phrases only a codex-WIRING
// finding can produce. A healthy codex-inited project must produce none of
// them — the CheckOK status assertion alone would also pass on a check that
// silently degraded to a skip, so the ABSENCE of finding text is asserted
// alongside the status (verification-completeness §1.1: the swept surface is
// named, not just the exit verdict).
var codexFindingPhrases = []string{
	"moai init --agent codex", // the unwired-project action directive
	"stale skill",             // the stale-home-skill sub-check finding
	"/hooks to re-trust",      // the hooks-divergence finding
	"mcp_servers.moai",        // the config-table-drift finding
}

// pinInitHome pins the environment the init tests run under: HOME to a fresh
// temp dir (the USER-scope settings write must never touch the real
// ~/.claude) plus the two sandbox-proof blanks the sibling init tests set.
func pinInitHome(t *testing.T) string {
	t.Helper()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	return homeDir
}

// TestRunInit_ThenDoctorCodexWiringHealthy — REQ-CEG-001 / AC-CEG-001: the
// REAL init path, answering codex, produces a project the doctor "Codex
// Wiring" judgment passes clean. The sanity leg (assertCodexArtifacts) runs
// FIRST so a silently-unwired init cannot vacuously pass the OK assertion.
func TestRunInit_ThenDoctorCodexWiringHealthy(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "codex", MCPProvision: true}
	homeDir := pinInitHome(t)

	projectDir, _ := runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, nil)

	// Sanity leg — the init actually wired the project (anti-vacuous).
	assertCodexArtifacts(t, projectDir, true)

	// Hermetic doctor judgment on the project init produced.
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(projectDir, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("init(codex)→doctor status = %v, want OK: %+v\nmessage: %q\ndetail: %q",
			check.Status, check, check.Message, check.Detail)
	}
	combined := check.Message + " " + codexDetailText(check)
	for _, phrase := range codexFindingPhrases {
		if strings.Contains(combined, phrase) {
			t.Errorf("healthy codex project carries a codex finding phrase %q: message %q / detail %q",
				phrase, check.Message, check.Detail)
		}
	}
}

// TestRunInit_ClaudeOnlyThenDoctorStaysSilent — REQ-CEG-002 / AC-CEG-002 (the
// companion negative): a claude-only init leaves NO codex wiring, and with no
// codex binary on PATH the doctor judgment stays an informational OK — never
// a Warn. The sanity leg asserts the wiring artifacts really are absent, so
// the OK cannot pass by way of a silent wiring failure.
func TestRunInit_ClaudeOnlyThenDoctorStaysSilent(t *testing.T) {
	wiz := &wizard.WizardResult{AgentWiring: "claude", MCPProvision: true}
	homeDir := pinInitHome(t)

	projectDir, _ := runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, nil)

	// Sanity leg — the claude-only init really left no codex wiring.
	assertCodexArtifacts(t, projectDir, false)

	// No codex binary on this machine (moai present, codex absent).
	stubCodexLookup(t, true, false)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(projectDir, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("init(claude)→doctor status = %v, want OK (informational): %+v\nmessage: %q\ndetail: %q",
			check.Status, check, check.Message, check.Detail)
	}
	combined := check.Message + " " + codexDetailText(check)
	for _, phrase := range codexFindingPhrases {
		if strings.Contains(combined, phrase) {
			t.Errorf("claude-only machine carries a codex finding phrase %q: message %q / detail %q",
				phrase, check.Message, check.Detail)
		}
	}
}
