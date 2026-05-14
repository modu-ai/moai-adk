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

// TestHarnessRetirement asserts that rootCmd does not contain any command with
// Use: "harness" (or any alias resolving to "harness"). Per SPEC-V3R4-HARNESS-001
// REQ-HRN-FND-002, the harness CLI verb path is retired; invoking
// `moai harness <verb>` MUST produce cobra's standard `unknown command` error.
func TestHarnessRetirement(t *testing.T) {
	t.Parallel()

	for _, cmd := range rootCmd.Commands() {
		// Direct match on the Use field — cobra's primary registration key.
		// Strip any argument suffix (e.g., "harness <verb>") to compare the verb noun.
		useFirst := strings.SplitN(cmd.Use, " ", 2)[0]
		if useFirst == "harness" {
			t.Fatalf(
				"SPEC-V3R4-HARNESS-001 / REQ-HRN-FND-002 violation: command with Use=%q "+
					"is registered as a subcommand of rootCmd. The harness CLI verb path is "+
					"retired per BC-V3R4-HARNESS-001-CLI-RETIREMENT. Remove the "+
					"rootCmd.AddCommand(newHarnessCmd()) call (or equivalent) from "+
					"internal/cli/root.go. The harness lifecycle is owned by the /moai:harness "+
					"slash command surface and the moai skill workflow body. See "+
					".claude/skills/moai/workflows/harness.md for the slash-command-only "+
					"implementation contract.",
				cmd.Use,
			)
		}

		// Defensive: any alias also forbidden.
		for _, alias := range cmd.Aliases {
			if alias == "harness" {
				t.Fatalf(
					"SPEC-V3R4-HARNESS-001 / REQ-HRN-FND-002 violation: command %q registers "+
						"%q as an alias. The harness verb name is retired as a public CLI surface.",
					cmd.Use, alias,
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
