package cli

// N56 (audit 1c05083f71, and N69 last bullet of a36ea8c368; residual classification): the help says "two spellings of one path
// are one repository", but the bound identity is a string and a spelling that differs only in case (the default volume of
// macOS is case-insensitive) is another repository. The same card, the same directory, refused. The expectation: the
// directory spelled with another case is the repository the ledger is bound to.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAFResidual_N56AProjectDirectorySpelledWithAnotherCaseIsTheSameRepository(t *testing.T) {
	afSkipWithoutSh(t)
	root := t.TempDir()
	p, _ := afResRepo(t, filepath.Join(root, "Repo"))
	variant := filepath.Join(root, "REPO")
	a, aerr := os.Stat(p)
	b, berr := os.Stat(variant)
	if aerr != nil || berr != nil || !os.SameFile(a, b) {
		t.Skipf("the volume of %s is case-sensitive (stat of the variant: %v)", root, berr)
	}
	db := filepath.Join(root, "central", "ledger.db")
	afResLedger(t, db)
	t.Setenv("CLAUDE_PROJECT_DIR", p)
	if text, err := afResVerb(db, "reap", "--bind-repo", p); err != nil {
		t.Fatalf("the first bind from %s: %v (%.200s)", p, err, text)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", variant)
	text, err := afResVerb(db, "reap")
	t.Logf("reap with the project spelled %s: err=%v (%.240s)", variant, err, text)
	if err != nil {
		t.Errorf("the same directory spelled with another case is refused as another repository: %v", err)
	}
}
