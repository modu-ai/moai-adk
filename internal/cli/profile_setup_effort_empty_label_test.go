package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// effortEmptyOptionLabel returns the wizard's empty effort_level option label
// for one locale, failing the test when the wizard offers no empty option.
func effortEmptyOptionLabel(t *testing.T, loc string) string {
	t.Helper()
	opts := schemaSelectOptions(getProfileText(loc), "effort_level", true)
	if len(opts) == 0 || opts[0].Value != "" {
		t.Fatalf("%s: wizard offers no empty effort_level option", loc)
	}
	return opts[0].Label
}

// TestEffortEmptyLabelLocalized: the ko/ja/zh wizard renders the empty effort
// option in its own language rather than the English schema literal, and each
// localized label still names both launch fallbacks — the model policy first,
// then Claude Code's own default.
func TestEffortEmptyLabelLocalized(t *testing.T) {
	t.Parallel()
	english := settings.EmptyLabelFor("effort_level")
	policyWord := map[string]string{"ko": "모델 정책", "ja": "モデルポリシー", "zh": "模型策略"}
	for _, loc := range []string{"ko", "ja", "zh"} {
		label := effortEmptyOptionLabel(t, loc)
		if label == english {
			t.Errorf("%s: empty effort label is the untranslated English schema literal %q", loc, label)
			continue
		}
		for _, want := range []string{policyWord[loc], "Claude Code"} {
			if !strings.Contains(label, want) {
				t.Errorf("%s: empty effort label %q does not name %q", loc, label, want)
			}
		}
	}
}

// TestEffortEmptyLabelCarriesRuntimeDefaultFact is the wizard half of the
// runtime-default drift guard (the console half is
// TestRuntimeDefaultI18nCarriesEffortFact in internal/web): every locale's
// empty effort label states the Claude Code default effort and the model it
// applies to, taken from the settings constants. When the default model or
// effort changes, the constant edit makes this test name every label that
// still carries the old fact.
func TestEffortEmptyLabelCarriesRuntimeDefaultFact(t *testing.T) {
	t.Parallel()
	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		label := effortEmptyOptionLabel(t, loc)
		for _, want := range []string{settings.RuntimeDefaultEffort, settings.RuntimeDefaultEffortModel} {
			if !strings.Contains(label, want) {
				t.Errorf("%s: empty effort label %q does not carry %q", loc, label, want)
			}
		}
	}
}
