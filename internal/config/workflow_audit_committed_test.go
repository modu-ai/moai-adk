package config

import (
	"os"
	"path/filepath"
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
// .moai/config/sections/workflow.yaml (SPEC-AUDIT-MODEL-CONVERGE-001 REQ-ACV-019
// / AC-ACV-018): audit.model is the opt-in token multi, the claude pin effort
// equals the Go default, the other pins are unchanged and no gates key exists
// (the model token resolves the gates). The distributed template yaml is the
// neutral counterpart — no model key under audit and the claude pin already at
// the default effort. The repository root comes from this file's location, not
// from the working directory or CLAUDE_PROJECT_DIR.
func TestCommittedWorkflowYamlAuditAlignment(t *testing.T) {
	repoRoot := findRepoRootFromCaller(t)
	committedPath := filepath.Join(repoRoot, ".moai", "config", "sections", "workflow.yaml")
	templatePath := filepath.Join(repoRoot,
		"internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml")

	committed, body := readRawAuditBlock(t, committedPath)

	if committed.Model != AuditModelMulti {
		t.Errorf("committed workflow.audit.model = %q, want %q", committed.Model, AuditModelMulti)
	}
	if committed.Claude.Effort != DefaultClaudeAuditEffort {
		t.Errorf("committed claude pin effort = %q, want %q (DefaultClaudeAuditEffort)",
			committed.Claude.Effort, DefaultClaudeAuditEffort)
	}
	if committed.Claude.Model != DefaultClaudeAuditModel {
		t.Errorf("committed claude pin model = %q, want %q", committed.Claude.Model, DefaultClaudeAuditModel)
	}
	if committed.Codex.Model != "gpt-6.1-sol" || committed.Codex.Effort != "high" {
		t.Errorf("committed codex pin = {%s %s}, want {gpt-6.1-sol high}", committed.Codex.Model, committed.Codex.Effort)
	}
	if committed.GLM.Model != "glm-5.3" || committed.GLM.Effort != "max" {
		t.Errorf("committed glm pin = {%s %s}, want {glm-5.3 max}", committed.GLM.Model, committed.GLM.Effort)
	}
	if committed.Gates != nil {
		t.Errorf("committed workflow.audit carries a gates key %v; the model token resolves the gates", committed.Gates)
	}
	if strings.Contains(body, "claude-opus-5-5/medium") {
		t.Errorf("committed workflow.yaml header comment still names claude-opus-5-5/medium")
	}

	// The committed file loads through the config loader to the same values.
	cfg, err := NewLoader().Load(filepath.Join(repoRoot, ".moai"))
	if err != nil {
		t.Fatalf("Loader.Load(committed tree): %v", err)
	}
	if got := cfg.Workflow.Audit.Model; got != AuditModelMulti {
		t.Errorf("loaded Audit.Model = %q, want %q", got, AuditModelMulti)
	}
	if got := cfg.Workflow.Audit.Claude.Effort; got != DefaultClaudeAuditEffort {
		t.Errorf("loaded Audit.Claude.Effort = %q, want %q", got, DefaultClaudeAuditEffort)
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
