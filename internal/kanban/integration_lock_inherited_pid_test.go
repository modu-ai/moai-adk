package kanban

import "testing"

// TestReleasableBy_InheritedOwnerPIDCrossesSessions is the consequence arm of
// card t958's reproduction, read-only with respect to integration_lock.go.
//
// The session layer's ResolveOwnerPID honors MOAI_SESSION_PID unconditionally,
// so two independent sessions that INHERITED the same value resolve to the
// same owner pid (see
// internal/session.TestResolveOwnerPID_InheritedStampCollapsesDistinctSessions).
// This test states what that collapse buys at the lock layer: the pid key is
// admitted on a PIDSourceSessionOwner record, so session B satisfies the
// holder judgment on session A's window.
//
// Attribution: PIDSourceSessionOwner predates this, from commit 3f3465369
// (card t298); commit ea939d9a7 (card t951) added releasableBy, making pid a
// SECOND key on the RELEASE holder judgment and thereby WIDENING the exposure
// surface of that pre-existing acquire-era property. Nothing here asserts a
// t951 regression — the lock layer's pid key is doing exactly what it says.
// The separable defect is upstream, in what ResolveOwnerPID hands it.
//
// So this is a CHARACTERIZATION, not a criterion the lock layer must move to
// satisfy: it asserts the true that releasableBy returns today, and it will
// keep asserting it after card t958's fix, because the fix lands in
// internal/session and nothing in this file changes. What the fix removes is
// the PREMISE — after it, two independent sessions no longer resolve to a
// shared owner pid, so the reachable input to this judgment goes away while
// the judgment itself is untouched. A future change that makes this
// assertion fail has altered the release holder judgment and owes its own
// reasoning.
func TestReleasableBy_InheritedOwnerPIDCrossesSessions(t *testing.T) {
	// The window recorded by session A. inheritedPID is the value BOTH
	// sessions resolved to, because both inherited the same stamp.
	const inheritedPID = 7000
	held := &IntegrationLock{
		SessionID: "lane-a",
		PIDSource: PIDSourceSessionOwner,
		PID:       inheritedPID,
	}

	// Control: session A's own release is admitted by the session-id key, so
	// the pid key is not the only thing that could be answering below.
	if !held.releasableBy("lane-a", inheritedPID) {
		t.Fatal("the recorded holder cannot release its own window; the fixture is wrong, not the code")
	}

	// Negative control: without the marker the pid key is not admitted at
	// all, so the true below is the pid key answering rather than some other
	// path admitting every caller.
	legacy := &IntegrationLock{SessionID: "lane-a", PID: inheritedPID}
	if legacy.releasableBy("lane-b", inheritedPID) {
		t.Fatal("a record with no PIDSource marker admitted the pid key; this fixture measures nothing")
	}

	if !held.releasableBy("lane-b", inheritedPID) {
		t.Errorf("releasableBy(lane-b, %d) = false; the consequence this test documents is gone, and the reasoning above no longer describes the code", inheritedPID)
	} else {
		t.Logf("documented consequence: session lane-b satisfies the holder judgment on lane-a's window on the strength of the shared owner pid %d — closed upstream, in what ResolveOwnerPID resolves", inheritedPID)
	}
}
