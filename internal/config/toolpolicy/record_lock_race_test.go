package toolpolicy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func waitRecordRaceSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatal("record-lock race barrier did not complete")
	}
}

func TestRecordLockSecondStaleObserverCannotDeleteNewWinner(t *testing.T) {
	settings := filepath.Join(t.TempDir(), ".claude", "settings.json")
	path := managedAllowLockPath(settings)
	seedDeadPidLock(t, path)
	seen := [2]chan struct{}{make(chan struct{}), make(chan struct{})}
	resume := [2]chan struct{}{make(chan struct{}), make(chan struct{})}
	results := [2]chan error{make(chan error, 1), make(chan error, 1)}
	done := [2]chan struct{}{make(chan struct{}), make(chan struct{})}
	for i := range seen {
		i := i
		go func() {
			defer close(done[i])
			results[i] <- takeoverStaleRecordLockObserved(path, func() {
				close(seen[i])
				<-resume[i]
			})
		}()
	}
	// Always unblock and reap both waiters, even when a barrier assertion fails.
	t.Cleanup(func() {
		for _, ch := range resume {
			select {
			case <-ch:
			default:
				close(ch)
			}
		}
		for _, ch := range done {
			waitRecordRaceSignal(t, ch)
		}
	})
	waitRecordRaceSignal(t, seen[0])
	waitRecordRaceSignal(t, seen[1])
	close(resume[0])
	if err := <-results[0]; err != nil {
		t.Fatalf("first stale observer: %v", err)
	}
	release, err := acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatalf("acquire first winner: %v", err)
	}
	defer release()
	// The observation itself must allow native deletion; a default Windows
	// os.Open handle would hide the very disposal race this test exercises.
	winner, err := openRecordObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	defer winner.Close()
	winnerInfo, err := winner.Stat()
	if err != nil {
		t.Fatal(err)
	}
	winnerBody, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	close(resume[1])
	if err := <-results[1]; err == nil {
		t.Error("second observer accepted a replaced lock instance")
	}
	current, err := os.Stat(path)
	if err != nil {
		t.Fatalf("second observer removed first winner: %v", err)
	}
	if !os.SameFile(winnerInfo, current) {
		t.Error("second observer changed first winner's lock identity")
	}
	if body, err := os.ReadFile(path); err != nil || !bytes.Equal(body, winnerBody) {
		t.Errorf("second observer changed winner stamp: %q, %v", body, err)
	}
}

func TestRecordLockOldReleaseCannotDeleteSuccessor(t *testing.T) {
	settings := filepath.Join(t.TempDir(), ".claude", "settings.json")
	path := managedAllowLockPath(settings)
	oldRelease, err := acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatal(err)
	}
	defer oldRelease()
	stale := time.Now().Add(-recordLockStaleAge - time.Second)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}
	if err := takeoverStaleRecordLock(path); err != nil {
		t.Fatalf("take over old instance: %v", err)
	}
	newRelease, err := acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatalf("acquire successor: %v", err)
	}
	defer newRelease()
	winner, err := openRecordObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	defer winner.Close()
	winnerInfo, err := winner.Stat()
	if err != nil {
		t.Fatal(err)
	}
	winnerBody, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	oldRelease()
	current, err := os.Stat(path)
	if err != nil {
		t.Fatalf("old release removed successor: %v", err)
	}
	if !os.SameFile(winnerInfo, current) {
		t.Error("old release changed successor identity")
	}
	if body, err := os.ReadFile(path); err != nil || !bytes.Equal(body, winnerBody) {
		t.Errorf("old release changed successor stamp: %q, %v", body, err)
	}
}
