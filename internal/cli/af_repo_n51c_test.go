package cli

// N51 (c) (audit of 1c05083f71, TestA4Bind_SymlinkedLedgerFile): the ledger file is a clue to the repository it sits in,
// and the directory asked was the directory of the path as given. A ledger file in repository P reached through a link
// outside any repository was therefore asked about the link's directory, found no repository there, and bound whatever
// repository the first caller stood in. The link is resolved before the file's location is asked.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAFVerbs_N51cASymlinkedLedgerFileIsAskedWhereItReallyIs(t *testing.T) {
	afSkipWithoutSh(t)
	p, _ := afN38Other(t)
	o, _ := afN38Other(t)
	idP, idO := afN38GitDir(t, p), afN38GitDir(t, o)
	real := filepath.Join(p, ".moai", "af", "ledger.db")
	st := afN49Ledger(t, real)
	link := filepath.Join(t.TempDir(), "ledger.db")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", o)
	wants := []string{"not bound yet", idP, idO}
	afN49Refused(t, st, "reap through the link from O", wants, "reap", "--db", link)
	// the same file by its direct path is refused the same way
	afN49Refused(t, st, "reap by the direct path from O", wants, "reap", "--db", real)

	// the control: the project is the repository the file really sits in, reached through the link. The reap has nothing to
	// reap, so it binds only by declaring the repository (N55, the explicit first bind).
	t.Setenv("CLAUDE_PROJECT_DIR", p)
	if out, errOut, err := afRunBoth("reap", "--db", link, "--bind-repo", p); err != nil {
		t.Fatalf("reap through the link from P: %v\n%.300s %.300s", err, out, errOut)
	}
	if got := afN49Bound(t, st); got != idP {
		t.Errorf("bound %q, want %q", got, idP)
	}
}
