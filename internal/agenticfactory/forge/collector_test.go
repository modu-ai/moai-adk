package forge

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func adaptive() Adaptive {
	return Adaptive{Min: 10 * time.Second, Max: 5 * time.Minute, Safety: 30 * time.Minute}
}

func collectorWith(f *fakeT, now *time.Time, feed WebhookFeed) *Collector {
	c := &Client{T: f, B: &Budget{}, Backoff: cfg(), Now: func() time.Time { return *now }}
	return &Collector{Client: c, Path: "/repos/o/r/pulls", Interval: adaptive(), Differ: DiffByKey("number"), Feed: feed}
}

func TestAdaptiveCollectorSendsChangesOnlyAndStretchesWhenIdle(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{
		resp(200, `[{"number":1,"state":"open"}]`, "etag", `"a"`),
		resp(304, ""),
		resp(304, ""),
		resp(200, `[{"number":1,"state":"merged"},{"number":2,"state":"open"}]`, "etag", `"b"`),
	}}
	col := collectorWith(f, &now, nil)

	rep := col.Tick(t.Context(), now)
	if len(rep.Changes) != 1 || rep.Changes[0].Key != "1" || rep.Changes[0].Kind != ChangeAdded || !rep.Baseline {
		t.Fatalf("first tick = %+v", rep)
	}
	_ = col.Ack(rep)
	first := rep.Next.Sub(now)
	now = rep.Next
	rep = col.Tick(t.Context(), now) // 304: nothing changed
	if len(rep.Changes) != 0 || rep.Next.Sub(now) <= first {
		t.Fatalf("idle tick must stretch the interval: %+v first=%v next=%v", rep, first, rep.Next.Sub(now))
	}
	_ = col.Ack(rep)
	now = rep.Next
	rep = col.Tick(t.Context(), now)
	_ = col.Ack(rep)
	now = rep.Next
	rep = col.Tick(t.Context(), now) // content changed
	kinds := map[string]ChangeKind{}
	for _, c := range rep.Changes {
		kinds[c.Key] = c.Kind
	}
	if kinds["1"] != ChangeModified || kinds["2"] != ChangeAdded || len(rep.Changes) != 2 {
		t.Fatalf("changes = %+v", rep.Changes)
	}
	if rep.Next.Sub(now) > 5*time.Minute || rep.Next.Sub(now) < 10*time.Second {
		t.Errorf("interval out of bounds: %v", rep.Next.Sub(now))
	}
}

func TestTickBeforeNextMakesNoCall(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(200, `[]`)}}
	col := collectorWith(f, &now, nil)
	rep := col.Tick(t.Context(), now)
	_ = col.Tick(t.Context(), now.Add(time.Second))
	_ = col.Tick(t.Context(), now.Add(2*time.Second))
	if len(f.calls) != 1 || rep.Next.IsZero() {
		t.Errorf("calls = %d", len(f.calls))
	}
}

type fakeFeed struct {
	healthy bool
	events  []Change
}

func (f *fakeFeed) Healthy() bool   { return f.healthy }
func (f *fakeFeed) Drain() []Change { e := f.events; f.events = nil; return e }

func TestWebhookFirstPollsOnlyAsSafetyNet(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(200, `[]`)}}
	feed := &fakeFeed{healthy: true, events: []Change{{Key: "7", Kind: ChangeModified}}}
	col := collectorWith(f, &now, feed)
	rep := col.Tick(t.Context(), now)
	if len(rep.Changes) != 1 || rep.Changes[0].Key != "7" {
		t.Fatalf("webhook events not delivered: %+v", rep)
	}
	if len(f.calls) != 0 {
		t.Errorf("polled %d times while the webhook is healthy and no safety poll is due", len(f.calls))
	}
	if rep.Next.Sub(now) > adaptive().Safety {
		t.Errorf("next wake %v beyond the safety interval", rep.Next.Sub(now))
	}
	feed.healthy = false // webhook lost: fall back to polling
	now = now.Add(time.Second)
	_ = col.Tick(t.Context(), now.Add(adaptive().Safety))
	if len(f.calls) != 1 {
		t.Errorf("no polling fallback when the webhook is unhealthy: calls=%d", len(f.calls))
	}
}

func TestDeferredPollRespectsNotBeforeAndPollIntervalHeader(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(429, "", "retry-after", "120")}}
	col := collectorWith(f, &now, nil)
	rep := col.Tick(t.Context(), now)
	if rep.Next.Sub(now) != 120*time.Second || len(rep.Changes) != 0 {
		t.Errorf("rep = %+v", rep)
	}
	f2 := &fakeT{resps: []Response{resp(200, `[]`, "x-poll-interval", "600")}}
	col2 := collectorWith(f2, &now, nil)
	if r := col2.Tick(t.Context(), now); r.Next.Sub(now) < 600*time.Second {
		t.Errorf("x-poll-interval ignored: %v", r.Next.Sub(now))
	}
}

func TestDeniedIsParkedNotRetried(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(403, `{"message":"Resource not accessible"}`)}}
	col := collectorWith(f, &now, nil)
	rep := col.Tick(t.Context(), now)
	if !rep.Parked || !strings.Contains(rep.Reason, "denied") {
		t.Fatalf("rep = %+v", rep)
	}
	_ = col.Tick(t.Context(), now.Add(24*time.Hour))
	if len(f.calls) != 1 {
		t.Errorf("a parked collector kept calling: %d", len(f.calls))
	}
}

func TestRegistryGivesOneCollectorPerRepo(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(200, `[]`)}}
	reg := &Registry{New: func(repo string) *Collector { return collectorWith(f, &now, nil) }}
	a, b := reg.For("o/r"), reg.For("o/r")
	if a != b || reg.For("o/other") == a {
		t.Error("registry must return the same collector for a repo and a different one for another")
	}
}

func TestDiffByKeyDetectsAddModifyRemove(t *testing.T) {
	d := DiffByKey("number")
	ch, err := d([]byte(`[{"number":1,"s":"a"},{"number":2,"s":"a"}]`), []byte(`[{"number":2,"s":"b"},{"number":3,"s":"a"}]`))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]ChangeKind{}
	for _, c := range ch {
		got[c.Key] = c.Kind
	}
	if got["1"] != ChangeRemoved || got["2"] != ChangeModified || got["3"] != ChangeAdded || len(ch) != 3 {
		t.Errorf("changes = %+v", ch)
	}
	if _, err := d(nil, []byte(`not json`)); err == nil {
		t.Error("bad json accepted")
	}
}

func TestUnacknowledgedChangesComeBackEvenWhenTheHostAnswers304(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(200, `[{"number":1,"state":"open"}]`, "etag", `"a"`), resp(304, "")}}
	col := collectorWith(f, &now, nil)
	rep := col.Tick(t.Context(), now) // delivery fails: no Ack
	now = rep.Next
	again := col.Tick(t.Context(), now)
	if len(again.Changes) != 1 || again.Changes[0].Key != "1" {
		t.Fatalf("a change that was never acknowledged was lost: %+v", again.Changes)
	}
	_ = col.Ack(again)
	now = again.Next
	if third := col.Tick(t.Context(), now); len(third.Changes) != 0 {
		t.Errorf("an acknowledged change was delivered again: %+v", third.Changes)
	}
}

func TestRestartReportsABaselineNotNewEvents(t *testing.T) {
	now := t0
	body := `[{"number":1},{"number":2}]`
	col := collectorWith(&fakeT{resps: []Response{resp(200, body, "etag", `"a"`)}}, &now, nil)
	rep := col.Tick(t.Context(), now)
	_ = col.Ack(rep)
	restarted := collectorWith(&fakeT{resps: []Response{resp(200, body, "etag", `"a"`)}}, &now, nil) // in-memory state is gone
	r2 := restarted.Tick(t.Context(), now)
	if !r2.Baseline || len(r2.Changes) != 2 {
		t.Fatalf("restart report = %+v, want a flagged baseline", r2)
	}
	_ = restarted.Ack(r2)
	later := restarted.Tick(t.Context(), r2.Next)
	if later.Baseline {
		t.Error("baseline flag stuck after acknowledgement")
	}
}

func TestWebhookEventsSurviveUntilAcknowledged(t *testing.T) {
	now := t0
	feed := &fakeFeed{healthy: true, events: []Change{{Key: "7", Kind: ChangeModified}}}
	col := collectorWith(&fakeT{resps: []Response{resp(200, `[]`)}}, &now, feed)
	r1 := col.Tick(t.Context(), now)
	r2 := col.Tick(t.Context(), now.Add(time.Second)) // not acknowledged
	if len(r1.Changes) != 1 || len(r2.Changes) != 1 {
		t.Fatalf("r1=%v r2=%v", r1.Changes, r2.Changes)
	}
	_ = col.Ack(r2)
	if r3 := col.Tick(t.Context(), now.Add(2*time.Second)); len(r3.Changes) != 0 {
		t.Errorf("acknowledged event repeated: %v", r3.Changes)
	}
}

// --- persisted cursor: the acknowledged baseline survives a restart --------------------------------

type memCursors struct {
	m       map[string][]byte
	saves   int
	loadErr error
	saveErr error
}

func (s *memCursors) LoadCursor(key string) ([]byte, bool, error) {
	if s.loadErr != nil {
		return nil, false, s.loadErr
	}
	b, ok := s.m[key]
	return b, ok, nil
}

func (s *memCursors) SaveCursor(key string, body []byte) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	if s.m == nil {
		s.m = map[string][]byte{}
	}
	s.m[key] = append([]byte(nil), body...)
	s.saves++
	return nil
}

func withCursors(f *fakeT, now *time.Time, cs CursorStore) *Collector {
	col := collectorWith(f, now, nil)
	col.Cursors, col.CursorKey = cs, "o/r/pulls"
	return col
}

func TestARestartResumesFromTheAcknowledgedCursorAndReportsOnlyWhatChanged(t *testing.T) {
	now := t0
	cs := &memCursors{}
	first := withCursors(&fakeT{resps: []Response{resp(200, `[{"number":1,"s":"open"},{"number":2,"s":"open"}]`, "etag", `"a"`)}}, &now, cs)
	rep := first.Tick(t.Context(), now)
	if err := first.Ack(rep); err != nil || cs.saves != 1 {
		t.Fatalf("ack err=%v saves=%d", err, cs.saves)
	}

	// a new process: no memory, same cursor store; the host now shows one change
	restarted := withCursors(&fakeT{resps: []Response{resp(200, `[{"number":1,"s":"merged"},{"number":2,"s":"open"}]`, "etag", `"b"`)}}, &now, cs)
	r2 := restarted.Tick(t.Context(), now)
	if r2.Baseline {
		t.Error("a restart with an acknowledged cursor must not report a baseline")
	}
	if len(r2.Changes) != 1 || r2.Changes[0].Key != "1" || r2.Changes[0].Kind != ChangeModified {
		t.Errorf("changes = %+v, want only issue 1 modified", r2.Changes)
	}
}

func TestACursorIsSavedOnlyWhenTheConsumerAcknowledges(t *testing.T) {
	now := t0
	cs := &memCursors{}
	col := withCursors(&fakeT{resps: []Response{resp(200, `[{"number":1}]`)}}, &now, cs)
	col.Tick(t.Context(), now)
	if cs.saves != 0 {
		t.Fatalf("a cursor was saved before any acknowledgement (%d saves)", cs.saves)
	}
}

func TestAFailedSaveKeepsTheChangesComingBack(t *testing.T) {
	now := t0
	cs := &memCursors{saveErr: errors.New("disk full")}
	col := withCursors(&fakeT{resps: []Response{resp(200, `[{"number":1}]`, "etag", `"a"`), resp(304, "")}}, &now, cs)
	rep := col.Tick(t.Context(), now)
	if err := col.Ack(rep); err == nil {
		t.Fatal("a failed save was reported as acknowledged")
	}
	now = rep.Next
	again := col.Tick(t.Context(), now)
	if len(again.Changes) != 1 || !again.Baseline {
		t.Errorf("an unsaved acknowledgement must not advance the cursor: %+v", again)
	}
}

func TestAnUnreadableCursorPausesInsteadOfInventingABaseline(t *testing.T) {
	now := t0
	cs := &memCursors{loadErr: errors.New("store unavailable")}
	f := &fakeT{resps: []Response{resp(200, `[{"number":1}]`)}}
	rep := withCursors(f, &now, cs).Tick(t.Context(), now)
	if len(f.calls) != 0 || len(rep.Changes) != 0 || rep.Reason == "" || rep.Next.IsZero() {
		t.Errorf("rep=%+v calls=%d: with the cursor unreadable nothing may be reported or polled", rep, len(f.calls))
	}
	cs.loadErr = nil // the store recovers; the next tick loads it and proceeds
	later := withCursors(f, &now, cs)
	if r := later.Tick(t.Context(), rep.Next); r.Reason != "" || len(f.calls) != 1 {
		t.Errorf("after recovery: %+v calls=%d", r, len(f.calls))
	}
}

func TestNoCursorStoreKeepsTheOldInMemoryBehaviour(t *testing.T) {
	now := t0
	col := collectorWith(&fakeT{resps: []Response{resp(200, `[{"number":1}]`)}}, &now, nil)
	rep := col.Tick(t.Context(), now)
	if err := col.Ack(rep); err != nil {
		t.Errorf("ack without a store: %v", err)
	}
}

// --- durable push channel: events leave the queue only after the consumer acknowledged them -----------

type memDurable struct {
	healthy    bool
	events     []PendingChange
	pendingErr error
	commitErr  error
	commits    [][]string
}

func (d *memDurable) Healthy() bool { return d.healthy }

func (d *memDurable) Pending() ([]PendingChange, error) {
	if d.pendingErr != nil {
		return nil, d.pendingErr
	}
	return append([]PendingChange(nil), d.events...), nil
}

func (d *memDurable) Commit(ids []string) error {
	if d.commitErr != nil {
		return d.commitErr
	}
	d.commits = append(d.commits, ids)
	var keep []PendingChange
	for _, e := range d.events {
		drop := false
		for _, id := range ids {
			drop = drop || e.ID == id
		}
		if !drop {
			keep = append(keep, e)
		}
	}
	d.events = keep
	return nil
}

func durableCollector(d DurableFeed) (*Collector, *time.Time) {
	now := t0
	col := collectorWith(&fakeT{resps: []Response{resp(200, `[]`)}}, &now, nil)
	col.Durable = d
	return col, &now
}

func TestDurableEventsAreRepeatedUntilAcknowledgedThenCommitted(t *testing.T) {
	d := &memDurable{healthy: true, events: []PendingChange{{ID: "d1", Change: Change{Key: "7", Kind: ChangeModified}}}}
	col, now := durableCollector(d)
	r1 := col.Tick(t.Context(), *now)
	r2 := col.Tick(t.Context(), now.Add(time.Second)) // not acknowledged
	if len(r1.Changes) != 1 || len(r2.Changes) != 1 || len(d.commits) != 0 {
		t.Fatalf("r1=%v r2=%v commits=%v", r1.Changes, r2.Changes, d.commits)
	}
	if err := col.Ack(r2); err != nil {
		t.Fatal(err)
	}
	if len(d.commits) != 1 || d.commits[0][0] != "d1" {
		t.Errorf("commits = %v, want d1 committed after the ack", d.commits)
	}
	if r3 := col.Tick(t.Context(), now.Add(2*time.Second)); len(r3.Changes) != 0 {
		t.Errorf("committed event repeated: %v", r3.Changes)
	}
}

// A commit that fails is reported, and the event stays queued: it comes back, it is never dropped.
func TestAFailedCommitKeepsTheEventQueued(t *testing.T) {
	d := &memDurable{healthy: true, events: []PendingChange{{ID: "d1", Change: Change{Key: "7", Kind: ChangeModified}}}, commitErr: errors.New("store down")}
	col, now := durableCollector(d)
	rep := col.Tick(t.Context(), *now)
	if err := col.Ack(rep); err == nil {
		t.Fatal("a failed commit was reported as acknowledged")
	}
	d.commitErr = nil
	if again := col.Tick(t.Context(), now.Add(time.Second)); len(again.Changes) != 1 {
		t.Errorf("the event must come back: %v", again.Changes)
	}
}

// A queue that cannot be read is said so; the collector does not pretend there is nothing pending.
func TestAnUnreadableDurableQueueIsReported(t *testing.T) {
	d := &memDurable{healthy: true, pendingErr: errors.New("corrupt queue")}
	col, now := durableCollector(d)
	rep := col.Tick(t.Context(), *now)
	if !strings.Contains(rep.Reason, "corrupt queue") {
		t.Errorf("reason = %q, want the queue error", rep.Reason)
	}
}

// A healthy durable channel turns polling into a safety net, like the in-memory feed.
func TestAHealthyDurableFeedMakesThePollASafetyNet(t *testing.T) {
	d := &memDurable{healthy: true}
	now := t0
	f := &fakeT{resps: []Response{resp(200, `[]`)}}
	col := collectorWith(f, &now, nil)
	col.Durable = d
	rep := col.Tick(t.Context(), now)
	if len(f.calls) != 0 || rep.Next.Sub(now) != adaptive().Safety {
		t.Errorf("calls=%d next=%v, want no call and the safety interval", len(f.calls), rep.Next.Sub(now))
	}
}

// --- a cursor that was resumed from an older generation says so --------------------------------------

type notedCursors struct {
	memCursors
	note string
}

func (n *notedCursors) LoadCursorNoted(key string) ([]byte, bool, string, error) {
	b, ok, err := n.LoadCursor(key)
	return b, ok, n.note, err
}

func TestACursorNoteIsOnTheReportsUntilAnAcknowledgementIsSaved(t *testing.T) {
	now := t0
	cs := &notedCursors{memCursors: memCursors{m: map[string][]byte{"o/r/pulls": []byte(`[{"number":1,"state":"open"}]`)}}, note: "resumed from an older generation"}
	f := &fakeT{resps: []Response{resp(200, `[{"number":1,"state":"merged"}]`, "etag", `"a"`), resp(304, "")}}
	col := withCursors(f, &now, cs)
	r1 := col.Tick(t.Context(), now)
	if r1.Notice != cs.note || r1.Baseline {
		t.Fatalf("first report: notice=%q baseline=%v", r1.Notice, r1.Baseline)
	}
	if err := col.Ack(r1); err != nil {
		t.Fatal(err)
	}
	if r2 := col.Tick(t.Context(), r1.Next); r2.Notice != "" {
		t.Errorf("the note outlived the saved acknowledgement: %q", r2.Notice)
	}
}
