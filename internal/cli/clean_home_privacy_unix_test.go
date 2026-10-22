//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
)

func TestSecureHomeDirectoriesKeepsUnixFilePolicy(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "user-file")
	if err := os.WriteFile(path, []byte("preserved"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := secureHomeDirectories(root); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.CheckDirectory(root); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("Unix file mode = %o, want 0644 preserved", info.Mode().Perm())
	}
}
