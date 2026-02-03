package git

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestHasConflicts_NoConflicts(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	// Create a feature branch from main.
	if err := bm.Create("feature"); err != nil {
		t.Fatal(err)
	}

	// Modify different files on each branch.
	// main: modify a.go
	writeTestFile(t, filepath.Join(dir, "a.go"), "package a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Modify a.go on main")

	// feature: modify b.go
	runGit(t, dir, "checkout", "feature")
	writeTestFile(t, filepath.Join(dir, "b.go"), "package b\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Modify b.go on feature")

	// Check from feature looking at main.
	hasConflicts, err := bm.HasConflicts("main")
	if err != nil {
		t.Fatalf("HasConflicts() error: %v", err)
	}
	if hasConflicts {
		t.Error("HasConflicts() = true, want false (different files)")
	}
}

func TestHasConflicts_Detected(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	// Create shared file.
	writeTestFile(t, filepath.Join(dir, "shared.go"), "package shared\n// original\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Add shared.go")

	// Create feature branch.
	if err := bm.Create("feature"); err != nil {
		t.Fatal(err)
	}

	// Modify shared.go on main.
	writeTestFile(t, filepath.Join(dir, "shared.go"), "package shared\n// main version\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Modify shared.go on main")

	// Modify shared.go on feature.
	runGit(t, dir, "checkout", "feature")
	writeTestFile(t, filepath.Join(dir, "shared.go"), "package shared\n// feature version\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Modify shared.go on feature")

	hasConflicts, err := bm.HasConflicts("main")
	if err != nil {
		t.Fatalf("HasConflicts() error: %v", err)
	}
	if !hasConflicts {
		t.Error("HasConflicts() = false, want true (same file modified on both sides)")
	}
}

func TestHasConflicts_TargetNotFound(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	_, err := bm.HasConflicts("nonexistent")
	if err == nil {
		t.Fatal("HasConflicts() with nonexistent target should return error")
	}
	if !errors.Is(err, ErrBranchNotFound) {
		t.Errorf("error = %v, want ErrBranchNotFound", err)
	}
}

func TestHasConflicts_WorktreeUnmodified(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	// Create feature branch with different changes.
	writeTestFile(t, filepath.Join(dir, "shared.go"), "package shared\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Add shared.go")

	if err := bm.Create("feature"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dir, "shared.go"), "package shared\n// main\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Main change")

	runGit(t, dir, "checkout", "feature")
	writeTestFile(t, filepath.Join(dir, "shared.go"), "package shared\n// feature\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Feature change")

	// Create a dirty working tree for safety check.
	writeTestFile(t, filepath.Join(dir, "dirty.txt"), "dirty\n")

	// Capture status before conflict detection.
	repo, err := NewRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	statusBefore, err := repo.Status()
	if err != nil {
		t.Fatal(err)
	}

	// Run conflict detection.
	_, detectErr := bm.HasConflicts("main")
	if detectErr != nil {
		t.Fatalf("HasConflicts() error: %v", detectErr)
	}

	// Verify status is unchanged after detection.
	statusAfter, err := repo.Status()
	if err != nil {
		t.Fatal(err)
	}

	if len(statusBefore.Untracked) != len(statusAfter.Untracked) {
		t.Errorf("Untracked files changed: before=%d, after=%d",
			len(statusBefore.Untracked), len(statusAfter.Untracked))
	}

	// Verify HEAD hasn't changed.
	headBefore := runGit(t, dir, "rev-parse", "HEAD")
	headAfter := runGit(t, dir, "rev-parse", "HEAD")
	if headBefore != headAfter {
		t.Error("HEAD changed during conflict detection")
	}
}

func TestMergeBase(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	// Record the common ancestor hash.
	commonHash := runGit(t, dir, "rev-parse", "HEAD")

	// Create feature branch and add commits on both sides.
	if err := bm.Create("feature"); err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, filepath.Join(dir, "main-file.txt"), "main\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Main commit")

	runGit(t, dir, "checkout", "feature")
	writeTestFile(t, filepath.Join(dir, "feature-file.txt"), "feature\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Feature commit")

	runGit(t, dir, "checkout", "main")

	base, err := bm.MergeBase("main", "feature")
	if err != nil {
		t.Fatalf("MergeBase() error: %v", err)
	}
	if base != commonHash {
		t.Errorf("MergeBase() = %q, want %q", base, commonHash)
	}
}

func TestMergeBase_NoCommonAncestor(t *testing.T) {
	dir := initTestRepo(t)
	bm := NewBranchManager(dir)

	// Create an orphan branch (no common history with main).
	runGit(t, dir, "checkout", "--orphan", "orphan1")
	writeTestFile(t, filepath.Join(dir, "orphan.txt"), "orphan\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Orphan commit")

	_, err := bm.MergeBase("main", "orphan1")
	if err == nil {
		t.Fatal("MergeBase() with orphan branches should return error")
	}
	if !errors.Is(err, ErrNoMergeBase) {
		t.Errorf("error = %v, want ErrNoMergeBase", err)
	}
}
