package factory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestone M1, AC-FLR-008: the queue claim
// lease ignores the card lease duration key (REQ-FLR-008). Given
// workflow.factory_lease.duration = 20m in the project's configuration, a queue claim
// writes an expiry 15 minutes after the claim, the compiled queue lease.
func TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration(t *testing.T) {
	t.Parallel()
	store := claimFixture(t)

	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"),
		[]byte("workflow:\n  factory_lease:\n    duration: 20m\n"), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	seedQueuedCards(t, store, "queued card")
	before := time.Now()
	result, err := store.Claim("lane-1")
	after := time.Now()
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if result.Item.LeaseExpiresAt == nil {
		t.Fatal("claimed lease_expires_at = nil, want the queue lease expiry")
	}
	expiresAt, perr := time.Parse(time.RFC3339, *result.Item.LeaseExpiresAt)
	if perr != nil {
		t.Fatalf("claimed expiry %q is not RFC 3339: %v", *result.Item.LeaseExpiresAt, perr)
	}
	// The expiry is written at second precision, so it lies between the claim
	// window's start and end, each shifted by the 15-minute queue lease.
	lo := before.Add(15 * time.Minute).Truncate(time.Second)
	hi := after.Add(15 * time.Minute)
	if expiresAt.Before(lo) || expiresAt.After(hi) {
		t.Fatalf("claim expiry = %s, want within [%s, %s] (15 minutes after the claim, not the card key's 20 minutes)",
			expiresAt.Format(time.RFC3339), lo.Format(time.RFC3339), hi.Format(time.RFC3339))
	}
}
