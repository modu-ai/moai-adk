//go:build windows

// integration_lock_mutation_windows_test.go — AC-ILA-007's behavioural half
// (REQ-ILA-006, card t336): on the atomic-create substrate a mutation-lock
// holder killed mid-section must not wedge the record permanently, and a LIVE
// holder's artifact must be left strictly alone.
//
// These cases run on a Windows runner. On other hosts the build tag excludes
// them entirely — no darwin-lane command compiles this file, so no darwin
// result may be cited for it; `GOOS=windows go vet ./internal/kanban/...`
// verifies compilation only, and CI's windows job is the behavioural verdict,
// running these alongside the board's pre-existing clear criteria
// (board_lock_clear_windows_test.go), which must not regress.
package kanban

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeMutationArtifact plants a mutation-lock artifact naming pid as its owner
// — the shape a process killed inside the critical section leaves behind.
func writeMutationArtifact(t *testing.T, root string, pid int) string {
	t.Helper()
	path := integrationMutationLockPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	raw, err := json.MarshalIndent(BoardLockOwner{
		PID:       pid,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal owner: %v", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("write mutation artifact: %v", err)
	}
	return path
}

// TestIntegrationMutationLock_DeadOwnerDoesNotWedge — a wedged artifact whose
// recorded owner is positively dead is cleared, and the mutation that was
// blocked then runs. Without this the first killed `moai integration acquire`
// would block every subsequent one on the machine, for a lock whose whole
// lifetime is supposed to be one CLI invocation.
func TestIntegrationMutationLock_DeadOwnerDoesNotWedge(t *testing.T) {
	root := t.TempDir()
	path := writeMutationArtifact(t, root, deadPIDWin(t))

	ran := false
	if err := withIntegrationLockMutation(root, func() error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("withIntegrationLockMutation over a dead-owner artifact: %v — a killed short-lived holder has wedged the record", err)
	}
	if !ran {
		t.Fatal("the critical section never ran, yet no error was returned")
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("the mutation artifact survived the section — it must be released on every path")
	}
}

// TestIntegrationMutationLock_LiveOwnerIsNotCleared — an artifact whose
// recorded owner is LIVE is left alone and the contender reports busy, not
// held. Clearing on a guess would unlink a lock a live process holds and admit
// two writers to the critical section: the clear causing precisely the
// concurrency the lock exists to prevent.
func TestIntegrationMutationLock_LiveOwnerIsNotCleared(t *testing.T) {
	root := t.TempDir()
	// This test process is unambiguously alive.
	path := writeMutationArtifact(t, root, os.Getpid())

	ran := false
	err := withIntegrationLockMutation(root, func() error {
		ran = true
		return nil
	})
	if err == nil {
		t.Fatal("the section ran against a LIVE holder's artifact — the clear removed a lock it could not prove absent")
	}
	if ran {
		t.Fatal("the critical section ran despite the error")
	}
	if !IsIntegrationLockBusy(err) {
		t.Fatalf("error = %v, want the busy sentinel — a wedged mutation lock is transient contention, never a statement that another session owns the window", err)
	}
	if IsIntegrationLockHeld(err) {
		t.Fatalf("error = %v also satisfies IsIntegrationLockHeld — a lane would be told a live peer owns a window that is free", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("the live owner's artifact was removed: %v", statErr)
	}
}
