package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Doctor must observe the same per-user install root as init/update. Both
// candidate roots are temporary; the resolver's error must never fall back to
// another account/environment root or turn its install into a healthy verdict.
func TestUserCodexInstallPresenceUsesInstallerHome(t *testing.T) {
	for _, tc := range []struct {
		name          string
		installerFile bool
		ambientFile   bool
		resolverError bool
		want          bool
	}{
		{"installer only", true, false, false, true},
		{"ambient only", false, true, false, false},
		{"neither", false, false, false, false},
		{"resolver failure", false, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			installer, ambient := t.TempDir(), t.TempDir()
			t.Setenv("HOME", ambient)
			t.Setenv("USERPROFILE", ambient)
			old := userHomeDirFn
			t.Cleanup(func() { userHomeDirFn = old })
			userHomeDirFn = func() (string, error) {
				if tc.resolverError {
					return "", errors.New("installer home unavailable")
				}
				return installer, nil
			}
			for home, create := range map[string]bool{installer: tc.installerFile, ambient: tc.ambientFile} {
				if !create {
					continue
				}
				path := filepath.Join(home, ".agents", "skills", "moai", "SKILL.md")
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("# fixture\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := userCodexInstallPresent(); got != tc.want {
				t.Errorf("installer-root presence = %v, want %v (ambient file=%v)", got, tc.want, tc.ambientFile)
			}
		})
	}
}
