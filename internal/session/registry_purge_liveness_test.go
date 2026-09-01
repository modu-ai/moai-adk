package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// livenessProbeFake is the per-PID probe outcome table the liveness-aware
// purge tests install. A pid absent from the table probes as (dead, true).
type livenessProbeFake struct {
	alive       map[int]bool
	undetermined map[int]bool
}

// install replaces the probeLiveness seam for the duration of the test.
func (f livenessProbeFake) install(t *testing.T) {
	t.Helper()
	orig := probeLiveness
	t.Cleanup(func() { probeLiveness = orig })
	probeLiveness = func(pid int) (bool, bool) {
		if f.undetermined[pid] {
			return false, false
		}
		return f.alive[pid], true
	}
}

// newLivenessTestRegistry builds a registry bound to a fresh t.TempDir() with
// a FakeClock at a fixed reference time, and returns the registry file path
// so tests can write entries with explicit PIDs directly.
func newLivenessTestRegistry(t *testing.T) (*Registry, *FakeClock, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "active-sessions.json")
	clock := &FakeClock{Current: time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)}
	return NewRegistry(path, clock), clock, path
}

// writeEntries replaces the registry file content with the given entries.
func writeEntries(t *testing.T, path string, entries []Entry) {
	t.Helper()
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

// TestPurgeKeepsLiveSessionPastHeartbeatFloor is the regression test for the
// live-session omission defect (card t404): a session whose heartbeat froze
// at registration (no component sends heartbeats automatically) and that is
// still running must survive the next session's SessionStart purge.
//
// Before the repair, Purge removed every entry past DefaultStaleMinutes by
// heartbeat age alone, so any session alive longer than 30 minutes vanished
// from the registry — and from `moai session list --json` — the moment
// another session started, exactly when the orchestrator's pre-spawn sync
// check most needs the entry.
func TestPurgeKeepsLiveSessionPastHeartbeatFloor(t *testing.T) {
	r, clock, path := newLivenessTestRegistry(t)
	livenessProbeFake{alive: map[int]bool{65001: true}}.install(t)

	stale := clock.Current.Add(-2 * time.Hour)
	writeEntries(t, path, []Entry{{
		SessionID:     "uuid-live-stale",
		SpecID:        "SPEC-A",
		Phase:         "run",
		StartedAt:     stale,
		LastHeartbeat: stale,
		PID:           65001,
		Host:          "test-host",
	}})

	purged, err := r.Purge(30)
	if err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if purged != 0 {
		t.Fatalf("purged count: got %d, want 0 (live session must survive the heartbeat floor)", purged)
	}
	entries := mustRead(t, r)
	if len(entries) != 1 || entries[0].SessionID != "uuid-live-stale" {
		t.Fatalf("post-purge entries: got %+v, want the live-stale entry kept", entries)
	}
}

// TestPurgeRemovesDeadStaleSession verifies the zombie-cleanup contract the
// purge exists for: a heartbeat-stale entry whose session process is
// positively gone (ESRCH) is removed.
func TestPurgeRemovesDeadStaleSession(t *testing.T) {
	r, clock, path := newLivenessTestRegistry(t)
	livenessProbeFake{}.install(t) // every probe: (dead, determined)

	stale := clock.Current.Add(-2 * time.Hour)
	writeEntries(t, path, []Entry{{
		SessionID:     "uuid-dead-stale",
		SpecID:        "SPEC-B",
		Phase:         "run",
		StartedAt:     stale,
		LastHeartbeat: stale,
		PID:           65002,
		Host:          "test-host",
	}})

	purged, err := r.Purge(30)
	if err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if purged != 1 {
		t.Fatalf("purged count: got %d, want 1 (dead stale entry is purge food)", purged)
	}
	if entries := mustRead(t, r); len(entries) != 0 {
		t.Fatalf("post-purge entries: got %+v, want none", entries)
	}
}

// TestPurgeFreshDeadEntryKept verifies the heartbeat floor still gates: an
// entry inside the threshold is never probed or removed, dead PID or not.
func TestPurgeFreshDeadEntryKept(t *testing.T) {
	r, clock, path := newLivenessTestRegistry(t)
	livenessProbeFake{}.install(t)

	fresh := clock.Current.Add(-1 * time.Minute)
	writeEntries(t, path, []Entry{{
		SessionID:     "uuid-fresh-dead",
		SpecID:        "SPEC-C",
		Phase:         "plan",
		StartedAt:     fresh,
		LastHeartbeat: fresh,
		PID:           65003,
		Host:          "test-host",
	}})

	purged, err := r.Purge(30)
	if err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if purged != 0 {
		t.Fatalf("purged count: got %d, want 0 (fresh entry is below the floor)", purged)
	}
}

// TestPurgeUndeterminedProbeFallsBackToHeartbeat documents the fallback:
// when the platform cannot determine liveness (the Windows probe reports
// nothing), the pre-repair heartbeat verdict stands so the existing
// stale-entry hygiene is preserved there.
func TestPurgeUndeterminedProbeFallsBackToHeartbeat(t *testing.T) {
	r, clock, path := newLivenessTestRegistry(t)
	livenessProbeFake{undetermined: map[int]bool{65004: true}}.install(t)

	stale := clock.Current.Add(-2 * time.Hour)
	writeEntries(t, path, []Entry{{
		SessionID:     "uuid-undetermined",
		SpecID:        "SPEC-D",
		Phase:         "run",
		StartedAt:     stale,
		LastHeartbeat: stale,
		PID:           65004,
		Host:          "test-host",
	}})

	purged, err := r.Purge(30)
	if err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if purged != 1 {
		t.Fatalf("purged count: got %d, want 1 (undetermined probe defers to the heartbeat verdict)", purged)
	}
}

// TestPurgeCorruptPIDUsesHeartbeatVerdict covers the corrupt-entry path: a
// stale entry with no usable PID carries no liveness signal, so the
// heartbeat verdict stands.
func TestPurgeCorruptPIDUsesHeartbeatVerdict(t *testing.T) {
	r, clock, path := newLivenessTestRegistry(t)
	livenessProbeFake{}.install(t)

	stale := clock.Current.Add(-2 * time.Hour)
	writeEntries(t, path, []Entry{{
		SessionID:     "uuid-no-pid",
		SpecID:        "SPEC-E",
		Phase:         "run",
		StartedAt:     stale,
		LastHeartbeat: stale,
		PID:           0,
		Host:          "test-host",
	}})

	purged, err := r.Purge(30)
	if err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if purged != 1 {
		t.Fatalf("purged count: got %d, want 1 (corrupt PID defers to the heartbeat verdict)", purged)
	}
}
