package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// TestHandoffInjectRender pins the goal-arming guidance line the auto-resume
// injection renders, per locale. The emitted surface is the programmatic
// `/moai goal` directive: the native `/goal` command is a HUMAN-ONLY TUI
// command the model cannot set on the user's behalf, so it is not an emission
// surface. Each locale is a named subtest so a regression names the locale
// that drifted rather than failing the whole renderer opaquely.
func TestHandoffInjectRender(t *testing.T) {
	const wantPrefix = "  • /moai goal "
	const retiredPrefix = "  • /goal "

	for _, locale := range []string{"ko", "ja", "zh", "en"} {
		t.Run(locale, func(t *testing.T) {
			// The locale strings table carries the prefix.
			if got := handoffLocaleStrings(locale).goalPrefix; got != wantPrefix {
				t.Errorf("handoffLocaleStrings(%q).goalPrefix = %q, want %q", locale, got, wantPrefix)
			}

			// The prefix is reachable through the render path, not merely
			// present in the table (a table entry no renderer emits is inert).
			rec := &handoff.PendingRecord{
				ConversationLanguage: locale,
				Directives:           handoff.Directives{Goal: "the target test suite is green"},
				Body:                 "resume body",
			}
			rendered := renderHandoffContext(rec)

			if want := wantPrefix + "the target test suite is green"; !strings.Contains(rendered, want) {
				t.Errorf("rendered context does not carry %q\n--- rendered ---\n%s", want, rendered)
			}
			if strings.Contains(rendered, retiredPrefix) {
				t.Errorf("rendered context still emits the retired native-goal token %q\n--- rendered ---\n%s", retiredPrefix, rendered)
			}
		})
	}
}
