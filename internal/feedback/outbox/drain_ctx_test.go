package outbox

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/bugreport"
)

// TestDrainContextHonorsDeadlineUnderLockContention pins the review-gate
// P2 finding: DrainContext discarded the caller's context, and the
// queue-lock waits accumulated PER ITEM — a live lock holder plus twelve
// spool items kept a 50ms-deadline drain running for over thirteen
// seconds, error-free. Cancellation must be honored throughout: the
// per-item loop stops on ctx.Done(), and the lock acquisition's retry
// loop selects on it, so a cancelled drain returns by its deadline with
// the spool batch unconsumed (the next drain retries it).
func TestDrainContextHonorsDeadlineUnderLockContention(t *testing.T) {
	spoolFixture(t,
		bugreport.KindPanic,
		bugreport.KindHookHandlerFailure,
		bugreport.KindInternalError,
		bugreport.KindPanic,
		bugreport.KindHookHandlerFailure,
		bugreport.KindInternalError,
		bugreport.KindPanic,
		bugreport.KindHookHandlerFailure,
		bugreport.KindInternalError,
		bugreport.KindPanic,
		bugreport.KindHookHandlerFailure,
		bugreport.KindInternalError,
	)
	consentOn(t)
	validBuildForTest(t)

	// A LIVE lock holder: this process claims the queue's sibling lock,
	// labels it with its own live identity, and holds it well past the
	// drain's deadline. Every drain mutation contends against it.
	home := os.Getenv("MOAI_HOME")
	lockPath := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir), "queue.lock")
	label, err := json.Marshal(atomicfile.LockOwner{PID: os.Getpid(), BootID: atomicfile.BootIDIdentity()})
	if err != nil {
		t.Fatalf("marshal label: %v", err)
	}
	if err := atomicfile.Claim(lockPath, 0o600); err != nil {
		t.Fatalf("claim holder lock: %v", err)
	}
	if err := os.WriteFile(lockPath, label, 0o600); err != nil {
		t.Fatalf("label holder lock: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(lockPath) })

	spoolPath, err := bugreport.SpoolPath()
	if err != nil {
		t.Fatalf("SpoolPath: %v", err)
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := DrainContext(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("a 200ms-deadline drain ran %s against a live lock holder — cancellation was not honored", elapsed)
	}

	// The spool batch was never consumed: the next drain retries it.
	raw, err := os.ReadFile(spoolPath)
	if err != nil || len(raw) == 0 {
		t.Fatalf("the cancelled drain consumed the spool batch: %v", err)
	}
	var entry bugreport.SpoolEntry
	if err := json.Unmarshal(raw[:firstLineLen(raw)], &entry); err != nil {
		t.Fatalf("decode surviving spool line: %v (%s)", err, raw)
	}
	if !entry.Kind.Valid() {
		t.Fatalf("the surviving spool line is not a valid entry: %s", raw)
	}
}

func firstLineLen(raw []byte) int {
	for i, b := range raw {
		if b == '\n' {
			return i
		}
	}
	return len(raw)
}
