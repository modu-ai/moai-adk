package config

import "testing"

// TestRequiredChecks_ReleaseGateSubsetOfMain verifies that the real SSoT carries a
// release_gate list of exactly the three pinned names (no duplicates) and that
// every entry is also a required context on main. This checks the repository
// configuration contract; it does not prove a release-script consumer executes
// that contract. A missing or shortened list must fail this test.
func TestRequiredChecks_ReleaseGateSubsetOfMain(t *testing.T) {
	rc, err := LoadRequiredChecks(findProjectRoot(t))
	if err != nil {
		t.Fatalf("LoadRequiredChecks() unexpected error: %v", err)
	}

	if len(rc.ReleaseGate) == 0 {
		t.Fatal("release_gate must be a non-empty list in " + RequiredChecksFile)
	}

	// CI-enforced floor. A list shortened to "- Lint" would allow a consumer
	// to demand only one check. This pin stops contract drift: a PR that
	// shortens, reorders-with-duplicates or renames the list turns this required
	// Test job red before the change can land. Changing the demanded set is a
	// deliberate two-place edit (the SSoT and this list).
	want := []string{"Lint", "Test (ubuntu-latest)", "Build (linux/amd64)"}
	seen := make(map[string]bool, len(rc.ReleaseGate))
	for _, name := range rc.ReleaseGate {
		if seen[name] {
			t.Errorf("release_gate lists %q more than once: %v", name, rc.ReleaseGate)
		}
		seen[name] = true
	}
	for _, name := range want {
		if !seen[name] {
			t.Errorf("release_gate lacks the pinned check %q: %v", name, rc.ReleaseGate)
		}
	}
	if len(seen) != len(want) {
		t.Errorf("release_gate must hold exactly %v, got %v", want, rc.ReleaseGate)
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
