package bugreport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// TestSpoolSectionReclaimsVerifiedDeadOwner pins the spool lock's D37
// contract (review-gate addition): the spool section lock is owner-labelled
// and its contention path performs the SAME owner-verified stale-break as
// the queue lock. Before the fix the lock was a bare exclusive create with
// no owner record and no reclaim: a process dying while it held the
// section left an unlabelled/owned lock forever, every later capture write
// failed its claim budget, and purge did not remove the artifact — the
// spool went silently dead for the life of the boot.
func TestSpoolSectionReclaimsVerifiedDeadOwner(t *testing.T) {
	path := seedConsent(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}

	// A lock labelled by a verified-dead owner: a previous boot's identity,
	// self-consistent otherwise. The next writer must reclaim it.
	identity := atomicfile.BootIDIdentity()
	if identity == "" {
		t.Skip("no boot identity on this platform; the boot-comparison path is not exercised")
	}
	owner := atomicfile.LockOwner{PID: os.Getpid(), BootID: "previous-boot-" + identity}
	raw, err := json.Marshal(owner)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	lockPath := filepath.Join(filepath.Dir(path), spoolSectionClaim)
	if err := os.WriteFile(lockPath, raw, 0o600); err != nil {
		t.Fatalf("seed stale lock: %v", err)
	}

	if err := appendSpoolLine(SpoolEntry{
		Kind:    KindHookTimeout,
		Verdict: VerdictAmbiguous,
		Frames:  []string{"internal/cli.Execute"},
	}); err != nil {
		t.Fatalf("the next writer could not reclaim the verified-dead spool lock: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spool: %v", err)
	}
	if !strings.Contains(string(written), "hook_timeout") {
		t.Fatalf("entry not written after the reclaim: %s", written)
	}
	// The reclaimed lock must be gone once the writer's section closed.
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("the section lock survived its release: %v", err)
	}
}
