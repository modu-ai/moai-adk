package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeAuditTemplateSurfacesAndCatalogHash(t *testing.T) {
	templates, err := EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		".claude/rules/moai/core/moai-mcp-tools.md",
		".claude/rules/moai/core/moai-mcp-tools-catalogue.md",
		".claude/skills/moai-ref-cross-model-audit/SKILL.md",
		".claude/agents/moai/plan-auditor.md",
		".claude/agents/moai/sync-auditor.md",
		".codex/agents/moai/plan-auditor.toml",
		".codex/agents/moai/sync-auditor.toml",
		".moai/config/sections/workflow.yaml",
	} {
		data, err := fs.ReadFile(templates, path)
		if err != nil {
			t.Errorf("read %s: %v", path, err)
			continue
		}
		if !strings.Contains(string(data), "claude_audit") && !strings.Contains(string(data), "claude:") {
			t.Errorf("%s does not carry the Claude audit contract", path)
		}
	}

	var stored string
	for _, entry := range allCatalogEntries(loadCatalog(t)) {
		if entry.Name == "moai-ref-cross-model-audit" {
			stored = entry.Hash
			break
		}
	}
	if stored == "" {
		t.Fatal("moai-ref-cross-model-audit is absent from catalog.yaml")
	}
	computed, err := ComputeDirTreeHash(templates, ".claude/skills/moai-ref-cross-model-audit")
	if err != nil {
		t.Fatal(err)
	}
	if stored != computed {
		t.Fatalf("cross-model audit catalog hash = %s, computed %s", stored, computed)
	}

	for _, name := range []string{"plan-auditor.md", "sync-auditor.md"} {
		templatePath := ".claude/agents/moai/" + name
		templateData, err := fs.ReadFile(templates, templatePath)
		if err != nil {
			t.Fatal(err)
		}
		localData, err := os.ReadFile(filepath.Join("..", "..", templatePath))
		if err != nil {
			t.Fatal(err)
		}
		if claudeAuditDocSection(string(templateData)) != claudeAuditDocSection(string(localData)) {
			t.Errorf("%s Claude audit section differs between template and local carrier", name)
		}
	}
}

func claudeAuditDocSection(body string) string {
	const heading = "## MCP Audit Tools (cross-model second opinion)"
	start := strings.Index(body, heading)
	if start < 0 {
		return ""
	}
	rest := body[start+len(heading):]
	if end := strings.Index(rest, "\n## "); end >= 0 {
		rest = rest[:end]
	}
	return strings.TrimSpace(rest)
}
