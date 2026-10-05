package guardliveness

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The persisted result is what makes the render a separate act from the
// refresh (REQ-GDL-011): the host surface reads THIS, and issues no query of
// its own.
func TestStoreRoundTripsAResultWithItsTimestamp(t *testing.T) {
	store := NewStore(t.TempDir())
	root := t.TempDir()
	takenAt := time.Now().Add(-42 * time.Minute).Truncate(time.Second)

	if err := store.Save(root, resultA(), takenAt); err != nil {
		t.Fatalf("Save: %v", err)
	}

	snap, err := store.Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !snap.TakenAt.Equal(takenAt) {
		t.Errorf("TakenAt = %v, want %v — the age would be derived from the wrong moment", snap.TakenAt, takenAt)
	}

	nonClean, err := snap.Result.Partition()
	if err != nil {
		t.Fatalf("Partition of the loaded result: %v", err)
	}
	if got, want := subjectsOf(nonClean), []string{"subject-2", "subject-3"}; !equalStrings(got, want) {
		t.Errorf("loaded result partitions to %v, want %v — the classification designation did not survive the round trip", got, want)
	}
}

// A contract-violating result must survive persistence unchanged, or the
// advisory that has to REPORT the violation (AC-GDL-013) never sees one: a
// store that quietly normalizes a malformed result renders it green.
func TestStorePreservesAContractViolatingResult(t *testing.T) {
	store := NewStore(t.TempDir())
	root := t.TempDir()
	violating := Result{Clean: nil, Entries: []Entry{entry("subject-1", "alpha", "settled")}}

	if err := store.Save(root, violating, time.Now()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	snap, err := store.Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := snap.Result.Partition(); !errors.Is(err, ErrDesignationAbsent) {
		t.Fatalf("loaded result partitions with err = %v, want %v — the violation was normalized away in transit", err, ErrDesignationAbsent)
	}
}

// Before any refresh has completed there is nothing to read, and that is a
// distinct outcome from an all-clear.
func TestStoreLoadReportsAnAbsentResult(t *testing.T) {
	store := NewStore(t.TempDir())
	if _, err := store.Load(t.TempDir()); !errors.Is(err, ErrNoPersistedResult) {
		t.Fatalf("Load of an unwritten root: err = %v, want %v", err, ErrNoPersistedResult)
	}
}

// One store serves every tree on the machine, so two roots must not read each
// other's verdicts.
func TestStoreKeysByRootSoTreesDoNotCollide(t *testing.T) {
	store := NewStore(t.TempDir())
	first, second := t.TempDir(), t.TempDir()

	if err := store.Save(first, resultA(), time.Now()); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	if err := store.Save(second, resultB(), time.Now()); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	snap, err := store.Load(first)
	if err != nil {
		t.Fatalf("Load first: %v", err)
	}
	nonClean, err := snap.Result.Partition()
	if err != nil {
		t.Fatalf("Partition: %v", err)
	}
	if got, want := subjectsOf(nonClean), []string{"subject-2", "subject-3"}; !equalStrings(got, want) {
		t.Fatalf("first root read %v, want %v — the two trees share a key", got, want)
	}
}

// REQ-GDL-008 — the persistence lives outside the working tree. Asserted as the
// evaluated root being byte-identical across a save, which is the property
// AC-GDL-008(b) measures with `git status --porcelain`.
func TestStoreWritesNothingIntoTheEvaluatedRoot(t *testing.T) {
	store := NewStore(t.TempDir())
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "tracked.txt"), []byte("before\n"), 0o644); err != nil {
		t.Fatalf("seed root: %v", err)
	}

	before := treeListing(t, root)
	if err := store.Save(root, resultA(), time.Now()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if after := treeListing(t, root); !equalStrings(before, after) {
		t.Fatalf("the evaluated root changed across a save:\nbefore %v\nafter  %v", before, after)
	}
}

func treeListing(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		out = append(out, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}
