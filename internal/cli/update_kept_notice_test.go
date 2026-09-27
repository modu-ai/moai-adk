package cli

// Card t1216 (sync-audit F3): the kept-over-default notice lists every value
// that differs from the current template, which includes the user's own
// customizations. Its wording must not assert that a template default changed;
// it may only say what happens where one did.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/tui"
)

func TestRenderRetainedKeyAdvisory_KeptNoticeIsConditional(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderRetainedKeyAdvisory(&buf, []backup.RetainedKeyRef{
		{Section: "git-strategy.yaml", Key: "git_strategy.manual.workflow", KeptOverDefault: true},
	}, false, tui.LightTheme())
	got := buf.String()

	if strings.Contains(got, "a changed default was not applied") {
		t.Errorf("notice asserts a default changed, but the list may hold only user-set keys:\n%s", got)
	}
	for _, want := range []string{
		"where a default itself changed",
		"git-strategy.yaml: git_strategy.manual.workflow",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("notice lacks %q:\n%s", want, got)
		}
	}
}
