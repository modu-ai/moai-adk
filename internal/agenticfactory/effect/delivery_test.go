package effect

import (
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// AC-AF-005 / design section 3: the four delivery states (sent, received, executed, completed) move only on
// the evidence the design names. A send that went through, an idle notice and a reply message are not
// evidence of anything past "sent".

func TestDelivery_SendSuccessAloneLeavesOnlySent(t *testing.T) {
	st := NewMemStore()
	tr := NewTracker(st)
	rec, err := tr.Advance("cmd-1", 7, Sent, ev(SendRecord))
	if err != nil || rec.Stage != Sent {
		t.Fatalf("%+v %v", rec, err)
	}
	for _, k := range []EvidenceKind{SendSuccess, IdleNotice, ReplyMessage} {
		if _, err := tr.Advance("cmd-1", 7, Received, ev(k)); !errors.Is(err, ErrEvidence) {
			t.Errorf("received on %s: err = %v, want ErrEvidence", k, err)
		}
	}
	if got, _ := st.Get("cmd-1"); got.Stage != Sent {
		t.Errorf("stage = %s, want it still sent (received and above: none)", got.Stage)
	}
}

func TestDelivery_IdleNoticeReplyAndSendSuccessNeverCompleteACommand(t *testing.T) {
	tr := NewTracker(NewMemStore())
	steps := []struct {
		to Stage
		k  EvidenceKind
	}{{Sent, SendRecord}, {Received, ReceiptRow}, {Executed, AttemptStart}}
	for _, s := range steps {
		if _, err := tr.Advance("cmd-2", 7, s.to, ev(s.k)); err != nil {
			t.Fatal(err)
		}
	}
	for _, k := range []EvidenceKind{IdleNotice, ReplyMessage, SendSuccess, AttemptStart, ReceiptRow} {
		if _, err := tr.Advance("cmd-2", 7, Completed, ev(k)); !errors.Is(err, ErrEvidence) {
			t.Errorf("completed on %s: err = %v, want ErrEvidence", k, err)
		}
	}
	short := Evidence{Kind: ResultFile, Ref: "result.json", HeadSHA: "0123456"}
	if _, err := tr.Advance("cmd-2", 7, Completed, short); !errors.Is(err, ErrEvidence) {
		t.Errorf("completed on a short HEAD sha: err = %v, want ErrEvidence", err)
	}
	rec, err := tr.Advance("cmd-2", 7, Completed, ev(ResultFile))
	if err != nil || rec.Stage != Completed {
		t.Errorf("the result file with the full HEAD sha must complete it: %+v %v", rec, err)
	}
}

// The reply is lost in the transport but the agent finished: progress is recovered from the result record,
// not from a reply, and nothing is sent a second time. The wait kept while the reply was awaited carries all
// six answers of design 2.1c. With no record either, nothing becomes completed.
func TestDelivery_ReplyLoss_RecoveredFromRecordNotReply(t *testing.T) {
	st := NewMemStore()
	tr := NewTracker(st)
	for _, s := range []struct {
		to Stage
		k  EvidenceKind
	}{{Sent, SendRecord}, {Received, ReceiptRow}, {Executed, AttemptStart}} {
		if _, err := tr.Advance("cmd-3", 7, s.to, ev(s.k)); err != nil {
			t.Fatal(err)
		}
	}

	// While the reply is awaited the waiting party leaves a wait record: all six fields, or it is invalid.
	w := core.Wait{Owner: "controlservice", NextAction: "read the result record", RecheckAt: time.Unix(1_000_000, 0),
		Diagnosis: "the completion reply has not arrived", MaxRetries: 3, Escalate: "leader"}
	if err := w.Validate(); err != nil {
		t.Fatalf("a wait with all six answers was refused: %v", err)
	}
	for name, mut := range map[string]func(*core.Wait){
		"owner": func(w *core.Wait) { w.Owner = "" }, "next action": func(w *core.Wait) { w.NextAction = "" },
		"recheck time": func(w *core.Wait) { w.RecheckAt = time.Time{} }, "diagnosis": func(w *core.Wait) { w.Diagnosis = "" },
		"retry limit": func(w *core.Wait) { w.MaxRetries = 0 }, "escalation target": func(w *core.Wait) { w.Escalate = "" },
	} {
		bad := w
		mut(&bad)
		if bad.Validate() == nil {
			t.Errorf("a wait without its %s was accepted", name)
		}
	}

	// No result record and no reply: still not completed.
	if got, _ := st.Get("cmd-3"); got.Stage == Completed {
		t.Fatal("completed without any record")
	}
	// The result record is found on disk: the command completes. The tracker keeps one record per command
	// id, so re-recording the send (what a resend would do) is a no-op that writes nothing.
	rec, err := tr.Advance("cmd-3", 7, Completed, ev(ResultFile))
	if err != nil || rec.Stage != Completed {
		t.Fatalf("recovery from the record failed: %+v %v", rec, err)
	}
	seq := st.Seq()
	if _, err := tr.Advance("cmd-3", 7, Sent, ev(SendRecord)); !errors.Is(err, ErrBackward) || st.Seq() != seq {
		t.Errorf("re-recording the send after completion: err=%v, writes=%d; want a refusal and no write", err, st.Seq()-seq)
	}
}
