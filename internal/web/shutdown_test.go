package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- handleShutdown unit tests (handler layer) ---

// TestShutdown_POST_Returns200AndTriggersShutdown verifies the POST /__shutdown__
// route returns 200 with a readable shutdown message AND invokes the injected
// triggerShutdown seam. The shutdown MUST be observable (the seam fires) and the
// response MUST be human-readable (never blank).
//
// The handler fires triggerShutdown in a goroutine (to avoid a deadlock where
// httpSrv.Shutdown waits for the handler to return). The test therefore signals
// via a channel inside the fake trigger and waits on it with a timeout — this is
// deterministic rather than racing the goroutine schedule.
func TestShutdown_POST_Returns200AndTriggersShutdown(t *testing.T) {
	a := newTestApp(t)

	triggered := make(chan struct{}, 1)
	a.triggerShutdown = func() {
		triggered <- struct{}{}
	}
	h := a.routes()

	rec := servePostShutdown(t, h)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /__shutdown__ status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	// The response must be a readable HTML page (never blank) — mirrors the
	// renderError minimal-HTML discipline but for a neutral/success banner.
	if strings.TrimSpace(body) == "" {
		t.Fatal("POST /__shutdown__ produced a blank page")
	}
	if !strings.Contains(strings.ToLower(body), "shut") {
		t.Errorf("shutdown response body does not mention shutdown:\n%s", body)
	}

	select {
	case <-triggered:
		// good — the goroutine fired the seam
	case <-time.After(2 * time.Second):
		t.Fatal("triggerShutdown seam was NOT invoked by POST /__shutdown__ within 2s")
	}
}

// TestShutdown_GET_Returns405 verifies the route rejects non-POST methods with
// 405 Method Not Allowed (same gate pattern as handleSave).
func TestShutdown_GET_Returns405(t *testing.T) {
	a := newTestApp(t)
	a.triggerShutdown = func() {} // wired but must NOT fire on GET
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/__shutdown__", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /__shutdown__ status = %d, want 405", rec.Code)
	}
}

// TestShutdown_NilTriggerSafe verifies a bare app (no triggerShutdown wired, as
// in a unit test that builds app directly) does NOT panic on POST — the handler
// nil-checks the seam before calling.
func TestShutdown_NilTriggerSafe(t *testing.T) {
	a := newTestApp(t)
	// a.triggerShutdown deliberately left nil.
	h := a.routes()

	rec := servePostShutdown(t, h)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /__shutdown__ with nil trigger status = %d, want 200 (nil must not panic)", rec.Code)
	}
}

// TestShutdown_ForeignHostRejected403 verifies the hostCheckMiddleware gates the
// shutdown route the same way it gates /save: a POST with a foreign Host is
// rejected 403 before reaching the handler.
func TestShutdown_ForeignHostRejected403(t *testing.T) {
	a := newTestApp(t)
	var triggered bool
	a.triggerShutdown = func() { triggered = true }
	h := a.routes()

	req := httptest.NewRequest(http.MethodPost, "/__shutdown__", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "attacker.example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("foreign-Host POST /__shutdown__ status = %d, want 403", rec.Code)
	}
	if triggered {
		t.Error("triggerShutdown fired on a foreign-Host POST (must be rejected before the handler)")
	}
}

// --- integration test (server layer) ---

// TestShutdown_RouteStopsServer verifies the full path: a real POST to
// /__shutdown__ against a running server causes ListenAndServe to return nil
// (clean graceful drain) within the 5s drain window. This is the end-to-end
// proof that the in-page button reuses the existing signal drain path rather
// than adding a parallel shutdown.
func TestShutdown_RouteStopsServer(t *testing.T) {
	srv, err := NewServer(newTestConfig(t))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var wg sync.WaitGroup
	var serveErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		serveErr = srv.ListenAndServe(ctx)
	}()
	waitForAddr(t, srv)
	addr := srv.Addr()

	// POST /__shutdown__ from a loopback Host — this should trigger the
	// existing signal/drain path and cause ListenAndServe to return.
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		req, _ := http.NewRequest(http.MethodPost, "http://"+addr+"/__shutdown__", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Host = "127.0.0.1"
		resp, perr := client.Do(req)
		if perr == nil {
			_ = resp.Body.Close()
		}
		// A connection-reset is also acceptable: the server may close the
		// connection mid-drain. The assertion that matters is serveErr below.
	}()

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		if serveErr != nil {
			t.Errorf("ListenAndServe returned %v after /__shutdown__, want nil", serveErr)
		}
	case <-time.After(shutdownDrain + 3*time.Second):
		t.Fatal("ListenAndServe did not return within the drain window after POST /__shutdown__")
	}
}

// --- helpers ---

// servePostShutdown is a thin helper that POSTs to /__shutdown__ with a loopback
// Host (so it passes hostCheckMiddleware) and returns the recorder.
func servePostShutdown(t *testing.T, h http.Handler) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/__shutdown__", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
