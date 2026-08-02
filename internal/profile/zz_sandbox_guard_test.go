package profile

// Trailing sandbox guard (SPEC-PROFILE-MEMORY-001 REQ-PM-021/022, AC-PM-021).
//
// WHY THE FILENAME STARTS WITH zz_.
//
// Go runs test files in lexicographic order. TestProfileBaseDirIsSandboxed
// lives in main_test.go, which sorts FIRST — so it observes the package
// sandbox before any test has had a chance to disturb it, and it passes
// unconditionally. It cannot detect a test that clears BaseDirOverride and
// never restores it; every test that runs afterwards is then writing to the
// real ~/.moai/claude-profiles with no signal at all.
//
// TestGetBaseDir_Default (profile_test.go) is exactly such a test: verifying
// the home-derived default REQUIRES clearing the override. It must restore it,
// and this guard is what makes that requirement falsifiable — the file sorts
// after profile_test.go, so it observes the state that test left behind.
//
// DO NOT rename this file to something that sorts earlier. Doing so does not
// break the guard loudly; it makes it vacuous, which is worse.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSandboxSurvivesPackageRun asserts the package sandbox is still intact
// after the tests that mutate BaseDirOverride have run.
//
// Removing the t.Cleanup restore in TestGetBaseDir_Default makes this fail
// deterministically under a full-package run (`go test ./internal/profile/`).
// Note it will NOT fail under a targeted `-run` that excludes the offending
// test — the contamination has to actually happen for the guard to see it,
// which is why AC-PM-021 judges this with a whole-package run.
func TestSandboxSurvivesPackageRun(t *testing.T) {
	if BaseDirOverride == "" {
		t.Fatal("BaseDirOverride was cleared by an earlier test and never restored. " +
			"Every test running after that point writes the developer's real " +
			"~/.moai/claude-profiles. The test that clears it (TestGetBaseDir_Default " +
			"needs the override empty to check the home-derived default) must save " +
			"the original and restore it via t.Cleanup.")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory; sandbox comparison unavailable")
	}
	realBase := filepath.Join(home, ".moai", "claude-profiles")
	if got := GetBaseDir(); got == realBase {
		t.Fatalf("GetBaseDir() = %q after the package run — the sandbox was "+
			"replaced with the real user profile base.", got)
	}
}
