package web

// codex_model_select_test.go — card t1980: the codex audit model pin renders
// as a select (the closed set grew to 4 ids), the newest id is selectable, and
// the validator still rejects non-member ids.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestCodexModelPinRendersAsSelectWithTerra is the render golden: the live
// schema field routes to a <select> whose options carry every closed-set id
// (gpt-6.1-terra included) plus the empty "unset" option — radio groups render
// no <select>, so a regression to the t1278 widget fails the first assertion.
func TestCodexModelPinRendersAsSelectWithTerra(t *testing.T) {
	f, ok := settings.Field("workflow.audit.codex.model")
	if !ok {
		t.Fatal("workflow.audit.codex.model missing from the schema")
	}
	html := mustRender(t, func(sb *strings.Builder) error {
		return schemaFieldWidget(pageView{}, f).Render(context.Background(), sb)
	})
	if !strings.Contains(html, "<select") {
		t.Errorf("workflow.audit.codex.model did not render as a select:\n%s", html)
	}
	if !strings.Contains(html, `value="gpt-6.1-terra"`) {
		t.Errorf("the select is missing the gpt-6.1-terra option (card t1980):\n%s", html)
	}
	if !strings.Contains(html, `value=""`) {
		t.Errorf("the select lost the empty 'unset' option:\n%s", html)
	}
	for _, id := range []string{"gpt-6.1-sol", "gpt-5.6-sol", "gpt-5.6"} {
		if !strings.Contains(html, `value="`+id+`"`) {
			t.Errorf("the select lost the %q option:\n%s", id, html)
		}
	}
}
