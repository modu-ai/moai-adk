package cli

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestMainSandboxesProfileLeaseEnv pins the TestMain premise that keeps this
// package's SessionStart runs off the real profile-lease database (card t1229):
// MOAI_HOME is the binary's own sandbox and the lane lease variables are gone.
func TestMainSandboxesProfileLeaseEnv(t *testing.T) {
	home := os.Getenv(config.EnvHome)
	if home == "" || home != os.Getenv(moaiHomeSandboxEnv) {
		t.Fatalf("MOAI_HOME=%q is not the TestMain sandbox %q", home, os.Getenv(moaiHomeSandboxEnv))
	}
	for _, key := range []string{"MOAI_PROFILE_LEASE_TOKEN", config.EnvClaudeConfigDir} {
		if v, ok := os.LookupEnv(key); ok {
			t.Fatalf("%s=%q survived TestMain", key, v)
		}
	}
}
