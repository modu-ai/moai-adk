// Package template — retired_wrappers_test.go
// SPEC-V3R2-MIG-002 T-MIG002-04
// Verifies that the 4 retired-event wrapper templates ship with the
// RETIRE-OBS-ONLY inline comment documenting their forward-compat contract.
package template

import (
	"io/fs"
	"strings"
	"testing"
)

// retiredWrapperTmpls lists the 4 templates that are kept for forward-compat
// with the system.yaml hook.observability_events opt-in path (SPEC-V3R2-RT-006),
// but are NOT registered in settings.json.
var retiredWrapperTmpls = []string{
	".claude/hooks/moai/handle-notification.sh.tmpl",
	".claude/hooks/moai/handle-elicitation.sh.tmpl",
	".claude/hooks/moai/handle-elicitation-result.sh.tmpl",
	".claude/hooks/moai/handle-task-created.sh.tmpl",
}

// TestRenderRetiredWrappers asserts that each of the 4 retired-wrapper templates:
// 1. Exists in the embedded template filesystem.
// 2. Contains the "RETIRE-OBS-ONLY" inline comment documenting the contract.
//
// SPEC-V3R2-MIG-002 T-MIG002-04 (M1 RED → M2 GREEN after T-MIG002-10 adds the comment).
// AC-MIG002-A5 (indirectly — presence guarantees observability opt-in path is preserved).
func TestRenderRetiredWrappers(t *testing.T) {
	// Use the real embedded FS so the test reflects the shipped artifact.
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}

	for _, tmplPath := range retiredWrapperTmpls {
		t.Run(tmplPath, func(t *testing.T) {
			data, err := fs.ReadFile(fsys, tmplPath)
			if err != nil {
				t.Fatalf("retired wrapper template not found: %s (%v)", tmplPath, err)
			}

			content := string(data)

			if !strings.Contains(content, "RETIRE-OBS-ONLY") {
				t.Errorf("%s: missing RETIRE-OBS-ONLY inline comment — M2.2 decision (keep for observability opt-in) must be documented in the template (SPEC-V3R2-MIG-002 T-MIG002-10)", tmplPath)
			}
		})
	}
}

// TestRetiredWrapperTemplates_Renderable verifies each retired wrapper renders
// without error under a representative TemplateContext.
func TestRetiredWrapperTemplates_Renderable(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	r := NewRenderer(fsys)

	ctx := struct {
		GoBinPath string
		HomeDir   string
	}{
		GoBinPath: "/tmp/go/bin",
		HomeDir:   "/tmp/home",
	}

	for _, tmplPath := range retiredWrapperTmpls {
		t.Run(tmplPath, func(t *testing.T) {
			_, err := r.Render(tmplPath, ctx)
			if err != nil {
				t.Errorf("%s: render failed: %v (retired wrappers must remain renderable — SPEC-V3R2-MIG-002 M2.2 Option Keep)", tmplPath, err)
			}
		})
	}
}
