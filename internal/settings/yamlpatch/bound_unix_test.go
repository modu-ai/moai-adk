//go:build unix

package yamlpatch

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestBoundPatchRejectsFIFOLeafBeforeRead(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "section.yaml")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	err = checkRootSectionPath(r, "section.yaml")
	if closeErr := r.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err == nil {
		// Do not enter a known-blocking read when the pre-read predicate is broken.
		t.Fatal("FIFO accepted by pre-read shape check")
	}
	err = PatchFileAtRoot(root, "section.yaml", []KeyEdit{{Path: []string{"a"}, Value: "new"}}, RootOptions{})
	if err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
		t.Fatalf("FIFO = %v; want refusal before read", err)
	}
}
