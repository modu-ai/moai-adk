package cli

// SPEC-INIT-HARNESS-001 M4 — doctor harness-conditional checks (REQ-IH-011,
// AC-IH-010): on a codex-only project the claude-surface findings downgrade
// to the explicit informational line, and the run's verdict is not gated on
// surfaces that were never deployed.
//
// Filename note: doctor_harness_test.go is the pre-existing
// SPEC-V3R3-PROJECT-HARNESS-001 suite — a different harness axis entirely.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// initLegacyCodexOnlyProject seeds the codex-only project shape the doctor
// downgrade serves: llm.harness recorded as gpt and no .claude/ surface. Init
// is claude-only since the general codex execution surface removal (card
// t1909), so the shape a pre-removal `--llm gpt` init produced is seeded by
// hand — the doctor behavior under test is the one existing gpt trees still
// get.
func initLegacyCodexOnlyProject(t *testing.T) string {
	t.Helper()
	projectDir, _ := runInitForAutonomy(t, nil, nil)

	llmPath := filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml")
	data, err := os.ReadFile(llmPath)
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	updated := strings.Replace(string(data), `harness: "claude"`, `harness: "gpt"`, 1)
	if updated == string(data) {
		updated = strings.Replace(string(data), "harness: claude", "harness: gpt", 1)
	}
	if updated == string(data) {
		t.Fatalf("llm.yaml carries no harness line to rewrite:\n%s", data)
	}
	if err := os.WriteFile(llmPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("rewrite llm.yaml harness: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(projectDir, ".claude")); err != nil {
		t.Fatalf("drop the .claude/ surface: %v", err)
	}
	return projectDir
}

// TestDoctorCodexOnlyDowngradesClaudeSurfaces verifies AC-IH-010 on the
// check-result level: after a codex-only init, every claude-surface check
// reports the informational downgrade instead of a warning/failure.
func TestDoctorCodexOnlyDowngradesClaudeSurfaces(t *testing.T) {
	projectDir := initLegacyCodexOnlyProject(t)

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
	projectDir := initLegacyCodexOnlyProject(t)

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
