package testguard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileModePolicyChecksActualPermissionsAndExclusiveSeed(t *testing.T) {
	for _, perm := range []os.FileMode{0o600, 0o644, 0o755} {
		t.Run(perm.String(), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "fixture")
			if err := SeedFileModePolicy(path, []byte("fixture bytes\n"), perm); err != nil {
				t.Fatal(err)
			}
			if err := CheckFileModePolicy(path, perm); err != nil {
				t.Fatal(err)
			}
			wrongMode := os.FileMode(0o600)
			if perm == wrongMode {
				wrongMode = 0o644
			}
			if err := CheckFileModePolicy(path, wrongMode); err == nil {
				t.Fatalf("actual %04o policy was accepted as %04o", perm, wrongMode)
			}
			if err := SeedFileModePolicy(path, []byte("clobber"), perm); !os.IsExist(err) {
				t.Errorf("exclusive seed must refuse existing fixture: %v", err)
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != "fixture bytes\n" {
				t.Fatalf("real fixture bytes changed or unreadable: %q err=%v", data, err)
			}
		})
	}
}
