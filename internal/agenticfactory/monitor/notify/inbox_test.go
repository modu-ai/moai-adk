package notify

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/docstore"
)

func (e *env) inbox() Inbox {
	return Inbox{Docs: e.st, Kind: docstore.CollectorKind, SeenLimit: 8, MaxIDBytes: 32}
}

func notice(issue int, holder string, kind Kind, detail string) Notice {
	return Notice{Issue: issue, Holder: holder, Agent: "agent-" + holder, Kind: kind, Detail: detail}
}

func ids[T any](rs []Received[T]) (out []string) {
	for _, r := range rs {
		out = append(out, r.ID)
	}
	return out
}

// A notice is durable in its holder's inbox before Notify returns, only that holder sees it, and it
// leaves only when the holder acknowledges it.
func TestNoticesAreQueuedForTheirHolderAndLeaveOnlyWhenAcknowledged(t *testing.T) {
	e := newEnv(t)
	in := e.inbox()
	ctx := context.Background()
	a1, a2, b1 := notice(7, "sess-a", CIFailed, "check failed"), notice(7, "sess-a", ChangesRequested, "changes"), notice(8, "sess-b", PRMerged, "merged")
	for _, n := range []Notice{a1, a2, a1, b1} { // a1 twice: a retry after a crash
		if err := in.Notify(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	got, err := in.Notices("sess-a")
	if err != nil || len(got) != 2 || got[0].Item != a1 || got[1].Item != a2 {
		t.Fatalf("sess-a inbox: %+v %v", got, err)
	}
	if other, _ := in.Notices("sess-b"); len(other) != 1 || other[0].Item != b1 {
		t.Fatalf("sess-b inbox: %+v", other)
	}
	if none, _ := in.Notices("sess-c"); len(none) != 0 {
		t.Fatalf("a stranger's inbox: %+v", none)
	}
	if err := in.AckNotices("sess-a", []string{got[0].ID}); err != nil {
		t.Fatal(err)
	}
	if err := in.AckNotices("sess-a", []string{got[0].ID}); err != nil { // a repeated acknowledgement is harmless
		t.Fatal(err)
	}
	if left, _ := in.Notices("sess-a"); len(left) != 1 || left[0].Item != a2 {
		t.Fatalf("after the acknowledgement: %+v", left)
	}
	_ = in.Notify(ctx, a1) // a consumed notice resent is not queued again
	if left, _ := in.Notices("sess-a"); len(left) != 1 {
		t.Errorf("a consumed notice came back: %+v", left)
	}
}

func TestEscalationsGoToTheirTargetAndNotToAgents(t *testing.T) {
	e := newEnv(t)
	in := e.inbox()
	esc := Escalation{Issue: 7, Holder: "sess-a", To: "leader", Retries: 3, MaxRetries: 3, Diagnosis: "no progress", NextAction: "wait"}
	if err := in.Escalate(context.Background(), esc); err != nil {
		t.Fatal(err)
	}
	got, err := in.Escalations("leader")
	if err != nil || len(got) != 1 || got[0].Item != esc {
		t.Fatalf("leader inbox: %+v %v", got, err)
	}
	if n, _ := in.Notices("sess-a"); len(n) != 0 {
		t.Errorf("an escalation reached the agent's notices: %+v", n)
	}
	if err := in.AckEscalations("leader", ids(got)); err != nil {
		t.Fatal(err)
	}
	if left, _ := in.Escalations("leader"); len(left) != 0 {
		t.Errorf("after the acknowledgement: %+v", left)
	}
}

func TestAMessageWithoutAnAddresseeIsRefused(t *testing.T) {
	in := newEnv(t).inbox()
	if err := in.Notify(context.Background(), notice(7, "", CIFailed, "x")); err == nil {
		t.Error("a notice without a holder was queued")
	}
	if err := in.Escalate(context.Background(), Escalation{Issue: 7}); err == nil {
		t.Error("an escalation without a target was queued")
	}
}

// A full inbox refuses and keeps what it holds; the reminder protocol treats the refusal as a failed
// delivery (it spends the retry), so a stuffed inbox still ends in an escalation.
func TestAFullInboxRefusesAndKeepsWhatItHolds(t *testing.T) {
	e := newEnv(t)
	in := e.inbox()
	in.MaxDocBytes, in.SeenLimit = 700, 1
	var held []string
	var full error
	for i := range 20 {
		n := notice(7, "sess-a", CIFailed, fmt.Sprintf("failure number %02d", i))
		if err := in.Notify(context.Background(), n); err != nil {
			full = err
			break
		}
		held = append(held, n.Detail)
	}
	if !errors.Is(full, docstore.ErrQueueFull) || len(held) == 0 {
		t.Fatalf("full=%v held=%d", full, len(held))
	}
	got, _ := in.Notices("sess-a")
	if len(got) != len(held) {
		t.Fatalf("holds %d, accepted %d", len(got), len(held))
	}
	for i, r := range got {
		if r.Item.Detail != held[i] {
			t.Errorf("entry %d changed: %q", i, r.Item.Detail)
		}
	}
}

// An id longer than the configured bound is refused, not truncated.
func TestAnInboxWhoseIdBoundIsTooSmallRefusesInsteadOfTruncating(t *testing.T) {
	in := newEnv(t).inbox()
	in.MaxIDBytes = 16
	if err := in.Notify(context.Background(), notice(7, "sess-a", CIFailed, "x")); !errors.Is(err, docstore.ErrBadDeliveryID) {
		t.Errorf("err = %v", err)
	}
}

// The notice is queued before Notify returns: a crash right after the write leaves it queued, and the
// retry is absorbed.
func TestACrashAfterTheInboxWriteIsAbsorbedByTheRetry(t *testing.T) {
	e := newEnv(t)
	cd := &crashDocs{inner: e.st, crashAt: 1}
	in := e.inbox()
	in.Docs = cd
	n := notice(7, "sess-a", CIFailed, "x")
	if err := in.Notify(context.Background(), n); err == nil {
		t.Fatal("the simulated crash was not reported")
	}
	in.Docs = e.st // restart
	if got, _ := in.Notices("sess-a"); len(got) != 1 {
		t.Fatalf("the durable write was lost: %+v", got)
	}
	if err := in.Notify(context.Background(), n); err != nil {
		t.Fatal(err)
	}
	if got, _ := in.Notices("sess-a"); len(got) != 1 {
		t.Errorf("the retry duplicated the notice: %+v", got)
	}
}

type crashDocs struct {
	inner   *claim.Store
	puts    int
	crashAt int
}

func (c *crashDocs) GetDoc(kind, name string) (claim.Doc, error) { return c.inner.GetDoc(kind, name) }

func (c *crashDocs) PutDoc(kind, name string, rev int64, data []byte) (claim.Doc, error) {
	c.puts++
	d, err := c.inner.PutDoc(kind, name, rev, data)
	if err == nil && c.puts == c.crashAt {
		return claim.Doc{}, errors.New("simulated crash")
	}
	return d, err
}

// The whole chain through the real inbox: the reminders reach the holder's inbox, the escalation reaches
// the leader's, once; a restart repeats nothing; the owner is unchanged.
func TestTheReminderChainThroughTheInbox(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	if _, err := e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3)); err != nil {
		t.Fatal(err)
	}
	in := e.inbox()
	svc := func() *Service {
		s := e.service()
		s.Notifier, s.Escalator = in, in
		return s
	}
	for _, d := range []time.Duration{time.Hour, 90 * time.Minute, 2 * time.Hour, 3 * time.Hour} {
		e.tickAt(svc(), t0.Add(d)) // a fresh service each pass: a restart each time
	}
	got, err := in.Notices(holder)
	if err != nil || len(got) != 3 {
		t.Fatalf("holder inbox: %d notices, err=%v", len(got), err)
	}
	for i, r := range got {
		if r.Item.Kind != WaitOverdue || r.Item.Attempt != i+1 || r.Item.Issue != 7 || r.Item.Holder != holder {
			t.Errorf("notice %d: %+v", i, r.Item)
		}
	}
	esc, err := in.Escalations("leader")
	if err != nil || len(esc) != 1 || esc[0].Item.Issue != 7 || esc[0].Item.Holder != holder || esc[0].Item.Retries != 3 {
		t.Fatalf("leader inbox: %+v %v", esc, err)
	}
	if err := in.AckNotices(holder, ids(got)); err != nil {
		t.Fatal(err)
	}
	e.tickAt(svc(), t0.Add(5*time.Hour))
	if more, _ := in.Notices(holder); len(more) != 0 {
		t.Errorf("after escalation and acknowledgement: %+v", more)
	}
	if rec := e.record(7); rec.Holder != holder || rec.Epoch != tk.Epoch {
		t.Errorf("the owner changed: %q epoch %d", rec.Holder, rec.Epoch)
	}
}

// The reading side: a message is acknowledged only after it was handled; a failing handler leaves it and
// the ones behind it queued; a lost acknowledgement means one repeat, never a loss.
func TestConsumeAcknowledgesOnlyWhatWasHandled(t *testing.T) {
	e := newEnv(t)
	in := e.inbox()
	ctx := context.Background()
	for i := range 3 {
		if err := in.Notify(ctx, notice(7, "sess-a", CIFailed, fmt.Sprintf("failure %d", i))); err != nil {
			t.Fatal(err)
		}
	}
	var seen []string
	n, err := in.ConsumeNotices("sess-a", func(x Notice) error {
		if x.Detail == "failure 1" {
			return errors.New("cannot take this one in")
		}
		seen = append(seen, x.Detail)
		return nil
	})
	if n != 1 || err == nil || len(seen) != 1 || seen[0] != "failure 0" {
		t.Fatalf("n=%d err=%v seen=%v", n, err, seen)
	}
	left, _ := in.Notices("sess-a")
	if len(left) != 2 || left[0].Item.Detail != "failure 1" || left[1].Item.Detail != "failure 2" {
		t.Fatalf("after the failed pass: %+v", left)
	}
	seen = nil
	n, err = in.ConsumeNotices("sess-a", func(x Notice) error { seen = append(seen, x.Detail); return nil })
	if n != 2 || err != nil || len(seen) != 2 {
		t.Fatalf("second pass: n=%d err=%v seen=%v", n, err, seen)
	}
	if rest, _ := in.Notices("sess-a"); len(rest) != 0 {
		t.Errorf("handled notices stayed queued: %+v", rest)
	}
	if n, err := in.ConsumeNotices("sess-a", func(Notice) error { t.Error("handler called on an empty inbox"); return nil }); n != 0 || err != nil {
		t.Errorf("empty inbox: n=%d err=%v", n, err)
	}
}

// The acknowledgement's write lands and then the process dies: the notice is gone, it is not handled twice.
func TestAnAcknowledgementThatLandedBeforeACrashIsNotRepeated(t *testing.T) {
	e := newEnv(t)
	in := e.inbox()
	if err := in.Notify(context.Background(), notice(7, "sess-a", CIFailed, "x")); err != nil {
		t.Fatal(err)
	}
	handled := 0
	in.Docs = &crashDocs{inner: e.st, crashAt: 1}
	if _, err := in.ConsumeNotices("sess-a", func(Notice) error { handled++; return nil }); err == nil {
		t.Fatal("the crash was not reported")
	}
	in.Docs = e.st // restart
	if n, err := in.ConsumeNotices("sess-a", func(Notice) error { handled++; return nil }); n != 0 || err != nil || handled != 1 {
		t.Errorf("n=%d err=%v handled=%d, want it handled exactly once", n, err, handled)
	}
}

type refuseDocs struct {
	inner *claim.Store
	puts  int
}

func (r *refuseDocs) GetDoc(kind, name string) (claim.Doc, error) { return r.inner.GetDoc(kind, name) }
func (r *refuseDocs) PutDoc(kind, name string, rev int64, data []byte) (claim.Doc, error) {
	r.puts++
	return claim.Doc{}, errors.New("store refused the write")
}

func TestAnAcknowledgementThatCannotBeWrittenLeavesTheMessageForARepeat(t *testing.T) {
	e := newEnv(t)
	in := e.inbox()
	if err := in.Notify(context.Background(), notice(7, "sess-a", CIFailed, "x")); err != nil {
		t.Fatal(err)
	}
	handled := 0
	in.Docs = &refuseDocs{inner: e.st}
	if n, err := in.ConsumeNotices("sess-a", func(Notice) error { handled++; return nil }); n != 0 || err == nil {
		t.Fatalf("n=%d err=%v: an unwritten acknowledgement was counted", n, err)
	}
	in.Docs = e.st
	if n, err := in.ConsumeNotices("sess-a", func(Notice) error { handled++; return nil }); n != 1 || err != nil || handled != 2 {
		t.Errorf("repeat: n=%d err=%v handled=%d, want the one repeat", n, err, handled)
	}
}

func TestTheLeaderConsumesItsEscalations(t *testing.T) {
	e := newEnv(t)
	in := e.inbox()
	esc := Escalation{Issue: 7, Holder: "sess-a", To: "leader", Retries: 3, MaxRetries: 3}
	if err := in.Escalate(context.Background(), esc); err != nil {
		t.Fatal(err)
	}
	var got []Escalation
	if n, err := in.ConsumeEscalations("leader", func(x Escalation) error { got = append(got, x); return nil }); n != 1 || err != nil || got[0] != esc {
		t.Fatalf("n=%d err=%v got=%+v", n, err, got)
	}
	if left, _ := in.Escalations("leader"); len(left) != 0 {
		t.Errorf("left: %+v", left)
	}
}
