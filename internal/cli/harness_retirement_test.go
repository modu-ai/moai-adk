// Package cli — harness CLI surface CI guard.
//
// HISTORY:
//   - V3R4 (SPEC-V3R4-HARNESS-001 REQ-HRN-FND-002): retired the lifecycle verbs
//     (status/apply/rollback/disable) from the `moai harness` CLI tree. This test
//     historically asserted that none of those verbs appeared under any harness
//     command, allowing only the SPEC-V3R2-HRN-001 routing verbs (route/validate).
//   - V3R5 (SPEC-V3R5-HARNESS-AUTONOMY-001 §6.4 + AC-HRA-009): supersedes the V3R4
//     retirement. The lifecycle verbs are un-retired and MUST be registered under
//     `moai harness`, alongside the new proposal-management verbs (mute/mute-list/
//     unmute/verify). The unified Cobra tree must satisfy:
//
//         ./moai harness --help | grep -E '(status|apply|rollback|disable|mute|verify)'
//
//     yielding at least 6 matches.
//
// @MX:NOTE: [AUTO] V3R5 supersedence — this guard now asserts that all 10 V3R5 verbs
// are registered (route + validate + 4 lifecycle + 4 proposal-management) under the
// single `moai harness` parent command.
package cli

import (
	"strings"
	"testing"
)

// v3r5RequiredHarnessVerbs is the set of 10 verbs that must be registered under
// the `moai harness` tree, per SPEC-V3R5-HARNESS-AUTONOMY-001 §6.4 and AC-HRA-009.
var v3r5RequiredHarnessVerbs = map[string]bool{
	// SPEC-V3R2-HRN-001 routing verbs.
	"route":    true,
	"validate": true,
	// SPEC-V3R5-HARNESS-AUTONOMY-001 §6 lifecycle verbs (un-retired).
	"status":   true,
	"apply":    true,
	"rollback": true,
	"disable":  true,
	// SPEC-V3R5-HARNESS-AUTONOMY-001 §6 M4 proposal-management verbs (new).
	"mute":      true,
	"mute-list": true,
	"unmute":    true,
	"verify":    true,
}

// TestHarnessV3R5VerbSurface asserts that the `moai harness` command tree exposes
// every required V3R5 verb. AC-HRA-009 (acceptance.md L286-L320) mandates that:
//
//	./moai harness --help | grep -E '(status|apply|rollback|disable|mute|verify)'
//
// must produce at least 6 matches. This test enforces the upstream Cobra registration
// that makes the grep contract reachable.
//
// SPEC-V3R5-HARNESS-AUTONOMY-001 plan.md §6.4 explicitly mandates all 8 lifecycle/
// proposal verbs alongside the V3R2-HRN-001 routing verbs (route/validate).
func TestHarnessV3R5VerbSurface(t *testing.T) {
	t.Parallel()

	// Locate the harness command in rootCmd.
	var found bool
	for _, cmd := range rootCmd.Commands() {
		useFirst := strings.SplitN(cmd.Use, " ", 2)[0]
		if useFirst != "harness" {
			continue
		}
		found = true

		// Collect all registered subcommand verbs (first token of Use).
		registered := make(map[string]bool)
		for _, sub := range cmd.Commands() {
			subVerb := strings.SplitN(sub.Use, " ", 2)[0]
			registered[subVerb] = true
		}

		// Assert every required V3R5 verb is registered.
		for verb := range v3r5RequiredHarnessVerbs {
			if !registered[verb] {
				t.Errorf(
					"SPEC-V3R5-HARNESS-AUTONOMY-001 §6.4 / AC-HRA-009 violation: required harness verb %q "+
						"is NOT registered under the `harness` command. The V3R5 unified Cobra tree must expose "+
						"route + validate + 4 lifecycle verbs (status/apply/rollback/disable) + 4 proposal-management "+
						"verbs (mute/mute-list/unmute/verify). Verify newHarnessRouterCmd() in harness_route.go.",
					verb,
				)
			}
		}

		// Warn (non-fatal) for any unexpected registered verb — encourages explicit
		// acceptance via the v3r5RequiredHarnessVerbs set.
		for verb := range registered {
			if !v3r5RequiredHarnessVerbs[verb] {
				t.Logf(
					"NOTE: unexpected harness sub-verb %q registered — not in V3R5 required set. "+
						"If this is intentional, add it to v3r5RequiredHarnessVerbs in harness_retirement_test.go.",
					verb,
				)
			}
		}
	}

	if !found {
		t.Fatal(
			"`harness` command not registered in rootCmd. V3R5 mandates newHarnessRouterCmd() " +
				"to be added via rootCmd.AddCommand() in internal/cli/root.go.",
		)
	}
}

// TestHarnessFactoryStillCompiles asserts that the legacy newHarnessCmd factory remains
// present and constructible. Preserved per SPEC-V3R4-HARNESS-001 §2.1's deprecation-marker
// contract — the factory MUST continue to compile even though it is no longer the
// surface registered into rootCmd (V3R5 uses newHarnessRouterCmd() instead).
//
// The factory is invoked only for its return-value type assertion; it is not added to
// the command tree.
func TestHarnessFactoryStillCompiles(t *testing.T) {
	t.Parallel()

	cmd := newHarnessCmd()
	if cmd == nil {
		t.Fatal("newHarnessCmd returned nil; the factory must remain operational as a deprecation marker per SPEC-V3R4-HARNESS-001 §2.1")
	}
	if cmd.Use == "" {
		t.Error("newHarnessCmd returned a command with empty Use; expected the historical \"harness\" identifier preserved")
	}
}
