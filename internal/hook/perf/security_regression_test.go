package perf

import (
	"strings"
	"testing"
)

// BASH_SUBCOMMAND_SOFT_CAP is the compound-subcommand count above which a
// Bash command is considered too complex for any fast-path short-circuit
// (per the Bash Risk-Amplifier Doctrine, .claude/rules/moai/development/
// coding-standards.md § Bash Risk-Amplifier Doctrine (3)). Any command
// whose compound-subcommand count (|, &&, ||, ;, backtick, $(...)) exceeds
// this cap MUST route through the full config + security scan path.
const BASH_SUBCOMMAND_SOFT_CAP = 5

// destructivePrimitives is the canonical set of Bash destructive primitives
// that MUST NEVER be eligible for a fast-path short-circuit, per the Bash
// Risk-Amplifier Doctrine (coding-standards.md § Bash Risk-Amplifier
// Doctrine (3)). REQ-PERF-007 binds any future fast-path to this set.
//
// Each entry is a representative command text containing the primitive.
var destructivePrimitives = []string{
	"rm -rf /",
	"rm -rf ~",
	"rm -rf .git",
	"rm -rf node_modules",
	"git push --force origin main",
	"git push -f origin main",
	"git push --no-verify origin main",
	"git commit --no-verify -m 'bypass'",
	"git reset --hard origin/main",
	"echo 'DROP TABLE users;' | psql",
	"echo 'TRUNCATE TABLE logs;' | psql",
	"chmod -R 777 /",
	"chmod -R 777 .",
}

// compoundSeparators counts the number of compound-subcommand separators
// in a Bash command string: pipe |, &&, ||, semicolon ;, backtick `, and
// command substitution $(...). A count exceeding BASH_SUBCOMMAND_SOFT_CAP
// means the command is too complex for any fast-path.
func compoundSeparators(command string) int {
	count := 0
	count += strings.Count(command, "|") // covers |, ||, && uses different char
	count += strings.Count(command, "&&")
	count += strings.Count(command, ";")
	count += strings.Count(command, "`")
	count += strings.Count(command, "$(")
	// Note: || is counted by | (2 occurrences) + the || itself; we use a
	// simple over-counting approach that is conservative (over-counts rather
	// than under-counts), which is safe for a "must not fast-path" predicate.
	return count
}

// IsDestructiveOrComplex reports whether a Bash command matches any
// destructive primitive in the Bash Risk-Amplifier Doctrine set, OR has a
// compound-subcommand count exceeding BASH_SUBCOMMAND_SOFT_CAP.
//
// This function encodes the constraint that REQ-PERF-007 / AC-PERF-007 binds
// to any future fast-path: a fast-path MUST call this function and MUST NOT
// short-circuit any command for which it returns true.
func IsDestructiveOrComplex(command string) bool {
	// Check compound-subcommand overflow first (cheapest).
	if compoundSeparators(command) > BASH_SUBCOMMAND_SOFT_CAP {
		return true
	}

	lower := strings.ToLower(command)

	// Destructive primitive set (case-insensitive substring match).
	destructiveSubstrings := []string{
		"rm -rf",
		"git push --force",
		"git push -f",
		"git push --no-verify",
		"git commit --no-verify",
		"git reset --hard",
		"drop table",
		"truncate table",
		"truncate",
		"chmod -r 777",
	}

	for _, substr := range destructiveSubstrings {
		if strings.Contains(lower, substr) {
			return true
		}
	}

	return false
}

// FastPathEligible reports whether a Bash command is eligible for a
// hypothetical fast-path short-circuit. Currently always returns false
// because NO fast-path is implemented in SPEC-HOOK-PRETOOL-PERF-001.
//
// When a future SPEC introduces a fast-path, it MUST implement this function
// to return true ONLY for commands where IsDestructiveOrComplex returns
// false. The AC-PERF-007 regression test below guards this contract.
func FastPathEligible(command string) bool {
	// No fast-path implemented — every command routes through the full
	// config + security scan path. This stub exists as the integration
	// point that a future fast-path MUST populate correctly.
	return false
}

// TestAC_PERF_007_DestructivePrimitivesNotFastPath (AC-PERF-007, SECURITY,
// make-or-break) enumerates EVERY destructive primitive in the Bash
// Risk-Amplifier Doctrine set as a SEPARATE test case and asserts that a
// hypothetical fast-path does NOT fire for any of them.
//
// This test is a FORWARD GUARD: no fast-path is implemented in this SPEC,
// but the test lands now so any future fast-path inherits the constraint.
// A future fast-path that lets any destructive primitive through will fail
// this test.
func TestAC_PERF_007_DestructivePrimitivesNotFastPath(t *testing.T) {
	t.Parallel()

	for _, cmd := range destructivePrimitives {
		cmd := cmd // capture for closure
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()

			// The destructive primitive MUST be detected.
			if !IsDestructiveOrComplex(cmd) {
				t.Fatalf("IsDestructiveOrComplex(%q) = false; want true (destructive primitive must be detected)", cmd)
			}

			// A hypothetical fast-path MUST NOT short-circuit this command.
			if FastPathEligible(cmd) {
				t.Fatalf("FastPathEligible(%q) = true; want false (fast-path must NOT fire for destructive primitive)", cmd)
			}
		})
	}
}

// TestAC_PERF_007_CompoundSubcommandOverflowNotFastPath (AC-PERF-007)
// verifies that Bash commands with a compound-subcommand count exceeding
// BASH_SUBCOMMAND_SOFT_CAP (5) are NOT eligible for a fast-path, even if
// no individual subcommand is destructive.
func TestAC_PERF_007_CompoundSubcommandOverflowNotFastPath(t *testing.T) {
	t.Parallel()

	// A command with 6 compound separators (exceeds the cap of 5).
	complexCmd := "echo a && echo b && echo c && echo d && echo e && echo f && echo g"

	if compoundSeparators(complexCmd) <= BASH_SUBCOMMAND_SOFT_CAP {
		t.Fatalf("test command should exceed BASH_SUBCOMMAND_SOFT_CAP; got %d separators",
			compoundSeparators(complexCmd))
	}

	if !IsDestructiveOrComplex(complexCmd) {
		t.Fatal("IsDestructiveOrComplex should return true for compound-subcommand overflow")
	}

	if FastPathEligible(complexCmd) {
		t.Fatal("FastPathEligible must NOT fire for compound-subcommand overflow")
	}
}

// TestAC_PERF_007_SafeCommandEligible verifies that a known-safe command
// (no destructive primitive, no compound overflow) IS classified as
// non-destructive. This proves the detection function works bidirectionally
// — it does not over-block safe commands.
//
// NOTE: FastPathEligible still returns false because no fast-path is
// implemented. This test only verifies IsDestructiveOrComplex returns false
// for safe commands.
func TestAC_PERF_007_SafeCommandNotDestructive(t *testing.T) {
	t.Parallel()

	safeCmds := []string{
		"echo hello",
		"ls -la",
		"go test ./internal/config/...",
		"git status",
		"cat README.md",
	}

	for _, cmd := range safeCmds {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			if IsDestructiveOrComplex(cmd) {
				t.Fatalf("IsDestructiveOrComplex(%q) = true; want false (safe command must not be flagged)", cmd)
			}
		})
	}
}
