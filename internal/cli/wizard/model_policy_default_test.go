package wizard

// SPEC-CLI-WIZARD-RESTRUCTURE-001 M2 — model_policy default High -> Medium
// (AC-WIZ-005, REQ-WIZ-008/011). The recommended marker moves with the default;
// ModelPolicyHigh survives as a selectable OPTION, only the DEFAULT moves.

import (
	"strings"
	"testing"
)

// recommendedMarkers are the per-locale "(Recommended)" renderings carried by
// the model_policy option labels. English lives inline in questions.go; the
// other three live in the translation table.
var recommendedMarkers = map[string]string{
	"en": "(Recommended)",
	"ko": "(권장)",
	"ja": "(推奨)",
	"zh": "(推荐)",
}

// TestModelPolicyDefaultIsMedium pins AC-WIZ-005: the pre-selected default is
// Medium, the Medium option carries the recommended marker, Max does not, and
// "high" remains a selectable option value.
func TestModelPolicyDefaultIsMedium(t *testing.T) {
	t.Parallel()
	q := QuestionByID(DefaultQuestions(t.TempDir()), "model_policy")
	if q == nil {
		t.Fatal("model_policy question not found")
	}

	if q.Default != "medium" {
		t.Errorf("model_policy default = %q, want %q", q.Default, "medium")
	}

	marker := recommendedMarkers["en"]
	var marked []string
	for _, opt := range q.Options {
		if strings.Contains(opt.Label, marker) {
			marked = append(marked, opt.Value)
		}
	}
	if len(marked) != 1 || marked[0] != "medium" {
		t.Errorf("recommended marker %q sits on option value(s) %v, want exactly [medium]", marker, marked)
	}

	// The Max/High tier stays selectable — only the DEFAULT moved.
	values := make([]string, len(q.Options))
	for i, opt := range q.Options {
		values[i] = opt.Value
	}
	if len(values) != 3 || values[0] != "high" || values[1] != "medium" || values[2] != "low" {
		t.Errorf("model_policy option values = %v, want [high medium low]", values)
	}
}

// TestModelPolicyRecommendedMarkerPerLocale pins the C12 half of AC-WIZ-005:
// the marker moved from Max to Medium in ko/ja/zh too. Each locale is checked
// explicitly (never with a shared halfwidth-punctuation pattern) because the
// CJK locales use their own native marker text.
func TestModelPolicyRecommendedMarkerPerLocale(t *testing.T) {
	t.Parallel()
	src := QuestionByID(DefaultQuestions(t.TempDir()), "model_policy")
	if src == nil {
		t.Fatal("model_policy question not found")
	}

	for _, locale := range localizableLocales {
		marker, ok := recommendedMarkers[locale]
		if !ok {
			t.Fatalf("no recommended marker recorded for locale %q", locale)
		}
		localized := GetLocalizedQuestion(src, locale)
		if len(localized.Options) != 3 {
			t.Fatalf("locale %q: %d localized options, want 3", locale, len(localized.Options))
		}
		var marked []string
		for _, opt := range localized.Options {
			if strings.Contains(opt.Label, marker) {
				marked = append(marked, opt.Value)
			}
		}
		if len(marked) != 1 || marked[0] != "medium" {
			t.Errorf("locale %q: marker %q sits on %v, want exactly [medium]", locale, marker, marked)
		}
	}
}
