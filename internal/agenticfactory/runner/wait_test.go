package runner

import (
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func wait1() core.Wait {
	return core.Wait{
		Owner: "sess-a", NextAction: "re-run the failed check", RecheckAt: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC),
		Diagnosis: "build failed on head " + headA[:12], MaxRetries: 2, Escalate: "leader",
	}
}

func TestWaitIsRecordedByItsOwnerAndGoesWhenTheIssueIsDone(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.HandOff(tk, goodHandoff()); err != nil {
		t.Fatal(err)
	}
	other := wait1()
	other.Owner = "someone-else"
	if _, err := r.SetWait(tk, other); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("a wait owned by another session: %v", err)
	}
	bad := wait1()
	bad.Escalate = ""
	if _, err := r.SetWait(tk, bad); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("an incomplete wait: %v", err)
	}
	rec, err := r.SetWait(tk, wait1())
	if err != nil || rec.Wait == nil || rec.Wait.Owner != "sess-a" || rec.State != core.StateLanding {
		t.Fatalf("set wait: %+v %v", rec, err)
	}
	// a wait is not a gate or a park: the position does not change
	if rec.Gate != nil || rec.Park != nil {
		t.Errorf("a wait must not look like a gate or a park: %+v %+v", rec.Gate, rec.Park)
	}
	done, err := r.Complete(tk, Landing{RemoteMerged: true, LocalSyncRecorded: true, IssueClosed: true, Deploy: "N/A: x"})
	if err != nil || done.Wait != nil {
		t.Errorf("a finished Issue waits for nothing: %+v %v", done.Wait, err)
	}
}

func TestRetryingAWaitEscalatesAtTheLimit(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.SetWait(tk, wait1()); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	rec, err := r.RetryWait(tk, "still red", "re-run once more", at)
	if err != nil || rec.Wait.Retries != 1 || rec.Wait.Escalated {
		t.Fatalf("first retry: %+v %v", rec.Wait, err)
	}
	rec, err = r.RetryWait(tk, "still red", "ask for help", at)
	if !errors.Is(err, core.ErrRetryLimit) || rec.Wait == nil || !rec.Wait.Escalated || rec.Wait.Retries != 2 {
		t.Fatalf("the retry that reaches the limit is recorded and reported: %+v %v", rec.Wait, err)
	}
	if _, err := r.RetryWait(tk, "x", "y", at); !errors.Is(err, core.ErrRetryLimit) {
		t.Errorf("no retry past the limit: %v", err)
	}
	cleared, err := r.ClearWait(tk)
	if err != nil || cleared.Wait != nil {
		t.Errorf("clear: %+v %v", cleared.Wait, err)
	}
}

// A refused retry writes nothing. The retry that reaches the limit is recorded
// once, as escalated; every retry after that is refused with ErrRetryLimit and
// leaves the stored record (its Seq and its wait) exactly as it was.
func TestARetryPastTheLimitWritesNothing(t *testing.T) {
	r, ref := newRunner(t)
	tk := landed(t, r, ref)
	if _, err := r.SetWait(tk, wait1()); err != nil { // MaxRetries 2
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	if _, err := r.RetryWait(tk, "still red", "re-run once more", at); err != nil {
		t.Fatal(err)
	}
	before, _ := r.Store.Read(ref)
	// regression: the retry that reaches the limit is recorded, once, escalated
	rec, err := r.RetryWait(tk, "still red", "ask for help", at)
	if !errors.Is(err, core.ErrRetryLimit) || rec.Wait == nil || !rec.Wait.Escalated || rec.Wait.Retries != 2 {
		t.Fatalf("the limit-reaching retry: %+v %v", rec.Wait, err)
	}
	if rec.Seq != before.Seq+1 {
		t.Fatalf("the limit-reaching retry is one write: Seq %d -> %d", before.Seq, rec.Seq)
	}
	escalated, _ := r.Store.Read(ref)
	for i := 0; i < 2; i++ {
		got, err := r.RetryWait(tk, "again", "again", at.Add(time.Hour))
		if !errors.Is(err, core.ErrRetryLimit) {
			t.Fatalf("retry past the limit #%d: %v", i+1, err)
		}
		cur, _ := r.Store.Read(ref)
		if cur.Seq != escalated.Seq {
			t.Errorf("a refused retry wrote the record again: Seq %d -> %d", escalated.Seq, cur.Seq)
		}
		if cur.Wait == nil || *cur.Wait != *escalated.Wait {
			t.Errorf("a refused retry changed the wait: %+v -> %+v", escalated.Wait, cur.Wait)
		}
		_ = got // a refused call returns no record, as every other refusal here does
	}
}
