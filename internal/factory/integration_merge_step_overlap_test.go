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
// once it has returned: the status change that S-abort is measured against.
func stepGitFailingMergeThenAbort(f *stepFixture, after func()) func(args ...string) (string, error) {
	return func(args ...string) (string, error) {
		isMerge := len(args) >= 1 && args[0] == "merge"
		if isMerge && !containsArg(args, "--abort") {
			return "", errors.New("simulated merge failure")
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
