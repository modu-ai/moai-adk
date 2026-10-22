package atomicfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// A claim must never truncate another owner's record. Returning an open
// handle also lets the caller publish its identity before releasing the name.
func TestCreateExclusivePreservesExistingOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "owner.lock")
	first, err := CreateExclusive(path, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.WriteString("first owner"); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := CreateExclusive(path, 0o600)
	if second != nil {
		second.Close()
		t.Fatal("second owner acquired an existing claim")
	}
	if !errors.Is(err, fs.ErrExist) {
		t.Fatalf("claim = %v, want ErrExist", err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "first owner" {
		t.Fatalf("owner = %q, %v", got, err)
	}
}
