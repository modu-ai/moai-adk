package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// rawAuditPin is one backend pin of a raw-decoded workflow.audit block.
type rawAuditPin struct {
	Model  string `yaml:"model"`
	Effort string `yaml:"effort"`
}

// rawAuditBlock decodes workflow.audit without the loader, so a key that the
// loader would fill from defaults (gates, model) is visible as absent.
type rawAuditBlock struct {
	Model  string         `yaml:"model"`
	Gates  map[string]any `yaml:"gates"`
	Claude rawAuditPin    `yaml:"claude"`
	Codex  rawAuditPin    `yaml:"codex"`
	GLM    rawAuditPin    `yaml:"glm"`
}

func readRawAuditBlock(t *testing.T, path string) (rawAuditBlock, string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", path, err)
	}
	var doc struct {
		Workflow struct {
			Audit rawAuditBlock `yaml:"audit"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("yaml.Unmarshal(%q): %v", path, err)
	}
	return doc.Workflow.Audit, string(data)
}

// TestCommittedWorkflowYamlAuditAlignment pins the repository's committed
// .moai/config/sections/workflow.yaml against the D8-adopted audit posture
// (operator decision D8, commit "adopt codex-single audit gates"): audit.model
// is the codex-single token with an EXPLICIT gates key (claude off, codex
// required, glm off), and the claude pin deliberately diverges from the Go
// defaults (opus[1m]/medium). The codex and glm pins are unchanged from the
// multi era. The distributed template yaml stays neutral — no model key, no
// gates, and the claude pin at the default effort. The repository root comes
// from this file's location, not from the working directory or
// CLAUDE_PROJECT_DIR.
func TestCommittedWorkflowYamlAuditAlignment(t *testing.T) {
	repoRoot := findRepoRootFromCaller(t)
	committedPath := filepath.Join(repoRoot, ".moai", "config", "sections", "workflow.yaml")
	templatePath := filepath.Join(repoRoot,
		"internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml")

	committed, body := readRawAuditBlock(t, committedPath)

	if committed.Model != AuditModelCodex {
		t.Errorf("committed workflow.audit.model = %q, want %q (D8 codex-single)", committed.Model, AuditModelCodex)
	}
	if committed.Claude.Effort != "medium" {
		t.Errorf("committed claude pin effort = %q, want %q (the D8-adopted divergence from DefaultClaudeAuditEffort)", committed.Claude.Effort, "medium")
	}
	if committed.Claude.Model != "opus[1m]" {
		t.Errorf("committed claude pin model = %q, want %q (D8)", committed.Claude.Model, "opus[1m]")
	}
	if committed.Codex.Model != "gpt-6.1-sol" || committed.Codex.Effort != "high" {
		t.Errorf("committed codex pin = {%s %s}, want {gpt-6.1-sol high}", committed.Codex.Model, committed.Codex.Effort)
	}
	if committed.GLM.Model != "glm-5.3" || committed.GLM.Effort != "max" {
		t.Errorf("committed glm pin = {%s %s}, want {glm-5.3 max}", committed.GLM.Model, committed.GLM.Effort)
	}
	wantGates := map[string]any{"claude": "off", "codex": "required", "glm": "off"}
	if !reflect.DeepEqual(committed.Gates, wantGates) {
		t.Errorf("committed workflow.audit.gates = %v, want %v (the D8 explicit posture — not the model-token resolution)", committed.Gates, wantGates)
	}
	if strings.Contains(body, "claude-opus-5-5/medium") {
		t.Errorf("committed workflow.yaml header comment still names claude-opus-5-5/medium")
	}

	// The committed file loads through the config loader to the same values.
	cfg, err := NewLoader().Load(filepath.Join(repoRoot, ".moai"))
	if err != nil {
		t.Fatalf("Loader.Load(committed tree): %v", err)
	}
	if got := cfg.Workflow.Audit.Model; got != AuditModelCodex {
		t.Errorf("loaded Audit.Model = %q, want %q (D8 codex-single)", got, AuditModelCodex)
	}
	if got := cfg.Workflow.Audit.Claude.Effort; got != "medium" {
		t.Errorf("loaded Audit.Claude.Effort = %q, want %q (D8)", got, "medium")
	}

	// The Go default stays claude: the opt-in lives in this repository's yaml only.
	if got := NewDefaultConfig().Workflow.Audit.Model; got != AuditModelClaude {
		t.Errorf("Go default Audit.Model = %q, want %q", got, AuditModelClaude)
	}

	// Template neutrality: pins only, claude pin already at the default effort.
	tmpl, _ := readRawAuditBlock(t, templatePath)
	if tmpl.Model != "" {
		t.Errorf("template workflow.audit.model = %q, want no model key (distributed default lives in code)", tmpl.Model)
	}
	if tmpl.Gates != nil {
		t.Errorf("template workflow.audit carries a gates key %v, want none", tmpl.Gates)
	}
	if tmpl.Claude.Effort != DefaultClaudeAuditEffort {
		t.Errorf("template claude pin effort = %q, want %q", tmpl.Claude.Effort, DefaultClaudeAuditEffort)
	}
}
