package factory

// integration_merge_step_overlap_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616)
// M2 RED tests for the separate-worktree dirty decision (REQ-LMF-005, plan §B5)
// and the status-set comparison (REQ-LMF-006, plan §B5). The fixture is
// newMergeFixture from integration_merge_step_test.go: its base .gitignore
// ignores card.txt, the candidate's added path, so an ignored byte there is the
// ignored overlap I of §B5.

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorylane"
)

// stepDirtyCase is one dirty-state shape. dirty arranges the integration
// worktree, and the card branch when the card must change a path, before the
// step runs.
type stepDirtyCase struct {
	name  string
	dirty func(t *testing.T, f *stepFixture)
}

// stepWriteFile writes body to name under dir, creating parent directories.
func stepWriteFile(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// stepStatus reads the integration worktree's status with untracked files
// listed one by one: the reading M1's status-set test takes (lmfStatus).
func stepStatus(t *testing.T, f *stepFixture) string {
	t.Helper()
	return stepMustGit(t, f.integ, "status", "--porcelain=v1", "--untracked-files=all")
}

// recommitCard adds one commit on the card's WT- branch that writes files, then
// re-keys the re-measure record to the new candidate tree. The base does not
// move, so the pinned SHA still descends from it.
func (f *stepFixture) recommitCard(t *testing.T, files map[string]string) {
	t.Helper()
	var names []string
	for name, body := range files {
		stepWriteFile(t, f.cardTree, name, body)
		names = append(names, name)
	}
	stepMustGit(t, f.cardTree, append([]string{"add", "-f"}, names...)...)
	stepMustGit(t, f.cardTree, "commit", "-q", "-m", "card recommit")
	f.cardSHA = strings.TrimSpace(stepMustGit(t, f.cardTree, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(stepMustGit(t, f.cardTree, "rev-parse", "HEAD^{tree}"))
	rec := *f.record
	rec.Tree = tree
	if err := WriteRemeasureRecord(f.root, tree, rec); err != nil {
		t.Fatal(err)
	}
	f.record = &rec
}

// requireMergeOfCard asserts that sha is a --no-ff merge whose second parent is
// the pinned card SHA.
func requireMergeOfCard(t *testing.T, f *stepFixture, sha string) {
	t.Helper()
	parents := strings.Fields(stepMustGit(t, f.integ, "rev-list", "--parents", "-n", "1", sha))
	if len(parents) != 3 || parents[2] != f.cardSHA {
		t.Fatalf("the merge commit's second parent must be the pinned SHA %s: %v", f.cardSHA, parents)
	}
}

// requireNoMerge asserts that the integration HEAD is still the record's base:
// a refusal performs no merge.
func requireNoMerge(t *testing.T, f *stepFixture) {
	t.Helper()
	if head := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD")); head != f.record.Base {
		t.Fatalf("a refusal must perform no merge: HEAD %s, want the base %s", head, f.record.Base)
	}
}

// requireHeldWithoutPromotion asserts that the window carries a hold and that C
// was not promoted onto it, and returns the hold's reason.
func requireHeldWithoutPromotion(t *testing.T, f *stepFixture) string {
	t.Helper()
	policy, err := ReadIntegrationWindowPolicy(f.root)
	if err != nil || policy.Policy != PolicyHold {
		t.Fatalf("the window must be held: %+v err=%v", policy, err)
	}
	if lock, _ := ReadIntegrationLock(f.root); lock.SessionID == "sess-c" {
		t.Fatalf("C must not be promoted onto a held window")
	}
	return policy.Reason
}

// stepGitAfterMerge runs every git call in the integration worktree and runs
// after once a merge (not an abort) has succeeded, so the status change lands
// between the merge commit and the step's post-merge reads.
func stepGitAfterMerge(f *stepFixture, after func()) func(args ...string) (string, error) {
	return func(args ...string) (string, error) {
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		if err == nil && len(args) >= 1 && args[0] == "merge" && !containsArg(args, "--abort") {
			after()
		}
		return string(out), err
	}
}

// stepGitFailingMergeThenAbort returns a git seam whose merge fails before it
// writes anything, so the step must abort. The real abort runs, and after runs
// once it has returned: the status change that S-abort is measured against. git
// refuses such a merge with exit 2 and writes no MERGE_HEAD (measured on git
// 2.54.0, for a local change the merge would overwrite), so the seam reports that
// status: the call ran and did not stop with a merge in progress.
func stepGitFailingMergeThenAbort(f *stepFixture, after func()) func(args ...string) (string, error) {
	return func(args ...string) (string, error) {
		isMerge := len(args) >= 1 && args[0] == "merge"
		if isMerge && !containsArg(args, "--abort") {
			return "", &factorylane.GitExitError{ExitCode: 2, Stderr: "simulated merge failure"}
		}
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		if isMerge {
			after()
		}
		return string(out), err
	}
}

func TestMergeStepDirtyDisjointPathsProceeds(t *testing.T) {
	// REQ-LMF-005 (plan §B5): on a separate integration worktree the dirty
	// precondition is the overlap decision, not the fully-clean rule. A dirty path
	// that neither equals nor nests with a card target path does not refuse, and
	// the merge lands with the status set unchanged (REQ-LMF-006).
	cases := []stepDirtyCase{
		{"untracked file outside the card's paths", func(t *testing.T, f *stepFixture) {
			stepWriteFile(t, f.integ, "untracked.txt", "local\n")
		}},
		{"unstaged change to a tracked file outside the card's paths", func(t *testing.T, f *stepFixture) {
			stepWriteFile(t, f.integ, "base.txt", "local edit\n")
		}},
		{"path sharing only a textual prefix with a card path", func(t *testing.T, f *stepFixture) {
			stepWriteFile(t, f.integ, "card.txt.bak", "local\n")
		}},
		{"ignored file outside the card's paths", func(t *testing.T, f *stepFixture) {
			exclude := filepath.Join(f.integ, ".git", "info", "exclude")
			file, err := os.OpenFile(exclude, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.WriteString("scratch.bin\n"); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			stepWriteFile(t, f.integ, "scratch.bin", "ignored\n")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newMergeFixture(t)
			card := f.withCardTree(readyCardPtr())
			tc.dirty(t, f)
			before := stepStatus(t, f)
			sha, err := RunMergeStep(f.input(), f.seams(card))
			if err != nil {
				t.Fatalf("a dirty path outside the card's target paths must not refuse the merge: %v", err)
			}
			requireMergeOfCard(t, f, sha)
			if after := stepStatus(t, f); after != before {
				t.Fatalf("the status set must be unchanged by the merge (REQ-LMF-006): before %q, after %q", before, after)
			}
			requireWindowReleasedAndCPromoted(t, f)
		})
	}
}

func TestMergeStepDirtyPathOverlapRefuses(t *testing.T) {
	// REQ-LMF-005 (plan §B5): a dirty path that equals a card target path refuses
	// with MergeExitWorktreeDirty before any merge call. A staged rename dirties
	// its source as well as its destination (plan §B5, D: both paths of an R or C
	// record belong to D), so the rename case refuses on the source path, which
	// the rename's destination does not name.
	cases := []stepDirtyCase{
		{"unstaged change to a card target path", func(t *testing.T, f *stepFixture) {
			f.recommitCard(t, map[string]string{"base.txt": "card edit\n"})
			stepWriteFile(t, f.integ, "base.txt", "local edit\n")
		}},
		{"staged change to a card target path", func(t *testing.T, f *stepFixture) {
			f.recommitCard(t, map[string]string{"base.txt": "card edit\n"})
			stepWriteFile(t, f.integ, "base.txt", "local edit\n")
			stepMustGit(t, f.integ, "add", "base.txt")
		}},
		{"staged rename whose source is a card target path", func(t *testing.T, f *stepFixture) {
			f.recommitCard(t, map[string]string{"base.txt": "card edit\n"})
			stepMustGit(t, f.integ, "mv", "base.txt", "renamed.txt")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newMergeFixture(t)
			card := f.withCardTree(readyCardPtr())
			tc.dirty(t, f)
			_, err := RunMergeStep(f.input(), f.seams(card))
			requireCode(t, err, MergeExitWorktreeDirty)
			requireNoMerge(t, f)
			requireWindowReleasedAndCPromoted(t, f)
		})
	}
}

func TestMergeStepDirtyDirectoryPrefixOverlapRefuses(t *testing.T) {
	// REQ-LMF-005 (plan §B5): Overlap also holds when one path names a directory
	// that contains the other, in either direction.
	cases := []stepDirtyCase{
		{"untracked file where the card adds files under a directory", func(t *testing.T, f *stepFixture) {
			f.recommitCard(t, map[string]string{"shared/x.txt": "card\n"})
			stepWriteFile(t, f.integ, "shared", "local file\n")
		}},
		{"untracked file under a directory the card replaces with a file", func(t *testing.T, f *stepFixture) {
			f.recommitCard(t, map[string]string{"dir": "card file\n"})
			stepWriteFile(t, f.integ, "dir/file.txt", "local\n")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newMergeFixture(t)
			card := f.withCardTree(readyCardPtr())
			tc.dirty(t, f)
			_, err := RunMergeStep(f.input(), f.seams(card))
			requireCode(t, err, MergeExitWorktreeDirty)
			requireNoMerge(t, f)
			requireWindowReleasedAndCPromoted(t, f)
		})
	}
}

func TestMergeStepDirtyIgnoredTargetRefuses(t *testing.T) {
	// REQ-LMF-005 (plan §B5): an ignored file at a card target path is the ignored
	// overlap I and refuses with MergeExitWorktreeDirty. git status does not list
	// ignored files, so the status set cannot see this byte. The decision is taken
	// before the collision check (13), which is why the refusal is 12, not 13.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	stepWriteFile(t, f.integ, "card.txt", "late ignored byte")
	_, err := RunMergeStep(f.input(), f.seams(card))
	requireCode(t, err, MergeExitWorktreeDirty)
	got, readErr := os.ReadFile(filepath.Join(f.integ, "card.txt"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "late ignored byte" {
		t.Fatalf("the refusal must leave the ignored byte untouched, got %q", got)
	}
	requireNoMerge(t, f)
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepStatusSetChangedAfterMergeHolds(t *testing.T) {
	// REQ-LMF-006 (plan §B5): S-after is compared with S-before record by record,
	// as exact record bytes. A change across the merge holds with
	// MergeExitPostMerge; the merge commit stays, and the hold names it.
	cases := []struct {
		name   string
		before func(t *testing.T, f *stepFixture)
		after  func(t *testing.T, f *stepFixture)
	}{
		{"clean before, a new untracked file after the merge", nil, func(t *testing.T, f *stepFixture) {
			stepWriteFile(t, f.integ, "late.txt", "late\n")
		}},
		{"untracked file before, the same file staged after the merge",
			func(t *testing.T, f *stepFixture) {
				stepWriteFile(t, f.integ, "untracked.txt", "local\n")
			}, func(t *testing.T, f *stepFixture) {
				stepMustGit(t, f.integ, "add", "untracked.txt")
			}},
		{"untracked file before, a new untracked file after the merge",
			func(t *testing.T, f *stepFixture) {
				stepWriteFile(t, f.integ, "untracked.txt", "local\n")
			}, func(t *testing.T, f *stepFixture) {
				stepWriteFile(t, f.integ, "late.txt", "late\n")
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newMergeFixture(t)
			card := f.withCardTree(readyCardPtr())
			if tc.before != nil {
				tc.before(t, f)
			}
			seams := f.seams(card)
			seams.Git = stepGitAfterMerge(f, func() { tc.after(t, f) })
			_, err := RunMergeStep(f.input(), seams)
			requireCode(t, err, MergeExitPostMerge)
			head := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD"))
			requireMergeOfCard(t, f, head)
			if reason := requireHeldWithoutPromotion(t, f); !strings.Contains(reason, head[:12]) {
				t.Fatalf("the hold must name the merge commit %s: %q", head[:12], reason)
			}
		})
	}
}

func TestMergeStepStatusSetChangedAfterAbortHolds(t *testing.T) {
	// REQ-LMF-006 (plan §B5): S-abort is compared with S-before after
	// `git merge --abort`. A change there holds with MergeExitMergeDirty, and the
	// hold is written before the release.
	cases := []struct {
		name   string
		before func(t *testing.T, f *stepFixture)
		after  func(t *testing.T, f *stepFixture)
	}{
		{"clean before, a new untracked file after the abort", nil, func(t *testing.T, f *stepFixture) {
			stepWriteFile(t, f.integ, "late.txt", "late\n")
		}},
		{"untracked file before, the same file staged after the abort",
			func(t *testing.T, f *stepFixture) {
				stepWriteFile(t, f.integ, "untracked.txt", "local\n")
			}, func(t *testing.T, f *stepFixture) {
				stepMustGit(t, f.integ, "add", "untracked.txt")
			}},
		{"untracked file before, a new untracked file after the abort",
			func(t *testing.T, f *stepFixture) {
				stepWriteFile(t, f.integ, "untracked.txt", "local\n")
			}, func(t *testing.T, f *stepFixture) {
				stepWriteFile(t, f.integ, "late.txt", "late\n")
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newMergeFixture(t)
			card := f.withCardTree(readyCardPtr())
			if tc.before != nil {
				tc.before(t, f)
			}
			seams := f.seams(card)
			seams.Git = stepGitFailingMergeThenAbort(f, func() { tc.after(t, f) })
			_, err := RunMergeStep(f.input(), seams)
			requireCode(t, err, MergeExitMergeDirty)
			requireNoMerge(t, f)
			requireHeldWithoutPromotion(t, f)
		})
	}
}

func TestMergeStepStagedDisjointChangeRefusesFailClosed(t *testing.T) {
	// Characterization test (card t1616, plan §B5). The staged path base.txt is
	// outside T, and no ignored file sits at a target path, so the overlap decision
	// proceeds. git's merge then refuses the staged index ("Your local changes to
	// the following files would be overwritten by merge: base.txt"), and the step
	// fails closed with MergeExitMergeFailed: the abort restores the status
	// records, HEAD does not move, no merge commit exists, and the window is
	// released. This pins the observed behaviour. §B5 says a disjoint change
	// proceeds, so the refusal is a Gap (progress.md §59).
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	stepWriteFile(t, f.integ, "base.txt", "staged edit\n")
	stepMustGit(t, f.integ, "add", "base.txt")
	statusBefore := stepStatus(t, f)
	headBefore := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD"))
	_, err := RunMergeStep(f.input(), f.seams(card))
	requireCode(t, err, MergeExitMergeFailed)
	if statusAfter := stepStatus(t, f); statusAfter != statusBefore {
		t.Fatalf("the status records must be identical before and after the refused merge: before %q, after %q", statusBefore, statusAfter)
	}
	if headAfter := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD")); headAfter != headBefore {
		t.Fatalf("HEAD must not move on a refused merge: %s -> %s", headBefore, headAfter)
	}
	requireNoMerge(t, f)
	if _, statErr := os.Stat(filepath.Join(f.integ, ".git", "MERGE_HEAD")); !os.IsNotExist(statErr) {
		t.Fatalf("the abort must leave no merge in progress: %v", statErr)
	}
	requireWindowReleasedAndCPromoted(t, f)
}

// stepPreexistingMerge leaves the integration worktree in a merge that another
// session began and has not finished: MERGE_HEAD names an unrelated commit, and
// base.txt is an add/add conflict between develop and that commit. develop stays
// at the record's base, so every gate before the merge call passes, and base.txt
// is outside the card's target path card.txt, so the overlap decision does not
// refuse on the conflict.
func stepPreexistingMerge(t *testing.T, f *stepFixture) {
	t.Helper()
	scratch := filepath.Join(t.TempDir(), "unrelated")
	stepMustGit(t, f.integ, "worktree", "add", "-q", "--detach", scratch)
	stepMustGit(t, scratch, "checkout", "-q", "--orphan", "unrelated")
	stepWriteFile(t, scratch, "base.txt", "unrelated side\n")
	stepMustGit(t, scratch, "add", "base.txt")
	stepMustGit(t, scratch, "commit", "-q", "-m", "unrelated side")
	unrelated := strings.TrimSpace(stepMustGit(t, scratch, "rev-parse", "HEAD"))
	stepMustGit(t, f.integ, "worktree", "remove", "--force", scratch)
	// git merge exits 1 on the conflict, which is the state this fixture needs.
	conflict := exec.Command("git", "merge", "--no-commit", "--no-ff", "--allow-unrelated-histories", unrelated)
	conflict.Dir = f.integ
	_, _ = conflict.CombinedOutput()
	if _, err := os.Stat(filepath.Join(f.integ, ".git", "MERGE_HEAD")); err != nil {
		t.Fatalf("fixture: the conflicted merge must leave MERGE_HEAD: %v", err)
	}
	if unmerged := stepMustGit(t, f.integ, "ls-files", "-u"); !strings.Contains(unmerged, "base.txt") {
		t.Fatalf("fixture: base.txt must be the conflicted path, got unmerged entries %q", unmerged)
	}
}

func TestMergeStepPreexistingMergeHeadRefusesWithoutAbort(t *testing.T) {
	// F1 (sync audit, card t1616): a separate integration worktree that is already in
	// an unfinished merge must refuse the landing with MergeExitWorktreeDirty, and the
	// refusal must leave that merge as it was. integration_merge_step.go runs
	// `git merge --abort` after its own merge fails. git refuses to start a merge
	// while MERGE_HEAD exists, so that abort discards the merge another session began.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	stepPreexistingMerge(t, f)
	statusBefore := stepStatus(t, f)
	unmergedBefore := stepMustGit(t, f.integ, "ls-files", "-u")
	bodyBefore, err := os.ReadFile(filepath.Join(f.integ, "base.txt"))
	if err != nil {
		t.Fatal(err)
	}
	_, runErr := RunMergeStep(f.input(), f.seams(card))
	if _, statErr := os.Stat(filepath.Join(f.integ, ".git", "MERGE_HEAD")); statErr != nil {
		t.Fatalf("the step discarded the merge already in progress: MERGE_HEAD is gone (%v)", statErr)
	}
	if after := stepStatus(t, f); after != statusBefore {
		t.Fatalf("the conflicted status must be unchanged by the refusal: before %q, after %q", statusBefore, after)
	}
	if after := stepMustGit(t, f.integ, "ls-files", "-u"); after != unmergedBefore {
		t.Fatalf("the unmerged entries must be unchanged by the refusal: before %q, after %q", unmergedBefore, after)
	}
	if body, readErr := os.ReadFile(filepath.Join(f.integ, "base.txt")); readErr != nil || string(body) != string(bodyBefore) {
		t.Fatalf("the conflicted file must keep its bytes across the refusal: before %q, after %q (err %v)", bodyBefore, body, readErr)
	}
	requireCode(t, runErr, MergeExitWorktreeDirty)
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepRaceMergeHeadAfterProbeIsNotAborted(t *testing.T) {
	// F1 follow-up (sync audit, card t1616): the in-section probe runs before the merge
	// call, so a merge another actor begins after it is not caught there. Here the
	// foreign MERGE_HEAD appears inside the merge call, which is where the race lands:
	// git refuses to start the merge over it, and the step's abort would then discard
	// it. The step must not abort a merge it did not start. It refuses with
	// MergeExitWorktreeDirty, leaves MERGE_HEAD naming the foreign commit, and releases
	// the window.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	tree := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD^{tree}"))
	foreign := strings.TrimSpace(stepMustGit(t, f.integ, "commit-tree", tree, "-m", "foreign merge head"))
	seams := f.seams(card)
	seams.Git = func(args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "merge" && !containsArg(args, "--abort") {
			// The other actor's merge begins here: MERGE_HEAD names its commit, and
			// the real merge call that follows is refused by git over that merge.
			if err := os.WriteFile(filepath.Join(f.integ, ".git", "MERGE_HEAD"), []byte(foreign+"\n"), 0o644); err != nil {
				return "", err
			}
		}
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		return string(out), err
	}
	_, runErr := RunMergeStep(f.input(), seams)
	body, readErr := os.ReadFile(filepath.Join(f.integ, ".git", "MERGE_HEAD"))
	if readErr != nil {
		t.Fatalf("the step aborted a merge another actor began after the probe: MERGE_HEAD is gone (%v)", readErr)
	}
	if got := strings.TrimSpace(string(body)); got != foreign {
		t.Fatalf("MERGE_HEAD must still name the foreign commit %s, got %q", foreign, got)
	}
	requireCode(t, runErr, MergeExitWorktreeDirty)
	requireNoMerge(t, f)
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepRaceSameShaMergeHeadAfterProbeIsNotAborted(t *testing.T) {
	// F1 same-SHA (sync audit, card t1616): another actor begins a merge of the SAME
	// pinned SHA after the in-section probe. Its MERGE_HEAD names the pin, so the
	// different-SHA check does not see it, and the message is the step's own, so the
	// message does not separate the two merges either. Git refuses our merge call
	// while that merge is unfinished (exit 128) and writes nothing, so the step must
	// not abort a merge it did not start. It refuses with MergeExitWorktreeDirty,
	// leaves MERGE_HEAD naming the pin, and keeps the other actor's staged result.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	merged := false
	var foreignStatus string
	seams.Git = func(args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "merge" && !containsArg(args, "--abort") {
			// The other actor's merge begins here, with the step's own message: it
			// stages the card's change and leaves MERGE_HEAD naming the pin. Our merge
			// call below then runs over it.
			msg := ""
			for i := 0; i+1 < len(args); i++ {
				if args[i] == "-m" {
					msg = args[i+1]
				}
			}
			other := exec.Command("git", "merge", "--no-commit", "--no-ff", "-m", msg, f.cardSHA)
			other.Dir = f.integ
			if out, err := other.CombinedOutput(); err != nil {
				t.Fatalf("fixture: the other actor's merge must begin: %v: %s", err, out)
			}
			merged = true
			foreignStatus = stepStatus(t, f)
		}
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		if err != nil {
			// The production runner reports a non-zero exit as *GitExitError; the seam
			// reports the same shape, so the step classifies it as production does.
			code := 1
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				code = exitErr.ExitCode()
			}
			return string(out), &factorylane.GitExitError{ExitCode: code, Stderr: string(out)}
		}
		return string(out), nil
	}
	_, runErr := RunMergeStep(f.input(), seams)
	if !merged {
		t.Fatal("fixture: the step never reached its merge call")
	}
	body, readErr := os.ReadFile(filepath.Join(f.integ, ".git", "MERGE_HEAD"))
	if readErr != nil {
		t.Fatalf("the step aborted a merge another actor began after the probe: MERGE_HEAD is gone (%v)", readErr)
	}
	if got := strings.TrimSpace(string(body)); got != f.cardSHA {
		t.Fatalf("MERGE_HEAD must still name the other actor's commit %s, got %q", f.cardSHA, got)
	}
	if after := stepStatus(t, f); after != foreignStatus {
		t.Fatalf("the other actor's staged result must survive the refusal: before %q, after %q", foreignStatus, after)
	}
	requireCode(t, runErr, MergeExitWorktreeDirty)
	requireNoMerge(t, f)
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepStartFailureSameShaMergeIsNotAborted(t *testing.T) {
	// F1 start-failure (sync audit, card t1616): our merge call fails to START while
	// another actor's merge of the SAME pinned SHA is in progress. The production runner
	// reports a start failure as exit 1: ExecGitRunner.Git defaults the code to 1 for any
	// error that is not an *exec.ExitError. mergeCallDidNotStop reads exit 1 as "this call
	// began a merge and stopped", so the step falls through to `git merge --abort` and
	// discards the other actor's merge. A call that never started began nothing, so the
	// step must refuse with MergeExitWorktreeDirty and leave MERGE_HEAD naming the pin.
	// The merge call runs through the real runner, and PATH is emptied around that call
	// only, so the abort that follows starts normally (the transient case).
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	runner := factorylane.ExecGitRunner{Dir: f.integ}
	merged := false
	var foreignStatus string
	seams.Git = func(args ...string) (string, error) {
		if len(args) < 1 || args[0] != "merge" || containsArg(args, "--abort") {
			return runner.Git(args...)
		}
		// The other actor's merge of the pin begins here, with the step's own message: it
		// stages the card's change and leaves MERGE_HEAD naming the pin.
		msg := ""
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "-m" {
				msg = args[i+1]
			}
		}
		other := exec.Command("git", "merge", "--no-commit", "--no-ff", "-m", msg, f.cardSHA)
		other.Dir = f.integ
		if out, err := other.CombinedOutput(); err != nil {
			t.Fatalf("fixture: the other actor's merge must begin: %v: %s", err, out)
		}
		merged = true
		foreignStatus = stepStatus(t, f)
		// Our merge call fails to START: PATH names an empty directory, so git cannot be
		// found. PATH is restored before any other call.
		prev := os.Getenv("PATH")
		t.Setenv("PATH", t.TempDir())
		if _, lookErr := exec.LookPath("git"); lookErr == nil {
			t.Fatal("fixture: git must not resolve on the emptied PATH")
		}
		out, err := runner.Git(args...)
		t.Setenv("PATH", prev)
		return out, err
	}
	_, runErr := RunMergeStep(f.input(), seams)
	if !merged {
		t.Fatal("fixture: the step never reached its merge call")
	}
	body, readErr := os.ReadFile(filepath.Join(f.integ, ".git", "MERGE_HEAD"))
	if readErr != nil {
		t.Fatalf("the step aborted a same-SHA merge another actor began after the probe, when our merge call failed to start: MERGE_HEAD is gone (%v)", readErr)
	}
	if got := strings.TrimSpace(string(body)); got != f.cardSHA {
		t.Fatalf("MERGE_HEAD must still name the other actor's commit %s, got %q", f.cardSHA, got)
	}
	if after := stepStatus(t, f); after != foreignStatus {
		t.Fatalf("the other actor's staged result must survive the refusal: before %q, after %q", foreignStatus, after)
	}
	requireCode(t, runErr, MergeExitWorktreeDirty)
	requireNoMerge(t, f)
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepStartFailureWithoutMergeHeadRefusesWithoutAbort(t *testing.T) {
	// F1 start-failure, no MERGE_HEAD (sync audit, card t1616): our merge call fails to
	// start, so git began no merge and no MERGE_HEAD exists. The step must refuse with
	// MergeExitOther and must not run `git merge --abort` on the strength of a call that
	// began nothing. The refusal releases the window.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	aborted := false
	seams.Git = func(args ...string) (string, error) {
		if len(args) >= 1 && args[0] == "merge" {
			if containsArg(args, "--abort") {
				aborted = true
			} else {
				// A process-start failure: the exec error, which carries no exit status.
				return "", &exec.Error{Name: "git", Err: exec.ErrNotFound}
			}
		}
		runner := exec.Command("git", args...)
		runner.Dir = f.integ
		out, err := runner.CombinedOutput()
		return string(out), err
	}
	_, err := RunMergeStep(f.input(), seams)
	if aborted {
		t.Error("the step ran `git merge --abort` although its merge call never started")
	}
	requireCode(t, err, MergeExitOther)
	if _, statErr := os.Stat(filepath.Join(f.integ, ".git", "MERGE_HEAD")); !os.IsNotExist(statErr) {
		t.Fatalf("no MERGE_HEAD may exist after a merge call that never started: %v", statErr)
	}
	requireNoMerge(t, f)
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeStepHookStoppedOwnMergeIsAbortedCause6(t *testing.T) {
	// Characterization (card t1616, F1 same-SHA): the step's own merge can begin and
	// stop without a conflict. A pre-merge-commit hook that refuses the merge commit
	// leaves MERGE_HEAD naming the pin, and git exits 1 after the merge began. That
	// merge is this call's own, so the step aborts it and the worktree reads clean
	// (cause 6). The same-SHA refusal must not reach it, because exit 1 marks a merge
	// that began here.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	hook := filepath.Join(f.integ, ".git", "hooks", "pre-merge-commit")
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hook, []byte("#!/bin/sh\necho 'stopped by the fixture hook' >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := RunMergeStep(f.input(), f.seams(card))
	requireCode(t, err, MergeExitMergeFailed)
	if _, statErr := os.Stat(filepath.Join(f.integ, ".git", "MERGE_HEAD")); !os.IsNotExist(statErr) {
		t.Fatalf("the abort must clear the merge this call began: %v", statErr)
	}
	if status := stepStatus(t, f); status != "" {
		t.Fatalf("the worktree must read clean after the abort: %q", status)
	}
	requireNoMerge(t, f)
	requireWindowReleasedAndCPromoted(t, f)
}

// stepRelocateToPrimary turns the separate-surface fixture into the primary surface.
// The primary is the integration worktree and the root the window record and the
// re-measure store live under at once (integration_merge.go resolves the primary as
// the integration worktree), so the record store moves under the integration
// worktree's .moai/state, and git excludes .moai/ the way a real primary's ignore
// rules do. The card worktree stays a sibling outside the primary, so its checkout is
// not an untracked entry there. f.root then names the primary, and f.input() carries
// the primary's operands.
func stepRelocateToPrimary(t *testing.T, f *stepFixture) {
	t.Helper()
	if err := os.Rename(filepath.Join(f.root, ".moai"), filepath.Join(f.integ, ".moai")); err != nil {
		t.Fatal(err)
	}
	exclude := filepath.Join(f.integ, ".git", "info", "exclude")
	file, err := os.OpenFile(exclude, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(".moai/\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	f.root = f.integ
}

func TestMergeStepPrimaryDirtyBeforeMergeRefusesWithoutMerge(t *testing.T) {
	// SPEC-LOCAL-MAIN-FLOW-001 (card t1616): on the primary surface the dirty precondition is
	// read once, before the mutation section. Another session that does not take the mutation
	// lock can create an untracked file in the primary after that read, and the merge then
	// runs over the dirty primary; only the post-merge checks would see it, after main has
	// advanced. The gap is modelled at the first git call through the seam after the clean
	// check: the unfinished-merge probe's `rev-parse --git-path MERGE_HEAD`. The clean check
	// runs through exec, not the seam, so the file lands after it and before any merge call,
	// and it does not sit inside the merge call itself, which would skip the gap the
	// re-read has to close. The step must re-read the primary inside the section,
	// immediately before the merge, and refuse with MergeExitWorktreeDirty without merging.
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	stepRelocateToPrimary(t, f)
	seams := f.seams(card)
	runner := factorylane.ExecGitRunner{Dir: f.integ}
	dirtied := false
	seams.Git = func(args ...string) (string, error) {
		if !dirtied {
			dirtied = true
			stepWriteFile(t, f.integ, "other-session.txt", "another session's bytes\n")
		}
		return runner.Git(args...)
	}
	_, err := RunMergeStep(f.input(), seams)
	if !dirtied {
		t.Fatal("fixture: the step never reached a git call after its clean check")
	}
	if code, ok := MergeExitCode(err); !ok || code != MergeExitWorktreeDirty {
		t.Errorf("a primary dirtied after its clean check must refuse with MergeExitWorktreeDirty (%d), got code %d (ok %v, err %v)", MergeExitWorktreeDirty, code, ok, err)
	}
	if head := strings.TrimSpace(stepMustGit(t, f.integ, "rev-parse", "HEAD")); head != f.record.Base {
		t.Errorf("the refusal must perform no merge: HEAD %s, want the base %s", head, f.record.Base)
	}
	if body, readErr := os.ReadFile(filepath.Join(f.integ, "other-session.txt")); readErr != nil || string(body) != "another session's bytes\n" {
		t.Errorf("the other session's untracked file must stay in place: body %q (err %v)", body, readErr)
	}
	requireWindowReleasedAndCPromoted(t, f)
}
