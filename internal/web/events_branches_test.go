package web

// events_branches_test.go — branches of the SSE hub that no test reached
// (Issue #1948, B15): a slow client is skipped, not blocked on; /events refuses
// non-GET and a writer that cannot stream. Tests only.
//
// Not tested, reported instead: ServeEvents' `case ev, ok := <-ch: if !ok` branch
// can only run if something other than ServeEvents' own deferred remove closes a
// client channel, and nothing does; remove closes unconditionally, so such a
// close would end in a double close (panic). The 25-second keepalive branch is
// clock-driven and is not worth a 25-second test.
//
// Defect answer: these tests catch a publish that blocks on a full client (one
// stuck browser freezing every screen), an /events that accepts a POST or that
// answers 200 where it cannot flush.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHubPublishSkipsAFullClientInsteadOfBlocking(t *testing.T) {
	h := NewHub()
	slow := h.add() // buffered 8, never drained
	fast := h.add()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 20; i++ { // more than the buffer holds
			h.Publish("factory")
			select { // keep the fast client drained
			case <-fast:
			default:
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Publish blocked on a client whose buffer was full")
	}
	if got := len(slow); got != cap(slow) {
		t.Errorf("the slow client holds %d events, want its full buffer of %d", got, cap(slow))
	}
	h.remove(slow)
	h.remove(fast)
}

func TestServeEventsRefusesWritesAndNonStreamingWriters(t *testing.T) {
	h := NewHub()
	rec := httptest.NewRecorder()
	h.ServeEvents(rec, httptest.NewRequest(http.MethodPost, "/events", nil))
	if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodGet {
		t.Errorf("POST /events = %d Allow %q, want 405 GET", rec.Code, rec.Header().Get("Allow"))
	}
	// A ResponseWriter without Flush cannot carry a stream.
	bare := &noFlushWriter{header: http.Header{}}
	h.ServeEvents(bare, httptest.NewRequest(http.MethodGet, "/events", nil))
	if bare.code != http.StatusInternalServerError || !strings.Contains(bare.body.String(), "streaming unsupported") {
		t.Errorf("a writer that cannot flush = %d %q, want 500 streaming unsupported", bare.code, bare.body.String())
	}
}

type noFlushWriter struct {
	header http.Header
	code   int
	body   strings.Builder
}

func (w *noFlushWriter) Header() http.Header         { return w.header }
func (w *noFlushWriter) WriteHeader(code int)        { w.code = code }
func (w *noFlushWriter) Write(b []byte) (int, error) { return w.body.Write(b) }
