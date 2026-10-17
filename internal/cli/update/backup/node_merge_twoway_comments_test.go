package backup

import (
	"strings"
	"testing"
)

// TestMergeYAMLDeepCommentsOldWins pins the comment rule of the 2-way fallback
// (deepMergeMapsTo, reached through MergeYAMLDeep). The fallback has no base, so
// a user's comment cannot be told apart from a template comment: per comment
// slot the user's block wins, and the template's block fills only a slot the
// user left empty. Values keep the 2-way value policy, the user's value wins,
// except system fields, which take the template's value.
func TestMergeYAMLDeepCommentsOldWins(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		newYAML  string
		userYAML string
		keepOnce []string
		dropped  []string
	}{
		{
			name:     "user head comment beats the template's on a shared key",
			newYAML:  "gate:\n  # TPL-HEAD\n  skip_tests: false\n",
			userYAML: "gate:\n  # USER-HEAD\n  skip_tests: true\n",
			keepOnce: []string{"# USER-HEAD", "skip_tests: true"},
			dropped:  []string{"# TPL-HEAD"},
		},
		{
			name:     "template head comment fills a shared key the user left bare",
			newYAML:  "gate:\n  # TPL-HEAD\n  skip_tests: false\n",
			userYAML: "gate:\n  skip_tests: true\n",
			keepOnce: []string{"# TPL-HEAD", "skip_tests: true"},
		},
		{
			name:     "user line comment survives a system field the template changed",
			newYAML:  "system:\n  template_version: \"2.0.0\"\n",
			userYAML: "system:\n  template_version: \"1.0.0\"  # USER-LINE\n",
			keepOnce: []string{"# USER-LINE", "template_version: \"2.0.0\""},
			dropped:  []string{"1.0.0"},
		},
		{
			name:     "user comment above a key the template does not carry is kept",
			newYAML:  "gate:\n  skip_tests: false\n",
			userYAML: "gate:\n  # USER-ONLY\n  user_key: v\n  skip_tests: false\n",
			keepOnce: []string{"# USER-ONLY", "user_key: v"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			merged, err := MergeYAMLDeep([]byte(tc.newYAML), []byte(tc.userYAML))
			if err != nil {
				t.Fatalf("MergeYAMLDeep: %v", err)
			}
			out := string(merged)
			for _, comment := range tc.keepOnce {
				if n := strings.Count(out, comment); n != 1 {
					t.Errorf("comment %q appears %d times, want exactly 1; merged:\n%s", comment, n, out)
				}
			}
			for _, comment := range tc.dropped {
				if strings.Contains(out, comment) {
					t.Errorf("comment %q should be gone after the merge; merged:\n%s", comment, out)
				}
			}
		})
	}
}
