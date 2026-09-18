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
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
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
// deployed — AGENTS.md, the wiring trio, the codex agent TOMLs, the 16
// published skills, the remapped catalog skills, the config sections, and the
// git infrastructure.
func TestInitCodexOnlyRequiredSurfaces(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"llm": "gpt"})

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

	// .codex/agents/moai/*.toml — 12 template TOMLs (mission-governor added).
	tomls, err := filepath.Glob(filepath.Join(projectDir, ".codex", "agents", "moai", "*.toml"))
	if err != nil {
		t.Fatalf("glob codex agent tomls: %v", err)
	}
	if len(tomls) != 12 {
		t.Errorf(".codex/agents/moai/*.toml count = %d, want 12", len(tomls))
	}

	// 16 published skills — real template files.
	for _, cmd := range []string{"plan", "run", "sync", "fix", "gate", "goal", "loop",
		"mx", "clean", "codemaps", "e2e", "feedback", "harness", "project", "review", "todo"} {
		p := filepath.Join(projectDir, ".agents", "skills", "moai-"+cmd, "SKILL.md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("published skill missing after codex-only init: %s: %v", p, err)
		}
	}

	// Catalog skills remapped to .agents/skills/<name> — every catalog skill
	// directory present under the new root. The catalog source of truth is the
	// embedded FS's .claude/skills listing (34 directories).
	embeddedFS, fsErr := template.EmbeddedTemplates()
	if fsErr != nil {
		t.Fatalf("embedded templates: %v", fsErr)
	}
	entries, rdErr := fs.ReadDir(embeddedFS, filepath.ToSlash(filepath.Join(".claude", "skills")))
	if rdErr != nil {
		t.Fatalf("read embedded catalog skills: %v", rdErr)
	}
	catalogNames := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			catalogNames = append(catalogNames, e.Name())
		}
	}
	if len(catalogNames) == 0 {
		t.Fatal("embedded catalog skill listing is empty — precondition broken")
	}
	for _, name := range catalogNames {
		p := filepath.Join(projectDir, ".agents", "skills", name, "SKILL.md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("remapped catalog skill missing after codex-only init: %s: %v", name, err)
		}
	}

	// Universal config + git infrastructure.
	for _, rel := range []string{".moai/config/sections/llm.yaml", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Errorf("%s missing after codex-only init: %v", rel, err)
		}
	}
}
