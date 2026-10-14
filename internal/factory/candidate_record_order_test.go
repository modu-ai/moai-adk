package factory

// candidate_record_order_test.go — card t1478 Finding 2 (SPEC-CANDIDATE-CI-001
// REQ-CCI-010): an observation must not overwrite a verdict recorded from a
// NEWER run. Runs order by execution: GitHub run ids increase monotonically,
// and a re-run keeps its run id while carrying a higher attempt.

import (
	"testing"
	"time"
)

var runOrderNow = time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)

// runOrderSeedPending writes the card's pending candidate record the
// observations are judged against.
func runOrderSeedPending(t *testing.T, root string) {
	t.Helper()
	if err := WriteCandidateRecord(root, CandidateRecord{
		CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1",
		Verdict: CandidateVerdictPending, PushedAt: "2026-10-10T08:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
}

// runOrderRun is one completed observation of the candidate's own run.
func runOrderRun(runID string, attempt int, conclusion string) CandidateRunState {
	return CandidateRunState{RunID: runID, Attempt: attempt, HeadSHA: "c1", Ref: "ci/t1", Status: "completed", Conclusion: conclusion}
}

func TestObserveCandidateRunOrder(t *testing.T) {
	t.Run("a stale successful run cannot overwrite a newer red verdict", func(t *testing.T) {
		root := t.TempDir()
		runOrderSeedPending(t, root)
		if _, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("200", 1, "failure"), runOrderNow); err != nil || !wrote {
			t.Fatalf("seed red from run 200: wrote=%v err=%v", wrote, err)
		}
		_, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("100", 1, "success"), runOrderNow)
		if err != nil {
			t.Fatalf("observe run 100: %v", err)
		}
		stored, err := ReadCandidateRecord(root, "t1", "p1")
		if err != nil {
			t.Fatal(err)
		}
		if wrote || stored.Verdict != CandidateVerdictRed || stored.RunID != "200" {
			t.Errorf("wrote=%v, stored verdict %q run %q: run 100 is older than the recorded run 200 — want wrote=false, red, run 200", wrote, stored.Verdict, stored.RunID)
		}
	})

	t.Run("a re-run attempt may replace its earlier attempt, never the reverse", func(t *testing.T) {
		root := t.TempDir()
		runOrderSeedPending(t, root)
		if _, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("200", 1, "failure"), runOrderNow); err != nil || !wrote {
			t.Fatalf("seed red from run 200 attempt 1: wrote=%v err=%v", wrote, err)
		}
		if _, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("200", 2, "success"), runOrderNow); err != nil || !wrote {
			t.Fatalf("run 200 attempt 2 must replace attempt 1: wrote=%v err=%v", wrote, err)
		}
		_, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("200", 1, "failure"), runOrderNow)
		if err != nil {
			t.Fatalf("observe run 200 attempt 1 after attempt 2: %v", err)
		}
		stored, err := ReadCandidateRecord(root, "t1", "p1")
		if err != nil {
			t.Fatal(err)
		}
		if wrote || stored.Verdict != CandidateVerdictGreen || stored.RunAttempt != 2 {
			t.Errorf("wrote=%v, stored verdict %q attempt %d: attempt 1 arriving after attempt 2 must be refused — want wrote=false, green, attempt 2", wrote, stored.Verdict, stored.RunAttempt)
		}
	})

	t.Run("a non-numeric or empty run id is refused and never green", func(t *testing.T) {
		for _, id := range []string{"", "r-9", " 9", "-9"} {
			root := t.TempDir()
			runOrderSeedPending(t, root)
			_, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun(id, 1, "success"), runOrderNow)
			if err == nil || wrote {
				t.Errorf("run id %q: wrote=%v err=%v — an unorderable run must be refused with an error", id, wrote, err)
			}
			stored, readErr := ReadCandidateRecord(root, "t1", "p1")
			if readErr != nil {
				t.Fatal(readErr)
			}
			if stored.Verdict != CandidateVerdictPending {
				t.Errorf("run id %q: stored verdict %q, want pending", id, stored.Verdict)
			}
		}
	})

	t.Run("the first observation of a record with no prior run keeps working", func(t *testing.T) {
		root := t.TempDir()
		runOrderSeedPending(t, root)
		rec, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("300", 1, "success"), runOrderNow)
		if err != nil || !wrote || rec.Verdict != CandidateVerdictGreen || rec.RunID != "300" || rec.RunAttempt != 1 {
			t.Errorf("first observation: wrote=%v err=%v verdict %q run %q attempt %d, want green from run 300 attempt 1", wrote, err, rec.Verdict, rec.RunID, rec.RunAttempt)
		}
	})

	t.Run("a record written before attempts reads as attempt 1 of its run", func(t *testing.T) {
		root := t.TempDir()
		if err := WriteCandidateRecord(root, CandidateRecord{
			CardID: "t1", PinnedSHA: "p1", CandidateSHA: "c1", CandidateBranch: "ci/t1",
			Verdict: CandidateVerdictRed, RunID: "200", PushedAt: "2026-10-10T08:00:00Z",
		}); err != nil {
			t.Fatal(err)
		}
		// Attempt 1 of the recorded run 200 is the same attempt, not an older one.
		if _, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("200", 1, "success"), runOrderNow); err != nil || !wrote {
			t.Errorf("attempt 1 of the legacy run 200: wrote=%v err=%v, want the observation admitted", wrote, err)
		}
		// An older run is refused against the legacy record.
		if _, wrote, err := ObserveCandidateVerdict(root, "t1", "p1", runOrderRun("199", 1, "failure"), runOrderNow); err != nil || wrote {
			t.Errorf("run 199 against the legacy run 200: wrote=%v err=%v, want refused", wrote, err)
		}
	})
}
