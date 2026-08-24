package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadWorkflowAuditSection (SPEC-V3R6-AUDIT-MODEL-PIN-001 M1 / AC-AMP-001)
// verifies the workflow-section load helper: populated pins load verbatim,
// an absent file returns the zero AuditConfig with NO error, and an unparseable
// file returns an error the audit resolvers treat as an absent pin (N3
// fail-open — tested at the resolver level in mcp_codex_test.go /
// mcp_glm_test.go).
func TestLoadWorkflowAuditSection(t *testing.T) {
	t.Parallel()

	writeWorkflow := func(t *testing.T, body string) string {
		t.Helper()
		dir := t.TempDir()
		sections := filepath.Join(dir, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}

	t.Run("populated pins load verbatim", func(t *testing.T) {
		dir := writeWorkflow(t, "workflow:\n"+
			"    audit:\n"+
			"        model: multi\n"+
			"        codex:\n"+
			"            model: gpt-5.6-sol\n"+
			"            effort: high\n"+
			"        glm:\n"+
			"            model: glm-5.3\n"+
			"            effort: max\n")
		got, err := loadWorkflowAuditSection(dir)
		if err != nil {
			t.Fatalf("loadWorkflowAuditSection: %v", err)
		}
		if got.Model != "multi" {
			t.Errorf("Model: got %q, want multi", got.Model)
		}
		if got.Codex.Model != "gpt-5.6-sol" || got.Codex.Effort != "high" {
			t.Errorf("Codex pin: got %+v, want {gpt-5.6-sol high}", got.Codex)
		}
		if got.GLM.Model != "glm-5.3" || got.GLM.Effort != "max" {
			t.Errorf("GLM pin: got %+v, want {glm-5.3 max}", got.GLM)
		}
	})

	t.Run("absent file is zero value with no error", func(t *testing.T) {
		got, err := loadWorkflowAuditSection(t.TempDir())
		if err != nil {
			t.Fatalf("absent workflow.yaml must not error: %v", err)
		}
		if got.Model != "" || got.Codex.Model != "" || got.GLM.Model != "" {
			t.Errorf("absent file must yield zero AuditConfig, got %+v", got)
		}
	})

	t.Run("file without audit block is zero pins", func(t *testing.T) {
		dir := writeWorkflow(t, "workflow:\n    default_mode: \"\"\n")
		got, err := loadWorkflowAuditSection(dir)
		if err != nil {
			t.Fatalf("audit-less workflow.yaml must not error: %v", err)
		}
		if got.Codex.Model != "" || got.GLM.Model != "" {
			t.Errorf("audit-less file must yield zero pins, got %+v", got)
		}
	})

	t.Run("unparseable file errors so callers fail open", func(t *testing.T) {
		dir := writeWorkflow(t, "workflow: [not: a: mapping\n")
		if _, err := loadWorkflowAuditSection(dir); err == nil {
			t.Fatal("unparseable workflow.yaml must return an error (the resolver then fails open to the legacy SSOT path)")
		}
	})
}
