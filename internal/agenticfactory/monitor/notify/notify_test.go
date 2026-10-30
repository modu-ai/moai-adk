package notify

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/docstore"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// Everything here runs on a claim store in a temp directory, with fake channels: nothing is sent anywhere
// and no forge is written.

var t0 = time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)

type fakeNotifier struct {
	notices []Notice
	err     error // when set, every delivery fails
}

func (f *fakeNotifier) Notify(_ context.Context, n Notice) error {
	f.notices = append(f.notices, n)
	return f.err
}

func (f *fakeNotifier) count(k Kind) (n int) {
	for _, x := range f.notices {
		if x.Kind == k {
			n++
		}
	}
	return n
}

type fakeEscalator struct {
	got []Escalation
	err error
}

func (f *fakeEscalator) Escalate(_ context.Context, e Escalation) error {
	if f.err != nil {
		return f.err
	}
	f.got = append(f.got, e)
	return nil
}

type fakePRs struct {
	view  PRView
	found bool
	err   error
}

func (f *fakePRs) PR(context.Context, int) (PRView, bool, error) { return f.view, f.found, f.err }

type env struct {
	t   *testing.T
	st  *claim.Store
	rn  *runner.Runner
	now time.Time
	n   *fakeNotifier
	esc *fakeEscalator
	prs *fakePRs
}

const holder = "sess-a"

func newEnv(t *testing.T) *env {
	t.Helper()
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	e := &env{t: t, st: st, rn: runner.New(st), now: t0, n: &fakeNotifier{}, esc: &fakeEscalator{}, prs: &fakePRs{}}
	return e
}

// service is a fresh Service over the same store and memory: what a restarted monitor is.
func (e *env) service() *Service {
	return &Service{
		Records: e.st, Waits: e.rn, PRs: e.prs, Repairs: e.st, Notifier: e.n, Escalator: e.esc,
		State: docstore.CollectorCursors(e.st), Now: func() time.Time { return e.now }, ReNotifyAfter: 30 * time.Minute,
	}
}

func (e *env) acquire(issue int) runner.Ticket {
	e.t.Helper()
	ref, _ := core.NewIssueRef(issue)
	run, _ := core.ParseRunID("run-1")
	rec, err := e.st.Acquire(ref, holder, "cmd-1", run)
	if err != nil {
		e.t.Fatal(err)
	}
	return runner.Ticket{Issue: ref, Holder: holder, Epoch: rec.Epoch}
}

func waitAt(recheck time.Time, max int) core.Wait {
	return core.Wait{Owner: holder, NextAction: "re-run the failed check", RecheckAt: recheck,
		Diagnosis: "CI red on the latest head", MaxRetries: max, Escalate: "leader"}
}

func (e *env) tickAt(s *Service, at time.Time) []Action {
	e.t.Helper()
	e.now = at
	acts, err := s.Tick(context.Background())
	if err != nil {
		e.t.Fatal(err)
	}
	return acts
}

func (e *env) record(issue int) claim.Record {
	e.t.Helper()
	ref, _ := core.NewIssueRef(issue)
	r, err := e.st.Load(ref)
	if err != nil {
		e.t.Fatal(err)
	}
	return r
}

// A reminder that never gets an answer is repeated at each recheck, one retry is spent each time, and the
// retry that reaches the limit escalates exactly once — to the target the wait names — without moving the
// owner. After that the monitor stays quiet.
func TestALostReminderIsRepeatedThenEscalatedAtTheLimit(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	if _, err := e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3)); err != nil {
		t.Fatal(err)
	}
	seq0 := e.record(7).Seq
	s := e.service()

	if acts := e.tickAt(s, t0); len(e.n.notices) != 0 {
		t.Fatalf("before the recheck time: %v %v", acts, e.n.notices)
	}
	e.tickAt(s, t0.Add(time.Hour))
	if e.n.count(WaitOverdue) != 1 || e.n.notices[0].Attempt != 1 || e.n.notices[0].Holder != holder {
		t.Fatalf("first reminder: %+v", e.n.notices)
	}
	e.tickAt(s, t0.Add(time.Hour))
	if e.n.count(WaitOverdue) != 1 {
		t.Fatalf("a second pass at the same time must not remind again: %v", e.n.notices)
	}
	e.tickAt(s, t0.Add(90*time.Minute))
	e.tickAt(s, t0.Add(2*time.Hour))
	if e.n.count(WaitOverdue) != 3 || e.n.notices[2].Attempt != 3 {
		t.Fatalf("three reminders expected: %+v", e.n.notices)
	}
	rec := e.record(7)
	if rec.Wait == nil || !rec.Wait.Escalated || rec.Wait.Retries != 3 {
		t.Fatalf("the wait after the last retry: %+v", rec.Wait)
	}
	if len(e.esc.got) != 1 || e.esc.got[0].To != "leader" || e.esc.got[0].Issue != 7 || e.esc.got[0].Retries != 3 || e.esc.got[0].Holder != holder {
		t.Fatalf("escalation: %+v", e.esc.got)
	}
	for _, at := range []time.Duration{3 * time.Hour, 4 * time.Hour} {
		e.tickAt(s, t0.Add(at))
	}
	if e.n.count(WaitOverdue) != 3 || len(e.esc.got) != 1 {
		t.Errorf("after escalation: %d reminders, %d escalations", e.n.count(WaitOverdue), len(e.esc.got))
	}
	if rec.Holder != holder || e.record(7).Holder != holder || e.record(7).Epoch != tk.Epoch {
		t.Errorf("the owner changed: holder=%q epoch=%d", e.record(7).Holder, e.record(7).Epoch)
	}
	if got := e.record(7).Seq; got != seq0+3 {
		t.Errorf("seq = %d, want %d: the only writes are the three retries", got, seq0+3)
	}
}

// The owner showing progress means no reminder: it clears the wait, sets a later recheck, or writes the
// record (which earns one more period before the next reminder).
func TestOwnerProgressMeansNoReminder(t *testing.T) {
	t.Run("clears the wait", func(t *testing.T) {
		e := newEnv(t)
		tk := e.acquire(7)
		_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3))
		s := e.service()
		e.tickAt(s, t0)
		if _, err := e.rn.ClearWait(tk); err != nil {
			t.Fatal(err)
		}
		e.tickAt(s, t0.Add(5*time.Hour))
		if len(e.n.notices) != 0 || len(e.esc.got) != 0 {
			t.Errorf("notices=%v escalations=%v", e.n.notices, e.esc.got)
		}
	})
	t.Run("sets a later recheck", func(t *testing.T) {
		e := newEnv(t)
		tk := e.acquire(7)
		_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3))
		s := e.service()
		e.tickAt(s, t0)
		_, _ = e.rn.SetWait(tk, waitAt(t0.Add(4*time.Hour), 3))
		e.tickAt(s, t0.Add(2*time.Hour))
		e.tickAt(s, t0.Add(3*time.Hour))
		if len(e.n.notices) != 0 {
			t.Errorf("reminded before the new recheck time: %v", e.n.notices)
		}
	})
	t.Run("writes the record while the wait is stale", func(t *testing.T) {
		e := newEnv(t)
		tk := e.acquire(7)
		_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3))
		s := e.service()
		e.tickAt(s, t0) // the monitor has seen the record
		if _, err := e.st.Update(tk.Issue, tk.Holder, tk.Epoch, func(r *claim.Record) error { r.Head = core.HeadSHA(""); return nil }); err != nil {
			t.Fatal(err)
		}
		acts := e.tickAt(s, t0.Add(time.Hour))
		if len(e.n.notices) != 0 || len(acts) != 1 || acts[0].What != "progress-seen" {
			t.Fatalf("progress must be seen, not reminded: %v %v", acts, e.n.notices)
		}
		e.tickAt(s, t0.Add(time.Hour+10*time.Minute))
		if len(e.n.notices) != 0 {
			t.Errorf("reminded inside the quiet period: %v", e.n.notices)
		}
		e.tickAt(s, t0.Add(time.Hour+31*time.Minute))
		if e.n.count(WaitOverdue) != 1 {
			t.Errorf("a wait that stays stale after the quiet period is reminded once: %v", e.n.notices)
		}
	})
}

// A reminder that is known not to be delivered still spends the retry, so a broken channel ends in an
// escalation and not in an endless loop.
func TestABrokenChannelStillEndsInAnEscalation(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 2))
	e.n.err = errors.New("inbox unreachable")
	s := e.service()
	e.tickAt(s, t0.Add(time.Hour))
	e.tickAt(s, t0.Add(90*time.Minute))
	e.tickAt(s, t0.Add(3*time.Hour))
	if got := e.n.count(WaitOverdue); got != 2 {
		t.Errorf("%d delivery attempts, want one per retry", got)
	}
	rec := e.record(7)
	if rec.Wait == nil || !rec.Wait.Escalated || len(e.esc.got) != 1 {
		t.Fatalf("wait=%+v escalations=%v", rec.Wait, e.esc.got)
	}
	if rec.Wait.Diagnosis == "" || rec.Wait.Retries != 2 {
		t.Errorf("diagnosis=%q retries=%d", rec.Wait.Diagnosis, rec.Wait.Retries)
	}
}

// An escalation that could not be delivered is tried again on the next pass, once delivered it is never
// repeated, and a restarted monitor does not repeat it either.
func TestAnEscalationIsDeliveredOnceEvenAcrossFailuresAndRestarts(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 1))
	e.esc.err = errors.New("leader unreachable")
	s := e.service()
	acts := e.tickAt(s, t0.Add(time.Hour))
	failed := false
	for _, a := range acts {
		failed = failed || (a.What == "escalation" && a.Err != nil)
	}
	if !failed || len(e.esc.got) != 0 {
		t.Fatalf("acts=%v got=%v", acts, e.esc.got)
	}
	e.esc.err = nil
	e.tickAt(s, t0.Add(2*time.Hour))
	e.tickAt(e.service(), t0.Add(3*time.Hour)) // a restart
	e.tickAt(e.service(), t0.Add(4*time.Hour))
	if len(e.esc.got) != 1 || e.n.count(WaitOverdue) != 1 {
		t.Errorf("escalations=%d reminders=%d, want 1 and 1", len(e.esc.got), e.n.count(WaitOverdue))
	}
}

// A restarted monitor continues where the memory says: no repeat of a reminder already sent.
func TestARestartDoesNotRepeatAReminder(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3))
	e.tickAt(e.service(), t0.Add(time.Hour))
	e.tickAt(e.service(), t0.Add(time.Hour+time.Minute))
	e.tickAt(e.service(), t0.Add(time.Hour+2*time.Minute))
	if e.n.count(WaitOverdue) != 1 {
		t.Errorf("%d reminders after two restarts inside one period", e.n.count(WaitOverdue))
	}
	e.tickAt(e.service(), t0.Add(2*time.Hour))
	if e.n.count(WaitOverdue) != 2 {
		t.Errorf("%d reminders, want the second one at the next recheck", e.n.count(WaitOverdue))
	}
}

type movedWaiter struct{ spent int }

func (m *movedWaiter) RetryWait(runner.Ticket, string, string, time.Time) (claim.Record, error) {
	m.spent++
	return claim.Record{}, fmt.Errorf("%w: epoch moved", core.ErrOwnershipUnverified)
}

// A record that moved on between the read and the retry (new epoch, released) is left alone.
func TestARecordThatMovedOnIsNotRetried(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3))
	s := e.service()
	w := &movedWaiter{}
	s.Waits = w
	acts := e.tickAt(s, t0.Add(time.Hour))
	moved := false
	for _, a := range acts {
		moved = moved || a.What == "record-moved"
	}
	if !moved || w.spent != 1 || e.record(7).Wait.Retries != 0 {
		t.Errorf("acts=%v spent=%d retries=%d", acts, w.spent, e.record(7).Wait.Retries)
	}
}

type brokenState struct{}

func (brokenState) LoadCursor(string) ([]byte, bool, error) {
	return nil, false, errors.New("state unreadable")
}
func (brokenState) SaveCursor(string, []byte) error { return errors.New("state unwritable") }

func TestUnreadableMemoryStopsThePassForThatIssueInsteadOfGuessing(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3))
	s := e.service()
	s.State = brokenState{}
	acts := e.tickAt(s, t0.Add(2*time.Hour))
	if len(acts) != 1 || acts[0].What != "state-error" || len(e.n.notices) != 0 || e.record(7).Wait.Retries != 0 {
		t.Errorf("acts=%v notices=%v", acts, e.n.notices)
	}
}

func TestMisconfigurationIsRefused(t *testing.T) {
	e := newEnv(t)
	s := e.service()
	s.ReNotifyAfter = 0
	if _, err := s.Tick(context.Background()); err == nil {
		t.Error("a reminder interval of zero (not configured) was accepted")
	}
}

// Pull-request changes reach the holder — and only the holder — once each; a failed delivery is retried
// and later events wait behind it; nothing is written to the claim record.
func TestPullRequestChangesAreToldToTheHolderOnceEach(t *testing.T) {
	e := newEnv(t)
	e.acquire(7)
	seq0 := e.record(7).Seq
	s := e.service()
	head := func(c string) string { return fmt.Sprintf("%040s", c) }

	e.prs.found, e.prs.view = true, PRView{Number: 9, Head: head("a"), CI: "failure"}
	e.tickAt(s, t0)
	e.tickAt(s, t0.Add(time.Minute))
	if e.n.count(CIFailed) != 1 {
		t.Fatalf("a failed check is told once: %+v", e.n.notices)
	}
	e.prs.view = PRView{Number: 9, Head: head("b"), CI: "failure"} // a new head failing is new news
	e.tickAt(s, t0.Add(2*time.Minute))
	e.prs.view = PRView{Number: 9, Head: head("b"), CI: "success", Review: "changes_requested"}
	e.tickAt(s, t0.Add(3*time.Minute))
	e.prs.view.Review = "approved"
	e.tickAt(s, t0.Add(4*time.Minute))
	e.prs.view.Merged = true
	e.tickAt(s, t0.Add(5*time.Minute))
	e.prs.view.IssueClosed = true
	e.tickAt(s, t0.Add(6*time.Minute))
	e.tickAt(s, t0.Add(7*time.Minute))
	for kind, want := range map[Kind]int{CIFailed: 2, ChangesRequested: 1, Approved: 1, PRMerged: 1, IssueClosed: 1} {
		if got := e.n.count(kind); got != want {
			t.Errorf("%s notices = %d, want %d", kind, got, want)
		}
	}
	for _, n := range e.n.notices {
		if n.Holder != holder || n.Issue != 7 {
			t.Errorf("a notice went to %q for issue %d", n.Holder, n.Issue)
		}
	}
	if got := e.record(7); got.Seq != seq0 || got.Holder != holder {
		t.Errorf("the claim record changed: seq %d -> %d holder %q", seq0, got.Seq, got.Holder)
	}
}

func TestAFailedPullRequestNoticeIsRetriedNotLost(t *testing.T) {
	e := newEnv(t)
	e.acquire(7)
	s := e.service()
	e.prs.found, e.prs.view = true, PRView{Number: 9, Head: "a", CI: "failure", Merged: true}
	e.n.err = errors.New("inbox unreachable")
	e.tickAt(s, t0)
	if e.n.count(PRMerged) != 0 {
		t.Fatalf("a later event was sent behind a failed one: %+v", e.n.notices)
	}
	e.n.err = nil
	e.tickAt(s, t0.Add(time.Minute))
	e.tickAt(s, t0.Add(2*time.Minute))
	if e.n.count(CIFailed) != 2 || e.n.count(PRMerged) != 1 { // one failed attempt, one delivered; merged once
		t.Errorf("notices: %+v", e.n.notices)
	}
}

func TestNobodyIsToldAboutAnIssueNobodyHolds(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	if err := e.st.Release(tk.Issue, tk.Holder); err != nil {
		t.Fatal(err)
	}
	e.prs.found, e.prs.view = true, PRView{Number: 9, Head: "a", CI: "failure"}
	e.tickAt(e.service(), t0)
	if len(e.n.notices) != 0 {
		t.Errorf("notices for a released Issue: %v", e.n.notices)
	}
}

func TestAPullRequestReadErrorIsReportedAndTheWaitStillRuns(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	_, _ = e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3))
	e.prs.err = errors.New("host unreachable")
	acts := e.tickAt(e.service(), t0.Add(time.Hour))
	readErr := false
	for _, a := range acts {
		readErr = readErr || (a.What == "pr-read-error" && a.Err != nil)
	}
	if !readErr || e.n.count(WaitOverdue) != 1 {
		t.Errorf("acts=%v notices=%v", acts, e.n.notices)
	}
}

var _ forge.CursorStore = brokenState{}

// handedOff gives the record the receipt of an open pull request, the way a published PR leaves it.
func (e *env) handedOff(tk runner.Ticket, pr int) {
	e.t.Helper()
	_, err := e.st.Update(tk.Issue, tk.Holder, tk.Epoch, func(r *claim.Record) error {
		r.Receipt = &claim.Receipt{HandoffID: "h1", PR: pr, Head: core.HeadSHA(strings.Repeat("a", 40)), Branch: "WT-x", Worktree: "/w/x", Session: holder}
		return nil
	})
	if err != nil {
		e.t.Fatal(err)
	}
}

// A failed check or a change request is recorded once as the repair signal, on the holder's own record;
// the holder, epoch and state stay. The monitor's own write is not mistaken for the owner's progress.
func TestAFailedCheckIsSignalledOnceAndIsNotTakenForOwnerProgress(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	e.handedOff(tk, 9)
	if _, err := e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 3)); err != nil {
		t.Fatal(err)
	}
	s := e.service()
	e.tickAt(s, t0) // the monitor has seen the record
	before := e.record(7)
	// The failure is first seen on the very pass where the wait comes due: the signal write and the
	// reminder happen together, and the signal must not be read as the owner's progress.
	e.prs.found, e.prs.view = true, PRView{Number: 9, Head: strings.Repeat("a", 40), CI: "failure"}
	acts := e.tickAt(s, t0.Add(time.Hour))
	rec := e.record(7)
	if !strings.Contains(rec.RepairNeeded, "required check failed") || rec.Holder != holder || rec.Epoch != before.Epoch || rec.State != before.State {
		t.Fatalf("repair signal: %+v (acts %v)", rec, acts)
	}
	if e.n.count(WaitOverdue) != 1 {
		t.Fatalf("the monitor's own signal write was taken for owner progress: %v (acts %v)", e.n.notices, acts)
	}
	if rec.Seq != before.Seq+2 { // the signal and the one retry
		t.Errorf("seq %d -> %d, want the signal and the retry only", before.Seq, rec.Seq)
	}
	e.tickAt(s, t0.Add(time.Hour+time.Minute))
	if e.record(7).Seq != rec.Seq {
		t.Errorf("the same failure was signalled again")
	}
	e.prs.view = PRView{Number: 9, Head: strings.Repeat("b", 40), CI: "failure"} // a new head failing is a new signal
	e.tickAt(s, t0.Add(time.Hour+time.Minute))
	e.prs.view = PRView{Number: 9, Head: strings.Repeat("b", 40), CI: "success", Review: "changes_requested"}
	e.tickAt(s, t0.Add(time.Hour+2*time.Minute))
	if got := e.record(7).RepairNeeded; !strings.Contains(got, "changes were requested") {
		t.Errorf("a change request is a repair signal too: %q", got)
	}
}

type failingRepairs struct{ calls int }

func (f *failingRepairs) MarkRepairNeeded(core.IssueRef, string) (claim.Record, error) {
	f.calls++
	return claim.Record{}, errors.New("store unavailable")
}

func TestARepairSignalThatFailedIsRetriedAndAnIssueWithoutAnOpenPRIsLeftAlone(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	e.handedOff(tk, 9)
	e.prs.found, e.prs.view = true, PRView{Number: 9, Head: strings.Repeat("a", 40), CI: "failure"}
	s := e.service()
	f := &failingRepairs{}
	s.Repairs = f
	e.tickAt(s, t0)
	e.tickAt(s, t0.Add(time.Minute))
	if f.calls != 2 || e.record(7).RepairNeeded != "" {
		t.Errorf("calls=%d signal=%q: a failed signal must be tried again", f.calls, e.record(7).RepairNeeded)
	}

	// no receipt: there is no open pull request on the record, so there is nothing to repair
	e2 := newEnv(t)
	e2.acquire(8)
	e2.prs.found, e2.prs.view = true, PRView{Number: 9, Head: strings.Repeat("a", 40), CI: "failure"}
	s2 := e2.service()
	acts := e2.tickAt(s2, t0)
	e2.tickAt(s2, t0.Add(time.Minute))
	for _, a := range acts {
		if a.What == "repair-signal" && a.Err != nil {
			t.Errorf("a record without an open PR is not an error: %v", a.Err)
		}
	}
	if e2.record(8).RepairNeeded != "" {
		t.Errorf("signal on a record without a receipt: %q", e2.record(8).RepairNeeded)
	}
}

type busyNotifier struct {
	busy  bool
	calls int
	sent  int
}

func (b *busyNotifier) Notify(context.Context, Notice) error {
	b.calls++
	if b.busy {
		return fmt.Errorf("%w: documents", core.ErrBusy)
	}
	b.sent++
	return nil
}

// A busy document lock is "retry", not a failed delivery: the reminder is tried again on the next pass
// and no retry of the wait is spent, so lock contention alone can never walk a wait to its escalation.
func TestABusyInboxDoesNotSpendARetry(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	if _, err := e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 2)); err != nil {
		t.Fatal(err)
	}
	b := &busyNotifier{busy: true}
	s := e.service()
	s.Notifier = b
	for i := range 4 { // more passes than the wait has retries
		acts := e.tickAt(s, t0.Add(time.Hour+time.Duration(i)*time.Minute))
		deferred := false
		for _, a := range acts {
			deferred = deferred || (a.What == "reminder-deferred" && errors.Is(a.Err, core.ErrBusy))
		}
		if !deferred {
			t.Fatalf("pass %d: %+v, want the reminder reported as deferred by Busy", i, acts)
		}
	}
	rec := e.record(7)
	if rec.Wait.Retries != 0 || rec.Wait.Escalated || len(e.esc.got) != 0 {
		t.Fatalf("Busy spent retries: %+v escalations=%v", rec.Wait, e.esc.got)
	}
	if b.calls != 4 {
		t.Errorf("delivery attempts = %d, want one per pass", b.calls)
	}
	b.busy = false
	e.tickAt(s, t0.Add(time.Hour+10*time.Minute))
	if b.sent != 1 || e.record(7).Wait.Retries != 1 {
		t.Errorf("after the lock cleared: sent=%d retries=%d, want the reminder delivered and one retry spent", b.sent, e.record(7).Wait.Retries)
	}
}

// F3 (notify side): contention alone never spends a retry, but a lock that stays busy pass after pass is
// reported — it may be a crashed holder's leftover, and silence would hide the lost reminders for good.
func TestALockBusyPassAfterPassIsReportedAsStuckWithoutSpendingARetry(t *testing.T) {
	e := newEnv(t)
	tk := e.acquire(7)
	if _, err := e.rn.SetWait(tk, waitAt(t0.Add(time.Hour), 2)); err != nil {
		t.Fatal(err)
	}
	b := &busyNotifier{busy: true}
	s := e.service()
	s.Notifier = b
	stuck := func(acts []Action) bool {
		for _, a := range acts {
			if a.What == "inbox-lock-stuck" && errors.Is(a.Err, core.ErrBusy) {
				return true
			}
		}
		return false
	}
	if stuck(e.tickAt(s, t0.Add(time.Hour))) {
		t.Fatal("one busy pass was reported as stuck")
	}
	for i := 1; i < 4; i++ {
		if !stuck(e.tickAt(s, t0.Add(time.Hour+time.Duration(i)*time.Minute))) {
			t.Fatalf("pass %d: the lock has been busy for %d passes (the wait's own retry count) and was not reported", i, i+1)
		}
	}
	if rec := e.record(7); rec.Wait.Retries != 0 || rec.Wait.Escalated {
		t.Fatalf("contention spent retries: %+v", rec.Wait)
	}
	b.busy = false
	e.tickAt(s, t0.Add(time.Hour+10*time.Minute))
	b.busy = true
	if stuck(e.tickAt(s, t0.Add(time.Hour+20*time.Minute))) {
		t.Error("the busy run was not reset by a delivery that went through")
	}
}
