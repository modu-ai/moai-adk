//go:build windows

package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestWindowsExclusiveClaimWaitsForDeletePendingHandle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claim.lock")
	prior, err := CreateExclusive(path, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer prior.Close()
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove open delete-sharing claim: %v", err)
	}
	done := make(chan error, 1)
	go func() { time.Sleep(20 * time.Millisecond); done <- prior.Close() }()
	next, err := CreateExclusive(path, 0o600)
	if err != nil {
		t.Fatalf("claim after delete pending: %v", err)
	}
	defer next.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestWindowsOwnerReadAllowsRemovalWhileOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "owner.lock")
	if err := os.WriteFile(path, []byte("owner"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := openRead(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := os.Remove(path); err != nil {
		t.Fatalf("verdict reader blocked removal: %v", err)
	}
}

func TestWindowsUnverifiableProcessRemainsAlive(t *testing.T) {
	old := openProcessForLock
	t.Cleanup(func() { openProcessForLock = old })
	openProcessForLock = func(uint32, bool, uint32) (windows.Handle, error) { return 0, windows.ERROR_ACCESS_DENIED }
	if !PidAlive(os.Getpid()) {
		t.Fatal("access denied was treated as death")
	}
	openProcessForLock = func(uint32, bool, uint32) (windows.Handle, error) { return 0, windows.ERROR_INVALID_PARAMETER }
	if PidAlive(2147483647) {
		t.Fatal("positively absent process was treated as alive")
	}
	if !errors.Is(windows.ERROR_ACCESS_DENIED, os.ErrPermission) {
		t.Fatal("Windows permission classifier fixture changed")
	}
}
