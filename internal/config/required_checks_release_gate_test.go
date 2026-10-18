package config

import "testing"

// TestRequiredChecks_ReleaseGateSubsetOfMain verifies that the real SSoT carries a
// non-empty release_gate list and that every entry is also a required context on
// main. scripts/release.sh reads release_gate instead of naming checks itself, so
// an empty list would make the release CI gate vacuous and an entry outside the
// main set would demand a check that branch protection never requires.
func TestRequiredChecks_ReleaseGateSubsetOfMain(t *testing.T) {
	rc, err := LoadRequiredChecks(findProjectRoot(t))
	if err != nil {
		t.Fatalf("LoadRequiredChecks() unexpected error: %v", err)
	}

	if len(rc.ReleaseGate) == 0 {
		t.Fatal("release_gate must be a non-empty list in " + RequiredChecksFile)
	}

	main, ok := rc.Branches["main"]
	if !ok {
		t.Fatalf("expected branch 'main' in RequiredChecks, branches: %v", rc.Branches)
	}
	inMain := make(map[string]bool, len(main.Contexts))
	for _, c := range main.Contexts {
		inMain[c] = true
	}
	for _, name := range rc.ReleaseGate {
		if !inMain[name] {
			t.Errorf("release_gate entry %q is not in branches.main.contexts: %v", name, main.Contexts)
		}
	}
}
