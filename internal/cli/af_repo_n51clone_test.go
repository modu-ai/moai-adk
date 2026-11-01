package cli

// N51, the clone form (audit of 1c05083f71, TestA4Bind_WrongFirstCallRunsToDone): a ledger given with --db whose file sits
// in no repository, with no lease worktree and no commit in its rows, knows nothing that says which repository it belongs
// to. Every call from a second clone of the project then ran to DONE with nothing ever refused, the commits landing in
// the clone's trunk and not in the project's. Nothing can be inferred for such a ledger, so its first bind needs the
// caller to say which repository it is: --bind-repo, the one flag that names a repository on every verb (--repo is
// absent from claim, move, release, reap and resume, and where it exists it is checked against the project directory, so
// it declares nothing the project directory has not said). The default ledger location needs no declaration.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestAFVerbs_N51WrongFirstCallFromASecondCloneNeedsADeclaration(t *testing.T) {
	afSkipWithoutSh(t)
	p, _ := afN38Other(t)
	clone := t.TempDir()
	if out, err := exec.Command("git", "clone", "-q", "--no-hardlinks", p, clone).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v %s", err, out)
	}
	other, _ := afN38Other(t)
	idClone, idOther := afN38GitDir(t, clone), afN38GitDir(t, other)
	db := filepath.Join(t.TempDir(), "central", "ledger.db") // in no repository
	st := afN49Ledger(t, db)
	t.Setenv("CLAUDE_PROJECT_DIR", clone)

	wants := []string{"not bound yet", "--bind-repo", idClone}
	afN49Refused(t, st, "claim from the clone", wants, "claim", "--db", db, "--session", "s1", "--worktree", clone)
	afN49Refused(t, st, "reap from the clone", wants, "reap", "--db", db)
	afN49Refused(t, st, "land from the clone", wants, "land", "t1", "--db", db, "--gate", "true")
	afN49Refused(t, st, "reconcile from the clone", wants, "reconcile", "--db", db)
	if cards, err := st.Cards(context.Background()); err != nil || len(cards) != 1 || cards[0].Position.Stage != core.StageReady {
		t.Errorf("cards %+v %v: the card must not have moved", cards, err)
	}
	// a declaration that names another repository than the one the call would bind is refused as well
	afN49Refused(t, st, "claim declaring another repository", []string{idClone, idOther},
		"claim", "--db", db, "--session", "s1", "--worktree", clone, "--bind-repo", other)

	// the control: the caller says which repository this ledger belongs to, and the first bind goes through
	out, errOut, err := afRunBoth("claim", "--db", db, "--session", "s1", "--worktree", clone, "--bind-repo", clone)
	if err != nil {
		t.Fatalf("claim declaring the clone: %v\n%.300s %.300s", err, out, errOut)
	}
	if got := afN49Bound(t, st); got != idClone {
		t.Errorf("bound %q, want %q", got, idClone)
	}
}

// What needs no declaration: a clue exists (a lease worktree, the repository the file sits in), or the ledger is at its
// default location.
func TestAFVerbs_N51ADeclarationIsNotAskedWhereAnythingElseSaysWhereTheLedgerBelongs(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	repo, _ := afN38Other(t)
	id := afN38GitDir(t, repo)
	t.Setenv("CLAUDE_PROJECT_DIR", repo)
	// N65 (audit of a36ea8c368): a lease worktree is the caller's own word (a claim from a directory that is no
	// repository records whatever worktree it is given), so it is no clue that says where the ledger belongs: the first
	// bind of such a ledger needs the declaration, and with it the bind goes through.
	t.Run("a lease worktree in the repository is not a clue", func(t *testing.T) {
		db := filepath.Join(t.TempDir(), "central", "ledger.db")
		st := afN49Ledger(t, db)
		if _, err := st.Claim(ctx, ledger.ClaimRequest{SessionID: "s", Worktree: repo, PID: os.Getpid(), PIDStart: "x"}); err != nil {
			t.Fatal(err)
		}
		afN49Refused(t, st, "reap on a lease alone", []string{"not bound yet", "--bind-repo", id}, "reap", "--db", db)
		if out, errOut, err := afRunBoth("reap", "--db", db, "--bind-repo", repo); err != nil {
			t.Fatalf("reap: %v\n%.300s %.300s", err, out, errOut)
		}
		if got := afN49Bound(t, st); got != id {
			t.Errorf("bound %q, want %q", got, id)
		}
	})
	t.Run("the ledger file sits in the repository", func(t *testing.T) {
		db := filepath.Join(repo, ".moai", "af", "ledger.db")
		st := afN49Ledger(t, db)
		if out, errOut, err := afRunBoth("reap", "--db", db); err != nil {
			t.Fatalf("reap: %v\n%.300s %.300s", err, out, errOut)
		}
		if got := afN49Bound(t, st); got != id {
			t.Errorf("bound %q, want %q", got, id)
		}
	})
	t.Run("the default location", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
		db, err := afLedgerPath("")
		if err != nil {
			t.Fatal(err)
		}
		if out, errOut, err := afRunBoth("reap"); err != nil {
			t.Fatalf("reap at the default location: %v\n%.300s %.300s", err, out, errOut)
		}
		st, err := ledger.Open(ctx, db)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = st.Close() }()
		if got := afN49Bound(t, st); got != id {
			t.Errorf("bound %q, want %q", got, id)
		}
	})
	t.Run("the help names the flag", func(t *testing.T) {
		for _, verb := range afN38Verbs() {
			flat := afHelpOf(t, verb)
			for _, want := range []string{"--bind-repo", "sits in no repository"} {
				if !strings.Contains(flat, want) {
					t.Errorf("`moai af %s --help` does not say %q", verb, want)
				}
			}
		}
	})
}
