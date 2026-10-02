package web

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Card t1446 (N2 — save-serialization hardening, from the t1411 sync-audit
// round-2 finding N2): two concurrent POST /save requests must not interleave
// their persistence steps. server.go's mutex guards only the listener field,
// so today a second request's write seams run while the first request is
// still mid-handler — a rolling-back request can revert another request's
// successful write. The serialized contract: request B's first persistence
// seam entry happens strictly after request A's handler (and therefore its
// own first seam) has finished.
//
// The discriminator is deterministic in BOTH directions without sleeps on the
// success path: request A parks inside its first write seam; request B is
// given a bounded window to (wrongly) enter the same seam while A is parked —
// that entry while A's seam has not exited is the serialization violation.
// With the save mutex, B cannot enter until A's handler returns, so the
// bounded window always times out, A is released, and B proceeds after it.

func TestSaveRequestsAreSerialized(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()

	// Every seam no-ops so a request that gets through runs to completion
	// without touching the real writers; writePreferences is the probe seam
	// both requests park on (first entry = request A, second = request B).
	nop := func(string, profile.ProfilePreferences) error { return nil }
	a.writePreferences = nop
	a.recordLastProfile = func(string) error { return nil }
	a.syncToProject = nop
	a.writeProjectConfig = func(string, string, string) error { return nil }
	a.writeProjectNestedConfig = func(string, projectNestedForm) error { return nil }
	a.applySchemaEdits = func(string, map[string]string) error { return nil }
	a.applyPerfTierEdits = func(string, string) error { return nil }
	a.patchAgentFM = func(string, map[string]config.ModelEffort, []string) error { return nil }
	a.glmcredSave = func(string) error { return nil }
	a.jevcredSave = func(string) error { return nil }

	var (
		entries      int32 // 1-based count of write-seam entries
		aSeamExited  int32 // set inside A's seam AFTER it unblocks — A's seam exit marker
		bEnteredEarly int32 // set when B's seam entry observed before A's seam exit
	)
	aReached := make(chan struct{})
	bReached := make(chan struct{})
	release := make(chan struct{})
	aDone := make(chan struct{})
	bDone := make(chan struct{})

	a.writePreferences = func(string, profile.ProfilePreferences) error {
		switch atomic.AddInt32(&entries, 1) {
		case 1:
			close(aReached)
			<-release
			atomic.StoreInt32(&aSeamExited, 1)
		case 2:
			close(bReached)
			if atomic.LoadInt32(&aSeamExited) == 0 {
				atomic.StoreInt32(&bEnteredEarly, 1)
			}
			<-release
		}
		return nil
	}

	// Request A: parks inside its first write seam until released.
	go func() {
		rec := servePost(t, h, "/save", reproForm())
		if rec.Code != http.StatusOK {
			t.Errorf("request A status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
		}
		close(aDone)
	}()
	<-aReached // A is now provably inside its write seam, parked.

	// Request B: under serialization it must BLOCK at the handler until A's
	// handler returns; without serialization its write seam entry interleaves
	// with A's parked seam — the violation.
	go func() {
		rec := servePost(t, h, "/save", reproForm())
		if rec.Code != http.StatusOK {
			t.Errorf("request B status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
		}
		close(bDone)
	}()

	// Bounded window for B to (wrongly) enter while A is parked. Under the
	// save mutex this window always times out — B cannot pass the handler
	// entry while A holds the save lock, so the wait is pure scheduling slack.
	select {
	case <-bReached:
		// B entered while A is parked; the ordering flag above captured
		// whether A's seam had already exited (it cannot have — A is parked).
	case <-time.After(2 * time.Second):
		// Serialized: B never reached its seam while A held the save.
	}

	// Let both proceed and finish; a stuck request means the lock deadlocked.
	close(release)
	select {
	case <-aDone:
	case <-time.After(5 * time.Second):
		t.Fatal("request A never completed — the save path deadlocked")
	}
	select {
	case <-bDone:
	case <-time.After(5 * time.Second):
		t.Fatal("request B never completed — the save path deadlocked")
	}

	if atomic.LoadInt32(&bEnteredEarly) == 1 {
		t.Fatal("request B entered its write seam while request A's handler was still in flight — concurrent saves are not serialized (card t1446 N2)")
	}
}
