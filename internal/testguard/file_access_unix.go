//go:build !windows

package testguard

import (
	"os"
	"testing"
)

// MakeFilePublic weakens a fixture file's access policy for rejection tests.
// It must only be used on test-owned files that contain no real secrets.
func MakeFilePublic(t testing.TB, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
}

// MakeDirReadOnly denies fixture file creation and restores cleanup access.
func MakeDirReadOnly(t testing.TB, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(path, 0o700); err != nil {
			t.Error(err)
		}
	})
}
