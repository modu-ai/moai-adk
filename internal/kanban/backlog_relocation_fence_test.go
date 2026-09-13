package kanban

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestPeerRelocationAlreadyPublishedFencesSource(t *testing.T) {
	base := t.TempDir()
	from := filepath.Join(base, "legacy")
	to := filepath.Join(base, "canonical")
	stale := NewBacklogStore(filepath.Join(from, "backlog.json"))
	if _, _, err := stale.Add("source"); err != nil {
		t.Fatal(err)
	}
	target := NewBacklogStore(filepath.Join(to, "backlog.json"))
	if _, _, err := target.Add("already published"); err != nil {
		t.Fatal(err)
	}
	if err := relocateQueueArtifacts(from, to); err != nil {
		t.Fatal(err)
	}
	_, _, err := stale.Add("late writer")
	if !errors.Is(err, ErrBacklogRelocated) {
		t.Fatalf("stale writer err=%v; expected relocated refusal after target already existed", err)
	}
}
