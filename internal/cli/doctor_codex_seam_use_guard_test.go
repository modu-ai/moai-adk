package cli

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Card t581 — the seam-USE guard, pinned under a '/' separator.
//
// The three t570 guards in doctor_codex_path_guard_test.go all pin
// configPathSeparator to '\\' and compare the stat target against a backslash
// literal. Under that pin, the seam call
//
//	fromConfigPath(e.Path, configPathSeparator)
//
// and a hand-rolled bypass
//
//	strings.ReplaceAll(e.Path, "/", "\\")
//
// produce BYTE-IDENTICAL output, so all three pass on the bypass. Those guards
// pin the converted VALUE; what needs pinning is that the seam was USED.
//
// A '/'-pinned run separates them, because the seam is the identity there
// while the bypass still converts. The two tests below are deliberately split:
//
//   - the reachability test asserts only that the absolute arm is REACHED
//     under a '/' pin — it must stay green under the bypass, since the bypass
//     changes the value the arm computes, not whether the arm runs;
//   - the identity test asserts the VALUE, and is the one the bypass breaks.
//
// Kept apart so a failure of the second cannot be read as "the arm was never
// reached". A guard that fails because its code path is unreachable pins
// nothing, and this repository has met that shape before.
//
// SERIAL (never t.Parallel): configPathSeparator and osStatFn are package-level
// vars; both overrides restore through t.Cleanup.
// ---------------------------------------------------------------------------

// overrideSeparatorSlash pins the config-path separator seam to '/' — the
// value on which fromConfigPath is the identity — and restores it via
// t.Cleanup. Sibling of overrideSeparator, which pins '\\'.
func overrideSeparatorSlash(t *testing.T) {
	t.Helper()
	orig := configPathSeparator
	t.Cleanup(func() { configPathSeparator = orig })
	configPathSeparator = '/'
}

// TestCodexStaleSkillFinding_SlashPinnedAbsoluteArmIsReached establishes the
// precondition the identity guard below rests on: under a '/' pin, an absolute
// slash-form declaration still classifies as codexPathAbsolute and still
// reaches the stat seam exactly once.
//
// This must hold whether or not the conversion is performed through the seam,
// which is what makes it a usable precondition rather than a restatement of
// the guard.
func TestCodexStaleSkillFinding_SlashPinnedAbsoluteArmIsReached(t *testing.T) {
	const declared = "/Users/u/skills/probe/SKILL.md"

	overrideSeparatorSlash(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: declared, EnabledKey: "true"}})
	stubCodexHome(t, home)
	rec := stubStatRecording(t)

	if _, ok := codexStaleSkillFinding(); !ok {
		t.Fatalf("absolute declaration produced no finding under a '/' pin")
	}
	if len(rec.paths) != 1 {
		t.Fatalf("recorded stat calls = %v, want exactly one — the absolute arm was not reached", rec.paths)
	}
}

// TestCodexStaleSkillFinding_SlashPinnedStatTargetIsUnconverted pins the seam
// USE. Under a '/' pin fromConfigPath returns its input unchanged, so the stat
// target must equal the declared form byte for byte.
//
// A bypass that converts unconditionally — strings.ReplaceAll(e.Path, "/", "\\")
// — hands back the backslash form here and fails, which is exactly what the
// three '\\'-pinned guards cannot see.
func TestCodexStaleSkillFinding_SlashPinnedStatTargetIsUnconverted(t *testing.T) {
	const declared = "/Users/u/skills/probe/SKILL.md"

	overrideSeparatorSlash(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: declared, EnabledKey: "true"}})
	stubCodexHome(t, home)
	rec := stubStatRecording(t)

	if _, ok := codexStaleSkillFinding(); !ok {
		t.Fatalf("absolute declaration produced no finding under a '/' pin")
	}
	if len(rec.paths) != 1 {
		t.Fatalf("recorded stat calls = %v, want exactly one", rec.paths)
	}
	if rec.paths[0] != declared {
		t.Errorf("stat target = %q, want the declared form %q unchanged — "+
			"the conversion did not go through the fromConfigPath seam", rec.paths[0], declared)
	}
}
