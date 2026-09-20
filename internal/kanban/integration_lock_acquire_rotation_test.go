package kanban

// integration_lock_acquire_rotation_test.go — card t959: the acquire-side
// counterpart of the t951 rotation measurement.
//
// AcquireIntegrationLock's doc comment declares "a lane that re-enters after a
// `/clear` is not locked out of its own window". These tests measure whether
// that declaration holds. They assert the OBSERVED behaviour, not the declared
// one, so the file records what the code does today and turns red the moment
// that changes — which is the point: the card's question is whether code and
// comment agree, and a test written against the comment would answer a
// different question.
//
// The record shape is the one the acquire verb writes (PID + PIDSourceSessionOwner
// from session.ResolveOwnerPID), identical to integration_lock_rotation_test.go.
//
// THIS FILE GOING RED IS A NORMAL OUTCOME, NOT A REGRESSION. It pins today's
// behaviour so that a future change to acquire has to turn it over on purpose.
// When that happens, read TestIntegrationLockAcquire_SerializedAcrossProcesses
// (integration_lock_cross_test.go) in the same pass before updating anything
// here: t959 measured that the obvious repair — giving acquire the release
// side's releasableBy — makes that test report two lanes holding the window at
// once, because its two children deliberately record the same parent pid. A
// change that turns this file green while turning that one red has not fixed
// the asymmetry; it has traded a loud refusal for a silent double-hold.

import (
	"errors"
	"os"
	"testing"
)

// THE measurement: the same owning process re-acquires its own window after its
// session id rotated.
//
// Stale() reads the holder as LIVE here, correctly — the process really is
// alive; only its session id changed. So the decision falls to the id
// comparison, which is the single key acquire has.
func TestAcquireIntegrationLock_SameOwnerProcessAfterSessionRotation(t *testing.T) {
	root := t.TempDir()
	owner := os.Getpid()
	mustAcquireOwned(t, root, "session-before-clear", owner)

	replaced, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "session-after-clear",
		PID:       owner,
		PIDSource: PIDSourceSessionOwner,
		Branch:    "release/v9.9.9",
	}, false)

	// Observed today: refused. The doc comment above AcquireIntegrationLock
	// declares the opposite. When this assertion starts failing, the code has
	// been fixed and the comment has become true — update this test then.
	if !errors.Is(err, ErrIntegrationLockHeld) {
		t.Fatalf("expected the refusal this test documents, got err=%v replaced=%+v", err, replaced)
	}

	// The refusal names the caller's own process as the blocking holder — the
	// same shape card t791 observed on the release side before t951 fixed it.
	if got := err.Error(); got == "" {
		t.Fatal("refusal carried no message")
	} else {
		t.Logf("acquire refusal after rotation: %v", err)
	}

	// The record is untouched: the refusal is clean, not a partial write.
	current, readErr := ReadIntegrationLock(root)
	if readErr != nil {
		t.Fatalf("read after refused acquire: %v", readErr)
	}
	if current.SessionID != "session-before-clear" {
		t.Errorf("refused acquire disturbed the record: %+v", current)
	}
}

// Control 1 — the id path still works. Without this, the test above cannot
// tell "rotation is refused" from "every acquire is refused".
func TestAcquireIntegrationLock_SameSessionIDReacquires(t *testing.T) {
	root := t.TempDir()
	owner := os.Getpid()
	mustAcquireOwned(t, root, "session-stable", owner)

	replaced, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "session-stable",
		PID:       owner,
		PIDSource: PIDSourceSessionOwner,
		Branch:    "release/v9.9.9",
	}, false)
	if err != nil {
		t.Fatalf("re-acquiring under the SAME id was refused: %v", err)
	}
	if replaced != nil {
		t.Errorf("re-acquiring one's own window reported a takeover: %+v", replaced)
	}
}

// Control 2 — release accepts exactly the case acquire refuses, from the same
// record. This is what makes the finding an ASYMMETRY rather than a policy:
// both verbs read the same bytes and disagree about who wrote them.
func TestAcquireAndReleaseDisagreeAfterSessionRotation(t *testing.T) {
	root := t.TempDir()
	owner := os.Getpid()
	mustAcquireOwned(t, root, "session-before-clear", owner)

	_, acqErr := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "session-after-clear",
		PID:       owner,
		PIDSource: PIDSourceSessionOwner,
		Branch:    "release/v9.9.9",
	}, false)
	if acqErr == nil {
		t.Fatal("acquire accepted the rotated caller; the asymmetry this test pins is gone")
	}

	released, relErr := ReleaseIntegrationLock(root, "session-after-clear", owner, false)
	if relErr != nil {
		t.Fatalf("release refused the same caller acquire refused: %v", relErr)
	}
	if released == nil || released.SessionID != "session-before-clear" {
		t.Errorf("release did not report the freed holder: %+v", released)
	}
}

// Control 3 — a genuinely foreign live holder must stay refused by acquire.
// A fix that admits the rotated caller must not admit this one.
func TestAcquireIntegrationLock_ForeignLiveHolderStillRefused(t *testing.T) {
	root := t.TempDir()
	foreign := os.Getppid()
	if foreign == os.Getpid() || foreign <= 1 || !FactoryProcessAlive(foreign) {
		t.Skip("no usable foreign live pid in this environment")
	}
	mustAcquireOwned(t, root, "session-foreign", foreign)

	if _, err := AcquireIntegrationLock(root, IntegrationLock{
		SessionID: "session-mine",
		PID:       os.Getpid(),
		PIDSource: PIDSourceSessionOwner,
		Branch:    "release/v9.9.9",
	}, false); !errors.Is(err, ErrIntegrationLockHeld) {
		t.Fatalf("a foreign live holder was not refused: %v", err)
	}
}
