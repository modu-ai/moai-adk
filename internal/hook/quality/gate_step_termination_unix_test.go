//go:build !windows

package quality

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// processAlive reports whether the PID names a live process. Signal 0 performs
// only the permission and existence checks, so it observes liveness without
// disturbing the process.
func processAlive(pid int) bool {
	err := unix.Kill(pid, 0)
	return err == nil || errors.Is(err, unix.EPERM)
}

// AC-GTA-009 — the descendant is actually terminated, not merely waited out.
// Bounding Wait alone satisfies AC-GTA-008 in full while leaving this orphan
// running forever, which is why the two criteria are separate.
func TestRunStep_TerminatesTheStepsDescendant(t *testing.T) {
	pidPath := orphanFixture(t)

	g := NewQualityGate(DefaultGateConfig())
	parent, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	name, args := orphanCommand()
	g.runStep(parent, "orphan", 2*time.Second, name, args...)

	pid, err := readOrphanPID(pidPath)
	if err != nil {
		t.Fatalf("the fixture did not record a descendant PID: %v", err)
	}

	deadline := time.Now().Add(stepWaitGrace)
	for processAlive(pid) {
		if time.Now().After(deadline) {
			t.Fatalf("descendant %d is still alive %s after the step returned", pid, stepWaitGrace)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
