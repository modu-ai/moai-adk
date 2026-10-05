package cli

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// SPEC-CODEX-DOCTOR-PATH-GUARD-001 (card t570) — the doctor-side declared-path
// conversion at codexStaleSkillFinding's codexPathAbsolute arm gets a guard
// that FAILS when the conversion is removed.
//
// t562 landed `statPath = fromConfigPath(e.Path, configPathSeparator)` on a
// structural argument and guarded it only under the host separator, where
// fromConfigPath is the identity — so reverting the line to `statPath = e.Path`
// left CI green. Every test below pins configPathSeparator to '\\' via
// overrideSeparator(t), which is what makes the two forms distinguishable on
// darwin.
//
// All three are SERIAL (never t.Parallel): configPathSeparator and osStatFn are
// package-level vars, and both overrides restore through t.Cleanup — the
// separator via overrideSeparator, the stat seam via stubStatRecording.
//
// No new recorder type is declared here: statRecorder already exists once in
// this package (doctor_codex_stale_skill_test.go) and a second declaration
// broke the package test binary during an earlier absorb merge.
// ---------------------------------------------------------------------------

// TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (AC-CDPG-001):
// under a pinned '\\' separator, the argument handed to osStatFn for an
// absolute slash-form declaration is the fromConfigPath-CONVERTED form, and is
// not the declared form. The fixture is pinned rather than "some absolute
// path" so the two operands cannot collapse onto each other: the declared form
// contains '/' and the expected form contains none, so the inequality half of
// the assertion is non-vacuous by inspection.
func TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm(t *testing.T) {
	const declared = "/Users/u/skills/probe/SKILL.md"
	const wantStat = `\Users\u\skills\probe\SKILL.md`

	overrideSeparator(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: declared, EnabledKey: "true"}})
	stubCodexHome(t, home)
	rec := stubStatRecording(t)

	if _, ok := codexStaleSkillFinding(); !ok {
		t.Fatalf("absolute declaration produced no finding")
	}
	if len(rec.paths) != 1 {
		t.Fatalf("recorded stat calls = %v, want exactly one", rec.paths)
	}
	if rec.paths[0] != wantStat {
		t.Errorf("stat target = %q, want the converted form %q", rec.paths[0], wantStat)
	}
	if rec.paths[0] == declared {
		t.Errorf("stat target = %q — the DECLARED form: the conversion is not applied", rec.paths[0])
	}
}

// TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm
// (AC-CDPG-002): the extended-length family is the one family on which the
// unconverted form is not merely untidy. Windows accepts '/' inside ordinary
// absolute paths, so every other absolute family degrades gracefully; inside a
// `\\?\` prefix Windows performs no separator normalization, so the declared
// slash form does not resolve there and a healthy registration would be
// reported stale.
//
// The second assertion is mandatory, not decorative: the first is meaningful
// only if the entry reaches the converting arm at all, which depends on
// classifyCodexSkillPath returning codexPathAbsolute for this declaration. It
// is asserted here rather than inherited from the SPEC, so a later change to
// the classifier turns this guard RED instead of silently making it vacuous.
func TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm(t *testing.T) {
	const declared = "//?/C:/Users/u/skills/probe/SKILL.md"
	const wantStat = `\\?\C:\Users\u\skills\probe\SKILL.md`

	if got := classifyCodexSkillPath(declared); got != codexPathAbsolute {
		t.Fatalf("classifyCodexSkillPath(%q) = %v, want codexPathAbsolute — "+
			"the reachability premise of this guard no longer holds", declared, got)
	}

	overrideSeparator(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: declared, EnabledKey: "true"}})
	stubCodexHome(t, home)
	rec := stubStatRecording(t)

	if _, ok := codexStaleSkillFinding(); !ok {
		t.Fatalf("extended-length declaration produced no finding")
	}
	if len(rec.paths) != 1 {
		t.Fatalf("recorded stat calls = %v, want exactly one", rec.paths)
	}
	if rec.paths[0] != wantStat {
		t.Errorf("stat target = %q, want the converted extended-length form %q", rec.paths[0], wantStat)
	}
}

// TestCodexStaleSkillFinding_ClassificationPrecedesConversion (AC-CDPG-003):
// classification reads the DECLARED string, never a converted one.
//
// The discriminator is the finding Detail string. On darwin `C:/Users/u/SKILL.md`
// has IsAbs false and carries no backslash, so the correct classify-first order
// renders it "relative". The forbidden convert-first order would rewrite it to
// `C:\Users\u\SKILL.md` first, which contains a backslash and therefore renders
// "oddly-formed" instead.
//
// [HARD] The zero-stat-call assertion below is explicitly NON-discriminating and
// is retained only as a supporting check: a zero-stat-call assertion does NOT
// discriminate the two classification orders. Both the codexPathRelative and the
// codexPathOddlyFormed branches `continue` BEFORE reaching osStatFn, so the
// recorder counts 0 under the correct order and under the forbidden one alike.
func TestCodexStaleSkillFinding_ClassificationPrecedesConversion(t *testing.T) {
	const declared = "C:/Users/u/SKILL.md"

	overrideSeparator(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: declared, EnabledKey: "true"}})
	stubCodexHome(t, home)
	rec := stubStatRecording(t)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("unresolvable-shape declaration produced no finding")
	}
	if !strings.Contains(f.detail, "relative entr") {
		t.Errorf("detail does not report the entry as relative:\n%s", f.detail)
	}
	if strings.Contains(f.detail, "oddly-formed") {
		t.Errorf("detail reports the entry as oddly-formed — the declared string was "+
			"converted BEFORE classification:\n%s", f.detail)
	}
	// Supporting, non-discriminating (see the header comment).
	if len(rec.paths) != 0 {
		t.Errorf("recorded stat calls = %v, want none", rec.paths)
	}
}
