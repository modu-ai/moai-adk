package cli

// N70 (audit a36ea8c368, residual classification): --bind-repo declares the repository the ledger belongs to. From a project
// directory that is no git repository the declaration is silently dropped (the verb exits 0 and the ledger stays unbound), and
// a declaration that names no repository at all is dropped the same way. The expectation: a declaration that cannot be
// honoured is refused, the ledger unchanged, instead of being ignored.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAFResidual_N70ADeclarationThatCannotBeHonouredIsRefusedNotDropped(t *testing.T) {
	afSkipWithoutSh(t)
	root := t.TempDir()
	p, _ := afResRepo(t, filepath.Join(root, "P"))
	nonrepo := filepath.Join(root, "nonrepo")
	if err := os.MkdirAll(nonrepo, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, declared string }{
		{"a path that is a repository, from a directory that is no repository", p},
		{"a path that is no repository, from a directory that is no repository", filepath.Join(root, "garbage")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := filepath.Join(root, "central", "ledger-"+filepath.Base(tc.declared)+".db")
			afResLedger(t, db)
			t.Setenv("CLAUDE_PROJECT_DIR", nonrepo)
			text, err := afResVerb(db, "reap", "--bind-repo", tc.declared)
			bound := afResBound(t, db)
			t.Logf("reap --bind-repo %s from no repository: err=%v bound=%q (%.160s)", filepath.Base(tc.declared), err, bound, text)
			if err == nil && bound == "" {
				t.Errorf("the declaration %q was dropped without a word: exit 0 and the ledger is still not bound", tc.declared)
			}
		})
	}
}
