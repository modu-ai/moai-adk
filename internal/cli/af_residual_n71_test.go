package cli

// N71 (audit a36ea8c368, residual classification): `moai af add` is an input boundary and accepts any card id. The id goes
// into the trunk commit message of the landing (merge(<id>): ...), into notes and into log file names. A newline or an escape
// sequence in an id reaches the commit message verbatim, a path separator or `..` reaches the name of a log, and the length is
// unbounded. The expectation: such an id is refused by the verb, the ledger unchanged.

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAFResidual_N71AnIdThatIsNotAPlainNameIsRefusedByAdd(t *testing.T) {
	afSkipWithoutSh(t)
	root := t.TempDir()
	p, _ := afResRepo(t, filepath.Join(root, "P"))
	t.Setenv("CLAUDE_PROJECT_DIR", p)
	for _, tc := range []struct{ name, id string }{
		{"a newline", "new\nline"},
		{"an escape sequence", "t1\x1b[2J"},
		{"parent directories", "../../../evil"},
		{"a path separator", "a/b"},
		{"5000 characters", strings.Repeat("x", 5000)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := filepath.Join(root, "central", strings.ReplaceAll(tc.name, " ", "-")+".db")
			text, err := afResVerb(db, "add", tc.id, "--bind-repo", p)
			stages := afResStages(t, db)
			t.Logf("add %.20q: err=%v cards in the ledger: %d bytes (%.100s)", tc.id, err, len(stages), text)
			if err == nil || stages != "" {
				t.Errorf("an id with %s was accepted: err=%v, cards in the ledger: %.40q (%d bytes)", tc.name, err, stages, len(stages))
			}
		})
	}
}
