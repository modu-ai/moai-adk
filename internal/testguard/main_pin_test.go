package testguard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// TestMain_ScrubsAndPinsGitConfigGlobal observes what TestMain did to this
// process: GIT_CONFIG_GLOBAL names an existing temp file the home guard accepts,
// and no repository-scoping variable survived the scrub (card t1801).
func TestMain_ScrubsAndPinsGitConfigGlobal(t *testing.T) {
	pin := os.Getenv(gitGlobalConfigEnv)
	if pin == "" {
		t.Fatalf("%s is unset; TestMain must pin it", gitGlobalConfigEnv)
	}
	if !filepath.IsAbs(pin) {
		t.Fatalf("%s = %q; want an absolute path", gitGlobalConfigEnv, pin)
	}
	if fi, err := os.Stat(pin); err != nil || !fi.Mode().IsRegular() {
		t.Fatalf("%s = %q; want an existing regular file (stat err %v)", gitGlobalConfigEnv, pin, err)
	}
	realHome, err := RealHome()
	if err != nil {
		t.Fatalf("RealHome: %v", err)
	}
	if err := CheckGlobalConfigPin(realHome, pin); err != nil {
		t.Fatalf("the pin %q is not acceptable to the home guard: %v", pin, err)
	}

	for _, name := range gitenv.RepoScopingVars {
		if v, ok := os.LookupEnv(name); ok {
			t.Errorf("%s=%q survived gitenv.ScrubProcess", name, v)
		}
	}
}
