package web

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

var errChipRead = errors.New("read failed")

// The chips reached only /settings before: shell.templ looped over ShellVM.Ctx
// and settings_shell.go filled it, but the four redesign screens never did. These
// cover the three branches of the filler, including the two that render nothing —
// an absent chip is the intended outcome there, not a failure to notice.
func TestScreenCtxChips(t *testing.T) {
	t.Parallel()

	okPrefs := func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{
			ConversationLang: "ko",
			Model:            "opus[1m]",
			EffortLevel:      "high",
		}, nil
	}

	tests := []struct {
		name    string
		prefs   func(string) (profile.ProfilePreferences, error)
		project func(string) (string, string, error)
		want    []KV
	}{
		{
			name:    "all four sources readable",
			prefs:   okPrefs,
			project: func(string) (string, string, error) { return "tdd", "angular", nil },
			want: []KV{
				{K: "lang", V: "ko"},
				{K: "model", V: "opus[1m]"},
				{K: "effort", V: "high"},
				{K: "dev", V: "tdd"},
			},
		},
		{
			// A read failure yields no chip rather than a failed page: these are
			// read-only status screens, and losing the whole view over a missing
			// chip trades a small gap for a blank screen.
			name:    "preferences unreadable drops its three chips",
			prefs:   func(string) (profile.ProfilePreferences, error) { return profile.ProfilePreferences{}, errChipRead },
			project: func(string) (string, string, error) { return "tdd", "", nil },
			want:    []KV{{K: "dev", V: "tdd"}},
		},
		{
			// An empty value is not "unset" — rendering an empty chip would read
			// as a configured blank, so the chip is omitted entirely.
			name: "empty values are omitted, not rendered blank",
			prefs: func(string) (profile.ProfilePreferences, error) {
				return profile.ProfilePreferences{Model: "opus[1m]"}, nil
			},
			project: func(string) (string, string, error) { return "", "", nil },
			want:    []KV{{K: "model", V: "opus[1m]"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := &app{readPreferences: tc.prefs, readProjectConfig: tc.project}

			got := a.screenCtxChips("default")

			if len(got) != len(tc.want) {
				t.Fatalf("chip count = %d (%v), want %d (%v)", len(got), got, len(tc.want), tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("chip[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// i18nSlug couples a heading's wording to its dictionary key, so the cases that
// would silently orphan a translation are worth pinning.
func TestI18nSlug(t *testing.T) {
	t.Parallel()

	tests := []struct{ in, want string }{
		{"In-progress SPECs", "in-progress-specs"},
		{"Needs attention", "needs-attention"},
		{"Must-fix drift", "must-fix-drift"},
		{"Close debt", "close-debt"},
		{"Sessions", "sessions"},
		// Runs of punctuation and space collapse to one dash, so a heading
		// reflowed with an extra space keeps the key it already had.
		{"Must fix  drift", "must-fix-drift"},
		{"", ""},
		{"---", ""},
	}

	for _, tc := range tests {
		if got := i18nSlug(tc.in); got != tc.want {
			t.Errorf("i18nSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
