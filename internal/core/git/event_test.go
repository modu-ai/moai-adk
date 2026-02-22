package git

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestEventDetector_BranchSwitch(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	if err := bm.Create("feature"); err != nil {
		t.Fatal(err)
	}

	detector := NewEventDetector(dir)

	// Snapshot the initial state.
	if err := detector.Snapshot(); err != nil {
		t.Fatalf("Snapshot() error: %v", err)
	}

	// Switch branches.
	if err := bm.Switch("feature"); err != nil {
		t.Fatal(err)
	}

	// Detect changes.
	events, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges() error: %v", err)
	}

	if len(events) == 0 {
		t.Fatal("DetectChanges() returned no events, want BranchSwitch event")
	}

	found := false
	for _, ev := range events {
		if ev.Type == EventBranchSwitch {
			found = true
			if ev.PreviousBranch != "main" {
				t.Errorf("PreviousBranch = %q, want %q", ev.PreviousBranch, "main")
			}
			if ev.CurrentBranch != "feature" {
				t.Errorf("CurrentBranch = %q, want %q", ev.CurrentBranch, "feature")
			}
			if ev.Timestamp.IsZero() {
				t.Error("Timestamp is zero")
			}
		}
	}
	if !found {
		t.Error("BranchSwitch event not found in detected events")
	}
}

func TestEventDetector_NewCommit(t *testing.T) {
	dir := initTestRepo(t)
	detector := NewEventDetector(dir)

	if err := detector.Snapshot(); err != nil {
		t.Fatal(err)
	}

	// Create a new commit.
	writeTestFile(t, filepath.Join(dir, "new-file.txt"), "content\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "New commit")

	events, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges() error: %v", err)
	}

	if len(events) == 0 {
		t.Fatal("DetectChanges() returned no events, want NewCommit event")
	}

	found := false
	for _, ev := range events {
		if ev.Type == EventNewCommit {
			found = true
			if ev.PreviousHEAD == "" {
				t.Error("PreviousHEAD is empty")
			}
			if ev.CurrentHEAD == "" {
				t.Error("CurrentHEAD is empty")
			}
			if ev.PreviousHEAD == ev.CurrentHEAD {
				t.Error("PreviousHEAD == CurrentHEAD, want different hashes")
			}
		}
	}
	if !found {
		t.Error("NewCommit event not found in detected events")
	}
}

func TestEventDetector_NoChanges(t *testing.T) {
	dir := initTestRepo(t)
	detector := NewEventDetector(dir)

	if err := detector.Snapshot(); err != nil {
		t.Fatal(err)
	}

	// No changes made.
	events, err := detector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges() error: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("DetectChanges() returned %d events, want 0", len(events))
	}
}

func TestEventDetector_PollingCancellation(t *testing.T) {
	dir := initTestRepo(t)
	detector := NewEventDetector(dir, WithPollInterval(50*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan GitEvent, 10)

	done := make(chan error, 1)
	go func() {
		done <- detector.Poll(ctx, ch)
	}()

	// Let it poll a few times.
	time.Sleep(200 * time.Millisecond)

	// Cancel the context.
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("Poll() error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Poll() did not terminate after context cancellation")
	}
}

func TestEventDetector_PollDetectsChanges(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	if err := bm.Create("feature"); err != nil {
		t.Fatal(err)
	}

	detector := NewEventDetector(dir, WithPollInterval(50*time.Millisecond))

	ctx := t.Context()

	ch := make(chan GitEvent, 10)

	go func() {
		// Ignore error; we cancel via context.
		_ = detector.Poll(ctx, ch)
	}()

	// Wait for initial snapshot to be captured.
	time.Sleep(100 * time.Millisecond)

	// Switch branches to trigger an event.
	if err := bm.Switch("feature"); err != nil {
		t.Fatal(err)
	}

	// Wait for poll to detect it.
	select {
	case ev := <-ch:
		if ev.Type != EventBranchSwitch {
			t.Errorf("event.Type = %q, want %q", ev.Type, EventBranchSwitch)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for branch switch event from Poll()")
	}
}
