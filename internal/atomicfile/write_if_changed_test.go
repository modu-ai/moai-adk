package atomicfile_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// card t2013: WriteFileIfChanged performs ZERO writes when the bytes on disk
// already equal the new bytes — mtime, inode, and bytes all stay put.
func TestWriteFileIfChanged_IdenticalPerformsNoWrite(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "record.md")
	if err := os.WriteFile(path, []byte("same\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	changed, err := atomicfile.WriteFileIfChanged(path, []byte("same\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFileIfChanged: %v", err)
	}
	if changed {
		t.Errorf("identical bytes reported changed=true")
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(past) {
		t.Errorf("mtime moved on an identical write: %v -> %v", past, after.ModTime())
	}
	if !os.SameFile(before, after) {
		t.Errorf("inode changed on an identical write")
	}
	if got, _ := os.ReadFile(path); string(got) != "same\n" {
		t.Errorf("bytes = %q", got)
	}
}

// A changed payload, and an absent destination, both still write.
func TestWriteFileIfChanged_ChangedAndAbsentStillWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "record.md")

	changed, err := atomicfile.WriteFileIfChanged(path, []byte("one\n"), 0o644)
	if err != nil || !changed {
		t.Fatalf("absent destination: changed=%v err=%v", changed, err)
	}
	changed, err = atomicfile.WriteFileIfChanged(path, []byte("two\n"), 0o644)
	if err != nil || !changed {
		t.Fatalf("different bytes: changed=%v err=%v", changed, err)
	}
	if got, _ := os.ReadFile(path); string(got) != "two\n" {
		t.Errorf("bytes = %q, want two", got)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("temp file left behind: %d entries", len(entries))
	}
}
