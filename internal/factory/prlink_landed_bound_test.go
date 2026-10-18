// prlink_landed_bound_test.go — card t1805 (t1746 follow-up): the landed
// check's query is a BOUNDED walk.
//
// The defect this card closes: the no-PR-link landed check ran
// `git log origin/main --format=%s` unbounded — 13,385 commits walked per
// card per query, 1.86 s at load ~200 (t1746 lane-verdict E5/R1). The fix
// bounds the walk with --max-count; the acceptance's "under 100 ms" is a
// MEASUREMENT recorded next to the change, never a wall-clock assertion
// (wall-clock asserts flake under load — t1746, t1807). What the tests
// assert is exactly two things:
//
//	structure   — the argv carries the bound (and nothing unbounded survives)
//	answers     — landed / not-landed are unchanged on a fixture history
//	              fully inside the window, against the unbounded reference
//
// The window's edge is pinned as documented behaviour: a landing older than
// the bound answers not-landed. 500 is the pinned bound — at t1746's
// load-200 anchor (~139 µs/commit) it measures ~70 ms, and origin/main grew
// ~5,892 commits in its last 30 days (measured 2026-10-18), so the window
// covers roughly 2.5 days of landing activity.
package factory

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// pinnedLandedMaxCount is the bound this card pins. The production constant
// must equal it; a deliberate change to the window updates both this literal
// and the constant's doc comment in the same change.
const pinnedLandedMaxCount = 500

var maxCountArgPattern = regexp.MustCompile(`^--max-count=([0-9]+)$`)

// maxCountFromArgs extracts the --max-count value the argv carries, or -1
// when the walk is unbounded.
func maxCountFromArgs(args []string) int {
	for _, a := range args {
		if m := maxCountArgPattern.FindStringSubmatch(a); m != nil {
			n, err := strconv.Atoi(m[1])
			if err != nil || n <= 0 {
				return -1
			}
			return n
		}
	}
	return -1
}

// Structure: the builder's argv carries a positive --max-count bound. This
// is the assertion that goes red against the unbounded builder.
func TestLandedSubjectArgs_CarriesABoundedWalk(t *testing.T) {
	args, err := LandedSubjectArgs("origin/main")
	if err != nil {
		t.Fatalf("argv: %v", err)
	}
	if n := maxCountFromArgs(args); n < 0 {
		t.Fatalf("argv %v carries no positive --max-count — the landed walk is unbounded", args)
	}
}

// Structure: one card query is exactly ONE git invocation. The bound must
// not arrive as a second query.
func TestLandedCheck_OneGitCall(t *testing.T) {
	calls := 0
	q := GitLandedQuerier{Run: func(name string, args ...string) (string, error) {
		calls++
		return "", nil
	}}
	if _, err := q.Landed("t199"); err != nil {
		t.Fatalf("landed: %v", err)
	}
	if calls != 1 {
		t.Errorf("Landed made %d git calls, want exactly 1", calls)
	}
}

// Answers: on a fixture history, the bounded querier's landed/not-landed
// verdicts match an UNBOUNDED reference scan of the same history — the
// predicate sees the same subjects, so the bound changes cost, not answers.
func TestLandedCheck_BoundedAnswersMatchUnboundedReference(t *testing.T) {
	dir := landedRepo(t)
	run := gitIn(dir)
	q := GitLandedQuerier{Run: run}

	// The reference: the historical unbounded query, attributed through the
	// same predicate the querier uses.
	raw, err := run("git", "log", DefaultLandedRef, LandedSubjectFormatFlag)
	if err != nil {
		t.Fatalf("reference query: %v", err)
	}
	branch := landedBranchFromRef(DefaultLandedRef)
	reference := map[string]LandingAnswer{}
	for _, subject := range strings.Split(raw, "\n") {
		if id := subjectAttribution(subject, branch); id != "" {
			reference[id] = LandingLanded
		}
	}

	for _, card := range []string{"t199", "t205"} {
		got, err := q.Landed(card)
		if err != nil {
			t.Fatalf("landed %s: %v", card, err)
		}
		want := reference[card]
		if want == "" {
			want = LandingNotLanded
		}
		if got != want {
			t.Errorf("%s: bounded answer %q != unbounded reference %q", card, got, want)
		}
	}
}

// deepLandedRepo builds a history of depth filler commits via plumbing
// (commit-tree — no working-tree churn), optionally carrying a landing
// commit BELOW the fillers (older than the window) or ABOVE them (newest).
func deepLandedRepo(t *testing.T, filler int, landing string) string {
	t.Helper()
	dir := t.TempDir()
	gitCfg := filepath.Join(dir, "gitconfig")
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(cmd.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
			"GIT_CONFIG_GLOBAL="+gitCfg, "GIT_CONFIG_NOSYSTEM=1",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("seed\n"), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	run("add", "seed.txt")
	rootTree := run("write-tree")
	parent := run("commit-tree", rootTree, "-m", "chore: seed the tree")
	if landing == "below" {
		parent = run("commit-tree", rootTree, "-p", parent, "-m", "fix(web): register signal handling before binding the listener (t199)")
	}
	for i := 0; i < filler; i++ {
		parent = run("commit-tree", rootTree, "-p", parent, "-m", fmt.Sprintf("chore: filler %d", i))
	}
	if landing == "above" {
		parent = run("commit-tree", rootTree, "-p", parent, "-m", "fix(web): register signal handling before binding the listener (t199)")
	}
	run("update-ref", "refs/remotes/origin/main", parent)
	return dir
}

// The window's edge is documented behaviour, pinned here: a landing OLDER
// than the bound is outside the walk and answers not-landed; the same
// landing at the tip stays inside and answers landed.
func TestLandedCheck_WindowEdge(t *testing.T) {
	t.Run("landing older than the bound reports not-landed", func(t *testing.T) {
		dir := deepLandedRepo(t, pinnedLandedMaxCount+5, "below")
		landed, err := (GitLandedQuerier{Run: gitIn(dir)}).Landed("t199")
		if err != nil {
			t.Fatalf("landed: %v", err)
		}
		if landed != LandingNotLanded {
			t.Errorf("a landing %d commits past the bound answered %q, want %q (the documented window trade-off)", pinnedLandedMaxCount+5, landed, LandingNotLanded)
		}
	})
	t.Run("landing at the tip reports landed", func(t *testing.T) {
		dir := deepLandedRepo(t, pinnedLandedMaxCount+5, "above")
		landed, err := (GitLandedQuerier{Run: gitIn(dir)}).Landed("t199")
		if err != nil {
			t.Fatalf("landed: %v", err)
		}
		if landed != LandingLanded {
			t.Errorf("a landing inside the window answered %q, want %q", landed, LandingLanded)
		}
	})
}
