package cli

import (
	"strings"
	"testing"
)

// t1710 sync-audit F2 — the GH_REPO drop compared the name exactly, but on
// Windows getenv is case-insensitive and gh is a Go binary: `Gh_Repo=org/x`
// passes an exact compare and still retargets the call there. The delivery
// env's GH_REPO filter follows the gitenv.Scrub precedent — the case rule is
// a parameter, so both rules are testable on any host.
func TestGhDeliveryEnvGHRepoCaseRule(t *testing.T) {
	env := []string{"HOME=/h", "Gh_Repo=org/stray", "PATH=/bin"}

	folded := ghDeliveryEnv(env, true)
	for _, kv := range folded {
		if strings.HasPrefix(strings.ToUpper(kv), "GH_REPO=") {
			t.Errorf("folded rule kept %q — Windows getenv is case-insensitive and gh would read it", kv)
		}
	}

	kept := ghDeliveryEnv(env, false)
	found := false
	for _, kv := range kept {
		if kv == "Gh_Repo=org/stray" {
			found = true
		}
	}
	if !found {
		t.Errorf("exact rule dropped %q — off Windows a differently-cased name is a different variable and is kept (the gitenv scrub precedent)", "Gh_Repo=org/stray")
	}
}
