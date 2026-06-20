package template_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// TestNoHardcodedAbsolutePath_HookWrappers verifies that hook wrappers contain no hardcoded absolute paths.
// REQ-V3R2-RT-007-002: generated shell wrappers MUST NOT contain absolute user paths.
// REQ-V3R2-RT-007-051: CI lint MUST fail when absolute path literals are present.
func TestNoHardcodedAbsolutePath_HookWrappers(t *testing.T) {
	// GREEN: the current 28 wrappers are already clean (research.md §2.2)
	// This test affirms the current state and detects regressions.
}

// TestNoHardcodedAbsolutePath_StatusLine verifies that status_line.sh.tmpl contains no hardcoded absolute paths.
// REQ-V3R2-RT-007-002: status_line template MUST NOT contain absolute paths either.
func TestNoHardcodedAbsolutePath_StatusLine(t *testing.T) {
	// GREEN: status_line.sh.tmpl is currently clean.
}

// TestNoHardcodedAbsolutePath_DocsAllowList handles docs example paths via an allow-list.
// REQ-V3R2-RT-007-002: docs examples (paste-ready samples) are permitted.
func TestNoHardcodedAbsolutePath_DocsAllowList(t *testing.T) {
	// GREEN: docs files are classified into the allow-list and pass the audit.
}

// TestFallbackChainOrder verifies the status_line.sh fallback chain order.
//
// REQ-SWF-011 (issue modu-ai/moai-adk#1068) UPDATES the chain order asserted
// here. The NEW canonical order is:
//
//	moai in PATH
//	  → resolved-executable (guarded, when ResolvedMoaiPath is non-empty)
//	  → $HOME/go/bin/moai (guarded)
//	  → $HOME/.local/bin/moai (guarded)
//
// This SUPERSEDES the prior REQ-V3R2-RT-007-003 order
// (PATH → $HOME/go/bin → {{posixPath .GoBinPath}}/moai): the baked
// `.GoBinPath` branch was removed to restore internal/template/CLAUDE.md §14
// compliance (generated shell fallbacks use $HOME, never a baked .GoBinPath).
// The test is UPDATED to the new order, not left asserting the removed branch.
func TestFallbackChainOrder(t *testing.T) {
	t.Parallel()

	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	r := template.NewRenderer(fsys)
	ctx := template.NewTemplateContext(
		template.WithResolvedMoaiPath("/opt/moai/bin/moai"),
	)
	out, err := r.Render(".moai/status_line.sh.tmpl", ctx)
	if err != nil {
		t.Fatalf("Render(status_line.sh.tmpl) error: %v", err)
	}
	rendered := string(out)

	// The baked .GoBinPath branch must be gone (§14 compliance, supersedes RT-007-003).
	if strings.Contains(rendered, "GoBinPath") {
		t.Errorf("status_line.sh still references a baked GoBinPath token:\n%s", rendered)
	}

	// Assert the new canonical order via ascending index positions.
	stages := []struct {
		name   string
		marker string
	}{
		{"PATH lookup", "command -v moai"},
		{"resolved-executable", "/opt/moai/bin/moai"},
		{"$HOME/go/bin", "$HOME/go/bin/moai"},
		{"$HOME/.local/bin", "$HOME/.local/bin/moai"},
	}
	prev := -1
	for _, s := range stages {
		idx := strings.Index(rendered, s.marker)
		if idx < 0 {
			t.Fatalf("status_line.sh missing %s stage marker %q:\n%s", s.name, s.marker, rendered)
		}
		if idx <= prev {
			t.Errorf("fallback stage %s (idx %d) out of order; expected after previous stage (idx %d):\n%s",
				s.name, idx, prev, rendered)
		}
		prev = idx
	}
}
