//go:build !windows

package migrations_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// The child confines the restrictive umask to this test's own process. Other
// migration tests must never inherit a temporary process-wide permission mask.
func TestM001PreservesOriginalModesUnderRestrictiveUmask(t *testing.T) {
	const child = "MOAI_TEST_M001_MODE_CHILD"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestM001PreservesOriginalModesUnderRestrictiveUmask$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("migration under restrictive umask: %v\n%s", err, out)
		}
		return
	}
	previous := syscall.Umask(0o077)
	defer syscall.Umask(previous)
	for _, tc := range []struct {
		name string
		mode os.FileMode
	}{
		{"private", 0o600},
		{"executable", 0o755},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			body := "#!/bin/sh\nexec " + hardcodedLiteral + " hook stop\n"
			p := writeWrapper(t, root, tc.name, body, tc.mode)
			if err := m001(t).Apply(root); err != nil {
				t.Fatalf("Apply: %v", err)
			}
			info, err := os.Stat(p)
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != tc.mode {
				t.Errorf("original mode changed: got %04o, want %04o", got, tc.mode)
			}
			if got, want := readFile(t, p), strings.ReplaceAll(body, hardcodedLiteral, portableForm); got != want {
				t.Errorf("rewritten body = %q, want %q", got, want)
			}
		})
	}
}

func TestM001DoesNotOverwriteAnExistingStagingName(t *testing.T) {
	root := t.TempDir()
	p := writeWrapper(t, root, "stop", "exec "+hardcodedLiteral+" hook stop\n", 0o600)
	foreign := p + ".tmp"
	const sentinel = "another writer's staging content"
	if err := os.WriteFile(foreign, []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := m001(t).Apply(root); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := readFile(t, foreign); got != sentinel {
		t.Errorf("unrelated staging content changed: got %q", got)
	}
	if got := readFile(t, p); got != "exec "+portableForm+" hook stop\n" {
		t.Errorf("rewritten wrapper = %q", got)
	}
	entries, err := filepath.Glob(filepath.Join(filepath.Dir(p), "handle-stop.sh.*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0] != foreign {
		t.Errorf("unexpected staging residue: %v", entries)
	}
}
