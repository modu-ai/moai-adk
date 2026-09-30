package factorylane

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/sessionmsg"
)

// A nil clock selects the real UTC clock — the production path.
func TestNewStoreNilClockSelectsRealClock(t *testing.T) {
	store := NewStore(t.TempDir(), nil)
	if _, err := store.RecordRequest("lane-9"); err != nil {
		t.Fatalf("RecordRequest with nil clock: %v", err)
	}
	obs, err := store.Observations("lane-9")
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("observation count = %d, want 1", len(obs))
	}
	if obs[0].RequestedAt.IsZero() {
		t.Fatal("RequestedAt is zero — the real clock did not fire")
	}
}

// A corrupt observation record is a read error, never a silent row drop.
func TestObservationsReportCorruptRecord(t *testing.T) {
	store, _ := newTestStore(t)
	dir := store.obsDir("lane-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "obs-0000000000000000001.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Observations("lane-1"); err == nil || !strings.Contains(err.Error(), "parse observation") {
		t.Fatalf("error = %v, want the parse-observation classification", err)
	}
}

// Non-record files in the lane directory are ignored, not fatal.
func TestObservationsIgnoreForeignFiles(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(store.obsDir("lane-1"), "readme.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	obs, err := store.Observations("lane-1")
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("observation count = %d, want 1 (foreign file ignored)", len(obs))
	}
}

// A store root where observations/ cannot be a directory surfaces as an
// error, never as an empty set.
func TestObservationsErrorWhenObservationsPathIsFile(t *testing.T) {
	root := t.TempDir()
	fbDir := filepath.Join(root, DefaultStateRoot)
	if err := os.MkdirAll(fbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fbDir, "observations"), []byte("a file, not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(root, nil)
	_, err := store.Observations("lane-1")
	if err == nil {
		t.Fatal("error = nil, want the read failure (fail loud, not empty)")
	}
}

// Two events minted inside one clock tick get distinct record files — the
// collision bump must not lose either event.
func TestDeclareFallbackBumpsFilenameOnCollision(t *testing.T) {
	store, clock := newTestStore(t)
	if _, err := store.Restore("lane-1"); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	// Pin the clock back so the next event stamps onto the same filename.
	clock.Current = base
	events, err := store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("Transitions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1 before the collision", len(events))
	}
	if _, err := store.DeclareFallback("lane-1", TriggerNoResponse, "t1"); err != nil {
		t.Fatalf("DeclareFallback on a taken filename: %v", err)
	}
	events, err = store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("Transitions after collision: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("event count = %d, want 2 — the collision must not lose either event", len(events))
	}
}

// A corrupt transition record is a read error, never a dropped event.
func TestTransitionsReportCorruptRecord(t *testing.T) {
	store, _ := newTestStore(t)
	dir := store.transitionDir("lane-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "evt-0000000000000000001.json"), []byte("]bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Transitions("lane-1"); err == nil || !strings.Contains(err.Error(), "parse transition") {
		t.Fatalf("error = %v, want the parse-transition classification", err)
	}
}

// An observation whose bound snapshot is malformed is skipped by the probe
// evaluation rather than fabricating an unavailability window.
func TestEvaluateSkipsMalformedBound(t *testing.T) {
	obs := noResponseObs()
	obs.Bound = "not-a-duration"
	got := EvaluateAvailability(ProbeInput{
		Now:          probeNow,
		LeadPeer:     "team-lead",
		OfflineBound: 30 * time.Minute,
		Registry:     []sessionmsg.AgentInfo{probePeer("team-lead", time.Minute)},
		Observations: []Observation{obs},
	})
	if got.Verdict != VerdictAvailable {
		t.Fatalf("verdict = %q, want %q (malformed bound must not fabricate unavailability): %s", got.Verdict, VerdictAvailable, got.Reason)
	}
}

// When the store root cannot be a directory (a file squats on it), every
// write fails loudly — RecordRequest, DeclareFallback, and Restore each
// surface the error instead of reporting success.
func TestWritesFailLoudWhenStoreRootIsFile(t *testing.T) {
	root := t.TempDir()
	fb := filepath.Join(root, DefaultStateRoot)
	if err := os.MkdirAll(filepath.Dir(fb), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fb, []byte("a file, not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewStore(root, nil)
	if _, err := store.RecordRequest("lane-1"); err == nil {
		t.Fatal("RecordRequest error = nil, want the mkdir failure")
	}
	if _, err := store.DeclareFallback("lane-1", TriggerNoResponse, "t1"); err == nil {
		t.Fatal("DeclareFallback error = nil, want the mkdir failure")
	}
	if _, err := store.Restore("lane-1"); err == nil {
		t.Fatal("Restore error = nil, want the mkdir failure")
	}
}

// A malformed timer snapshot is a sweep error, never a fabricated
// no-response record.
func TestSweepReportsMalformedTimerSnapshot(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest: %v", err)
	}
	obs, err := store.Observations("lane-1")
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	obs[0].Timer = "not-a-duration"
	data, err := marshalRecord(obs[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(store.obsPath("lane-1", obs[0].RequestedAt), data); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SweepNoResponse("lane-1"); err == nil || !strings.Contains(err.Error(), "parse timer") {
		t.Fatalf("error = %v, want the parse-timer classification", err)
	}
}

// A corrupt observation record also fails the ack path loudly.
func TestAckPendingReportsCorruptRecord(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.RecordRequest("lane-1"); err != nil {
		t.Fatalf("RecordRequest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(store.obsDir("lane-1"), "obs-0000000000000000009.json"), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AckPending("lane-1"); err == nil || !strings.Contains(err.Error(), "parse observation") {
		t.Fatalf("error = %v, want the parse-observation classification", err)
	}
}
