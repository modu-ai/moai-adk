package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestMXRawViewRendered verifies SPEC-WEB-CONSOLE-014 M4 (REQ-WC14-030,
// AC-WC14-030b): the GET / response renders the mx raw-view group. mx is a
// raw-only section (0 editable fields — mx stays RouteExcluded), so the render
// path exercises the generic fieldsetSchemaSection with only RawViewBlocks:
// the section header (sec.mx.title), the mx panel container (data-panel="mx"),
// and the two raw-block code chips (mx.danger_categories / mx.test_paths) are
// present in the rendered body.
func TestMXRawViewRendered(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	for _, want := range []string{
		`data-i18n="sec.mx.title"`,
		`data-panel="mx"`,
		`mx.danger_categories`,
		`mx.test_paths`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing mx raw-view marker %q", want)
		}
	}
}
