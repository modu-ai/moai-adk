package cli

// N55 (audits 1c05083f71 and a36ea8c368, residual classification): the help says "the first verb that writes it binds", but
// afRun binds before the verb's own work, so a verb that is refused, or that writes nothing, binds the ledger to the
// repository it ran in all the same. The binding is never undone. The expectation: a verb the ledger refuses (no such card)
// has written nothing, so the ledger is still not bound.

import (
	"path/filepath"
	"testing"
)

func TestAFResidual_N55ARefusedVerbThatWroteNothingDoesNotBindTheLedger(t *testing.T) {
	afSkipWithoutSh(t)
	root := t.TempDir()
	o, _ := afResRepo(t, filepath.Join(root, "O"))
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"release of a card that does not exist", []string{"release", "nosuch", "--epoch", "1"}},
		{"resume of a card that does not exist", []string{"resume", "nosuch", "--epoch", "1", "--guard", "reclaim-cap", "--reason", "x"}},
		// "reap with nothing to reap" is not a case: a verb that completes binds, and reap is the verb the identity tests and
		// the N56, N69 and N70 tests use to bind a ledger (`reap --bind-repo`), so it binds by design.
	} {
		t.Run(tc.name, func(t *testing.T) {
			// the ledger file sits in O, so O is a clue and the first bind is accepted from O
			db := filepath.Join(o, ".moai-af", tc.args[0]+".db")
			afResLedger(t, db)
			t.Setenv("CLAUDE_PROJECT_DIR", o)
			text, err := afResVerb(db, tc.args...)
			bound := afResBound(t, db)
			t.Logf("%v: err=%v bound=%q (%.160s)", tc.args[:1], err, bound, text)
			if bound != "" {
				t.Errorf("the verb %q changed nothing and the ledger is bound to %q all the same", tc.args[0], bound)
			}
		})
	}
}
