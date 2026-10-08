package hook

// protected_zone_shell_parsing_test.go — card t1574: RED-first reproduction of
// the shell parsing bypasses in protected_zone_shell.go. Each subtest names
// the defect site and observes the allow-before-fix state; the parsing-matrix
// sweep (function declaration forms, wrapper prefixes, recursion, `--`
// operands) extends from these four seeds.

import (
	"testing"
)

// TestProtectedZoneShellParsingBypasses groups the card t1574 reproductions
// as named subtests of one top-level test: every subtest prints its own
// --- PASS line, so a selector that matches nothing cannot pass for a sweep.
func TestProtectedZoneShellParsingBypasses(t *testing.T) {
	t.Run("dash_dash_operand", testShellBypassDashDashOperand)
	t.Run("conditional_function_declaration", testShellBypassConditionalFuncDecl)
	t.Run("command_env_prefix", testShellBypassCommandEnvPrefix)
	t.Run("recursion_state_reset", testShellBypassRecursionStateReset)
	t.Run("executable_path_function_absorption", testShellBypassExecutablePathAbsorption)
}

// testShellBypassDashDashOperand is the :624 defect (repro material
// .moai/reports/t1574/repro-material.md, lane-9 gate overlay
// TestReviewDashOperand): everything after a `--` separator is an operand,
// but zonePathCandidates keeps skipping hyphen-leading words, so
// `rm -- -zone/secret.md` names zero candidates and deletes a protected
// file whose name legitimately starts with a hyphen.
func testShellBypassDashDashOperand(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"-zone/\"]\n"), "")
	h := zoneTestHandler(t, root)

	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": "rm -- -zone/secret.md"})
	wantZoneDeny(t, "rm -- -zone/secret.md", d, r, harnessLearnerIdentity, "category", "probe_zone")
}

// testShellBypassConditionalFuncDecl is the :432 defect (lane-9 gate round 8):
// a conditional declaration that never ran (`false && rm() { :; }`) is
// absorbed at the branch join as if the name were defined, so the following
// real `rm` walks the never-executed no-op body instead of being judged as a
// mutation — the protected file is deleted for real.
func testShellBypassConditionalFuncDecl(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
	h := zoneTestHandler(t, root)

	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{
		"command": "false && rm() { :; }; rm zone_dir/secret.md",
	})
	wantZoneDeny(t, "false && rm() { :; }; rm zone_dir/secret.md", d, r, harnessLearnerIdentity, "category", "probe_zone")
}

// testShellBypassCommandEnvPrefix is the :598 defect (lane-9 gate round 8):
// a static `command` or `env` wrapper prefix strips to the same verb the CC
// permission layer already strips, but zoneCall matches the raw head word —
// `command rm zone_dir/a.log` under-matches and the deletion goes through.
func testShellBypassCommandEnvPrefix(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
	h := zoneTestHandler(t, root)

	for _, cmd := range []string{
		"command rm zone_dir/a.log",
		"env rm zone_dir/a.log",
		"command -p rm zone_dir/a.log",
	} {
		d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
		wantZoneDeny(t, cmd, d, r, harnessLearnerIdentity, "category", "probe_zone")
	}
}

// testShellBypassRecursionStateReset is the :461/:468 recursion claim (gate
// rounds 7/9, undemonstrated before this card — this subtest is the
// demonstration attempt): an inner call's frame exit deletes calling[name],
// so the next call from an outer frame resets the depth to 1 and a chain
// whose real bash terminates can walk unbounded in the Go walker. Pre-fix,
// this subtest is expected NOT to terminate cleanly (a stack overflow kills
// the test binary — that hard failure IS the RED observation); post-fix the
// walk stays bounded and answers fail-closed (unbounded → deny, or allow
// with the recursion confined), and the command terminates.
func testShellBypassRecursionStateReset(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
	h := zoneTestHandler(t, root)

	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{
		"command": "f() { if false; then f; f; fi; }; f; rm zone_dir/a.log",
	})
	// Either answer is acceptable post-fix — the claim under test is
	// termination, not the verdict: unbounded → fail-closed deny, bounded
	// walk → the rm is judged on its own merits.
	if d == "" && r == "" {
		t.Log("recursion confined; rm judged normally (allowed)")
	}
}

// testShellBypassExecutablePathAbsorption is the :430 defect (plan-phase probe
// EV-3, independently listed by the merge-gate ledger round 1): the path fold
// (path.Base) runs BEFORE the funcs lookup, so `/bin/rm` resolves to the
// in-command `rm` function and the guard walks the no-op body. Real bash
// dispatches functions by bare name only — the external binary runs and the
// protected file is deleted for real.
func testShellBypassExecutablePathAbsorption(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc("  probe_zone:\n    paths: [\"zone_dir/\"]\n"), "")
	h := zoneTestHandler(t, root)

	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{
		"command": "rm() { :; }; /bin/rm zone_dir/secret.md",
	})
	wantZoneDeny(t, "rm() { :; }; /bin/rm zone_dir/secret.md", d, r, harnessLearnerIdentity, "category", "probe_zone")
}
