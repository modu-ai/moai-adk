package atomicfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// TestReplace_OntoAbsentDestination covers the first-write path: the
// destination does not exist yet.
func TestReplace_OntoAbsentDestination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.tmp")
	dst := filepath.Join(dir, "dst.json")

	if err := os.WriteFile(src, []byte("first"), 0o644); err != nil {
		t.Fatalf("seed src: %v", err)
	}
	if err := atomicfile.Replace(src, dst); err != nil {
		t.Fatalf("Replace: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "first" {
		t.Errorf("dst content = %q, want %q", got, "first")
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("src should be gone after replace; stat err=%v", err)
	}
}

// TestReplace_OverExistingDestination is the case that fails with a bare
// os.Rename on Windows when a handle is open on the destination.
func TestReplace_OverExistingDestination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.tmp")
	dst := filepath.Join(dir, "dst.json")

	if err := os.WriteFile(dst, []byte("stale"), 0o644); err != nil {
		t.Fatalf("seed dst: %v", err)
	}
	if err := os.WriteFile(src, []byte("fresh"), 0o644); err != nil {
		t.Fatalf("seed src: %v", err)
	}
	if err := atomicfile.Replace(src, dst); err != nil {
		t.Fatalf("Replace: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "fresh" {
		t.Errorf("dst content = %q, want %q", got, "fresh")
	}
}

// TestReplace_MissingSourceFails asserts a non-transient error is surfaced
// immediately rather than being swallowed by the Windows retry budget.
func TestReplace_MissingSourceFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	err := atomicfile.Replace(filepath.Join(dir, "nope.tmp"), filepath.Join(dir, "dst.json"))
	if err == nil {
		t.Fatal("Replace with missing source: want error, got nil")
	}
}
