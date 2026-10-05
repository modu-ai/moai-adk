package factory

// integration_window_queue_test.go — M1 data-model tests for the FIFO window
// queue (card t1479, SPEC-MERGE-WINDOW-QUEUE-001): legacy-record compatibility
// and ticket round-trip (AC-MWQ-001), lease stamping and disable (AC-MWQ-008's
// record layer), and the policy sibling record (REQ-MWQ-012's record layer).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fixedClock(t *testing.T) func() time.Time {
	t.Helper()
	at := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	return func() time.Time { return at }
}

func TestLegacyRecordReadsAsEmptyQueue(t *testing.T) {
	// AC-MWQ-001: a record written before the queue existed reads as a record
	// with an empty queue and unchanged holder fields, and no queue key is
	// written unless a ticket exists.
	root := t.TempDir()
	legacy := []byte(`{
  "session_id": "sess-a",
  "session_name": "lane-1",
  "pid": 4242,
  "pid_source": "session-owner",
  "branch": "develop",
  "worktree": "/repo/.claude/worktrees/develop",
  "acquired_at": "2026-10-01T10:00:00Z",
  "card": "t0002"
}`)
	path := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, IntegrationLockFileName), legacy, 0o644); err != nil {
		t.Fatal(err)
	}
	lock, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if !lock.Held() || lock.SessionID != "sess-a" || lock.PID != 4242 {
		t.Fatalf("holder fields changed by the queue introduction: %+v", lock)
	}
	if len(lock.Queue) != 0 {
		t.Fatalf("legacy record should read as an empty queue, got %d tickets", len(lock.Queue))
	}
	if lock.LeaseExpiresAt != "" {
		t.Fatalf("legacy record should carry no lease stamp, got %q", lock.LeaseExpiresAt)
	}
	// Round-trip through the new writer keeps the holder and stays key-absent
	// for the queue.
	if err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error { return nil }); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(path, IntegrationLockFileName))
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, has := raw["queue"]; has {
		t.Fatalf("no ticket exists, the record must not carry a queue key: %s", data)
	}
	if raw["session_id"] != "sess-a" {
		t.Fatalf("holder lost across the rewrite: %s", data)
	}
}

func TestIntegrationTicketFieldsRoundTrip(t *testing.T) {
	// AC-MWQ-001 (second half): a ticket written by the new code carries
	// session, name, card, enqueue instant, owning-session pid with
	// pid_source: session-owner, branch/branch_source/worktree, waiter pid,
	// waiter start time, and heartbeat instant.
	root := t.TempDir()
	ticket := IntegrationTicket{
		SessionID:    "sess-b",
		SessionName:  "lane-2",
		Card:         "t0003",
		OwnerPID:     4243,
		PIDSource:    PIDSourceSessionOwner,
		Branch:       "develop",
		BranchSource: BranchSourceConfig,
		Worktree:     "/repo/.claude/worktrees/develop",
		WaiterPID:    4244,
		WaiterStart:  "2026-10-05T08:59:30Z",
		Heartbeat:    "2026-10-05T09:00:00Z",
		EnqueuedAt:   "2026-10-05T09:00:00Z",
	}
	err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
		w.Queue = append(w.Queue, ticket)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	lock, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Queue) != 1 {
		t.Fatalf("expected exactly one ticket, got %d", len(lock.Queue))
	}
	got := lock.Queue[0]
	if got != ticket {
		t.Fatalf("ticket round-trip mismatch:\n got %+v\nwant %+v", got, ticket)
	}
	data, _ := os.ReadFile(filepath.Join(root, ".moai", "state", IntegrationLockFileName))
	var raw struct {
		Queue []struct {
			OwnerPID    int    `json:"owner_pid"`
			PIDSource   string `json:"pid_source"`
			WaiterPID   int    `json:"waiter_pid"`
			WaiterStart string `json:"waiter_start"`
			Heartbeat   string `json:"heartbeat"`
		} `json:"queue"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.Queue) != 1 || raw.Queue[0].OwnerPID != 4243 || raw.Queue[0].PIDSource != "session-owner" ||
		raw.Queue[0].WaiterPID != 4244 || raw.Queue[0].WaiterStart == "" || raw.Queue[0].Heartbeat == "" {
		t.Fatalf("on-disk ticket fields incomplete: %s", data)
	}
}

func TestLeaseStampAndExpiry(t *testing.T) {
	// AC-MWQ-008 (record layer): acquire/promotion stamp a lease expiry; a
	// configured duration of zero disables the lease entirely.
	clock := fixedClock(t)
	got := StampLease(&IntegrationLock{}, clock(), IntegrationLeaseDefault)
	want := clock().Add(IntegrationLeaseDefault).Format(time.RFC3339)
	if got.LeaseExpiresAt != want {
		t.Fatalf("lease stamp = %q, want %q", got.LeaseExpiresAt, want)
	}
	// Disabled lease: no stamp at all.
	got = StampLease(&IntegrationLock{LeaseExpiresAt: "2026-10-05T10:00:00Z"}, clock(), 0)
	if got.LeaseExpiresAt != "" {
		t.Fatalf("duration zero must clear the lease stamp, got %q", got.LeaseExpiresAt)
	}
	// A stamped lease that is still future is not expired.
	if (&IntegrationLock{LeaseExpiresAt: want}).LeaseExpired(clock()) {
		t.Fatalf("future lease must not read expired")
	}
	// A lapsed lease reads expired.
	if !(&IntegrationLock{LeaseExpiresAt: clock().Add(-time.Minute).Format(time.RFC3339)}).LeaseExpired(clock()) {
		t.Fatalf("lapsed lease must read expired")
	}
	// No stamp at all (legacy record / disabled lease) never reads expired:
	// validity is decided by owning-session liveness alone, as before.
	if (&IntegrationLock{SessionID: "sess"}).LeaseExpired(clock()) {
		t.Fatalf("absent lease must not read expired")
	}
}

func TestWindowPolicyRecord(t *testing.T) {
	// REQ-MWQ-012's record layer: an absent policy record reads as open; the
	// written policy round-trips beside the window record.
	root := t.TempDir()
	policy, err := ReadIntegrationWindowPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Policy != "open" {
		t.Fatalf("absent policy must read open, got %q", policy.Policy)
	}
	hold := IntegrationWindowPolicy{Policy: "hold", Reason: "release-cut", SetBy: "lead", SetAt: "2026-10-05T09:00:00Z"}
	if err := WriteIntegrationWindowPolicy(root, hold); err != nil {
		t.Fatal(err)
	}
	policy, err = ReadIntegrationWindowPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Policy != "hold" || policy.Reason != "release-cut" {
		t.Fatalf("policy round-trip mismatch: %+v", policy)
	}
	// The policy record sits beside the window record, under the same state
	// directory, and does not collide with the record file name.
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "integration-window-policy.json")); err != nil {
		t.Fatalf("policy record not beside the window record: %v", err)
	}
}
