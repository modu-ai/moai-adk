package effect

import (
	"errors"
	"strings"
	"testing"
)

const sha = "0123456789abcdef0123456789abcdef01234567"

func ev(k EvidenceKind) Evidence { return Evidence{Kind: k, Ref: "ref-" + string(k), HeadSHA: sha} }

func TestForwardChainAdvances(t *testing.T) {
	tr := NewTracker(NewMemStore())
	steps := []struct {
		to Stage
		ev Evidence
	}{
		{Sent, ev(SendRecord)},
		{Received, ev(ReceiptRow)},
		{Executed, ev(AttemptStart)},
		{Completed, ev(ResultFile)},
	}
	for _, s := range steps {
		got, err := tr.Advance("cmd-1", 7, s.to, s.ev)
		if err != nil {
			t.Fatalf("advance to %s: %v", s.to, err)
		}
		if got.Stage != s.to {
			t.Fatalf("stage = %s, want %s", got.Stage, s.to)
		}
	}
}

func TestSameCommandIDIsIdempotent(t *testing.T) {
	st := NewMemStore()
	tr := NewTracker(st)
	if _, err := tr.Advance("c", 1, Sent, ev(SendRecord)); err != nil {
		t.Fatal(err)
	}
	seq := st.Seq()
	rec, err := tr.Advance("c", 1, Sent, ev(SendRecord))
	if err != nil || rec.Stage != Sent {
		t.Fatalf("repeat = %v, %v", rec, err)
	}
	if st.Seq() != seq {
		t.Fatalf("idempotent repeat wrote: seq %d -> %d", seq, st.Seq())
	}
}

func TestBackwardAndSkipRejected(t *testing.T) {
	tr := NewTracker(NewMemStore())
	if _, err := tr.Advance("c", 1, Sent, ev(SendRecord)); err != nil {
		t.Fatal(err)
	}
	if _, err := tr.Advance("c", 1, Executed, ev(AttemptStart)); !errors.Is(err, ErrSkip) {
		t.Fatalf("skip err = %v", err)
	}
	if _, err := tr.Advance("c", 1, Received, ev(ReceiptRow)); err != nil {
		t.Fatal(err)
	}
	if _, err := tr.Advance("c", 1, Sent, ev(SendRecord)); !errors.Is(err, ErrBackward) {
		t.Fatalf("backward err = %v", err)
	}
	if _, err := tr.Advance("new", 1, Received, ev(ReceiptRow)); !errors.Is(err, ErrSkip) {
		t.Fatalf("first stage other than sent: %v", err)
	}
}

func TestNonEvidenceRefused(t *testing.T) {
	tr := NewTracker(NewMemStore())
	if _, err := tr.Advance("c", 1, Sent, ev(SendRecord)); err != nil {
		t.Fatal(err)
	}
	for _, k := range []EvidenceKind{IdleNotice, ReplyMessage, SendSuccess} {
		if _, err := tr.Advance("c", 1, Received, ev(k)); !errors.Is(err, ErrEvidence) {
			t.Errorf("%s accepted for received: %v", k, err)
		}
	}
}

func TestCompletedNeedsFullHeadSHA(t *testing.T) {
	tr := NewTracker(NewMemStore())
	for _, s := range []struct {
		to Stage
		k  EvidenceKind
	}{{Sent, SendRecord}, {Received, ReceiptRow}, {Executed, AttemptStart}} {
		if _, err := tr.Advance("c", 1, s.to, ev(s.k)); err != nil {
			t.Fatal(err)
		}
	}
	short := Evidence{Kind: ResultFile, Ref: "r", HeadSHA: "abc123"}
	if _, err := tr.Advance("c", 1, Completed, short); !errors.Is(err, ErrEvidence) {
		t.Fatalf("short sha err = %v", err)
	}
}

func TestSameIDDifferentIssueRefused(t *testing.T) {
	tr := NewTracker(NewMemStore())
	if _, err := tr.Advance("c", 1, Sent, ev(SendRecord)); err != nil {
		t.Fatal(err)
	}
	_, err := tr.Advance("c", 2, Received, ev(ReceiptRow))
	if err == nil || !strings.Contains(err.Error(), "issue") {
		t.Fatalf("cross-issue reuse err = %v", err)
	}
}

func TestInvalidInputRejected(t *testing.T) {
	tr := NewTracker(NewMemStore())
	if _, err := tr.Advance("", 1, Sent, ev(SendRecord)); err == nil {
		t.Error("empty command id accepted")
	}
	if _, err := tr.Advance("c", 1, Stage("bogus"), ev(SendRecord)); err == nil {
		t.Error("unknown stage accepted")
	}
}
