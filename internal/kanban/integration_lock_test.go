package kanban

// integration_lock_test.go — behaviour of the release-integration holder lock
// (card t194). The properties under test are the ones the doctrine leans on:
// a second lane is refused, a lane may re-enter its own window, a dead
// holder's window is reclaimable, and — the property this whole card exists
// for — an unreadable record never reads as a free window.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lockPathFor(t *testing.T, root string) string {
	t.Helper()
	return filepath.Join(root, ".moai", "state", IntegrationLockFileName)
}

func mustAcquire(t *testing.T, root, session string) {
	t.Helper()
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: session,
		Branch:    "release/v9.9.9",
		Worktree:  filepath.Join(root, "wt"),
	}, false); err != nil {
		t.Fatalf("acquire(%s): %v", session, err)
	}
}

// A free window is acquirable, and the record lands where every linked
// worktree can see it.
func TestAcquireIntegrationLock_RecordsHolder(t *testing.T) {
	root := t.TempDir()
	mustAcquire(t, root, "lane-8")

	data, err := os.ReadFile(lockPathFor(t, root))
	if err != nil {
		t.Fatalf("record not written: %v", err)
	}
	var got IntegrationLock
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("record is not valid JSON: %v", err)
	}
	if got.SessionID != "lane-8" {
		t.Errorf("session_id = %q, want lane-8", got.SessionID)
	}
	// The caller supplied no pid, and none is invented here. Filling it with
	// os.Getpid() is the defect card t298 removed: this process outlives
	// nothing, so its pid would be dead before any reader probes it. An unset
	// pid stays 0, which Stale() reads as live.
	if got.PID != 0 {
		t.Errorf("pid = %d, want 0; acquire must not substitute a pid the caller did not supply", got.PID)
	}
	if got.Stale() {
		t.Error("a record with no resolvable owner reads reclaimable; the conservative direction is live")
	}
	if got.AcquiredAt == "" {
		t.Error("acquired_at is empty; staleness triage has nothing to read")
	}
}

// The owner pid the caller resolved is what lands on disk, verbatim, alongside
// the marker naming whose pid it is.
func TestAcquireIntegrationLock_RecordsTheCallersOwnerPID(t *testing.T) {
	root := t.TempDir()
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "lane-8",
		PID:       os.Getpid(),
		PIDSource: PIDSourceSessionOwner,
	}, false); err != nil {
		t.Fatalf("acquire: %v", err)
	}

	got, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.PID != os.Getpid() {
		t.Errorf("pid = %d, want the supplied owner pid %d", got.PID, os.Getpid())
	}
	if got.PIDSource != PIDSourceSessionOwner {
		t.Errorf("pid_source = %q, want %q", got.PIDSource, PIDSourceSessionOwner)
	}
	if got.Stale() {
		t.Error("a live owner reads reclaimable")
	}
}

// AC-INL-004: an anchored record whose owner could not be resolved (pid 0)
// reads LIVE, and stays that way until an explicit release or a recorded
// --force. This is the mutant guard against "fill pid 0 so the field is never
// empty" reintroducing the defect under another name.
func TestAcquireIntegrationLock_AnchoredPIDZeroIsLiveNotStale(t *testing.T) {
	root := t.TempDir()
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "unresolvable-owner",
		PID:       0,
		PIDSource: PIDSourceSessionOwner,
	}, false); err != nil {
		t.Fatalf("acquire: %v", err)
	}

	got, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.PID != 0 || got.PIDSource != PIDSourceSessionOwner {
		t.Fatalf("record is not the anchored pid-0 shape: %+v", got)
	}
	if got.Stale() {
		t.Error("an anchored pid-0 record reads reclaimable; an unresolvable owner must read live")
	}
	// A bare acquire by another lane must therefore be refused.
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "lane-5",
		PID:       0,
		PIDSource: PIDSourceSessionOwner,
	}, false); !IsIntegrationLockHeld(err) {
		t.Errorf("bare acquire over an anchored pid-0 holder: err = %v, want the contention sentinel", err)
	}
}

// AC-INL-007: a record written before the anchor existed carries no
// pid_source, and is read exactly as it always was — it parses, a dead
// recorded pid still reads reclaimable, and a foreign acquire takes it over.
// The schema change is additive; there is no upgrade wedge.
func TestReadIntegrationLock_LegacyRecordWithoutPIDSource(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(lockPathFor(t, root)), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Hand-written in the pre-anchor shape: no pid_source key at all.
	legacy := `{
  "session_id": "legacy-lane",
  "session_name": "lane-legacy",
  "pid": 2147483632,
  "branch": "release/v9.9.9",
  "worktree": "/tmp/wt",
  "acquired_at": "2026-08-27T00:00:00Z"
}
`
	if err := os.WriteFile(lockPathFor(t, root), []byte(legacy), 0o644); err != nil {
		t.Fatalf("seed legacy record: %v", err)
	}

	got, err := ReadIntegrationLock(root)
	if err != nil {
		t.Fatalf("legacy record did not parse: %v", err)
	}
	if got.SessionID != "legacy-lane" || got.PID != 0x7FFFFFF0 {
		t.Errorf("legacy record parsed wrong: %+v", got)
	}
	if got.PIDSource != "" {
		t.Errorf("pid_source = %q on a legacy record, want empty", got.PIDSource)
	}
	if !got.Stale() {
		t.Skip("the seeded legacy pid is live on this machine; the reclaimable path is not exercisable here")
	}

	replaced, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "lane-5",
		PID:       os.Getpid(),
		PIDSource: PIDSourceSessionOwner,
	}, false)
	if err != nil {
		t.Fatalf("legacy stale window was not reclaimable: %v", err)
	}
	if replaced == nil || replaced.SessionID != "legacy-lane" {
		t.Errorf("takeover did not report the displaced legacy holder: %+v", replaced)
	}

	// Re-acquiring a window the caller already holds still refreshes it.
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "lane-5",
		PID:       os.Getpid(),
		PIDSource: PIDSourceSessionOwner,
	}, false); err != nil {
		t.Errorf("holder could not re-acquire its own window: %v", err)
	}
}

// The property the card is about: a second live session is refused, and the
// refusal names who holds it so the operator can address them.
func TestAcquireIntegrationLock_RefusesASecondLiveSession(t *testing.T) {
	root := t.TempDir()
	mustAcquire(t, root, "lane-8")

	_, err := AcquireIntegrationLock(root, IntegrationLock{SessionID: "lane-5"}, false)
	if err == nil {
		t.Fatal("second session acquired a held window")
	}
	if !IsIntegrationLockHeld(err) {
		t.Errorf("error is not the contention sentinel: %v", err)
	}
	if !strings.Contains(err.Error(), "lane-8") {
		t.Errorf("refusal does not name the holder: %v", err)
	}
}

// A lane that re-enters after a /clear is not locked out of its own window.
func TestAcquireIntegrationLock_SameSessionReacquires(t *testing.T) {
	root := t.TempDir()
	mustAcquire(t, root, "lane-8")
	if _, err := AcquireIntegrationLock(root, IntegrationLock{SessionID: "lane-8"}, false); err != nil {
		t.Fatalf("holder could not re-acquire its own window: %v", err)
	}
}

// A dead holder's window is reclaimable, and the takeover is reported rather
// than silent — the next lane must be able to say what it cleared.
func TestAcquireIntegrationLock_TakesOverAStaleHolder(t *testing.T) {
	root := t.TempDir()
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "dead-lane",
		PID:       -1, // never live; the probe rejects pid <= 0 up front
	}, false); err != nil {
		t.Fatalf("seed acquire: %v", err)
	}
	// A pid of -1 is not "stale" (Stale() ignores pid <= 0 rather than
	// guessing), so seed a plausible-but-dead pid through the record instead.
	seedDeadHolder(t, root, "dead-lane")

	replaced, err := AcquireIntegrationLock(root, IntegrationLock{SessionID: "lane-5"}, false)
	if err != nil {
		t.Fatalf("stale window was not reclaimable: %v", err)
	}
	if replaced == nil || replaced.SessionID != "dead-lane" {
		t.Errorf("takeover did not report what it replaced: %+v", replaced)
	}
}

// seedDeadHolder rewrites the record with a pid that is almost certainly not
// running. It is a seam for the staleness path, which cannot be exercised with
// a live pid by construction.
func seedDeadHolder(t *testing.T, root, session string) {
	t.Helper()
	lock := IntegrationLock{SessionID: session, PID: 0x7FFFFFF0, Branch: "release/v9.9.9"}
	data, err := json.MarshalIndent(&lock, "", "  ")
	if err != nil {
		t.Fatalf("seed marshal: %v", err)
	}
	if err := os.WriteFile(lockPathFor(t, root), data, 0o644); err != nil {
		t.Fatalf("seed write: %v", err)
	}
	if !lock.Stale() {
		t.Skip("seeded pid is live on this machine; staleness path not exercisable here")
	}
}

// force takes over a LIVE holder, because a wedged session must not block the
// batch forever.
func TestAcquireIntegrationLock_ForceTakesOverALiveHolder(t *testing.T) {
	root := t.TempDir()
	mustAcquire(t, root, "lane-8")

	replaced, err := AcquireIntegrationLock(root, IntegrationLock{SessionID: "lane-5"}, true)
	if err != nil {
		t.Fatalf("force acquire failed: %v", err)
	}
	if replaced == nil || replaced.SessionID != "lane-8" {
		t.Errorf("force did not report the displaced holder: %+v", replaced)
	}
}

// THE property this card exists for: an unreadable record must never read as a
// free window. Absence of a parseable holder is not evidence that nobody holds
// it — the same substitution t181 found in the MERGE_HEAD probe.
func TestReadIntegrationLock_CorruptRecordIsNotAFreeWindow(t *testing.T) {
	root := t.TempDir()
	mustAcquire(t, root, "lane-8")
	if err := os.WriteFile(lockPathFor(t, root), []byte("{ this is not json"), 0o644); err != nil {
		t.Fatalf("corrupt: %v", err)
	}

	if _, err := ReadIntegrationLock(root); err == nil {
		t.Error("a corrupt record read as a valid (empty) lock")
	}
	if _, err := AcquireIntegrationLock(root, IntegrationLock{SessionID: "lane-5"}, false); err == nil {
		t.Error("a second lane acquired the window over a corrupt record")
	}
}

// Releasing is the holder's act. A foreign release is refused for the same
// reason a foreign entry is.
func TestReleaseIntegrationLock_HolderAndForeign(t *testing.T) {
	root := t.TempDir()
	mustAcquire(t, root, "lane-8")

	if _, err := ReleaseIntegrationLock(root, "lane-5", false); err == nil {
		t.Error("a different session released the window")
	} else if !IsIntegrationLockForeign(err) {
		t.Errorf("error is not the foreign sentinel: %v", err)
	}

	released, err := ReleaseIntegrationLock(root, "lane-8", false)
	if err != nil {
		t.Fatalf("holder could not release its own window: %v", err)
	}
	if released == nil || released.SessionID != "lane-8" {
		t.Errorf("release did not report the freed holder: %+v", released)
	}
	if _, err := os.Stat(lockPathFor(t, root)); !os.IsNotExist(err) {
		t.Error("record survived its release")
	}
}

// Releasing a window nobody holds is reported, not silently accepted: a lane
// with a broken model of the board must not have it confirmed.
func TestReleaseIntegrationLock_EmptyIsReported(t *testing.T) {
	root := t.TempDir()
	if _, err := ReleaseIntegrationLock(root, "lane-8", false); !IsIntegrationLockNotHeld(err) {
		t.Errorf("releasing an unheld window: err = %v, want the not-held sentinel", err)
	}
}

// A free window reads as not-held rather than as an error.
func TestReadIntegrationLock_AbsentRecord(t *testing.T) {
	lock, err := ReadIntegrationLock(t.TempDir())
	if err != nil {
		t.Fatalf("absent record errored: %v", err)
	}
	if lock.Held() {
		t.Error("absent record reported a holder")
	}
}
