package wizard

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// TestModelPolicyDescsAgreeWithProfileMatrix guards the `moai init` tier
// descriptions against drifting away from the matrix they claim to mirror.
//
// WHY THIS EXISTS. The option block in questions.go carries a comment saying
// "Keep these in sync with that matrix, not with a marketing summary" — and the
// descriptions had drifted anyway, in all four locales at once: they advertised
// "Fable 5 (low)" and "Opus 4.8 (xhigh~low)" while defaultProfileMatrix holds
// zero fable cells, zero haiku cells, no xhigh cell, and resolves opus to
// claude-opus-5. A comment is not a guard; this is.
//
// The banned tokens are DERIVED from the matrix rather than listed by hand, so
// the day a model or effort genuinely enters the matrix, this test stops
// banning it instead of blocking the change.
func TestModelPolicyDescsAgreeWithProfileMatrix(t *testing.T) {
	t.Parallel()

	matrix := template.DefaultProfileMatrix()
	if len(matrix) == 0 {
		t.Fatal("DefaultProfileMatrix is empty: every assertion below would be " +
			"vacuous, since the derived token sets would be empty too")
	}

	// What the matrix actually contains, across every agent and column.
	presentModels := map[string]bool{}
	presentEfforts := map[string]bool{}
	for _, byProfile := range matrix {
		for _, cell := range byProfile {
			presentModels[cell.Model] = true
			presentEfforts[cell.Effort] = true
		}
	}

	// Tokens a description must not claim, because the matrix never assigns
	// them. Derived, not hardcoded — see the doc comment.
	var banned []string
	for _, m := range []string{"fable", "haiku"} {
		if !presentModels[m] {
			banned = append(banned, m)
		}
	}
	for _, e := range []string{"xhigh"} {
		if !presentEfforts[e] {
			banned = append(banned, e)
		}
	}
	if len(banned) == 0 {
		t.Fatal("no banned token survived derivation: the matrix now assigns " +
			"every model and effort this test screens for, so the ban assertion " +
			"below can no longer fail and must be re-derived")
	}

	// The canonical opus id is the version the descriptions must name. Taking
	// the marketing version from the id keeps the two in lockstep: bumping the
	// alias table to a future Opus fails this test until the copy follows.
	opusID := template.ModelAliasCanonicalID("opus") // e.g. "claude-opus-5"
	wantVersion := strings.TrimPrefix(opusID, "claude-opus-")
	if wantVersion == "" || wantVersion == opusID {
		t.Fatalf("cannot derive an opus version from canonical id %q; the "+
			"version assertion below would be vacuous", opusID)
	}

	src := QuestionByID(DefaultQuestions(t.TempDir()), "model_policy")
	if src == nil {
		t.Fatal("model_policy question not found")
	}

	// English lives inline in questions.go; the other three come from the
	// translation store. Both paths are checked — the drift hit all four.
	locales := append([]string{"en"}, localizableLocales...)
	for _, locale := range locales {
		opts := src.Options
		if locale != "en" {
			opts = GetLocalizedQuestion(src, locale).Options
		}
		if len(opts) != 3 {
			t.Fatalf("locale %q: %d options, want 3", locale, len(opts))
		}

		for _, opt := range opts {
			desc := strings.ToLower(opt.Desc)

			for _, tok := range banned {
				if strings.Contains(desc, tok) {
					t.Errorf("locale %q option %q: description names %q, which "+
						"defaultProfileMatrix never assigns\n  desc: %s",
						locale, opt.Value, tok, opt.Desc)
				}
			}

			if !strings.Contains(opt.Desc, "Opus "+wantVersion) {
				t.Errorf("locale %q option %q: description does not name "+
					"\"Opus %s\" (derived from canonical id %q)\n  desc: %s",
					locale, opt.Value, wantVersion, opusID, opt.Desc)
			}
		}
	}
}
