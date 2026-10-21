//go:build !windows

package privatefile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckRejectsWeakUnixPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Check(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Check(path); err == nil {
		t.Fatal("world-readable receipt accepted")
	}
}
