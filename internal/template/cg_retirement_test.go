package template

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCGEmbeddedRetirementPreservesRoutingAndAudit(t *testing.T) {
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	renderer := NewRenderer(embedded)
	render := func(path string) string {
		t.Helper()
		body, err := renderer.Render(path, nil)
		if err != nil {
			t.Fatalf("actual embedded render %s: %v", path, err)
		}
		return string(body)
	}
	web := render(".claude/rules/moai/core/glm-web-tooling.md")
	for _, required := range []string{"moai migrate cg", "--accept-role-change", "--apply", "claude-only", "claude-glm", "verified", "mcp__web_search_prime__webSearchPrime", "mcp__web_reader__webReader", "mcp__zai-mcp-server__analyze_image", "moai glm tools enable"} {
		if !strings.Contains(web, required) {
			t.Errorf("rendered web doctrine lacks %q", required)
		}
	}
	for _, retired := range []string{"cg → activate CG mode", "Enable CG mode inside tmux", "moai cg` injects these", "Leader performs evaluation inline"} {
		if strings.Contains(web, retired) {
			t.Errorf("rendered retired instruction: %s", retired)
		}
	}
	// AGENTS.md-primary product: the standing contract (and its CG-migration
	// row) ships as AGENTS.md.tmpl; the native Agent Teams allowance lives in
	// the orchestration-modes doctrine the contract cross-references.
	agents := render("AGENTS.md.tmpl")
	if !strings.Contains(agents, "moai migrate cg") {
		t.Error("explicit migration missing")
	}
	if strings.Contains(agents, "60-70% cost reduction") {
		t.Error("CG cost guarantee survived")
	}
	team := render(".claude/rules/moai/workflow/orchestration-mode-selection.md")
	if !strings.Contains(team, "Agent Teams") || !strings.Contains(team, "re-allowed") {
		t.Error("native team allowance missing")
	}
	var cfg struct {
		Harness struct {
			Modes map[string]string `yaml:"mode_defaults"`
			Audit struct {
				Always bool `yaml:"always_enabled"`
			} `yaml:"plan_audit_global"`
		} `yaml:"harness"`
	}
	if err := yaml.Unmarshal([]byte(render(".moai/config/sections/harness.yaml")), &cfg); err != nil {
		t.Fatal(err)
	}
	if _, exists := cfg.Harness.Modes["cg"]; exists {
		t.Error("CG still selects harness depth")
	}
	if cfg.Harness.Modes["solo"] != "auto" || cfg.Harness.Modes["team"] != "auto" || !cfg.Harness.Audit.Always {
		t.Error("ordinary mode or mandatory plan audit lost")
	}
	for _, path := range []string{".claude/skills/moai/workflows/run/task-decomposition.md", ".claude/skills/moai/workflows/run/phase-execution.md", ".claude/agents/moai/sync-auditor.md", ".codex/agents/moai/sync-auditor.toml"} {
		body := render(path)
		if !strings.Contains(body, "sync-auditor") {
			t.Errorf("independent auditor missing %s", path)
		}
		if strings.Contains(body, "CG mode:") || strings.Contains(body, "**CG mode**:") {
			t.Errorf("CG inline audit exception survives %s", path)
		}
	}
}
