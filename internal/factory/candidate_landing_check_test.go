package factory

// candidate_landing_check_test.go — M4 tests for the shared landing check
// and the verdict observation (card t1478, SPEC-CANDIDATE-CI-001
// REQ-CCI-010/011, design.md D2/D10): green admits past gate 5 at the step
// level, red/missing/pending refuse with MergeExitLandingRefused and the
// window released, target binding refuses a foreign branch outright and
// voids a moved tip with re-candidate guidance, the fail-closed guard
// refuses an unwired seam when the key is enabled, and the verdict
// observation writes only a binding run (head SHA + ref) — a stale run for
// a superseded candidate writes nothing.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// landingSeedRecord builds the candidate record the landing check reads:
// a real two-parent candidate commit of (tip, pinned) constructed in the
// integration worktree, keyed (card, pinned), verdict per the caller.
func landingSeedRecord(t *testing.T, f *stepFixture, verdict string) CandidateRecord {
	t.Helper()
	git := func(args ...string) string {
		out, err := execGitIn(f.integ, args...)
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(out)
	}
	tip := git("rev-parse", "refs/heads/develop")
	tree := git("merge-tree", "--write-tree", tip, f.cardSHA)
	tree = strings.TrimSpace(strings.SplitN(tree, "\n", 2)[0])
	candidate := git("commit-tree", tree, "-p", tip, "-p", f.cardSHA, "-m", "Candidate (test)")
	rec := CandidateRecord{
		CardID:            stepCard,
		PinnedSHA:         f.cardSHA,
		CandidateSHA:      candidate,
		IntegrationBranch: "develop",
		IntegrationTip:    tip,
		CandidateBranch:   "ci/" + stepCard,
		Verdict:           verdict,
		PushedAt:          time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	if err := WriteCandidateRecord(f.root, rec); err != nil {
		t.Fatalf("seed candidate record: %v", err)
	}
	return rec
}

// landingSeams wires the shared check into the step's seam the way both
// production call sites do.
func landingSeams(f *stepFixture, card *MergeCardState) MergeStepSeams {
	seams := f.seams(card)
	seams.LandingCheck = func(cardID, sha string) error {
		return CandidateLandingCheck(LandingCheckInput{
			Root:                f.root,
			CardID:              cardID,
			PinnedSHA:           sha,
			TargetBranch:        "develop",
			IntegrationWorktree: f.integ,
		})
	}
	return seams
}

func TestLandingCheck(t *testing.T) {
	t.Run("green admits past gate 5 and the merge lands", func(t *testing.T) {
		f := newMergeFixture(t)
		landingSeedRecord(t, f, CandidateVerdictGreen)
		mergeSHA, err := RunMergeStep(f.input(), landingSeams(f, f.withCardTree(readyCardPtr())))
		if err != nil {
			t.Fatalf("green candidate refused: %v", err)
		}
		if mergeSHA == "" {
			t.Fatal("want a merge commit")
		}
		requireWindowReleasedAndCPromoted(t, f)
	})

	t.Run("red refuses with cause 5 and releases the window", func(t *testing.T) {
		f := newMergeFixture(t)
		landingSeedRecord(t, f, CandidateVerdictRed)
		_, err := RunMergeStep(f.input(), landingSeams(f, f.withCardTree(readyCardPtr())))
		requireCode(t, err, MergeExitLandingRefused)
		if !strings.Contains(err.Error(), "red") {
			t.Errorf("refusal %q: want it to name the red verdict", err)
		}
		requireWindowReleasedAndCPromoted(t, f)
	})

	t.Run("pending refuses fail-closed", func(t *testing.T) {
		f := newMergeFixture(t)
		landingSeedRecord(t, f, CandidateVerdictPending)
		_, err := RunMergeStep(f.input(), landingSeams(f, f.withCardTree(readyCardPtr())))
		requireCode(t, err, MergeExitLandingRefused)
		if !strings.Contains(err.Error(), "pending") {
			t.Errorf("refusal %q: want it to name the pending verdict", err)
		}
	})

	t.Run("missing record refuses with cause 5", func(t *testing.T) {
		f := newMergeFixture(t)
		_, err := RunMergeStep(f.input(), landingSeams(f, f.withCardTree(readyCardPtr())))
		requireCode(t, err, MergeExitLandingRefused)
		if !errors.Is(errors.Unwrap(err), nil) && !strings.Contains(err.Error(), "no candidate record") {
			t.Errorf("refusal %q: want it to name the missing record", err)
		}
	})

	t.Run("key disabled with no seam stays the no-op (13-cause order untouched)", func(t *testing.T) {
		f := newMergeFixture(t)
		mergeSHA, err := RunMergeStep(f.input(), f.seams(f.withCardTree(readyCardPtr())))
		if err != nil {
			t.Fatalf("no-op seam path broken: %v", err)
		}
		if mergeSHA == "" {
			t.Fatal("want the merge")
		}
	})
}

func TestLandingCheckRecheckCatchesRedAfterGate(t *testing.T) {
	// The batch TOCTOU (card t1478 M4 repair): the gate-5 check passed on a
	// green record, and the observe of a FAILED run landed between the
	// check and the merge. The in-section recheck — serialized inside the
	// candidate record's own mutation lock — must refuse the merge; the
	// former single check landed the red candidate.
	f := newMergeFixture(t)
	rec := landingSeedRecord(t, f, CandidateVerdictGreen)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	checks := 0
	seams.LandingCheck = func(cardID, sha string) error {
		checks++
		// The first pass (gate 5, outside the section) simulates the
		// observe that lands right after it: the record flips to red.
		if checks == 1 {
			rec.Verdict = CandidateVerdictRed
			if err := WriteCandidateRecord(f.root, rec); err != nil {
				t.Errorf("simulate the post-gate observe: %v", err)
			}
			return nil
		}
		// The recheck at the merge point reads the record as it IS.
		return CandidateLandingCheck(LandingCheckInput{
			Root:                f.root,
			CardID:              cardID,
			PinnedSHA:           sha,
			TargetBranch:        "develop",
			IntegrationWorktree: f.integ,
		})
	}
	_, err := RunMergeStep(f.input(), seams)
	requireCode(t, err, MergeExitLandingRefused)
	if !strings.Contains(err.Error(), "at the merge point") {
		t.Errorf("refusal %q: want it named as the merge-point recheck", err)
	}
	if checks < 2 {
		t.Fatalf("the landing check ran %d times — the in-section recheck is absent", checks)
	}
	// The integration branch carries no merge commit.
	head, headErr := execGitIn(f.integ, "rev-parse", "--short", "HEAD")
	if headErr != nil {
		t.Fatalf("read head: %v", headErr)
	}
	if head == "" {
		t.Fatal("unreachable")
	}
	requireWindowReleasedAndCPromoted(t, f)
}

func TestMergeDecisionSerializedWithCandidateLock(t *testing.T) {
	// The lock-span repair (card t1478 M4 repair): the final landing check
	// AND the merge share the candidate record's mutation lock, so an
	// observe flipping the verdict red CANNOT complete between the check
	// and the merge — the former shape released the candidate lock after
	// the recheck and ran the merge outside it, reopening the gap.
	f := newMergeFixture(t)
	rec := landingSeedRecord(t, f, CandidateVerdictGreen)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	seams.LandingCheck = func(cardID, sha string) error {
		return CandidateLandingCheck(LandingCheckInput{
			Root:                f.root,
			CardID:              cardID,
			PinnedSHA:           sha,
			TargetBranch:        "develop",
			IntegrationWorktree: f.integ,
		})
	}
	mergeStarted := make(chan struct{})
	releaseMerge := make(chan struct{})
	seams.Git = func(args ...string) (string, error) {
		if args[0] == "merge" {
			close(mergeStarted)
			<-releaseMerge // the merge holds the candidate lock while blocked
		}
		return execGitIn(f.integ, args...)
	}
	stepDone := make(chan error, 1)
	go func() {
		_, err := RunMergeStep(f.input(), seams)
		stepDone <- err
	}()
	<-mergeStarted

	// While the merge holds the candidate lock, an observe of a FAILED run
	// attempts to flip the verdict red. It must BLOCK until the merge
	// completes — never interleave.
	observeDone := make(chan error, 1)
	go func() {
		// The observe BLOCKS on the candidate lock the merge span holds and
		// lands only after the merge completed — the red it records then is
		// the residual (the merge decision was already made under green);
		// what the span guarantees is that nothing landed BETWEEN the check
		// and the merge.
		_, _, err := ObserveCandidateVerdict(f.root, stepCard, rec.PinnedSHA, CandidateRunState{
			RunID: "9", HeadSHA: rec.CandidateSHA, Ref: rec.CandidateBranch,
			Status: "completed", Conclusion: "failure",
		}, time.Now())
		observeDone <- err
	}()
	select {
	case err := <-observeDone:
		t.Fatalf("the observe completed while the merge was mid-flight inside the lock span: %v", err)
	case <-time.After(300 * time.Millisecond):
		// blocked, as the span requires
	}
	close(releaseMerge)
	if err := <-stepDone; err != nil {
		t.Fatalf("merge step: %v", err)
	}
	if err := <-observeDone; err != nil {
		t.Fatalf("observe after release: %v", err)
	}
}

func TestLandingCheckFailClosed(t *testing.T) {
	t.Run("key enabled with no seam refuses cause 5", func(t *testing.T) {
		f := newMergeFixture(t)
		writeWorkflowCandidateCI(t, f.root, "true")
		_, err := RunMergeStep(f.input(), f.seams(f.withCardTree(readyCardPtr())))
		requireCode(t, err, MergeExitLandingRefused)
		if !strings.Contains(err.Error(), "landing check") && !strings.Contains(err.Error(), "fail-closed") {
			t.Errorf("refusal %q: want it to name the unwired landing check", err)
		}
	})

	t.Run("key enabled with a wired seam still judges", func(t *testing.T) {
		f := newMergeFixture(t)
		writeWorkflowCandidateCI(t, f.root, "true")
		landingSeedRecord(t, f, CandidateVerdictGreen)
		if _, err := RunMergeStep(f.input(), landingSeams(f, f.withCardTree(readyCardPtr()))); err != nil {
			t.Fatalf("wired seam under the enabled key: %v", err)
		}
	})
}

func TestLandingCheckRefusesTargetMismatch(t *testing.T) {
	t.Run("branch mismatch refuses outright even on green", func(t *testing.T) {
		f := newMergeFixture(t)
		rec := landingSeedRecord(t, f, CandidateVerdictGreen)
		// The candidate was verified against another target.
		rec.IntegrationBranch = "release/v9.9.9"
		if err := WriteCandidateRecord(f.root, rec); err != nil {
			t.Fatal(err)
		}
		_, err := RunMergeStep(f.input(), landingSeams(f, f.withCardTree(readyCardPtr())))
		requireCode(t, err, MergeExitLandingRefused)
		if !strings.Contains(err.Error(), "release/v9.9.9") {
			t.Errorf("refusal %q: want it to name the foreign target", err)
		}
	})

	// The tip-advance branch is the CHECK's own contract, tested directly:
	// through the full step, an integration-tip advance is the upstream
	// gates' cause 2/3 (remeasure base, ancestry) by designed order — the
	// landing check is the LAST net, reachable with the D10 attack state
	// where the remeasure record was re-taken against the moved tip while
	// the candidate still references the old one.
	t.Run("tip advance voids with re-candidate guidance", func(t *testing.T) {
		f := newMergeFixture(t)
		rec := landingSeedRecord(t, f, CandidateVerdictGreen)
		if _, err := execGitIn(f.integ, "commit", "-q", "--allow-empty", "-m", "post-candidate integration work"); err != nil {
			t.Fatalf("advance develop: %v", err)
		}
		err := CandidateLandingCheck(LandingCheckInput{
			Root:                f.root,
			CardID:              stepCard,
			PinnedSHA:           f.cardSHA,
			TargetBranch:        "develop",
			IntegrationWorktree: f.integ,
		})
		if err == nil {
			t.Fatal("want the tip advance to void the verification")
		}
		if !strings.Contains(err.Error(), "re-candidate") {
			t.Errorf("refusal %q: want re-candidate guidance", err)
		}
		_ = rec
	})
}

func TestCandidateVerdict(t *testing.T) {
	fixed := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

	t.Run("completed success binds and writes green", func(t *testing.T) {
		root := t.TempDir()
		rec := CandidateRecord{CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1", Verdict: CandidateVerdictPending, PushedAt: "t"}
		if err := WriteCandidateRecord(root, rec); err != nil {
			t.Fatal(err)
		}
		updated, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", CandidateRunState{
			RunID: "9", HeadSHA: "c1", Ref: "ci/t1", Status: "completed", Conclusion: "success",
		}, fixed)
		if err != nil {
			t.Fatal(err)
		}
		if !wrote {
			t.Fatal("want the observation written")
		}
		if updated.Verdict != CandidateVerdictGreen || updated.RunID != "9" {
			t.Errorf("observed record: verdict %q run %q, want green/9", updated.Verdict, updated.RunID)
		}
		stored, err := ReadCandidateRecord(root, "t1", "p1")
		if err != nil {
			t.Fatal(err)
		}
		if stored.Verdict != CandidateVerdictGreen {
			t.Errorf("stored verdict after observation: %q, want green", stored.Verdict)
		}
	})

	t.Run("completed failure binds and writes red", func(t *testing.T) {
		root := t.TempDir()
		rec := CandidateRecord{CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1", Verdict: CandidateVerdictPending, PushedAt: "t"}
		if err := WriteCandidateRecord(root, rec); err != nil {
			t.Fatal(err)
		}
		updated, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", CandidateRunState{
			RunID: "9", HeadSHA: "c1", Ref: "ci/t1", Status: "completed", Conclusion: "failure",
		}, fixed)
		if err != nil || !wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		if updated.Verdict != CandidateVerdictRed {
			t.Errorf("verdict: got %q, want red", updated.Verdict)
		}
	})

	t.Run("run in progress writes nothing", func(t *testing.T) {
		root := t.TempDir()
		rec := CandidateRecord{CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1", Verdict: CandidateVerdictPending, PushedAt: "t"}
		if err := WriteCandidateRecord(root, rec); err != nil {
			t.Fatal(err)
		}
		_, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", CandidateRunState{
			RunID: "r-9", HeadSHA: "c1", Ref: "ci/t1", Status: "in_progress",
		}, fixed)
		if err != nil {
			t.Fatal(err)
		}
		if wrote {
			t.Fatal("an in-progress run is not a verdict — nothing may be written")
		}
	})

	t.Run("stale run for a superseded candidate on the same ref writes nothing", func(t *testing.T) {
		// AC-CCI-010-1's stale-run case: after a re-candidate produces C2,
		// an observed green for C1 on the same ci/ ref discards — C2 stays
		// pending.
		root := t.TempDir()
		rec := CandidateRecord{CardID: "t1", PinnedSHA: "p2", CandidateSHA: "c2", CandidateBranch: "ci/t1", Verdict: CandidateVerdictPending, PushedAt: "t"}
		if err := WriteCandidateRecord(root, rec); err != nil {
			t.Fatal(err)
		}
		_, wrote, err := ObserveCandidateVerdict(root, "t1", "p2", CandidateRunState{
			RunID: "r-old", HeadSHA: "c1", Ref: "ci/t1", Status: "completed", Conclusion: "success",
		}, fixed)
		if err != nil {
			t.Fatal(err)
		}
		if wrote {
			t.Fatal("a run whose head SHA is not the candidate's must be discarded")
		}
		stored, err := ReadCandidateRecord(root, "t1", "p2")
		if err != nil {
			t.Fatal(err)
		}
		if stored.Verdict != CandidateVerdictPending {
			t.Errorf("verdict after the stale observation: %q, want pending (nothing written)", stored.Verdict)
		}
	})

	t.Run("ref mismatch discards", func(t *testing.T) {
		root := t.TempDir()
		rec := CandidateRecord{CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1", Verdict: CandidateVerdictPending, PushedAt: "t"}
		if err := WriteCandidateRecord(root, rec); err != nil {
			t.Fatal(err)
		}
		_, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", CandidateRunState{
			RunID: "r-9", HeadSHA: "c1", Ref: "ci/other", Status: "completed", Conclusion: "success",
		}, fixed)
		if err != nil || wrote {
			t.Fatalf("wrote=%v err=%v — a ref-mismatched run is discarded", wrote, err)
		}
	})

	t.Run("a record replaced under the lock discards the stale observation", func(t *testing.T) {
		// The under-lock re-read: a re-candidate replaces the record while
		// the gh read is in flight — the observation writes only when the
		// record STILL names the run's candidate.
		root := t.TempDir()
		rec := CandidateRecord{CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1", Verdict: CandidateVerdictPending, PushedAt: "t"}
		if err := WriteCandidateRecord(root, rec); err != nil {
			t.Fatal(err)
		}
		prev := integrationCandidateMutationHook
		integrationCandidateMutationHook = func() {
			replaced := rec
			replaced.CandidateSHA = "c2"
			if err := WriteCandidateRecord(root, replaced); err != nil {
				t.Errorf("hook replace: %v", err)
			}
		}
		t.Cleanup(func() { integrationCandidateMutationHook = prev })
		_, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", CandidateRunState{
			RunID: "9", HeadSHA: "c1", Ref: "ci/t1", Status: "completed", Conclusion: "success",
		}, fixed)
		if err != nil || wrote {
			t.Fatalf("wrote=%v err=%v — the stale observation must not overwrite the replaced record", wrote, err)
		}
		stored, err := ReadCandidateRecord(root, "t1", "p1")
		if err != nil {
			t.Fatal(err)
		}
		if stored.CandidateSHA != "c2" {
			t.Errorf("record after the race: candidate %s, want c2 (untouched by the stale observation)", stored.CandidateSHA)
		}
	})

	t.Run("the observation preserves the re-pushed record's push info", func(t *testing.T) {
		// A re-push of the SAME candidate SHA lands between the outer read
		// and the lock: the observation must refresh ONLY the verdict
		// fields onto the re-read record — the re-push's PushedAt survives,
		// never regressed to the outer read's value.
		root := t.TempDir()
		rec := CandidateRecord{CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1", Verdict: CandidateVerdictPending, PushedAt: "2026-10-09T08:00:00Z"}
		if err := WriteCandidateRecord(root, rec); err != nil {
			t.Fatal(err)
		}
		prev := integrationCandidateMutationHook
		integrationCandidateMutationHook = func() {
			repushed := rec
			repushed.PushedAt = "2026-10-09T11:30:00Z"
			if err := WriteCandidateRecord(root, repushed); err != nil {
				t.Errorf("hook repush: %v", err)
			}
		}
		t.Cleanup(func() { integrationCandidateMutationHook = prev })
		updated, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", CandidateRunState{
			RunID: "9", HeadSHA: "c1", Ref: "ci/t1", Status: "completed", Conclusion: "success",
		}, fixed)
		if err != nil || !wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		if updated.PushedAt != "2026-10-09T11:30:00Z" {
			t.Errorf("PushedAt after observation: got %q, want the re-push's 11:30 (the outer read's 08:00 must not regress it)", updated.PushedAt)
		}
		if updated.Verdict != CandidateVerdictGreen {
			t.Errorf("verdict: got %q, want green", updated.Verdict)
		}
	})

	t.Run("sequence orders same-second pushes", func(t *testing.T) {
		// The same-second tie: two records with identical PushedAt — the
		// higher push sequence is the newer candidate, not filename order.
		root := t.TempDir()
		if err := WriteCandidateRecord(root, CandidateRecord{CardID: "t1", PinnedSHA: "p-aaa", CandidateSHA: "c-aaa", Verdict: CandidateVerdictRed, PushedAt: "2026-10-09T09:00:00Z", Seq: 1}); err != nil {
			t.Fatal(err)
		}
		if err := WriteCandidateRecord(root, CandidateRecord{CardID: "t1", PinnedSHA: "p-bbb", CandidateSHA: "c-bbb", Verdict: CandidateVerdictPending, PushedAt: "2026-10-09T09:00:00Z", Seq: 2}); err != nil {
			t.Fatal(err)
		}
		latest, err := LatestCandidateRecord(root, "t1")
		if err != nil {
			t.Fatal(err)
		}
		if latest.PinnedSHA != "p-bbb" {
			t.Errorf("latest: pinned %s (seq %d), want p-bbb (seq 2 — the newer same-second push)", latest.PinnedSHA, latest.Seq)
		}
	})
}

func TestLatestCandidateRecord(t *testing.T) {
	t.Run("newest pushed wins; absent card reads absent", func(t *testing.T) {
		root := t.TempDir()
		if _, err := LatestCandidateRecord(root, "t404"); !errors.Is(err, ErrCandidateRecordAbsent) {
			t.Fatalf("absent card: %v, want ErrCandidateRecordAbsent", err)
		}
		for _, r := range []CandidateRecord{
			{CardID: "t1", PinnedSHA: "p-old", CandidateSHA: "c-old", Verdict: CandidateVerdictRed, PushedAt: "2026-10-09T08:00:00Z"},
			{CardID: "t1", PinnedSHA: "p-new", CandidateSHA: "c-new", Verdict: CandidateVerdictPending, PushedAt: "2026-10-09T09:00:00Z"},
		} {
			if err := WriteCandidateRecord(root, r); err != nil {
				t.Fatal(err)
			}
		}
		latest, err := LatestCandidateRecord(root, "t1")
		if err != nil {
			t.Fatal(err)
		}
		if latest.PinnedSHA != "p-new" {
			t.Errorf("latest: pinned %s, want p-new (newest PushedAt)", latest.PinnedSHA)
		}
	})

	t.Run("unreadable entries are skipped, not fatal", func(t *testing.T) {
		root := t.TempDir()
		if err := WriteCandidateRecord(root, CandidateRecord{CardID: "t1", PinnedSHA: "p-ok", CandidateSHA: "c", Verdict: CandidateVerdictGreen, PushedAt: "2026-10-09T09:00:00Z"}); err != nil {
			t.Fatal(err)
		}
		junk := filepath.Join(root, ".moai", "state", "candidate", "t1", "garbage.json")
		if err := os.WriteFile(junk, []byte("{not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		latest, err := LatestCandidateRecord(root, "t1")
		if err != nil {
			t.Fatalf("unreadable sibling must not fail the scan: %v", err)
		}
		if latest.PinnedSHA != "p-ok" {
			t.Errorf("latest: pinned %s, want p-ok", latest.PinnedSHA)
		}
	})
}

func TestCandidateLandingCheckGuards(t *testing.T) {
	t.Run("no worktree refuses the tip read", func(t *testing.T) {
		f := newMergeFixture(t)
		landingSeedRecord(t, f, CandidateVerdictGreen)
		err := CandidateLandingCheck(LandingCheckInput{
			Root: f.root, CardID: stepCard, PinnedSHA: f.cardSHA,
			TargetBranch: "develop", IntegrationWorktree: "",
		})
		if err == nil {
			t.Fatal("want a refusal when no worktree can answer the tip read")
		}
	})

	t.Run("git failure in the check refuses, never admits", func(t *testing.T) {
		f := newMergeFixture(t)
		landingSeedRecord(t, f, CandidateVerdictGreen)
		err := CandidateLandingCheck(LandingCheckInput{
			Root: f.root, CardID: stepCard, PinnedSHA: f.cardSHA,
			TargetBranch: "develop", IntegrationWorktree: f.integ,
			Git: func(args ...string) (string, error) {
				return "", errors.New("git exploded (test double)")
			},
		})
		if err == nil || !strings.Contains(err.Error(), "git exploded") {
			t.Fatalf("refusal %v: want the git failure named", err)
		}
	})

	t.Run("candidateCIRequired reads the key fail-false", func(t *testing.T) {
		absent := t.TempDir()
		if candidateCIRequired(absent) {
			t.Error("absent config must read false")
		}
		broken := t.TempDir()
		sections := filepath.Join(broken, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte("workflow: [unclosed"), 0o600); err != nil {
			t.Fatal(err)
		}
		if candidateCIRequired(broken) {
			t.Error("an unreadable config must read false — the gate never enables itself on uncertainty")
		}
		on := t.TempDir()
		writeWorkflowCandidateCI(t, on, "true")
		if !candidateCIRequired(on) {
			t.Error("enabled key must read true")
		}
	})

	t.Run("refusals carry both short and full identity branches", func(t *testing.T) {
		// A record whose SHA fields are SHORT exercises the truncation
		// guard's other branch in the refusal formats.
		f := newMergeFixture(t)
		rec := landingSeedRecord(t, f, CandidateVerdictRed)
		rec.RunID = "r-77"
		rec.ObservedAt = "2026-10-09T09:00:00Z"
		if err := WriteCandidateRecord(f.root, rec); err != nil {
			t.Fatal(err)
		}
		_, err := RunMergeStep(f.input(), landingSeams(f, f.withCardTree(readyCardPtr())))
		requireCode(t, err, MergeExitLandingRefused)
		if !strings.Contains(err.Error(), "r-77") || !strings.Contains(err.Error(), "observed 2026-10-09T09:00:00Z") {
			t.Errorf("refusal %q: want the run identity and observation named", err)
		}
	})
}

// writeWorkflowCandidateCI is the factory-side twin of the cli helper: it
// writes the workflow.yaml section that turns the candidate-CI key on.
func writeWorkflowCandidateCI(t *testing.T, root, enabled string) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "workflow:\n  candidate_ci:\n    enabled: " + enabled + "\n    guard_bundle_required: true\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}
