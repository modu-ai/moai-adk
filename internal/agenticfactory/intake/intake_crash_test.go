package intake

import "testing"

// t2072 (RC40 carried-over review finding, intake/intake.go:246): PublishIssue was
// sent with nothing recorded beforehand. A process that dies after the host
// created the Issue but before the request row says "unknown" restarts at
// StageAccepted with no Blocked row; GitHub gives no idempotency and its search
// can lag, so the duplicate pre-flight finds nothing and a second Issue is
// created for the same request.
func TestAProcessDyingDuringPublishDoesNotCreateASecondIssue(t *testing.T) {
	r := newRig(t, Config{})
	if _, err := r.in.Accept("crash-key", []byte("payload"), 1); err != nil {
		t.Fatal(err)
	}

	r.pub.mu.Lock()
	r.pub.crash = true
	r.pub.mu.Unlock()
	func() {
		defer func() { _ = recover() }() // the process died here
		_, _ = r.in.Advance(ctx(), "crash-key")
	}()
	if r.pub.createdCount() != 1 {
		t.Fatalf("setup: %d issues created before the restart, want 1", r.pub.createdCount())
	}

	// restart: the host answers, the duplicate search has not indexed the new Issue
	r.pub.mu.Lock()
	r.pub.crash = false
	r.pub.mu.Unlock()
	_, _ = r.in.Advance(ctx(), "crash-key")

	if got := r.pub.createdCount(); got != 1 {
		t.Errorf("%d issues exist for one request after a restart, want 1", got)
	}
}
