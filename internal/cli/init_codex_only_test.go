package cli

// SPEC-INIT-HARNESS-001 M2 — codex-only deployment (REQ-IH-005, AC-IH-002
// / AC-IH-003): `--llm gpt` deploys ONLY the universal + Codex surfaces.
// The claude-only surfaces (.claude/**, CLAUDE.md, .mcp.json, .claudeignore,
// .moai/status_line.sh) must not appear at the project root at all.
//
// Home-fingerprint discipline (t583 REQ-IQW-014/016): every test that runs a
// full init pins HOME to a per-test temp dir via runInitForAutonomy and never
// touches the real HOME.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInitCodexOnlyDeploysNoClaudeSurface is the negative half of REQ-IH-005
// (AC-IH-002): after a codex-only init, NONE of the claude-only surfaces
// exists at the project root.
func TestInitCodexOnlyDeploysNoClaudeSurface(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"llm": "gpt"})

	for _, rel := range []string{
		".claude",
		"CLAUDE.md",
		".mcp.json",
		".claudeignore",
		filepath.Join(".moai", "status_line.sh"),
	} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			t.Errorf("%s exists after codex-only init — claude-only surface must not deploy (REQ-IH-005)", rel)
		}
	}
}

// TestInitCodexOnlyRequiredSurfaces is the positive half of REQ-IH-005
// (AC-IH-003): after a codex-only init, every universal + Codex surface IS
// deployed — AGENTS.md, the wiring trio, the config sections, and the git
// infrastructure. SPEC-USER-ASSET-INSTALL-001 (REQ-005 / AC-011): the
// project tree carries NO common skill or agent file — the user folders
// (isolated HOME) carry the L0 set, and this test now pins the absence of
// the project placement alongside the surviving project surfaces.
func TestInitCodexOnlyRequiredSurfaces(t *testing.T) {
	projectDir, homeDir := runInitForAutonomy(t, nil, map[string]string{"llm": "gpt"})

	// AGENTS.md — universal contract surface.
	if _, err := os.Stat(filepath.Join(projectDir, "AGENTS.md")); err != nil {
		t.Errorf("AGENTS.md missing after codex-only init: %v", err)
	}

	// Codex wiring trio (codexwiring) + trust sidecar.
	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml", ".moai/state/codex-wiring.json"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Errorf("%s missing after codex-only init: %v", rel, err)
		}
	}

	// NO project-side common-asset placement (AC-011's placement set):
	// no .codex/agents/moai/, no .agents/skills/moai*, no .claude/skills/.
	for _, rel := range []string{".codex/agents/moai", ".agents/skills/moai", ".agents/skills/moai-plan", ".claude/skills/moai"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); !os.IsNotExist(err) {
			t.Errorf("common-asset placement %s exists project-side — REQ-005 forbids it", rel)
		}
	}

	// The user folders carry the L0 set instead (the M2 installer; the
	// dispatcher mirror included — AC-002).
	for _, p := range []string{
		filepath.Join(homeDir, ".agents", "skills", "moai", "SKILL.md"),
		filepath.Join(homeDir, ".agents", "skills", "moai-plan", "SKILL.md"),
		filepath.Join(homeDir, ".claude", "skills", "moai-workflow-tdd", "SKILL.md"),
		filepath.Join(homeDir, ".codex", "agents", "manager-develop.toml"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("user-folder install missing after codex-only init: %s: %v", p, err)
		}
	}

	// Universal config + git infrastructure.
	for _, rel := range []string{".moai/config/sections/llm.yaml", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Errorf("%s missing after codex-only init: %v", rel, err)
		}
	}
}
