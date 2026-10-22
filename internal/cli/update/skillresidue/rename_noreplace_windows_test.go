//go:build windows

package skillresidue

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// REQ-UBR-017: held-directory operations must keep their identity after a
// directory moves, and must never replace an independently created target.
func TestWindowsNoReplaceRenameUsesHeldDirectories(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "source")
	dst := filepath.Join(base, "destination")
	for _, name := range []string{src, dst} {
		if err := os.Mkdir(name, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	root := openRootT(t, base)
	s, err := holdDir(root, "source")
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	d, err := holdDir(root, "destination")
	if err != nil {
		t.Fatal(err)
	}
	defer d.close()
	if err := os.WriteFile(filepath.Join(src, "payload"), []byte("owned"), 0o600); err != nil {
		t.Fatal(err)
	}
	moved := dst + "-held"
	if err := os.Rename(dst, moved); err != nil {
		t.Fatalf("rename held directory: %v", err)
	}
	if err := os.Mkdir(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := noReplaceRenameAt(s.file, "payload", d.file, "payload"); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(moved, "payload")); err != nil || string(got) != "owned" {
		t.Fatalf("held destination = %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(dst, "payload")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("decoy destination received file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "payload"), []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := noReplaceRenameAt(s.file, "payload", d.file, "payload"); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("existing destination = %v, want ErrExist", err)
	}
	if got, err := os.ReadFile(filepath.Join(src, "payload")); err != nil || string(got) != "new" {
		t.Fatalf("refused source = %q, %v", got, err)
	}
}
