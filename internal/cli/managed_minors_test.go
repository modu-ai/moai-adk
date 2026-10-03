package cli

// managed_minors_test.go — card t1410: the three minor findings left by
// SPEC-FACTORY-MANAGED-SESSION-001 (F8 token dir cleanup, F9 handshake
// budget, F13 readiness redirect guard).

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestManagedCodexStartEarlyFailureLeavesNoTokenDir pins F8: a Start failure
// must not leave the capability token directory on disk.
func TestManagedCodexStartEarlyFailureLeavesNoTokenDir(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"write token": func(t *testing.T) {
			old := managedCodexWriteToken
			managedCodexWriteToken = func(string, []byte, os.FileMode) error { return errors.New("forced write failure") }
			t.Cleanup(func() { managedCodexWriteToken = old })
		},
		"allocate endpoint": func(t *testing.T) {
			old := managedCodexAllocateURL
			managedCodexAllocateURL = func() (string, error) { return "", errors.New("forced endpoint failure") }
			t.Cleanup(func() { managedCodexAllocateURL = old })
		},
		"spawn": func(t *testing.T) {},
	}
	for name, force := range cases {
		t.Run(name, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			force(t)
			s := &managedCodexSession{program: filepath.Join(tmp, "no-such-binary")}
			if err := s.Start(); err == nil {
				t.Fatal("Start succeeded, want a forced failure")
			}
			left, err := filepath.Glob(filepath.Join(tmp, "moai-factory-app-*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(left) != 0 {
				t.Fatalf("token directories left behind: %v", left)
			}
			if err := s.Close(); err != nil {
				t.Fatalf("Close after failed Start = %v, want nil", err)
			}
		})
	}
}
