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
			f := &fakeGH{results: append([]result{httpFail(status)}, okWrites(2)...)}
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
		httpFail(404),                // issue 2: gone
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
