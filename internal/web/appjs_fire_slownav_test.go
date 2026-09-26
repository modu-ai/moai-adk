package web

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Card t1167 — the AC-AFG-015 (ii) full-navigation control read the swap
// premise the instant the settle wait reported "document replaced", while the
// navigated document could still be mid-parse. On a slow runner the profile
// trigger was not yet in the new DOM, so leg (d) read false and the control
// went red for a harness race, not for anything the app did.
//
// The race is made deterministic here by a test-side proxy that stalls the
// /todo document mid-body, immediately before the profile trigger's markup:
// the new document commits (the old one — and the probe's marker — is gone),
// but the trigger does not exist until the stall ends. The stall is bounded
// and ends early when the browser abandons the request, so nothing outlives
// the test (REQ-AFG-006).

// fireStallAnchor is the byte sequence the proxy stalls in front of: the
// profile trigger's attribute. Cutting inside the unfinished tag keeps the
// element out of the DOM until the remainder arrives.
const fireStallAnchor = `data-pop="profile"`

// startFireGuardStallingProxy fronts upstream and stalls every GET /todo
// response for `stall` just before fireStallAnchor. It returns the proxy's
// base URL and a counter of stalls actually applied, so a caller can prove
// the fixture exercised anything — an anchor that vanished from the page
// would otherwise make the stall a silent no-op.
func startFireGuardStallingProxy(t *testing.T, upstream string, stall time.Duration) (string, *atomic.Int32) {
	t.Helper()
	target, err := url.Parse(upstream)
	if err != nil {
		t.Fatalf("parse upstream %q: %v", upstream, err)
	}
	passthrough := httputil.NewSingleHostReverseProxy(target)
	var stalls atomic.Int32
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/todo" {
			passthrough.ServeHTTP(w, r)
			return
		}
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream+r.URL.RequestURI(), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		idx := bytes.Index(body, []byte(fireStallAnchor))
		if idx < 0 {
			http.Error(w, "stall anchor absent from /todo", http.StatusInternalServerError)
			return
		}
		for k, vs := range resp.Header {
			if k == "Content-Length" {
				continue
			}
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(body[:idx])
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		stalls.Add(1)
		select {
		case <-r.Context().Done():
			return
		case <-time.After(stall):
		}
		_, _ = w.Write(body[idx:])
	}))
	t.Cleanup(hs.Close)
	return hs.URL, &stalls
}

// TestAppJsFireSwapPremiseStalledNavigation pins the (ii) control against a
// navigated document that is still loading when the settle wait ends.
//
//   - stalled below the wait bound: the probe waits for the replaced document
//     to be ready, so the verdict is exactly the (ii) verdict — (a)(b)(c)
//     false, (d) true, exit 1, the post-swap indicator not judged fired.
//   - stalled beyond the wait bound: the readiness wait expires, and the
//     report says so by name rather than reading the premise off a
//     half-parsed document and moving on.
func TestAppJsFireSwapPremiseStalledNavigation(t *testing.T) {
	chromePath := requireFireGuardPrereqs(t)
	baseURL := startFireGuardServer(t)
	cdpPort := launchFireGuardChrome(t, chromePath)

	t.Run("stall_within_bound", func(t *testing.T) {
		proxyURL, stalls := startFireGuardStallingProxy(t, baseURL, 3*time.Second)
		code, r := runFireGuardProbe(t, fireFixtureFullNavigation(t), cdpPort, proxyURL, "t1167-ii-stall-within-bound", "--primary-entries-only")
		t.Logf("t1167 within-bound: exit=%d premise=%v false_legs=%v settle=%s ready=%v failures=%v",
			code, r.P5SwapPremise, r.P5SwapPremiseFalseLegs, r.P5SettleWait, fireBoolPtr(r.P5ReplacedDocumentReady), fireFailureReasons(r))
		if n := stalls.Load(); n < 1 {
			t.Fatalf("the proxy stalled /todo %d times (want >= 1) — the fixture exercised nothing", n)
		}
		if code != 1 {
			t.Errorf("exited %d (want 1)", code)
		}
		if r.P5SettleWait != "document replaced" {
			t.Errorf("p5_settle_wait = %q (want %q) — the run did not take the full-navigation path", r.P5SettleWait, "document replaced")
		}
		want := []string{"a_boost_ancestor", "b_same_document", "c_swap_events"}
		if !slicesEqual(r.P5SwapPremiseFalseLegs, want) {
			t.Errorf("p5_swap_premise_false_legs = %v (want %v)", r.P5SwapPremiseFalseLegs, want)
		}
		if !r.P5SwapPremise["d_swap_inserted_trigger"] {
			t.Error("premise leg d_swap_inserted_trigger = false (want true) — the premise was read off a document still loading")
		}
		if r.P5ReplacedDocumentReady == nil || !*r.P5ReplacedDocumentReady {
			t.Errorf("p5_replaced_document_ready = %s (want true)", fireBoolPtr(r.P5ReplacedDocumentReady))
		}
		if r.P6PopoverAfterSwapFired {
			t.Error("the post-swap indicator was judged fired after a full navigation")
		}
	})

	t.Run("stall_beyond_bound", func(t *testing.T) {
		proxyURL, stalls := startFireGuardStallingProxy(t, baseURL, 12*time.Second)
		code, r := runFireGuardProbe(t, fireFixtureFullNavigation(t), cdpPort, proxyURL, "t1167-ii-stall-beyond-bound", "--primary-entries-only")
		t.Logf("t1167 beyond-bound: exit=%d premise=%v false_legs=%v settle=%s ready=%v failures=%v",
			code, r.P5SwapPremise, r.P5SwapPremiseFalseLegs, r.P5SettleWait, fireBoolPtr(r.P5ReplacedDocumentReady), fireFailureReasons(r))
		if n := stalls.Load(); n < 1 {
			t.Fatalf("the proxy stalled /todo %d times (want >= 1) — the fixture exercised nothing", n)
		}
		if code != 1 {
			t.Errorf("exited %d (want 1)", code)
		}
		if r.P5ReplacedDocumentReady == nil || *r.P5ReplacedDocumentReady {
			t.Errorf("p5_replaced_document_ready = %s (want false)", fireBoolPtr(r.P5ReplacedDocumentReady))
		}
		named := false
		for _, f := range r.Failures {
			if f.Entry == "popover_after_swap" && strings.Contains(f.Reason, "replaced document not ready") {
				named = true
			}
		}
		if !named {
			t.Errorf("the expired readiness wait is not named on popover_after_swap: %v", fireFailureReasons(r))
		}
		if r.P6PopoverAfterSwapFired {
			t.Error("the post-swap indicator was judged fired after an expired readiness wait")
		}
	})
}

func fireBoolPtr(b *bool) string {
	if b == nil {
		return "<absent>"
	}
	if *b {
		return "true"
	}
	return "false"
}
