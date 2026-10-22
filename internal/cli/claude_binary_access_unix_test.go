//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func nonExecutableClaudePinFixture(t *testing.T) (string, string) {
	t.Helper()
	plain := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(plain, []byte("not executable"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(plain)
	if err != nil || info.Mode()&0o111 != 0 {
		t.Fatalf("non-executable fixture mode: %v %v", info, err)
	}
	return plain, "chmod +x"
}
