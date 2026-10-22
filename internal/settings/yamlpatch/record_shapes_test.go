package yamlpatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A dangling or present symlink is an existing user entry, not an absent
// greenfield section. The writer must leave both the link and target untouched.
func TestAtomicWriteRefusesSymlinkTarget(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "outside.yaml")
	const old = "outside: preserved\n"
	if err := os.WriteFile(outside, []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{outside, filepath.Join(t.TempDir(), "missing.yaml")} {
		name := filepath.Join(t.TempDir(), "section.yaml")
		if err := os.Symlink(target, name); err != nil {
			t.Fatalf("create section symlink fixture: %v", err)
		}
		err := atomicWrite(name, []byte("outside: replaced\n"))
		if err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
			t.Errorf("linked section = %v; want stat refusal", err)
		}
		if info, err := os.Lstat(name); err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("linked section entry changed: %v, %v", info, err)
		}
	}
	if got, err := os.ReadFile(outside); err != nil || string(got) != old {
		t.Fatalf("outside payload changed: %q, %v", got, err)
	}
}
