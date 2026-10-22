package privatefile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestCreateExclusiveHoldsReadDataDeleteFence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.db")
	var f *os.File
	defer func() {
		if f != nil {
			if err := f.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	if err := CreateExclusive(path, func(owned *os.File) { f = owned }); err != nil {
		t.Fatal(err)
	}
	// File.Stat captures file ID while the original handle is still owned.
	before, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if err := Check(path); err != nil {
		t.Fatal(err)
	}
	p, err := windows.UTF16PtrFromString(nativePath(path))
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.DeleteFile(p); !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("expected delete sharing refusal, got %v", err)
	}
	after, err := f.Stat()
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("original identity changed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	f = nil
	if err := os.Remove(path); err != nil {
		t.Fatalf("normal completion prevented cleanup: %v", err)
	}
}
