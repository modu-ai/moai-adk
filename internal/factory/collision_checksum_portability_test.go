package factory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollisionChecksumReadsBytesWithoutExternalHasher(t *testing.T) {
	path := filepath.Join(t.TempDir(), "asset")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	r := &collisionRepo{t: t}
	if got := r.checksum(path); got != "a9993e364706816aba3e25717850c26c9cd0d89d" {
		t.Fatalf("existing asset checksum=%q, want the SHA-1 of its real bytes without external tools", got)
	}
	if got := r.checksum(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Fatalf("missing asset checksum=%q, want empty", got)
	}
}
