// todo_done_requirelanded_window_test.go — card t1805 iter2: a landing the
// bounded window cannot see must take the unknown path of
// `todo done --require-landed`, never the refusal.
//
// The bounded landed walk (LandedMaxCount) made a landing older than the
// window read as not-landed, and `--require-landed` refuses not-landed — a
// landed card's done would be refused on evidence the query never saw. The
// filled-window case is INDETERMINATE: the querier answers unknown, the
// refusal does not fire, and the done proceeds with the recorded verdict.
package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// windowFillFillers is the filler depth that guarantees the bounded walk is
// full while the landing commit sits below it — more than the bound, keeping
// the landing strictly outside the most-recent-500 subjects.
const windowFillFillers = factory.LandedMaxCount + 5

// seedWindowOutsideLandedRepo gives the fixture repo a landed ref whose
// subject stream attributes t1 in a commit OLDER than the bounded window:
// the landing, then windowFillFillers fillers on top of it.
func seedWindowOutsideLandedRepo(t *testing.T, root string) {
	t.Helper()
	runGitIn(t, root, "commit", "--allow-empty", "-q", "-m", "chore: seed the tree")
	tree := strings.TrimSpace(gitOutIn(t, root, "write-tree"))
	parent := strings.TrimSpace(gitOutIn(t, root, "commit-tree", tree, "-m", "feat: work the card (t1)"))
	for i := 0; i < windowFillFillers; i++ {
		parent = strings.TrimSpace(gitOutIn(t, root, "commit-tree", tree, "-p", parent, "-m", "chore: filler"))
	}
	runGitIn(t, root, "update-ref", "refs/remotes/origin/landed", parent)
	runGitIn(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/landed")
	_ = landedRefName(t, root)
}

// TestDoneVerdict_WindowOutsideLandingTakesTheUnknownPath — the regression:
// `--require-landed` must NOT refuse when the window filled without an
// attribution; the done proceeds and records the unknown verdict.
func TestDoneVerdict_WindowOutsideLandingTakesTheUnknownPath(t *testing.T) {
	root, store := todoFixture(t)
	seedWindowOutsideLandedRepo(t, root)

	if _, _, err := runTodo(t, "add", "--pick", "card landed outside the window"); err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	stdout, _, err := runTodo(t, "done", "t1", "--require-landed")
	if err != nil {
		t.Fatalf("done --require-landed refused on a filled window — the window-outside landing must take the unknown path: %v", err)
	}
	if !strings.Contains(stdout, "landing=unknown") {
		t.Fatalf("done line = %q, want landing=unknown", strings.TrimSpace(stdout))
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Archived) != 1 {
		t.Fatalf("archive holds %d entries, want 1", len(rec.Archived))
	}
	v := rec.Archived[0].LandingVerdict
	if v == nil {
		t.Fatalf("landing_verdict is NULL after the unknown-path done, want a record")
	}
	if v.Verdict != factory.LandingUnknown {
		t.Errorf("verdict = %q, want %q", v.Verdict, factory.LandingUnknown)
	}
}
