//go:build windows

package toolpolicy

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestRecordGuardReplacementFailsClosed(t *testing.T) {
	settings := filepath.Join(t.TempDir(), ".claude", "settings.json")
	path := managedAllowLockPath(settings)
	release, err := acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if err := os.Rename(path+".guard", path+".guard.replaced"); !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("held native guard did not deny replacement: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("failed replacement changed primary record: %v", err)
	}
	release()
	if _, err := os.Stat(path + ".guard"); err != nil {
		t.Errorf("release removed guard: %v", err)
	}
}
