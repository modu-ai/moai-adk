package cli

import (
	"strings"
	"testing"
)

// t1710 — the delivery edge's gh environment scrubs the git variables but
// keeps GH_REPO, and gh honors GH_REPO over the directory it is started in: a
// stray GH_REPO retargets the PR operations of the delivery edge to another
// repository. The gh environment must drop it the same way.
func TestFactoryGHEnvDropsGHRepo(t *testing.T) {
	t.Setenv("GH_REPO", "org/stray-repo")
	env := factoryGHEnv()
	for _, kv := range env {
		if strings.HasPrefix(kv, "GH_REPO=") {
			t.Errorf("factoryGHEnv carries %q — a stray GH_REPO would retarget the delivery edge's gh calls", kv)
		}
	}
	prompt := false
	for _, kv := range env {
		if kv == "GH_PROMPT_DISABLED=1" {
			prompt = true
		}
	}
	if !prompt {
		t.Error("factoryGHEnv lost GH_PROMPT_DISABLED=1")
	}
}
