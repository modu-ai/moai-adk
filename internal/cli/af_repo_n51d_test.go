package cli

// N51 (d) (audit of 1c05083f71, TestA4Bind_RelativeLeaseWorktreeAgreesWithEveryone): a lease worktree recorded as "."
// is looked up with `git -C .`, which is the working directory of whoever asks, so the comparison with the caller's own
// repository always agrees and the first bind is the caller's. A lease worktree must be an absolute path: `claim`
// refuses a relative --worktree where it is recorded, and a relative one a ledger already holds (an older build, a
// direct write) is refused where the first bind would trust it.

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestAFVerbs_N51dClaimRefusesARelativeWorktree(t *testing.T) {
	afSkipWithoutSh(t)
	p, _ := afN38Other(t)
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir()) // a claim binds its heartbeat under the project directory
	db := t.TempDir() + "/af/ledger.db"
	st := afN49Ledger(t, db)
	for _, rel := range []string{".", "sub/dir", "../elsewhere"} {
		before := afSnapshot(t, st)
		_, _, err := afRunBoth("claim", "--db", db, "--session", "s1", "--worktree", rel)
		if err == nil {
			t.Errorf("claim --worktree %q was accepted", rel)
			continue
		}
		if !strings.Contains(err.Error(), "absolute") || !strings.Contains(err.Error(), rel) {
			t.Errorf("claim --worktree %q: the refusal does not say that a worktree must be an absolute path and name %q: %v", rel, rel, err)
		}
		if after := afSnapshot(t, st); len(after.Leases) != 0 || len(before.Leases) != 0 {
			t.Errorf("claim --worktree %q left a lease: %+v", rel, after.Leases)
		}
	}
	// the control: an absolute worktree is recorded
	if got := afN49Claim(t, db, p); got == "" {
		t.Error("no epoch")
	}
}

func TestAFVerbs_N51dARelativeLeaseWorktreeAlreadyRecordedIsNotTrustedAtTheFirstBind(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	o, x := afN49Other(t)
	idO := afN38GitDir(t, o)
	claimAs := func(st *ledger.Store, worktree string) {
		t.Helper()
		if _, err := st.Claim(ctx, ledger.ClaimRequest{SessionID: "s", Worktree: worktree, PID: os.Getpid(), PIDStart: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	e := "1"

	db := t.TempDir() + "/af/ledger.db"
	st := afN49Ledger(t, db)
	claimAs(st, ".")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(o)
	wants := []string{"not bound yet", "not an absolute path", "t1"}
	afN49Refused(t, st, "evidence with a lease worktree recorded as .", wants,
		"evidence", "t1", "--db", db, "--epoch", e, "--kind", "gate", "--subject-sha", x, "--result", "pass")
	afN49Refused(t, st, "reap with a lease worktree recorded as .", wants, "reap", "--db", db)

	// the control: the same lease recorded with the absolute path agrees with the caller and the first bind goes through
	db2 := t.TempDir() + "/af/ledger.db"
	st2 := afN49Ledger(t, db2)
	claimAs(st2, o)
	if out, errOut, err := afRunBoth("evidence", "t1", "--db", db2, "--epoch", e, "--kind", "gate", "--subject-sha", x, "--result", "pass", "--bind-repo", o); err != nil {
		t.Fatalf("evidence with an absolute lease worktree: %v\n%.300s %.300s", err, out, errOut)
	}
	if got := afN49Bound(t, st2); got != idO {
		t.Errorf("bound %q, want %q", got, idO)
	}
}
