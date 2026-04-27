package template

import (
	"context"
	"testing"
)

// TestLLMPanelTemplate_Integration validates the actual llm-panel.yml.tmpl
// from the embedded filesystem can be parsed and rendered.
func TestLLMPanelTemplate_Integration(t *testing.T) {
	// Get the embedded filesystem
	embeddedFS, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("failed to get embedded templates: %v", err)
	}

	r := NewRenderer(embeddedFS)
	d := NewDeployerWithRenderer(embeddedFS, r)

	ctx := context.Background()
	tmplCtx := NewTemplateContext(WithProject("test-project", "/tmp/test"))

	// ValidateAll should succeed with the fixed template
	err = d.ValidateAll(ctx, tmplCtx)
	if err != nil {
		t.Fatalf("expected llm-panel.yml.tmpl to validate successfully, got: %v", err)
	}
}
