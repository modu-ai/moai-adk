// Package cli — harness CLI retirement CI guard (SPEC-V3R4-HARNESS-001).
//
// @MX:NOTE: [AUTO] CI regression guard for SPEC-V3R4-HARNESS-001 (REQ-HRN-FND-002).
// This test fails the build if newHarnessCmd (or any equivalent harness subcommand
// factory) is registered into the cobra command tree, enforcing the retirement
// contract declared by BC-V3R4-HARNESS-001-CLI-RETIREMENT.
//
// Why this guard exists:
//
//	SPEC-V3R4-HARNESS-001 retires the `moai harness <verb>` CLI verb path. The
//	implementation file internal/cli/harness.go remains in the tree as a
//	deprecation marker (factory function preserved for compatibility with any
//	internal introspection callers), but the cobra registration is removed.
//	This test verifies the registration absence so a future refactor cannot
//	silently re-introduce the public CLI surface.
package cli

import (
	"strings"
	"testing"
)

// retiredHarnessVerbs는 retired 상태의 harness lifecycle 동사 집합입니다.
// SPEC-V3R4-HARNESS-001 BC-V3R4-HARNESS-001-CLI-RETIREMENT.
// 이 동사들은 rootCmd.harness 하위에 절대 등록되면 안 됩니다.
var retiredHarnessVerbs = map[string]bool{
	"status":   true,
	"apply":    true,
	"rollback": true,
	"disable":  true,
}

// allowedHarnessVerbs는 SPEC-V3R2-HRN-001에서 새로 도입된 허용 동사 집합입니다.
// retirement guard는 이 동사들의 등록을 허용합니다.
var allowedHarnessVerbs = map[string]bool{
	"route":    true,
	"validate": true,
}

// TestHarnessRetirement asserts that the retired harness lifecycle verbs
// (status/apply/rollback/disable) are NOT registered under the harness command.
//
// SPEC-V3R4-HARNESS-001 REQ-HRN-FND-002: The harness lifecycle CLI verb path is
// retired. SPEC-V3R2-HRN-001 re-introduces a DISTINCT harness command with
// routing verbs (route/validate) only. This test enforces:
//
//  1. The "harness" command MAY exist if it only contains allowed routing verbs.
//  2. The retired lifecycle verbs (status/apply/rollback/disable) MUST NOT appear
//     under any "harness" command in rootCmd.
//  3. The retired newHarnessCmd() factory MUST NOT be the registered command
//     (distinguished by presence of retired verbs).
func TestHarnessRetirement(t *testing.T) {
	t.Parallel()

	for _, cmd := range rootCmd.Commands() {
		useFirst := strings.SplitN(cmd.Use, " ", 2)[0]
		if useFirst != "harness" {
			// 다른 커맨드는 확인 불필요
			continue
		}

		// "harness" 커맨드가 있다면 — retired 동사가 없는지 확인합니다.
		// SPEC-V3R2-HRN-001의 새 routing 커맨드 (route/validate)는 허용됩니다.
		for _, subCmd := range cmd.Commands() {
			subVerb := strings.SplitN(subCmd.Use, " ", 2)[0]
			if retiredHarnessVerbs[subVerb] {
				t.Fatalf(
					"SPEC-V3R4-HARNESS-001 / REQ-HRN-FND-002 violation: retired harness verb %q "+
						"is registered under 'harness' command. The lifecycle verbs "+
						"(status/apply/rollback/disable) are retired per BC-V3R4-HARNESS-001-CLI-RETIREMENT. "+
						"Only SPEC-V3R2-HRN-001 routing verbs (route/validate) are permitted. "+
						"Remove the rootCmd.AddCommand(newHarnessCmd()) call from internal/cli/root.go.",
					subVerb,
				)
			}
			if !allowedHarnessVerbs[subVerb] {
				t.Logf(
					"WARNING: unexpected harness sub-verb %q — not in retired set, not in allowed set. "+
						"If this is intentional, add to allowedHarnessVerbs in harness_retirement_test.go.",
					subVerb,
				)
			}
		}

		// Aliases도 확인
		for _, alias := range cmd.Aliases {
			if alias == "harness" && useFirst != "harness" {
				t.Fatalf(
					"SPEC-V3R4-HARNESS-001 / REQ-HRN-FND-002 violation: command %q registers "+
						"'harness' as an alias pointing to a retired command tree. "+
						"Only newHarnessRouterCmd() (route/validate verbs) is permitted to use 'harness'.",
					cmd.Use,
				)
			}
		}
	}
}

// TestHarnessFactoryStillCompiles asserts that the newHarnessCmd factory function
// remains present in the package, even though it is not registered. The file is
// preserved as a deprecation marker per SPEC-V3R4-HARNESS-001 §2.1; physical
// removal is deferred to a follow-up SPEC. If a future refactor deletes the
// factory, this test fails to flag the change as out of scope for the foundation
// SPEC's deprecation-marker contract.
//
// The factory is invoked only for its return-value type assertion; it is not
// added to the command tree.
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
