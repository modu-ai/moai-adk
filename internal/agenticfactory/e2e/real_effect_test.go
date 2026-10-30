package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// Scenario 3 (real, lane-20 effect.Tracker): a lost ACK leaves the command at
// sent. Non-evidence never advances it; a receipt row found by the probe does,
// and repeating a stage writes nothing.

const sha40 = "0123456789abcdef0123456789abcdef01234567"

func TestReal_ACKLoss_NonEvidenceDoesNotAdvance(t *testing.T) {
	store := effect.NewMemStore()
	tr := effect.NewTracker(store)
	if _, err := tr.Advance("cmd-1", 1948, effect.Sent, effect.Evidence{Kind: effect.SendRecord, Ref: "q/1"}); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []effect.EvidenceKind{effect.SendSuccess, effect.IdleNotice, effect.ReplyMessage} {
		_, err := tr.Advance("cmd-1", 1948, effect.Received, effect.Evidence{Kind: kind, Ref: "x"})
		if !errors.Is(err, effect.ErrEvidence) {
			t.Errorf("%s advanced past sent: err=%v", kind, err)
		}
	}
	if rec, _ := store.Get("cmd-1"); rec.Stage != effect.Sent {
		t.Errorf("stage = %s, want sent", rec.Stage)
	}
}

func TestReal_ACKLoss_ProbeFindsReceipt_NoResend(t *testing.T) {
	store := effect.NewMemStore()
	tr := effect.NewTracker(store)
	if _, err := tr.Advance("cmd-1", 1948, effect.Sent, effect.Evidence{Kind: effect.SendRecord, Ref: "q/1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tr.Advance("cmd-1", 1948, effect.Received, effect.Evidence{Kind: effect.ReceiptRow, Ref: "r/1"}); err != nil {
		t.Fatal(err)
	}
	writes := store.Seq()
	// Re-sending the same command id is a no-op at the same stage.
	if _, err := tr.Advance("cmd-1", 1948, effect.Received, effect.Evidence{Kind: effect.ReceiptRow, Ref: "r/1"}); err != nil {
		t.Fatal(err)
	}
	if store.Seq() != writes {
		t.Errorf("repeat wrote to the store: seq %d -> %d", writes, store.Seq())
	}
	// Going back to sent (a resend) is refused.
	if _, err := tr.Advance("cmd-1", 1948, effect.Sent, effect.Evidence{Kind: effect.SendRecord, Ref: "q/2"}); !errors.Is(err, effect.ErrBackward) {
		t.Errorf("resend after received: %v, want ErrBackward", err)
	}
}

func TestReal_Completed_NeedsFullHeadAndResultFile(t *testing.T) {
	tr := effect.NewTracker(effect.NewMemStore())
	steps := []struct {
		to effect.Stage
		ev effect.Evidence
	}{
		{effect.Sent, effect.Evidence{Kind: effect.SendRecord, Ref: "q"}},
		{effect.Received, effect.Evidence{Kind: effect.ReceiptRow, Ref: "r"}},
		{effect.Executed, effect.Evidence{Kind: effect.AttemptStart, Ref: "a"}},
	}
	for _, s := range steps {
		if _, err := tr.Advance("cmd-1", 1948, s.to, s.ev); err != nil {
			t.Fatalf("%s: %v", s.to, err)
		}
	}
	short := effect.Evidence{Kind: effect.ResultFile, Ref: "res.json", HeadSHA: sha40[:7]}
	if _, err := tr.Advance("cmd-1", 1948, effect.Completed, short); err == nil || !strings.Contains(err.Error(), "40-hex") {
		t.Errorf("abbreviated head accepted: %v", err)
	}
	reply := effect.Evidence{Kind: effect.ReplyMessage, Ref: "m", HeadSHA: sha40}
	if _, err := tr.Advance("cmd-1", 1948, effect.Completed, reply); !errors.Is(err, effect.ErrEvidence) {
		t.Errorf("reply accepted as completion: %v", err)
	}
	ok := effect.Evidence{Kind: effect.ResultFile, Ref: "res.json", HeadSHA: sha40}
	if _, err := tr.Advance("cmd-1", 1948, effect.Completed, ok); err != nil {
		t.Errorf("valid completion refused: %v", err)
	}
}
