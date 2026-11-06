package profile

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card t1702 audit iter6 B1: config.LauncherProfileDirName mirrors this
// package's isValidProfileName for every NAMED profile (config cannot import
// profile). The two must not drift, or the read-only reach judgment would stop
// noticing a profile the launcher can start.
func TestLauncherProfileDirName_MatchesTheLauncherRule(t *testing.T) {
	names := []string{
		"work", "a-b", "has space", "tab\tname", "__no_such_profile__", "UPPER",
		".hidden", ".", "..", "a/b", `a\b`, "/abs", "../x", "x.y", "ünï",
	}
	for _, n := range names {
		if got, want := config.LauncherProfileDirName(n), isValidProfileName(n); got != want {
			t.Errorf("name %q: config.LauncherProfileDirName = %v, isValidProfileName = %v", n, got, want)
		}
	}
	// "" is valid for the launcher (the base store) but is never a directory
	// entry name, so config rejects it by design.
	if config.LauncherProfileDirName("") {
		t.Error(`an empty name is not a store entry`)
	}
}
