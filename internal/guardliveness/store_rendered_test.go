package guardliveness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The render record survives the round trip that matters: written by one
// activation's render, read by the next one's. A shape that did not survive it
// would make every standing entry look changed, which is the full standing list
// rebuilt one layer down.
func TestRenderRecordRoundTripsThroughTheStore(t *testing.T) {
	store := NewStore(t.TempDir())
	root := t.TempDir()
	renderedAt := time.Now().Add(-20 * time.Minute).UTC().Truncate(time.Second)

	want := RenderRecord{
		RenderedAt:      renderedAt,
		Classifications: map[string]string{"subject-2": "beta", "subject-3": "gamma"},
	}
	if err := store.SaveRendered(root, want); err != nil {
		t.Fatalf("SaveRendered: %v", err)
	}

	got, err := store.LoadRendered(root)
	if err != nil {
		t.Fatalf("LoadRendered: %v", err)
	}
	if !got.RenderedAt.Equal(renderedAt) {
		t.Errorf("RenderedAt = %v, want %v", got.RenderedAt, renderedAt)
	}
	for subject, classification := range want.Classifications {
		if got.Classifications[subject] != classification {
			t.Errorf("classification for %q = %q, want %q", subject, got.Classifications[subject], classification)
		}
	}
}

// Before any render has happened there is no record, and the honest answer is
// an empty one rather than an error: every non-clean entry is then a change,
// which is what a reader who has seen nothing yet needs.
func TestLoadRenderedIsEmptyBeforeAnyRender(t *testing.T) {
	store := NewStore(t.TempDir())
	got, err := store.LoadRendered(t.TempDir())
	if err != nil {
		t.Fatalf("LoadRendered with nothing written: %v", err)
	}
	if len(got.Classifications) != 0 || !got.RenderedAt.IsZero() {
		t.Fatalf("LoadRendered returned %+v with nothing written", got)
	}
}

// The record and the persisted result are distinct files. Sharing a path would
// make each render overwrite the verdict it just read, so the next activation
// would see a measurement that never happened.
func TestRenderRecordDoesNotOverwriteThePersistedResult(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	root := t.TempDir()
	takenAt := time.Now().Add(-time.Hour)

	if err := store.Save(root, resultA(), takenAt); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := store.SaveRendered(root, RenderRecord{
		RenderedAt:      time.Now(),
		Classifications: map[string]string{"subject-2": "beta"},
	}); err != nil {
		t.Fatalf("SaveRendered: %v", err)
	}

	snap, err := store.Load(root)
	if err != nil {
		t.Fatalf("Load after SaveRendered: %v", err)
	}
	if !snap.TakenAt.Equal(takenAt) {
		t.Fatalf("the persisted result's timestamp changed to %v, want %v", snap.TakenAt, takenAt)
	}
	if len(snap.Result.Entries) != len(resultA().Entries) {
		t.Fatalf("the persisted result was overwritten by the render record: %+v", snap.Result)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read the persistence directory: %v", err)
	}
	if len(entries) != 2 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, filepath.Base(e.Name()))
		}
		t.Fatalf("persistence directory holds %d file(s) %v, want 2 (a result and a render record)", len(entries), names)
	}
}
