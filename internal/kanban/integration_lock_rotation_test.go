package kanban

// integration_lock_rotation_test.go — card t951: session rotation vs. the
// owning process.
//
// A lane that runs /clear keeps its owning PROCESS and gets a NEW session id.
// The record still names that process (PID + PIDSourceSessionOwner, written by
// the acquire verb from session.ResolveOwnerPID), so the evidence that the
// caller IS the holder is already on disk — the release decision simply never
// read it. Observed in production on card t791: release refused with
// "held by a different session: agent-8 (pid 48258)" while 48258 was the
// refused process itself.

import (
	"os"
	"testing"
)

// mustAcquireOwned records a window whose pid names the owning session — the
// record shape the acquire verb writes.
func mustAcquireOwned(t *testing.T, root, session string, pid int) {
	t.Helper()
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: session,
		PID:       pid,
		PIDSource: PIDSourceSessionOwner,
		Branch:    "release/v9.9.9",
	}, false); err != nil {
		t.Fatalf("acquire(%s): %v", session, err)
	}
}

// THE property this card exists for: the same owning process releases its own
// window after its session id rotated.
//
// Deliberately NOT routed through --force: that flag means "take a window from
// a DIFFERENT live holder", and it is recorded. Making a lane release its own
// window with it would stamp the ledger with a seizure that did not happen.
func TestReleaseIntegrationLock_SameOwnerProcessSurvivesSessionRotation(t *testing.T) {
	root := t.TempDir()
	owner := os.Getpid()
	mustAcquireOwned(t, root, "session-before-clear", owner)

	released, err := ReleaseIntegrationLock(root, "session-after-clear", owner, false)
	if err != nil {
		t.Fatalf("the owning process could not release its own window after /clear: %v", err)
	}
	if released == nil || released.SessionID != "session-before-clear" {
		t.Errorf("release did not report the freed holder: %+v", released)
	}
	if _, statErr := os.Stat(lockPathFor(t, root)); !os.IsNotExist(statErr) {
		t.Error("record survived its release")
	}
}

// The negative that makes the positive worth anything: a genuinely different
// live session — a different owning process — is still refused. A change that
// lets every release succeed is worse than the defect it replaces.
func TestReleaseIntegrationLock_DifferentLiveOwnerStillRefused(t *testing.T) {
	root := t.TempDir()
	mustAcquireOwned(t, root, "lane-8", os.Getpid())

	// A live pid that is not this process: this process's own parent.
	foreign := os.Getppid()
	if foreign == os.Getpid() || foreign <= 1 || !FactoryProcessAlive(foreign) {
		t.Skipf("no usable distinct live pid (ppid=%d)", foreign)
	}
	if _, err := ReleaseIntegrationLock(root, "lane-5", foreign, false); !IsIntegrationLockForeign(err) {
		t.Fatalf("a different owning process released another lane's window: err = %v", err)
	}
}

// A legacy record — a pid with no PIDSource marker — is NOT re-interpreted.
// Its pid does not mean "the owning session" (before the marker existed the
// field carried whatever the caller supplied), so it must not become a second
// ownership key retroactively.
func TestReleaseIntegrationLock_LegacyRecordPIDIsNotAnOwnershipKey(t *testing.T) {
	root := t.TempDir()
	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "lane-8",
		PID:       os.Getpid(), // no PIDSource: the pre-anchor record shape
		Branch:    "release/v9.9.9",
	}, false); err != nil {
		t.Fatalf("acquire: %v", err)
	}

	if _, err := ReleaseIntegrationLock(root, "lane-5", os.Getpid(), false); !IsIntegrationLockForeign(err) {
		t.Fatalf("a legacy record's pid was read as an ownership key: err = %v", err)
	}
}

// Unresolvable on either side never matches. pid 0 means "the owner could not
// be resolved", and two unknowns are not the same owner — reading them as
// equal would let ANY session release ANY unresolvable-owner window, which is
// a wider hole than the defect being fixed.
func TestReleaseIntegrationLock_UnresolvedPIDsDoNotMatch(t *testing.T) {
	root := t.TempDir()
	mustAcquireOwned(t, root, "lane-8", 0) // acquirer could not resolve its owner

	if _, err := ReleaseIntegrationLock(root, "lane-5", 0, false); !IsIntegrationLockForeign(err) {
		t.Errorf("two unresolved owners read as the same owner: err = %v", err)
	}
	if _, err := ReleaseIntegrationLock(root, "lane-5", os.Getpid(), false); !IsIntegrationLockForeign(err) {
		t.Errorf("a resolved caller matched an unresolvable record: err = %v", err)
	}
}

// A STALE record (the recorded owner's process is gone) is a different axis and
// is deliberately untouched: release still refuses a foreign stale window
// exactly as it did, because reclaiming a stale window is acquire's job.
func TestReleaseIntegrationLock_StaleForeignRecordUnchanged(t *testing.T) {
	root := t.TempDir()
	dead := deadPID(t)
	mustAcquireOwned(t, root, "lane-8", dead)

	if _, err := ReleaseIntegrationLock(root, "lane-5", os.Getpid(), false); !IsIntegrationLockForeign(err) {
		t.Errorf("stale-record release behavior changed: err = %v", err)
	}
}
