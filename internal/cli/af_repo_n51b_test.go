package cli

// N51 (b) (audit of 1c05083f71, TestA4Bind_RowsNameCommitsTheFirstCallerDoesNotHave): a ledger with history, no lease
// left and its file outside any repository has no lease worktree and no file location to compare, yet its rows name
// commits (evidence subjects, queue candidates). A first caller whose repository does not have those commits bound
// the ledger anyway, and every later call from the repository the commits are in was refused. Before the first bind
// the commits the rows name must exist in the repository the call would bind.

import (
	"context"
	"strconv"
	"testing"
)

func TestAFVerbs_N51bRowsNameCommitsTheCallerDoesNotHave(t *testing.T) {
	afSkipWithoutSh(t)
	for _, tc := range []struct {
		name  string
		queue bool
	}{
		{"evidence rows only", false},
		{"evidence rows and a queue row", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			a := afN38UnboundCard(t)
			a.storeMeasure(a.c1)
			epoch, _ := strconv.ParseInt(a.e2, 10, 64)
			if tc.queue {
				if err := a.st.Enqueue(ctx, "t1", epoch, a.c1); err != nil {
					t.Fatal(err)
				}
			}
			if err := a.st.Release(ctx, "t1", epoch); err != nil {
				t.Fatal(err)
			}
			if leases, err := a.st.Leases(ctx); err != nil || len(leases) != 0 {
				t.Fatalf("setup: leases %+v %v", leases, err)
			}
			other, _ := afN38Other(t) // a repository that does not have c1
			t.Setenv("CLAUDE_PROJECT_DIR", other)
			idOther := afN38GitDir(t, other)
			a.refusedAtFirstBind("reap from a repository without the commits", a.c1, idOther, "reap")
			a.refusedAtFirstBind("move from a repository without the commits", a.c1, idOther, "move", "t1", "--epoch", a.e2, "--stage", "REVIEWING")
			// the control: from the repository that has them the ledger binds, and the others are refused afterwards
			t.Setenv("CLAUDE_PROJECT_DIR", a.repo)
			a.mustRun("reap")
			if got, want := a.bound(), afN38GitDir(t, a.repo); got != want {
				t.Fatalf("bound %q, want %q", got, want)
			}
			t.Setenv("CLAUDE_PROJECT_DIR", other)
			a.refused("reap from the other repository once bound", idOther, "reap")
		})
	}
}
