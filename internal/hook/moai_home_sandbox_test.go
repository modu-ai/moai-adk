package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/sqliteguard"
)

// TestMainSandboxesProfileLeaseEnv pins the TestMain premise that keeps this
// package's SessionStart/SessionEnd runs off the real profile-lease database
// (card t1229): MOAI_HOME is the binary's own sandbox and the lane lease
// variables are gone.
func TestMainSandboxesProfileLeaseEnv(t *testing.T) {
	home := os.Getenv(config.EnvHome)
	if home == "" || home != os.Getenv(moaiHomeSandboxEnv) {
		t.Fatalf("MOAI_HOME=%q is not the TestMain sandbox %q", home, os.Getenv(moaiHomeSandboxEnv))
	}
	if err := privatefile.CheckDirectory(home); err != nil {
		t.Fatalf("TestMain home native privacy: %v", err)
	}
	// No CLAUDE_CONFIG_DIR check: profile.EnsureDir leaks it order-dependently, and the MOAI_HOME sandbox makes a leak harmless.
	if v, ok := os.LookupEnv("MOAI_PROFILE_LEASE_TOKEN"); ok {
		t.Fatalf("MOAI_PROFILE_LEASE_TOKEN=%q survived TestMain", v)
	}
}

// privateHookHome leaves the unrelated TempDir parent untouched. Windows
// tokens may assign an ordinary temp directory to a different default owner.
func privateHookHome(t *testing.T) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "managed-home")
	if err := sqliteguard.EnsureDirectory(home); err != nil {
		t.Fatal(err)
	}
	return home
}
