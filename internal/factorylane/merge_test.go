package factorylane

// merge_test.go — the lane-direct merge condition triple (AC-FLA-009).
//
// Two layers: scripted-runner unit tests for every named failure and the
// record/verify contract, and one real-repository test (t.TempDir) proving
// the load-bearing design fact — that the pre-merge tree-identity form
// (merge-tree result == the card branch tree) coincides with t1241's literal
// HEAD^{tree} == HEAD^2^{tree} on the resulting merge commit. The real repo
// test performs merges only inside its own throwaway repository, never into
// any develop branch.

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeGit scripts the GitRunner seam: the merge-tree reply and a per-command
// rev-parse answer table. Anything unscripted fails the test loudly.
type fakeGit struct {
	mergeTreeOut string
	mergeTreeErr error
	revParse     map[string]string // joined args -> stdout
	seen         []string
}

func (f *fakeGit) Git(args ...string) (string, error) {
	f.seen = append(f.seen, strings.Join(args, " "))
	switch args[0] {
	case "merge-tree":
		return f.mergeTreeOut, f.mergeTreeErr
	case "rev-parse":
		if v, ok := f.revParse[strings.Join(args, " ")]; ok {
			return v, nil
		}
		return "", fmt.Errorf("fakeGit: no rev-parse script for %v", args)
	}
	return "", fmt.Errorf("fakeGit: unexpected git invocation %v", args)
}

func exitErr(code int, stdout string) error {
	return &GitExitError{ExitCode: code, Stdout: stdout}
}

// tripleFixture builds a SPEC directory whose progress.md §E.4 carries the
// given sync_status value, and a fake runner with a clean, identity-preserving
// merge (the happy shape every per-condition test then perturbs).
func tripleFixture(t *testing.T, syncStatus string) (MergeTripleInput, *fakeGit) {
	t.Helper()
	specDir := t.TempDir()
	section := ""
	if syncStatus != "" {
		section = fmt.Sprintf("## §E.4 Sync-phase Audit-Ready Signal\n\nsync_status: %s\n", syncStatus)
	}
	progress := "## §E.2 Run-phase Evidence\n\n...\n\n" + section
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progress), 0o644); err != nil {
		t.Fatal(err)
	}
	in := MergeTripleInput{
		Lane:    "lane-1",
		Card:    "t9001",
		SpecDir: specDir,
		Branch:  "WT-card",
		Develop: "develop",
		RepoDir: t.TempDir(),
	}
	fake := &fakeGit{
		// Clean merge; the result tree equals the branch tree (the absorbed
		// shape) so tree-identity passes.
		mergeTreeOut: "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b\n",
		revParse: map[string]string{
			"rev-parse WT-card^{tree}": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b\n",
		},
	}
	return in, fake
}

// All three conditions pass on the absorbed shape, and the run records the
// three named checks with all_passed set.
func TestEvaluateMergeTriple_AllPassRecordsThreeChecks(t *testing.T) {
	in, fake := tripleFixture(t, "complete")

	run, err := EvaluateMergeTriple(in, fake)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple errored: %v", err)
	}
	if !run.AllPassed {
		t.Fatalf("all_passed false on the passing shape: %+v", run)
	}
	if run.FailedCondition != "" {
		t.Fatalf("failed_condition set on the passing shape: %q", run.FailedCondition)
	}
	if len(run.Checks) != 4 {
		t.Fatalf("got %d recorded checks, want 4 (the REQ-MWQ-021 record condition joined the triple): %+v", len(run.Checks), run.Checks)
	}
	for i, want := range []string{CheckSyncAudit, CheckConflictFree, CheckTreeIdentity} {
		if run.Checks[i].Name != want {
			t.Errorf("check[%d] = %q, want %q", i, run.Checks[i].Name, want)
		}
		if !run.Checks[i].Passed {
			t.Errorf("check[%d] %q did not pass on the passing shape: %s", i, want, run.Checks[i].Detail)
		}
	}
	if run.Lane != "lane-1" || run.Card != "t9001" || run.Branch != "WT-card" || run.Develop != "develop" {
		t.Errorf("run identity fields not carried: %+v", run)
	}
	// One stamping point: the evaluation carries no timestamp — the store
	// stamps the record at write time (the store-clock contract the
	// VerifyRunBeforeAcquire ordering reads).
	if !run.CheckedAt.IsZero() {
		t.Errorf("evaluated run pre-stamped checked_at %s, want zero until recorded", run.CheckedAt)
	}
}

// Condition (a): §E.4 not complete — the refusal names sync-audit, and the
// annotated "complete (3-phase close …)" spelling still passes.
func TestEvaluateMergeTriple_SyncAuditFailureNamed(t *testing.T) {
	cases := []struct {
		name       string
		syncStatus string // "" = §E.4 section missing entirely
	}{
		{name: "audit-ready is not closed", syncStatus: "audit-ready"},
		{name: "empty sync_status", syncStatus: "-"},
		{name: "section missing", syncStatus: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, fake := tripleFixture(t, tc.syncStatus)

			run, err := EvaluateMergeTriple(in, fake)
			if err != nil {
				t.Fatalf("EvaluateMergeTriple errored: %v", err)
			}
			if run.AllPassed {
				t.Fatalf("all_passed true while sync phase is not closed: %+v", run)
			}
			if run.FailedCondition != CheckSyncAudit {
				t.Fatalf("failing condition = %q, want %q", run.FailedCondition, CheckSyncAudit)
			}
			var chk *MergeCheck
			for i := range run.Checks {
				if run.Checks[i].Name == CheckSyncAudit {
					chk = &run.Checks[i]
				}
			}
			if chk == nil || chk.Passed {
				t.Fatalf("sync-audit check not recorded as failed: %+v", run.Checks)
			}
		})
	}

	// The annotated spelling passes: first whitespace token decides.
	in, fake := tripleFixture(t, "complete (3-phase close — in-progress → implemented → merged into the single sync commit)")
	run, err := EvaluateMergeTriple(in, fake)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple errored: %v", err)
	}
	if !run.AllPassed {
		t.Fatalf("annotated complete spelling refused: %+v", run)
	}
}

// Condition (b): merge-tree reports conflicts (exit 1) — the refusal names
// conflict-free and carries the conflicted paths, and no acquire-side
// action could have been justified by the run.
func TestEvaluateMergeTriple_ConflictNamed(t *testing.T) {
	in, fake := tripleFixture(t, "complete")
	fake.mergeTreeErr = exitErr(1, "0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e\n100644 1a2b 3c4d\tCHANGELOG.md\n100644 5e6f 7a8b\tinternal/config/defaults.go\n")

	run, err := EvaluateMergeTriple(in, fake)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple errored: %v", err)
	}
	if run.AllPassed || run.FailedCondition != CheckConflictFree {
		t.Fatalf("conflict not named: all_passed=%t failed=%q", run.AllPassed, run.FailedCondition)
	}
	var chk *MergeCheck
	for i := range run.Checks {
		if run.Checks[i].Name == CheckConflictFree {
			chk = &run.Checks[i]
		}
	}
	if chk == nil || chk.Passed {
		t.Fatalf("conflict-free check not recorded as failed: %+v", run.Checks)
	}
	if !strings.Contains(chk.Detail, "CHANGELOG.md") {
		t.Errorf("conflict detail does not carry the conflicted path: %s", chk.Detail)
	}
}

// Condition (b) fail-closed: a merge-tree tool failure (not an exit-1
// conflict verdict) refuses too — a check that cannot run is not a pass.
func TestEvaluateMergeTriple_MergeTreeToolFailureFailsClosed(t *testing.T) {
	in, fake := tripleFixture(t, "complete")
	fake.mergeTreeErr = &GitExitError{ExitCode: 129, Stderr: "usage: git merge-tree"}

	run, err := EvaluateMergeTriple(in, fake)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple errored: %v", err)
	}
	if run.AllPassed {
		t.Fatalf("tool failure read as a pass")
	}
	if run.FailedCondition != CheckConflictFree {
		t.Fatalf("failing condition = %q, want %q", run.FailedCondition, CheckConflictFree)
	}
}

// Condition (c), pre-merge form: the merge-tree result differs from the card
// branch tree (develop carries changes the branch does not) — the refusal
// names tree-identity.
func TestEvaluateMergeTriple_TreeIdentityPreMergeFailureNamed(t *testing.T) {
	in, fake := tripleFixture(t, "complete")
	fake.mergeTreeOut = "9f8e7d6c5b4a39281706f5e4d3c2b1a098765432\n" // != branch tree

	run, err := EvaluateMergeTriple(in, fake)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple errored: %v", err)
	}
	if run.AllPassed || run.FailedCondition != CheckTreeIdentity {
		t.Fatalf("tree identity mismatch not named: all_passed=%t failed=%q", run.AllPassed, run.FailedCondition)
	}
}

// Condition (c), literal form: with a prepared merge commit the check reads
// HEAD^{tree} vs HEAD^2^{tree} through rev-parse on that commit.
func TestEvaluateMergeTriple_TreeIdentityLiteralFormOnMergeCommit(t *testing.T) {
	in, fake := tripleFixture(t, "complete")
	in.MergeCommit = "abcabcabcabcabcabcabcabcabcabcabcabcabc1"
	fake.revParse["rev-parse abcabcabcabcabcabcabcabcabcabcabcabcabc1^{tree}"] = "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b\n"
	fake.revParse["rev-parse abcabcabcabcabcabcabcabcabcabcabcabcabc1^2^{tree}"] = "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b\n"

	run, err := EvaluateMergeTriple(in, fake)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple errored: %v", err)
	}
	if !run.AllPassed {
		t.Fatalf("identical merge-commit trees refused: %+v", run)
	}
	var sawLiteral bool
	for _, call := range fake.seen {
		if strings.Contains(call, "^2^{tree}") {
			sawLiteral = true
		}
	}
	if !sawLiteral {
		t.Errorf("literal form did not query ^2^{tree} via rev-parse: %v", fake.seen)
	}

	// Diverging trees on the merge commit fail and name the condition.
	fake2 := &fakeGit{
		mergeTreeOut: "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b\n",
		revParse: map[string]string{
			"rev-parse abcabcabcabcabcabcabcabcabcabcabcabcabc1^{tree}":   "ffffffffffffffffffffffffffffffffffffffff\n",
			"rev-parse abcabcabcabcabcabcabcabcabcabcabcabcabc1^2^{tree}": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b\n",
		},
	}
	in.MergeCommit = "abcabcabcabcabcabcabcabcabcabcabcabcabc1"
	run, err = EvaluateMergeTriple(in, fake2)
	if err != nil {
		t.Fatalf("EvaluateMergeTriple errored: %v", err)
	}
	if run.AllPassed || run.FailedCondition != CheckTreeIdentity {
		t.Fatalf("differing merge-commit trees not named: all_passed=%t failed=%q", run.AllPassed, run.FailedCondition)
	}
}

// AC-FLA-009 second half: the recorded run counts only when all three
// passed and it was recorded BEFORE the window acquire timestamp.
func TestVerifyRunBeforeAcquire(t *testing.T) {
	passed := &MergeCheckRun{AllPassed: true, CheckedAt: time.Now().UTC().Add(-time.Minute)}
	if ok, why := VerifyRunBeforeAcquire(passed, time.Now().UTC()); !ok {
		t.Fatalf("passed run before acquire refused: %s", why)
	}

	if ok, why := VerifyRunBeforeAcquire(nil, time.Now().UTC()); ok {
		t.Fatalf("missing run accepted")
	} else if !strings.Contains(why, "no recorded merge check run") {
		t.Errorf("missing-run refusal should name the absence: %s", why)
	}

	failed := &MergeCheckRun{AllPassed: false, FailedCondition: CheckConflictFree, CheckedAt: time.Now().UTC().Add(-time.Minute)}
	if ok, _ := VerifyRunBeforeAcquire(failed, time.Now().UTC()); ok {
		t.Fatalf("failed run accepted")
	}

	later := &MergeCheckRun{AllPassed: true, CheckedAt: time.Now().UTC().Add(time.Minute)}
	if ok, why := VerifyRunBeforeAcquire(later, time.Now().UTC()); ok {
		t.Fatalf("run recorded after the acquire accepted")
	} else if !strings.Contains(why, "not before") {
		t.Errorf("ordering refusal should name the ordering: %s", why)
	}
}

// The store round trip: RecordMergeCheckRun persists one file per run under
// merge-checks/<lane>/, and LatestMergeCheckRun returns the newest run for
// the lane and card.
func TestStoreMergeCheckRunRoundTrip(t *testing.T) {
	clock := &FakeClock{Current: time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)}
	store := NewStore(t.TempDir(), clock)

	run := MergeCheckRun{
		Lane: "lane-2", Card: "t9002", Branch: "WT-card", Develop: "develop",
		Checks:    []MergeCheck{{Name: CheckSyncAudit, Passed: true, Detail: "ok"}},
		AllPassed: true,
	}
	stored, err := store.RecordMergeCheckRun(run)
	if err != nil {
		t.Fatalf("RecordMergeCheckRun: %v", err)
	}
	if !stored.CheckedAt.Equal(clock.Current) {
		t.Errorf("checked_at = %s, want the store clock %s", stored.CheckedAt, clock.Current)
	}

	clock.Current = clock.Current.Add(time.Second)
	second := run
	second.AllPassed = false
	second.FailedCondition = CheckConflictFree
	if _, err := store.RecordMergeCheckRun(second); err != nil {
		t.Fatalf("RecordMergeCheckRun (second): %v", err)
	}

	got, err := store.LatestMergeCheckRun("lane-2", "t9002")
	if err != nil {
		t.Fatalf("LatestMergeCheckRun: %v", err)
	}
	if got.AllPassed {
		t.Errorf("latest run = the first run, want the second: %+v", got)
	}
	if missing, err := store.LatestMergeCheckRun("lane-2", "t9999"); err != nil || missing != nil {
		t.Errorf("unknown card: run=%+v err=%v, want nil,nil", missing, err)
	}
}

// gitInTempRepo initializes a throwaway repository and returns a helper that
// runs git inside it, failing the test on error.
func gitInTempRepo(t *testing.T) (dir string, run func(args ...string) string) {
	t.Helper()
	dir = t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
		return string(out)
	}
	git("init", "-b", "develop")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	return dir, git
}

// The load-bearing design fact: the pre-merge tree-identity form (merge-tree
// result == the card branch tree) decides exactly what t1241's literal
// HEAD^{tree} == HEAD^2^{tree} decides on the resulting merge commit — both
// before and after the card branch absorbs the integration branch.
func TestMergeTriple_PreMergeFormMatchesLiteralPostMergeForm(t *testing.T) {
	dir, git := gitInTempRepo(t)
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// develop: one base commit.
	write("base.txt", "base\n")
	git("add", ".")
	git("commit", "-m", "base")
	git("checkout", "-b", "WT-card")
	write("card.txt", "card\n")
	git("add", ".")
	git("commit", "-m", "card work")
	cardTree := strings.TrimSpace(git("rev-parse", "WT-card^{tree}"))

	// develop diverges: the branch does NOT contain develop.
	git("checkout", "develop")
	write("dev.txt", "dev\n")
	git("add", ".")
	git("commit", "-m", "develop work")

	runner := ExecGitRunner{Dir: dir}
	treeOf := func(rev string) string {
		out, err := runner.Git("rev-parse", rev)
		if err != nil {
			t.Fatalf("rev-parse %s: %v", rev, err)
		}
		return strings.TrimSpace(out)
	}

	// Unabsorbed: merge-tree's result carries dev.txt, so it differs from the
	// branch tree — the pre-merge form refuses.
	preMergeTree, err := runner.Git("merge-tree", "--write-tree", "develop", "WT-card")
	if err != nil {
		t.Fatalf("merge-tree (unabsorbed): %v", err)
	}
	if strings.TrimSpace(preMergeTree) == cardTree {
		t.Fatalf("unabsorbed pre-merge tree unexpectedly equals the branch tree")
	}
	if identity, why := checkTreeIdentityPreMerge(runner, "develop", "WT-card"); identity {
		t.Fatalf("pre-merge form passed an unabsorbed branch: %s", why)
	}

	// Now absorb develop into the card branch, and merge --no-ff back into
	// develop inside the throwaway repo.
	git("checkout", "WT-card")
	git("merge", "--no-ff", "-m", "absorb develop", "develop")
	absorbedCardTree := treeOf("WT-card^{tree}")
	preMergeTree, err = runner.Git("merge-tree", "--write-tree", "develop", "WT-card")
	if err != nil {
		t.Fatalf("merge-tree (absorbed): %v", err)
	}
	if strings.TrimSpace(preMergeTree) != absorbedCardTree {
		t.Fatalf("absorbed pre-merge tree %s != branch tree %s", strings.TrimSpace(preMergeTree), absorbedCardTree)
	}

	git("checkout", "develop")
	git("merge", "--no-ff", "-m", "merge card into develop", "WT-card")

	// The literal form on the real merge commit agrees with the pre-merge
	// form: HEAD^{tree} == HEAD^2^{tree}.
	literalTree, literalSecondParent := treeOf("HEAD^{tree}"), treeOf("HEAD^2^{tree}")
	if literalTree != literalSecondParent {
		t.Fatalf("literal form: HEAD^{tree} %s != HEAD^2^{tree} %s on the absorbed merge", literalTree, literalSecondParent)
	}
	if identity, why := checkTreeIdentityPreMerge(runner, "develop", "WT-card"); !identity {
		t.Fatalf("pre-merge form refused the absorbed merge: %s", why)
	}

	// And a tool failure of the runner (git absent) is a check failure, not
	// a pass — fail-closed on both forms.
	if ok, _ := checkTreeIdentityOnMergeCommit(failingRunner{}, "nonexistent"); ok {
		t.Fatalf("literal form passed on a runner error")
	}
	if ok, _ := checkTreeIdentityPreMerge(failingRunner{}, "develop", "WT-card"); ok {
		t.Fatalf("pre-merge form passed on a runner error")
	}
}

// failingRunner fails every invocation — the fail-closed control.
type failingRunner struct{}

func (failingRunner) Git(args ...string) (string, error) {
	return "", errors.New("git unavailable")
}

// AC-FLA-011 as a predicate: a live hold for this lane covering the moment
// proceeds; every missing property refuses and names itself.
func TestWindowCoversMerge(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	live := WindowSnapshot{
		Held: true, Live: true, HolderName: "lane-5",
		AcquiredAt: now.Add(-time.Minute), KnownAt: true,
	}
	if ok, why := WindowCoversMerge(live, "lane-5", now); !ok {
		t.Fatalf("live hold for the lane refused: %s", why)
	}

	cases := []struct {
		name    string
		snap    WindowSnapshot
		lane    string
		wantIn  string
		wantNot string
	}{
		{name: "no record", snap: WindowSnapshot{}, lane: "lane-5", wantIn: "no integration acquire record"},
		{name: "stale holder", snap: WindowSnapshot{Held: true, Live: false, HolderName: "lane-5", AcquiredAt: now.Add(-time.Minute), KnownAt: true}, lane: "lane-5", wantIn: "not a live hold"},
		{name: "foreign lane", snap: live, lane: "lane-6", wantIn: "held by lane-5", wantNot: "PROCEED"},
		{name: "no timestamp", snap: WindowSnapshot{Held: true, Live: true, HolderName: "lane-5"}, lane: "lane-5", wantIn: "no parseable acquired-at"},
		{name: "acquired after the moment", snap: WindowSnapshot{Held: true, Live: true, HolderName: "lane-5", AcquiredAt: now.Add(time.Minute), KnownAt: true}, lane: "lane-5", wantIn: "after the merge moment"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, why := WindowCoversMerge(tc.snap, tc.lane, now)
			if ok {
				t.Fatalf("negative case %q proceeded", tc.name)
			}
			if !strings.Contains(why, tc.wantIn) || (tc.wantNot != "" && strings.Contains(why, tc.wantNot)) {
				t.Errorf("refusal %q does not match wantIn=%q wantNot=%q", why, tc.wantIn, tc.wantNot)
			}
		})
	}
}

// ExecGitRunner wraps a failing git exit into *GitExitError carrying the exit
// code, so the check layer can read merge-tree's exit-1 verdict.
func TestExecGitRunnerWrapsExitError(t *testing.T) {
	dir, _ := gitInTempRepo(t)
	runner := ExecGitRunner{Dir: dir}
	_, err := runner.Git("rev-parse", "definitely-not-a-ref-9001")
	var exitErr *GitExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *GitExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode != 128 {
		t.Errorf("exit code = %d, want 128", exitErr.ExitCode)
	}
	// GitExitError with no captured output still names the exit code.
	empty := &GitExitError{ExitCode: 3}
	if msg := empty.Error(); !strings.Contains(msg, "git exited 3") {
		t.Errorf("empty-output Error() = %q, want the exit code named", msg)
	}
}

// A process that cannot start has no exit status, so ExecGitRunner must not
// invent one (SPEC-LOCAL-MAIN-FLOW-001 F1, card t1616). It reported every
// non-ExitError failure as exit 1, and the merge step reads exit 1 as "git began
// a merge and stopped", so it aborted a merge another actor had begun.
func TestExecGitRunnerStartFailureIsNotAnExitStatus(t *testing.T) {
	intact := os.Getenv("PATH")

	// Start failure: an empty PATH, so git cannot be found.
	t.Setenv("PATH", t.TempDir())
	_, err := ExecGitRunner{Dir: t.TempDir()}.Git("status")
	if err == nil {
		t.Fatal("git status on an empty PATH must fail")
	}
	var exitErr *GitExitError
	if errors.As(err, &exitErr) {
		t.Fatalf("a start failure is reported as an exit status: GitExitError{ExitCode: %d}, want the error as is: %v", exitErr.ExitCode, err)
	}

	// Positive control (passes before and after the fix): PATH restored, git runs
	// and exits non-zero, and that real status is still reported as one.
	t.Setenv("PATH", intact)
	dir, _ := gitInTempRepo(t)
	_, err = ExecGitRunner{Dir: dir}.Git("rev-parse", "--verify", "no-such-ref")
	if !errors.As(err, &exitErr) {
		t.Fatalf("control: a real non-zero exit must stay a *GitExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode != 128 {
		t.Errorf("control: exit code = %d, want git's 128", exitErr.ExitCode)
	}
}

// Two runs recorded inside one clock tick both persist: the filename stamp
// bumps so the second never overwrites the first.
func TestRecordMergeCheckRunSameTickCollision(t *testing.T) {
	clock := &FakeClock{Current: time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)}
	store := NewStore(t.TempDir(), clock)
	run := MergeCheckRun{Lane: "lane-3", Card: "t9003", AllPassed: true}
	if _, err := store.RecordMergeCheckRun(run); err != nil {
		t.Fatalf("first record: %v", err)
	}
	if _, err := store.RecordMergeCheckRun(run); err != nil {
		t.Fatalf("same-tick second record: %v", err)
	}
	got, err := store.LatestMergeCheckRun("lane-3", "t9003")
	if err != nil || got == nil {
		t.Fatalf("latest after two same-tick records: run=%+v err=%v", got, err)
	}
}

// An unreadable merge-check record is an error, never a silent "no runs".
func TestLatestMergeCheckRunUnreadableRecordErrors(t *testing.T) {
	clock := &FakeClock{Current: time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)}
	root := t.TempDir()
	store := NewStore(root, clock)
	if err := os.MkdirAll(filepath.Join(root, DefaultStateRoot, "merge-checks", "lane-4"), 0o755); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(root, DefaultStateRoot, "merge-checks", "lane-4", "chk-0000000000000000001.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.LatestMergeCheckRun("lane-4", "t9004"); err == nil {
		t.Fatalf("corrupt record read as no runs")
	}
}

// A progress.md that cannot be read (here: a directory in its place) is a
// named failure, not a pass — the fail-closed reading path.
func TestCheckSyncAuditReadErrorFailsClosed(t *testing.T) {
	specDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(specDir, "progress.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	detail, ok := checkSyncAudit(specDir)
	if ok {
		t.Fatalf("unreadable progress.md read as a pass")
	}
	if !strings.Contains(detail, "read progress.md") {
		t.Errorf("detail does not name the read failure: %s", detail)
	}
}
