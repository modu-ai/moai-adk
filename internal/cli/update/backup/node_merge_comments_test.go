package backup

import (
	"strings"
	"testing"
)

// TestMergeCommentsSurviveRestore pins the comment rule of the 3-way merge
// (mergeCommentBlock). The user's comment lines survive a restore above a key,
// beside a key or value, after a nested mapping, and at the document footer, and
// each branch of the template-versus-user rule holds. The assertions read the
// merged bytes, so they bind to what the user will see.
func TestMergeCommentsSurviveRestore(t *testing.T) {
	t.Parallel()

	const tpl = "gate:\n" +
		"  enabled: true\n" +
		"  skip_tests: false\n" +
		"  # Disable steps\n" +
		"  disabled_steps: {}\n" +
		"other:\n" +
		"  x: 1\n"
	const tplNoLine = "gate:\n" +
		"  enabled: true\n" +
		"  skip_tests: false\n" +
		"  disabled_steps: {}\n" +
		"other:\n" +
		"  x: 1\n"

	cases := []struct {
		name     string
		newYAML  string
		userYAML string
		baseYAML string
		keepOnce []string
		dropped  []string
	}{
		{
			name:     "user note above a template key without a template comment",
			newYAML:  tplNoLine,
			userYAML: strings.Replace(tplNoLine, "  skip_tests: false\n", "  # NOTE-ABOVE\n  skip_tests: true\n", 1),
			baseYAML: tplNoLine,
			keepOnce: []string{"# NOTE-ABOVE"},
		},
		{
			name:     "user note above a template key that carries a template comment",
			newYAML:  tpl,
			userYAML: strings.Replace(tpl, "  # Disable steps\n", "  # NOTE-ABOVE\n  # Disable steps\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-ABOVE", "# Disable steps"},
		},
		{
			name:     "trailing comment survives when the template value is unchanged",
			newYAML:  tpl,
			userYAML: strings.Replace(tpl, "  skip_tests: false\n", "  skip_tests: false  # NOTE-TRAIL\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-TRAIL"},
		},
		{
			name:     "standalone comment after a nested mapping survives",
			newYAML:  tpl,
			userYAML: strings.Replace(tpl, "other:\n", "  # NOTE-STANDALONE\nother:\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-STANDALONE"},
		},
		{
			name:     "user comment at the end of the file survives",
			newYAML:  tpl,
			userYAML: tpl + "\n# NOTE-EOF\n",
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-EOF"},
		},
		{
			name:     "template rewording of an untouched comment lands",
			newYAML:  strings.Replace(tpl, "# Disable steps", "# Disable steps (v2)", 1),
			userYAML: tpl,
			baseYAML: tpl,
			keepOnce: []string{"# Disable steps (v2)"},
			dropped:  []string{"# Disable steps\n"},
		},
		{
			name:     "user deletion of a template comment line stands",
			newYAML:  tpl,
			userYAML: tplNoLine,
			baseYAML: tpl,
			dropped:  []string{"# Disable steps"},
		},
		{
			name:     "template removal of an untouched comment line lands",
			newYAML:  tplNoLine,
			userYAML: tpl,
			baseYAML: tpl,
			dropped:  []string{"# Disable steps"},
		},
		{
			name:     "user line survives a template removal of the line beside it",
			newYAML:  tplNoLine,
			userYAML: strings.Replace(tpl, "  # Disable steps\n", "  # NOTE-KEPT\n  # Disable steps\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-KEPT"},
			dropped:  []string{"# Disable steps"},
		},
		{
			name:     "template-added comment line is kept beside the user's line",
			newYAML:  strings.Replace(tpl, "  # Disable steps\n", "  # Line A\n  # Disable steps\n", 1),
			userYAML: strings.Replace(tpl, "  disabled_steps: {}\n", "  # NOTE-MINE\n  disabled_steps: {}\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# Line A", "# Disable steps", "# NOTE-MINE"},
		},
		{
			name:     "user document header survives",
			newYAML:  tpl,
			userYAML: "# NOTE-HEADER\n" + tpl,
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-HEADER"},
		},
		{
			name:     "user comment directly after the last key survives",
			newYAML:  tpl,
			userYAML: tpl + "# NOTE-FOOT\n",
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-FOOT"},
		},
		{
			name:     "trailing comment on a mapping-valued key survives",
			newYAML:  tpl,
			userYAML: strings.Replace(tpl, "gate:\n", "gate:  # NOTE-KEY\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-KEY"},
		},
		{
			name:     "trailing comment on a flow value survives",
			newYAML:  tpl,
			userYAML: strings.Replace(tpl, "  disabled_steps: {}\n", "  disabled_steps: {}  # NOTE-FLOW\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-FLOW"},
		},
		{
			name:     "user comment above a user-added key survives",
			newYAML:  tpl,
			userYAML: strings.Replace(tpl, "other:\n", "  # NOTE-USER\n  user_key: v\nother:\n", 1),
			baseYAML: tpl,
			keepOnce: []string{"# NOTE-USER"},
		},
		{
			name:     "untouched file keeps each comment line exactly once",
			newYAML:  tpl,
			userYAML: tpl,
			baseYAML: tpl,
			keepOnce: []string{"# Disable steps"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			merged, _, err := MergeYAML3WayRetained([]byte(tc.newYAML), []byte(tc.userYAML), []byte(tc.baseYAML))
			if err != nil {
				t.Fatalf("MergeYAML3WayRetained: %v", err)
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
