package feedback

// TestKillMidMutateLeavesQueueWritable (AC-018, D37): a holder killed
// mid-mutation leaves its lock behind — the next mutation reclaims it
// through the owner-verified stale-lock break, so the queue never wedges.
import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

func TestKillMidMutateLeavesQueueWritable(t *testing.T) {
	dir := t.TempDir()
	store := NewQueueStore(filepath.Join(dir, "queue.json"))

	// Simulate a killed holder: a lock labelled with a DEAD owner (another
	// boot) sits beside the queue — exactly what a mid-mutate kill leaves.
	lockPath := store.LockPath()
	identity := atomicfile.BootIDIdentity()
	if identity == "" {
		t.Skip("no boot identity on this platform")
	}
	dead := atomicfile.LockOwner{PID: os.Getpid(), BootID: "previous-boot-" + identity}
	if err := atomicfile.Claim(lockPath, 0o600); err != nil {
		t.Fatalf("claim: %v", err)
	}
	raw, err := json.Marshal(dead)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(lockPath, raw, 0o600); err != nil {
		t.Fatalf("label: %v", err)
	}

	// The next mutation reclaims and succeeds — the queue is writable.
	if err := store.MutateContext(context.Background(), func(rec *QueueRecord) error {
		rec.LastSeq++
		return nil
	}); err != nil {
		t.Fatalf("the mutation wedged on a killed holder's lock: %v", err)
	}
	if _, serr := os.Stat(lockPath); serr == nil {
		t.Fatal("the reclaimed lock file survived its release")
	}
}
