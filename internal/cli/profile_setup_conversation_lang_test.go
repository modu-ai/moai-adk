package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestInitialProfileResult_ConversationLang pins card t2010 / PJ-13: the stored
// conversation language seeds the profile wizard's initial value. Only a truly
// empty value falls back to en.
func TestInitialProfileResult_ConversationLang(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stored string
		want   string
	}{
		{"ko is kept", "ko", "ko"},
		{"ja is kept", "ja", "ja"},
		{"zh is kept", "zh", "zh"},
		{"en is kept", "en", "en"},
		{"empty falls back to en", "", "en"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := initialProfileResult(profile.ProfilePreferences{ConversationLang: tc.stored})
			if got.ConversationLang != tc.want {
				t.Errorf("stored ConversationLang %q -> initial %q, want %q", tc.stored, got.ConversationLang, tc.want)
			}
		})
	}
}
