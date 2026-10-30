package mirror

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

func ledgerFor(n int, next string) core.Ledger {
	l := ledgerWith(next)
	l.Issue = core.IssueRef{Number: n}
	return l
}

const sensitive = "password=hunter2 and ghp_abcdefghijklmnopqrstuvwxyz0123456789"

// each write of a not yet remembered ledger lists the comments, then creates one
func okWrites(n int) []result {
	var out []result
	for i := 0; i < n; i++ {
		out = append(out, result{out: ""}, result{out: `{"id":` + string(rune('1'+i)) + `}`})
	}
	return out
}

// A ledger that is refused for good (a secret in it) is answered by a RESTRICTED
// record and leaves the queue: trying it again cannot help, and it must not hold
// the Issues behind it.
func TestFlushContinuesPastARefusedLedgerAndNamesIt(t *testing.T) {
	f := &fakeGH{results: okWrites(3)}
	m := writable(f)
	m.Enqueue(ledgerFor(1, sensitive))
	m.Enqueue(ledgerFor(2, "ordinary"))
	m.Enqueue(ledgerFor(3, "ordinary too"))

	n, err := m.Flush(context.Background())
	if n != 2 {
		t.Errorf("written = %d, want the two clean ledgers", n)
	}
	if !errors.Is(err, core.ErrUnsanitized) || !strings.Contains(err.Error(), "issue-1") {
		t.Errorf("the refusal must be reported with its Issue: %v", err)
	}
	if p := m.Pending(); len(p) != 0 {
		t.Errorf("nothing may stay queued behind a permanent refusal: %v", p)
	}
	all := strings.Join(calls(f), "\n")
	for _, want := range []string{"issues/2/comments", "issues/3/comments", "RESTRICTED"} {
		if !strings.Contains(all, want) {
			t.Errorf("%q never reached gh:\n%s", want, all)
		}
	}
	if strings.Contains(all, "hunter2") {
		t.Errorf("the secret reached gh")
	}
}

// A failure that may pass (a short budget, a block, the network) still stops the
// run and leaves the rest queued, so a refusal is not hammered; the permanent
// refusals before it are still reported and gone from the queue.
func TestFlushStillStopsAtATransientFailureAndReportsEarlierRefusals(t *testing.T) {
	f := &fakeGH{results: []result{
		{out: ""}, {out: `{"id":1}`}, // issue 1: the RESTRICTED record
		httpFail(503), // issue 2: the list call fails
	}}
	m := writable(f)
	m.Enqueue(ledgerFor(1, sensitive))
	m.Enqueue(ledgerFor(2, "ordinary"))
	m.Enqueue(ledgerFor(3, "ordinary too"))

	n, err := m.Flush(context.Background())
	if n != 0 {
		t.Errorf("written = %d", n)
	}
	if !errors.Is(err, core.ErrUnsanitized) || !strings.Contains(err.Error(), "issue-1") || !strings.Contains(err.Error(), "issue-2") {
		t.Errorf("both Issues must be named: %v", err)
	}
	if p := m.Pending(); len(p) != 2 || p[0] != 2 || p[1] != 3 {
		t.Errorf("pending = %v, want 2 and 3 still queued", p)
	}
}

// hookGH runs a callback on the first call, as another goroutine would.
type hookGH struct {
	*fakeGH
	onFirst func()
}

func (h *hookGH) Run(ctx context.Context, args ...string) ([]byte, []byte, int, error) {
	if h.onFirst != nil {
		f := h.onFirst
		h.onFirst = nil
		f()
	}
	return h.fakeGH.Run(ctx, args...)
}

// A newer ledger queued for the Issue while the refused one was being written
// stays queued: only the entry that was refused is dropped.
func TestFlushKeepsANewerLedgerQueuedWhileTheRefusedOneWasWritten(t *testing.T) {
	f := &fakeGH{results: okWrites(1)}
	m := writable(nil)
	h := &hookGH{fakeGH: f}
	m.Exec = h
	h.onFirst = func() { m.Enqueue(ledgerFor(1, "the secret is gone now")) }
	m.Enqueue(ledgerFor(1, sensitive))

	if _, err := m.Flush(context.Background()); !errors.Is(err, core.ErrUnsanitized) {
		t.Fatalf("flush: %v", err)
	}
	if p := m.Pending(); len(p) != 1 || p[0] != 1 {
		t.Errorf("the newer ledger must stay queued: %v", p)
	}
}

// A failure that belongs to one Issue for good (it is gone, or the host refuses its
// content) is a refusal like a secret: it leaves the queue and the run goes on.
func TestFlushContinuesPastAnIssueTheHostRefusesForGood(t *testing.T) {
	for _, status := range []int{404, 410, 422} {
		t.Run(statusName(status), func(t *testing.T) {
			first := []result{httpFail(status)}
			if status == 404 { // a 404 is only about this Issue when the repository itself can be read
				first = append(first, result{out: "{}"})
			}
			f := &fakeGH{results: append(first, okWrites(2)...)}
			m := writable(f)
			m.Enqueue(ledgerFor(1, "ordinary"))
			m.Enqueue(ledgerFor(2, "ordinary too"))
			m.Enqueue(ledgerFor(3, "and a third"))

			n, err := m.Flush(context.Background())
			var se *forge.StatusError
			if n != 2 || !errors.As(err, &se) || se.Status != status || !strings.Contains(err.Error(), "issue-1") {
				t.Errorf("written=%d err=%v", n, err)
			}
			if p := m.Pending(); len(p) != 0 {
				t.Errorf("nothing may stay queued behind it: %v", p)
			}
		})
	}
}

func statusName(n int) string {
	return map[int]string{404: "gone", 410: "deleted", 422: "unprocessable"}[n]
}

// Failures that are not about this Issue still stop the run: a permission or rate
// block, and a server that may recover.
func TestFlushStillStopsAtAGlobalOrTransientFailure(t *testing.T) {
	for name, fail := range map[string]result{
		"server error": httpFail(503),
		"forbidden":    httpFail(403, "body", `{"message":"Resource not accessible by integration"}`),
		"bad gateway":  httpFail(502),
	} {
		t.Run(name, func(t *testing.T) {
			f := &fakeGH{results: []result{fail}}
			m := writable(f)
			m.Enqueue(ledgerFor(1, "ordinary"))
			m.Enqueue(ledgerFor(2, "ordinary too"))
			if _, err := m.Flush(context.Background()); err == nil {
				t.Fatal("the failure must be reported")
			}
			if p := m.Pending(); len(p) != 2 {
				t.Errorf("both stay queued: %v", p)
			}
		})
	}
}

// With a store attached, a permanent refusal is also written to the Issue's
// mirror_state: refused for a secret, rejected(<status>) for the host's answer. An
// Issue the store does not know is skipped without an error of its own.
func TestFlushRecordsTheRefusalInTheClaimMirrorState(t *testing.T) {
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{1, 2, 3} {
		i, _ := core.NewIssueRef(n)
		if _, err := st.Acquire(i, "sess-a", "d"+string(rune('0'+n)), "run-1"); err != nil {
			t.Fatal(err)
		}
	}
	f := &fakeGH{results: []result{
		{out: ""}, {out: `{"id":1}`}, // issue 1: the RESTRICTED record
		httpFail(404), {out: "{}"}, // issue 2: gone, and the repository answers
		{out: ""}, {out: `{"id":3}`}, // issue 3: fine
	}}
	m := writable(f)
	m.Store = st
	m.Enqueue(ledgerFor(1, sensitive))
	m.Enqueue(ledgerFor(2, "ordinary"))
	m.Enqueue(ledgerFor(3, "and a third"))
	m.Enqueue(ledgerFor(9, sensitive)) // no claim record behind it
	f.results = append(f.results, result{out: ""}, result{out: `{"id":9}`})

	_, ferr := m.Flush(context.Background())
	if ferr == nil {
		t.Fatal("the refusals must be reported")
	}
	if strings.Contains(ferr.Error(), "recording") {
		t.Errorf("an Issue the store does not know is skipped, not reported: %v", ferr)
	}
	want := map[int]string{1: "refused", 2: "rejected(404)", 3: ""}
	for n, w := range want {
		i, _ := core.NewIssueRef(n)
		rec, err := st.Read(i)
		if err != nil || rec.Mirror != w {
			t.Errorf("issue-%d mirror_state = %q (%v), want %q", n, rec.Mirror, err, w)
		}
	}
}

// A 404 is also what the host answers when the token cannot see the repository at
// all. It is the Issue's own failure only if the repository can be read; when the
// repository read fails too, nothing is known about the Issue and the run stops with
// the rest queued.
func TestFlushDoesNotTakeA404ForAGoneIssueWhenTheRepositoryCannotBeRead(t *testing.T) {
	for name, probe := range map[string]result{
		"repository 404":      httpFail(404),
		"repository 403":      httpFail(403, "body", `{"message":"Resource not accessible by integration"}`),
		"repository 5xx":      httpFail(503),
		"repository no reply": {err: errors.New("network down")},
	} {
		t.Run(name, func(t *testing.T) {
			f := &fakeGH{results: []result{httpFail(404), probe}}
			m := writable(f)
			m.Enqueue(ledgerFor(1, "ordinary"))
			m.Enqueue(ledgerFor(2, "ordinary too"))
			n, err := m.Flush(context.Background())
			if n != 0 || err == nil {
				t.Fatalf("written=%d err=%v", n, err)
			}
			if p := m.Pending(); len(p) != 2 {
				t.Errorf("both stay queued: %v", p)
			}
		})
	}
}

// 410 and 422 are the Issue's own answer and need no repository read.
func TestFlushAsksNothingElseForA410Or422(t *testing.T) {
	for _, status := range []int{410, 422} {
		f := &fakeGH{results: []result{httpFail(status)}}
		m := writable(f)
		m.Enqueue(ledgerFor(1, "ordinary"))
		if _, err := m.Flush(context.Background()); err == nil {
			t.Fatal("reported")
		}
		if len(f.calls) != 1 {
			t.Errorf("status %d made %d calls, want only the failed one: %v", status, len(f.calls), calls(f))
		}
	}
}

// A locked Issue answers 403 with no limit signal and no permission wording, which
// arrives as an unclassified block. It is not dropped (the same answer can be a real
// denial, and dropping would empty the queue one Issue at a time) but it does not
// hold the Issues behind it either: it stays queued, is reported with its number and
// the run goes on.
func lockedIssue() result {
	return httpFail(403, "body", `{"message":"Unable to create comment because issue is locked."}`)
}

func TestFlushGoesPastALockedIssueAndKeepsItQueued(t *testing.T) {
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	i1, _ := core.NewIssueRef(1)
	if _, err := st.Acquire(i1, "sess-a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	f := &fakeGH{results: append([]result{lockedIssue()}, okWrites(2)...)}
	m := writable(f)
	m.Store = st
	m.Enqueue(ledgerFor(1, "locked"))
	m.Enqueue(ledgerFor(2, "ordinary"))
	m.Enqueue(ledgerFor(3, "ordinary too"))

	n, err := m.Flush(context.Background())
	var be *BlockedError
	if n != 2 || !errors.As(err, &be) || be.Reason != core.ReasonUnclassified || !strings.Contains(err.Error(), "issue-1") {
		t.Fatalf("written=%d err=%v", n, err)
	}
	if p := m.Pending(); len(p) != 1 || p[0] != 1 {
		t.Errorf("the locked Issue stays queued, the others are gone: %v", p)
	}
	if rec, _ := st.Read(i1); rec.Mirror != "blocked(unclassified-403)" {
		t.Errorf("mirror_state = %q", rec.Mirror)
	}
}

// Three of them in a row are not one locked Issue: the token is probably refused for
// the whole repository, so the run stops and the rest stays queued untouched.
func TestFlushStopsWhenTheSame403KeepsComing(t *testing.T) {
	f := &fakeGH{results: []result{lockedIssue(), lockedIssue(), lockedIssue(), lockedIssue()}}
	m := writable(f)
	for n := 1; n <= 5; n++ {
		m.Enqueue(ledgerFor(n, "ordinary"))
	}
	if _, err := m.Flush(context.Background()); err == nil {
		t.Fatal("reported")
	}
	if len(f.calls) != 3 {
		t.Errorf("the run must stop after the third 403, made %d calls: %v", len(f.calls), calls(f))
	}
	if p := m.Pending(); len(p) != 5 {
		t.Errorf("nothing is dropped: %v", p)
	}
}

// A success in between starts the count again.
func TestFlushCountsOnlyConsecutive403s(t *testing.T) {
	var seq []result
	for i := 0; i < 3; i++ {
		seq = append(seq, lockedIssue())
		seq = append(seq, okWrites(1)...)
	}
	f := &fakeGH{results: seq}
	m := writable(f)
	for n := 1; n <= 6; n++ {
		m.Enqueue(ledgerFor(n, "ordinary"))
	}
	n, err := m.Flush(context.Background())
	if n != 3 || err == nil {
		t.Fatalf("written=%d err=%v", n, err)
	}
	if p := m.Pending(); len(p) != 3 || p[0] != 1 || p[1] != 3 || p[2] != 5 {
		t.Errorf("the three locked Issues stay, in order: %v", p)
	}
}

// An Issue the store has no claim for has no mirror_state to record; that is skipped
// without an error of its own.
func TestFlushSkipsRecordingForALockedIssueNoClaimHolds(t *testing.T) {
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeGH{results: []result{lockedIssue()}}
	m := writable(f)
	m.Store = st
	m.Enqueue(ledgerFor(7, "locked"))
	_, ferr := m.Flush(context.Background())
	if ferr == nil || strings.Contains(ferr.Error(), "recording") {
		t.Errorf("the lock is reported, the missing claim is not: %v", ferr)
	}
}
