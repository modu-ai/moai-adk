// factory_ci_repair_base_test.go — card t2058: a repair card can carry the
// failed release-CI candidate SHA, and its worktree is cut from that SHA.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// t2058FailedCandidate makes a commit that is not on any branch — the shape of
// a failed candidate that the project's base branch has since moved past —
// and returns its full SHA.
func t2058FailedCandidate(t *testing.T, root string) string {
	t.Helper()
	return fcGit(t, root, "commit-tree", "HEAD^{tree}", "-p", "HEAD", "-m", "failed release candidate")
}

// TestT2058_RepairCardRecordsTheFailedCandidateSHA — `todo add --origin
// ci-repair --base-sha <sha>` stores the failed candidate on the card, so the
// recovery is not carried in prose.
func TestT2058_RepairCardRecordsTheFailedCandidateSHA(t *testing.T) {
	root, _ := fcFixture(t)
	sha := t2058FailedCandidate(t, root)
	if _, _, err := runTodo(t, "add", "repair the failed release CI", "--origin", "ci-repair", "--base-sha", sha); err != nil {
		t.Fatalf("todo add --base-sha: %v", err)
	}
	out, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("todo list --json: %v", err)
	}
	if !strings.Contains(out, `"base_sha":"`+sha+`"`) {
		t.Fatalf("the card does not record the failed candidate SHA %s:\n%s", sha, out)
	}
}

// TestT2058_LeaseCutsTheRepairWorktreeFromTheFailedCandidate — leasing the
// repair card creates its worktree from the card's base SHA, not from the
// project's configured base branch.
func TestT2058_LeaseCutsTheRepairWorktreeFromTheFailedCandidate(t *testing.T) {
	root, store := fcFixture(t)
	cepSelfDispatchYAML(t, root, true)
	sha := t2058FailedCandidate(t, root)
	if _, _, err := runTodo(t, "add", "repair the failed release CI", "--pick", "--origin", "ci-repair", "--base-sha", sha); err != nil {
		t.Fatalf("todo add --pick --base-sha: %v", err)
	}
	cepSetSpecID(t, store, "t1", "SPEC-T2058-REPAIR-001")
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	out, _, err := runFactory(t, "next", "--run", fcRun)
	if err != nil {
		t.Fatalf("next: %v\n%s", err, out)
	}
	homeRoot, rerr := worktreeroot.Root(root)
	if rerr != nil {
		t.Fatalf("resolve home root: %v", rerr)
	}
	wt := filepath.Join(homeRoot, "t1")
	if head := fcGit(t, wt, "rev-parse", "HEAD"); head != sha {
		t.Fatalf("repair worktree HEAD = %s, want the failed candidate %s", head, sha)
	}
	if got := fcCard(t, root, "t1"); got.WorktreePath != wt {
		t.Errorf("recorded worktree path = %q, want %q", got.WorktreePath, wt)
	}
}

// TestT2058_BaseSHAFlagIsValidated — the flag takes one full 40-hex SHA: a
// short or non-hex value and a repeated flag are refused and nothing is
// written to the queue.
func TestT2058_BaseSHAFlagIsValidated(t *testing.T) {
	root, store := fcFixture(t)
	good := t2058FailedCandidate(t, root)
	for name, args := range map[string][]string{
		"short":    {"add", "repair", "--base-sha", good[:12]},
		"non-hex":  {"add", "repair", "--base-sha", strings.Repeat("z", 40)},
		"repeated": {"add", "repair", "--base-sha", good, "--base-sha", good},
	} {
		if _, _, err := runTodo(t, args...); err == nil {
			t.Errorf("%s: todo add accepted %v", name, args)
		}
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("a refused add wrote %d card(s)", len(rec.Items))
	}
}

// TestT2058_UnresolvableBaseRefusesTheLeaseWorktree — the repair must never
// land on the wrong base: a well-formed SHA this repository does not have
// refuses the worktree creation (no tree is made, none is recorded) instead of
// falling back to the configured base branch.
func TestT2058_UnresolvableBaseRefusesTheLeaseWorktree(t *testing.T) {
	root, store := fcFixture(t)
	cepSelfDispatchYAML(t, root, true)
	missing := strings.Repeat("a", 40)
	if _, _, err := runTodo(t, "add", "repair the failed release CI", "--pick", "--origin", "ci-repair", "--base-sha", missing); err != nil {
		t.Fatalf("todo add --pick --base-sha: %v", err)
	}
	cepSetSpecID(t, store, "t1", "SPEC-T2058-REPAIR-002")
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	out, _, err := runFactory(t, "next", "--run", fcRun)
	if err == nil {
		t.Fatalf("the lease created a worktree for a base SHA this repository does not have:\n%s", out)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("the refusal does not name the missing SHA: %v", err)
	}
	homeRoot, rerr := worktreeroot.Root(root)
	if rerr != nil {
		t.Fatalf("resolve home root: %v", rerr)
	}
	if _, statErr := os.Stat(filepath.Join(homeRoot, "t1")); statErr == nil {
		t.Fatal("a worktree exists for the card whose base SHA could not be resolved")
	}
}

// TestT2058_CardWithoutABaseKeepsTheConfiguredCut — a card that carries no
// base SHA is cut exactly as before: the repair path changes nothing for it.
func TestT2058_CardWithoutABaseKeepsTheConfiguredCut(t *testing.T) {
	root, store := fcFixture(t)
	cepSelfDispatchYAML(t, root, true)
	if _, _, err := runTodo(t, "add", "an ordinary card", "--pick"); err != nil {
		t.Fatalf("todo add --pick: %v", err)
	}
	cepSetSpecID(t, store, "t1", "SPEC-T2058-ORDINARY-001")
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)
	want := fcGit(t, root, "rev-parse", "HEAD")

	if out, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("next: %v\n%s", err, out)
	}
	homeRoot, rerr := worktreeroot.Root(root)
	if rerr != nil {
		t.Fatalf("resolve home root: %v", rerr)
	}
	if head := fcGit(t, filepath.Join(homeRoot, "t1"), "rev-parse", "HEAD"); head != want {
		t.Fatalf("ordinary card worktree HEAD = %s, want the invoking tree's HEAD %s", head, want)
	}
}
