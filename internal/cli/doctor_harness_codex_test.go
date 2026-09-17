package cli

// SPEC-INIT-HARNESS-001 M4 — doctor harness-conditional checks (REQ-IH-011,
// AC-IH-010): on a codex-only project the claude-surface findings downgrade
// to the explicit informational line, and the run's verdict is not gated on
// surfaces that were never deployed.
//
// Filename note: doctor_harness_test.go is the pre-existing
// SPEC-V3R3-PROJECT-HARNESS-001 suite — a different harness axis entirely.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// TestDoctorCodexOnlyDowngradesClaudeSurfaces verifies AC-IH-010 on the
// check-result level: after a codex-only init, every claude-surface check
// reports the informational downgrade instead of a warning/failure.
func TestDoctorCodexOnlyDowngradesClaudeSurfaces(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"llm": "gpt"})

	t.Chdir(projectDir)
	groups := runGroupedChecks(false, "")

	foundClaude := map[string]bool{}
	for _, g := range groups {
		for _, c := range g.checks {
			if !claudeSurfaceCheckNames[c.Name] {
				continue
			}
			foundClaude[c.Name] = true
			if c.Status == uikit.CheckInfo && c.Message == claudeSurfaceDowngradedMessage {
				continue
			}
			if c.Status == uikit.CheckOK {
				// The user re-added the surface by hand — design.md §3 keeps
				// user files respected. Not the expected fresh-codex shape,
				// but never a false alarm either.
				continue
			}
			t.Errorf("claude-surface check %q reports %s/%q — want INFO %q",
				c.Name, c.Status, c.Message, claudeSurfaceDowngradedMessage)
		}
	}
	for _, want := range []string{"Claude Config", "Hooks Config", "Slash Commands"} {
		if !foundClaude[want] {
			t.Errorf("claude-surface check %q missing from the codex-only run", want)
		}
	}
}

// TestDoctorCodexOnlyOutputCarriesInfoLine asserts the downgrade produced an
// INFO finding naming harness=codex (the AC-IH-010 green-path message).
func TestDoctorCodexOnlyOutputCarriesInfoLine(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"llm": "gpt"})

	t.Chdir(projectDir)
	groups := runGroupedChecks(false, "")
	var infos []string
	for _, g := range groups {
		for _, c := range g.checks {
			if c.Status == uikit.CheckInfo && strings.Contains(c.Message, "harness=codex") {
				infos = append(infos, c.Name)
			}
		}
	}
	if len(infos) == 0 {
		t.Error("no INFO finding naming harness=codex found in the codex-only doctor run")
	}
}

// TestDowngradeLeavesOKAndCodexChecksUntouched pins the downgrade boundary:
// only failing/warning claude-surface checks are rewritten; OK checks and
// codex checks pass through unchanged.
func TestDowngradeLeavesOKAndCodexChecksUntouched(t *testing.T) {
	groups := []checkGroup{
		{title: "Workspace", checks: []DiagnosticCheck{
			{Name: "Claude Config", Status: uikit.CheckWarn, Message: ".claude/ directory not found"},
			{Name: "Codex Wiring", Status: uikit.CheckOK, Message: "codex wiring present"},
			{Name: "Hooks Config", Status: uikit.CheckOK, Message: "hooks configured"},
		}},
	}
	got := downgradeClaudeSurfaceChecks(groups)

	claude := got[0].checks[0]
	if claude.Status != uikit.CheckInfo || claude.Message != claudeSurfaceDowngradedMessage {
		t.Errorf("failing claude check not downgraded: %s/%q", claude.Status, claude.Message)
	}
	codex := got[0].checks[1]
	if codex.Status != uikit.CheckOK || codex.Message != "codex wiring present" {
		t.Errorf("codex check rewritten by the downgrade: %s/%q", codex.Status, codex.Message)
	}
	okClaude := got[0].checks[2]
	if okClaude.Status != uikit.CheckOK || okClaude.Message != "hooks configured" {
		t.Errorf("OK claude check rewritten by the downgrade: %s/%q", okClaude.Status, okClaude.Message)
	}
}
