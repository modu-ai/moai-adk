package privatefile

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestOpenPinnedAtomicPrivacyAndDeleteSharing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private.db")
	var f *os.File
	err := OpenPinned(path, func(pin io.Closer) {
		f = pin.(*os.File)
		t.Cleanup(func() {
			if err := f.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
				t.Errorf("cleanup acquired pin: %v", err)
			}
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := checkHandle(windows.Handle(f.Fd())); err != nil {
		t.Fatal(err)
	}
	info, err := f.Stat()
	if err != nil || info.Size() != 0 {
		t.Fatalf("before bytes info=%v error=%v", info, err)
	}
	if err := os.Remove(path); err == nil {
		t.Fatal("READ_DATA/no DELETE share pin did not prevent deletion")
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Check(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func TestOpenPinnedRefusesHardlinkBeforeNarrowing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "original")
	alias := filepath.Join(dir, "alias")
	if err := os.WriteFile(path, []byte("synthetic"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(path, alias); err != nil {
		t.Fatal(err)
	}
	var f *os.File
	err := OpenPinned(alias, func(pin io.Closer) {
		f = pin.(*os.File)
		t.Cleanup(func() {
			if err := f.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
				t.Errorf("cleanup refused pin: %v", err)
			}
		})
	})
	if f != nil {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err == nil {
		t.Fatal("multiply linked object accepted")
	}
}
