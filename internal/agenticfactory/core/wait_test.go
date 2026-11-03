package core

import (
	"errors"
	"testing"
	"time"
)

func goodWait() Wait {
	return Wait{
		Owner: "agent-1", NextAction: "re-run the failed check after the cache reset",
		RecheckAt: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC),
		Diagnosis: "ci job build failed on head aaaa; flaky network step", Retries: 1, MaxRetries: 3, Escalate: "leader",
	}
}

// A wait record names who is responsible, what happens next without anyone being
// asked, when it is looked at again, why it waits, how many retries are spent of
// how many are allowed, and who is told when they run out.
func TestWaitNeedsEveryField(t *testing.T) {
	if err := goodWait().Validate(); err != nil {
		t.Fatalf("a complete wait: %v", err)
	}
	for name, mut := range map[string]func(*Wait){
		"owner":        func(w *Wait) { w.Owner = "" },
		"next action":  func(w *Wait) { w.NextAction = " " },
		"recheck time": func(w *Wait) { w.RecheckAt = time.Time{} },
		"diagnosis":    func(w *Wait) { w.Diagnosis = "" },
		"retry limit":  func(w *Wait) { w.MaxRetries = 0 },
		"escalate to":  func(w *Wait) { w.Escalate = "" },
	} {
		w := goodWait()
		mut(&w)
		if err := w.Validate(); !errors.Is(err, ErrEvidenceIncomplete) {
			t.Errorf("without %s: %v", name, err)
		}
	}
	w := goodWait()
	w.Retries = -1
	if err := w.Validate(); !errors.Is(err, ErrEvidenceIncomplete) {
		t.Errorf("negative retries: %v", err)
	}
	w = goodWait()
	w.Retries = w.MaxRetries + 1
	if err := w.Validate(); !errors.Is(err, ErrEvidenceIncomplete) {
		t.Errorf("retries beyond the limit: %v", err)
	}
}

// The limit has no default: a zero or missing MaxRetries is refused, never
// replaced by a number of ours.
func TestWaitRetryCountsUpToTheLimitThenEscalates(t *testing.T) {
	w := goodWait()
	w.Retries = 0
	next := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	for i := 1; i < w.MaxRetries; i++ {
		var err error
		if w, err = w.Retry("still failing", "re-run again", next); err != nil {
			t.Fatalf("retry %d: %v", i, err)
		}
		if w.Retries != i || w.Escalated {
			t.Fatalf("retry %d: %+v", i, w)
		}
	}
	last, err := w.Retry("still failing", "re-run again", next)
	if !errors.Is(err, ErrRetryLimit) {
		t.Fatalf("the last allowed retry must report the limit: %v", err)
	}
	if last.Retries != last.MaxRetries || !last.Escalated || last.Escalate != "leader" {
		t.Errorf("the exhausted wait must name its escalation target and say it escalated: %+v", last)
	}
	if _, err := last.Retry("x", "y", next); !errors.Is(err, ErrRetryLimit) {
		t.Errorf("no retry past the limit: %v", err)
	}
}
