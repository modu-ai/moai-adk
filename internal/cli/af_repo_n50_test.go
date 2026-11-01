package cli

// N50 (audit of 1c05083f71) through the real `moai af` verbs, a real ledger and real git: the file-location clue is for a
// ledger the caller moved with --db; the default location under a MoaiHome that is itself a git repository made every
// legitimate project's first call a refusal.

import (
	"path/filepath"
	"strings"
	"testing"
)

// The default ledger location lies under MoaiHome. When that is itself a git repository (dotfiles, say), the location is
// a place the tool chose and not a statement about the project, so it is no clue: every project's first call used to be
// refused. The same file given with --db is a statement of the caller, and is still compared.
func TestAFVerbs_N50TheDefaultLedgerLocationIsNoFileClue(t *testing.T) {
	afSkipWithoutSh(t)
	home, _ := afN38Other(t)
	project, _ := afN38Other(t)
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
	t.Setenv("CLAUDE_PROJECT_DIR", project)
	db, err := afLedgerPath("")
	if err != nil {
		t.Fatal(err)
	}
	if real, rerr := filepath.EvalSymlinks(home); !strings.HasPrefix(db, home+string(filepath.Separator)) && (rerr != nil || !strings.HasPrefix(db, real+string(filepath.Separator))) {
		t.Fatalf("setup: the default ledger %s is not under the home repository %s", db, home)
	}
	idHome, idProject := afN38GitDir(t, home), afN38GitDir(t, project)

	// the control: the very same file given with --db is compared with the repository it sits in, and refused
	if out, errOut, err := afRunBoth("reap", "--db", db); err == nil {
		t.Fatalf("--db %s from a project elsewhere was accepted: %.200s %.200s", db, out, errOut)
	} else {
		for _, want := range []string{"not bound yet", idHome, idProject} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the refusal does not name %q: %v", want, err)
			}
		}
	}
	// the ledger at the default location holds one ready card, so that a writing verb has something to write (N55: a verb
	// that wrote nothing binds nothing without a declaration, so a first bind here is made by a verb that writes)
	st := afN49Ledger(t, db)
	if got := afN49Bound(t, st); got != "" {
		t.Fatalf("the refused --db call bound the ledger to %q", got)
	}

	// the default location, no --db: the project binds
	for _, args := range [][]string{{"reap"}, {"claim", "--session", "s"}} {
		out, errOut, err := afRunBoth(args...)
		if err != nil {
			t.Fatalf("%v with the default ledger under a home repository: %v\n%.300s %.300s", args, err, out, errOut)
		}
	}
	if got := afN49Bound(t, st); got != idProject {
		t.Errorf("bound %q, want the project's git directory %q", got, idProject)
	}
}

// The help says the file-location clue belongs to --db, and that the default location is not compared.
func TestAFVerbs_N50HelpSaysTheFileClueIsForADbFlagOnly(t *testing.T) {
	for _, verb := range afN38Verbs() {
		flat := afHelpOf(t, verb)
		for _, want := range []string{"given with --db", "default ledger location is not compared"} {
			if !strings.Contains(flat, want) {
				t.Errorf("`moai af %s --help` does not say %q", verb, want)
			}
		}
	}
}
