package cli

// codex_skills_path_shape_test.go — the asymmetry guard for
// classifyCodexSkillPath (SPEC-CODEX-HOME-BACKSLASH-001, card t571).
//
// The defect this file exists to keep closed: the classifier stripped the
// "~/" prefix BEFORE testing for a backslash, so a backslash-bearing
// home-relative declaration slipped past the oddly-formed refusal, was
// expanded, was stat'ed, and reached Eligible=true — a DELETION verdict that
// removes the entry's line range from the user's ~/.codex/config.toml. The
// same backslash in a RELATIVE declaration was refused correctly. The
// asymmetry IS the defect, which is why the binding test below asserts that
// the two declaration forms receive the SAME verdict rather than asserting
// one of them in isolation.
//
// [HARD] Nothing in this file opts into parallel execution (REQ-CHB-007).
// Several tests here reassign package-level seams (osStatFn,
// codexUserHomeDir, configPathSeparator), which is shared state; the file is
// new, so forgoing parallelism across all of it costs nothing and makes the
// discipline decidable by a single grep rather than by a human reading each
// function. Every seam override restores through t.Cleanup.
//
// AC-CHB-008 counts the parallel-opt-in token across this whole file, so this
// comment deliberately does NOT spell it out — prose that names the token it
// forbids makes the guard report a hit on a file that obeys it. That self-hit
// has now occurred three times on this card (once in a SPEC gate recipe, once
// in a Definition-of-Done line, once here), which is why the rule is stated
// by description rather than by literal.

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// shapeName maps a codexSkillPathShape to its constant name.
//
// codexSkillPathShape is a bare int with no String() method, so %v prints "1"
// rather than "codexPathHomeRelative". A mutant's failure message naming the
// integer teaches the reader nothing about which cell moved; the acceptance
// criteria (AC-CHB-002, AC-CHB-009) therefore require the constant name in the
// failure output, and this table is what supplies it. It is a test-local
// table, not a production seam and not a new shape constant.
func shapeName(s codexSkillPathShape) string {
	switch s {
	case codexPathAbsolute:
		return "codexPathAbsolute"
	case codexPathHomeRelative:
		return "codexPathHomeRelative"
	case codexPathRelative:
		return "codexPathRelative"
	case codexPathOddlyFormed:
		return "codexPathOddlyFormed"
	default:
		return fmt.Sprintf("codexSkillPathShape(%d)", int(s))
	}
}

// TestCodexSkillPathBackslashSymmetry is the binding discriminant (AC-CHB-001).
//
// Assertion (a) — the two arms agree — is the body of the test. Assertion (b)
// pins what they agree ON. Checking only (b) would pass for the wrong reason
// if a later change made every shape oddly-formed; checking only (a) would
// pass if both arms became home-relative. Both are required.
func TestCodexSkillPathBackslashSymmetry(t *testing.T) {
	const (
		homeArm = `~/x\SKILL.md`
		relArm  = `x\SKILL.md`
	)

	gotHome := classifyCodexSkillPath(homeArm)
	gotRel := classifyCodexSkillPath(relArm)

	t.Run("home_relative_arm_matches_relative_arm", func(t *testing.T) {
		if gotHome != gotRel {
			t.Fatalf("asymmetry: %q classified %s but %q classified %s — the same backslash must receive the same verdict in both declaration forms",
				homeArm, shapeName(gotHome), relArm, shapeName(gotRel))
		}
	})

	t.Run("both_arms_are_oddly_formed", func(t *testing.T) {
		if gotHome != codexPathOddlyFormed {
			t.Fatalf("%q classified %s, want %s", homeArm, shapeName(gotHome), shapeName(codexPathOddlyFormed))
		}
		if gotRel != codexPathOddlyFormed {
			t.Fatalf("%q classified %s, want %s", relArm, shapeName(gotRel), shapeName(codexPathOddlyFormed))
		}
	})
}

// TestCodexSkillPathPreservedShapes pins the cells the reordering must NOT
// move (AC-CHB-003). Only the home-relative-with-backslash cell changes.
//
// The first row is the load-bearing one. "IsAbs stays first" cannot be checked
// on darwin with a Windows-shaped C:\... path — filepath.IsAbs is false for it
// here, so it lands in oddly-formed under every candidate ordering and the
// check would be vacuous. A POSIX absolute carrying a backslash is the
// discriminant that actually works on this host: IsAbs is true, so the cell
// stays absolute — and goes red the moment the backslash check is hoisted
// above IsAbs (AC-CHB-009 observes exactly that).
func TestCodexSkillPathPreservedShapes(t *testing.T) {
	cases := []struct {
		name string
		path string
		want codexSkillPathShape
	}{
		{"posix_absolute_with_backslash", `/tmp/a\b`, codexPathAbsolute},
		{"other_user_home", "~someone/SKILL.md", codexPathOddlyFormed},
		{"plain_relative", "rel/SKILL.md", codexPathRelative},
		{"clean_home_relative", "~/ok/SKILL.md", codexPathHomeRelative},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyCodexSkillPath(tc.path)
			if got != tc.want {
				t.Fatalf("%q classified %s, want %s", tc.path, shapeName(got), shapeName(tc.want))
			}
		})
	}
}

// TestJudgeCodexSkillEntryRefusesHomeRelativeBackslash checks the consumer
// verdict and, just as importantly, that the refusal happens BEFORE stat
// (AC-CHB-004).
//
// Eligible=false alone does not separate "refused before stat" from "stat'ed
// and got a different answer", so the call count carries the second axis. A
// counter wired to nothing also reads zero, which is why the positive control
// sub-case is part of the same test rather than an optional extra: without it
// the zero is an absence, not a measurement.
//
// This test also has an inherited RED-now cell: under AC-CHB-002's old-ordering
// mutant the refusal case classifies home-relative, expands, and stats, so the
// count-0 assertion reddens inside that same window.
func TestJudgeCodexSkillEntryRefusesHomeRelativeBackslash(t *testing.T) {
	// Each sub-case installs its own counter and restores it, so the
	// refusal case's 0 can never be polluted by the control's 1.
	withCountingStat := func(t *testing.T, home string) *[]string {
		t.Helper()

		origHome := codexUserHomeDir
		codexUserHomeDir = func() (string, error) { return home, nil }
		t.Cleanup(func() { codexUserHomeDir = origHome })

		stated := make([]string, 0, 2)
		origStat := osStatFn
		osStatFn = func(p string) (os.FileInfo, error) {
			stated = append(stated, p)
			return nil, fs.ErrNotExist
		}
		t.Cleanup(func() { osStatFn = origStat })

		return &stated
	}

	t.Run("refuses_before_stat", func(t *testing.T) {
		stated := withCountingStat(t, "/Users/example")

		v := judgeCodexSkillEntry(codexwiring.SkillEntry{
			Path:                  `~/x\SKILL.md`,
			FirstUnrecognizedLine: -1,
		})

		if v.Eligible {
			t.Fatalf("Eligible = true for a backslash-bearing home-relative declaration; a deletion verdict must not be reachable here")
		}
		if !strings.Contains(v.SkipReason, "oddly-formed") {
			t.Fatalf("SkipReason = %q, want it to name the oddly-formed refusal", v.SkipReason)
		}
		if len(*stated) != 0 {
			t.Fatalf("osStatFn called %d time(s) with %v, want exactly 0 — the refusal must happen before any stat", len(*stated), *stated)
		}
	})

	t.Run("positive_control_clean_path_does_stat_once", func(t *testing.T) {
		stated := withCountingStat(t, "/Users/example")

		v := judgeCodexSkillEntry(codexwiring.SkillEntry{
			Path:                  "~/ok/SKILL.md",
			FirstUnrecognizedLine: -1,
		})

		if len(*stated) != 1 {
			t.Fatalf("osStatFn called %d time(s) with %v, want exactly 1 — without this the zero above would be an absence rather than a measurement", len(*stated), *stated)
		}
		if !v.Eligible {
			t.Fatalf("clean home-relative declaration was not eligible: %q", v.SkipReason)
		}
	})
}

// TestCodexBackslashRefusalIsSymmetricAcrossReadAndWrite checks that the read
// side now refuses what the write side already refused (AC-CHB-005).
//
// [HARD] The write side is exercised through its PRODUCTION entry point,
// upsertCodexSkillDisable — never by re-implementing its predicate here. A
// test that copies `strings.ContainsAny(skillPath, "\"\\\n\r")` inline would
// assert a copy against a copy and stay green after the production guard was
// deleted, which is the opposite of what this test is for.
func TestCodexBackslashRefusalIsSymmetricAcrossReadAndWrite(t *testing.T) {
	origSep := configPathSeparator
	configPathSeparator = '/'
	t.Cleanup(func() { configPathSeparator = origSep })

	const decl = `~/x\SKILL.md`
	fixture := []byte("[[skills.config]]\npath = \"/Users/example/other/SKILL.md\"\nenabled = true\n")

	out, verdict := upsertCodexSkillDisable(fixture, decl)

	if verdict.Action != codexSkillDisableSkipped {
		t.Fatalf("write side: Action = %v, want codexSkillDisableSkipped", verdict.Action)
	}
	if verdict.Reason == "" {
		t.Fatalf("write side: skipped with an empty Reason — a refusal the user cannot read is not a refusal")
	}
	if string(out) != string(fixture) {
		t.Fatalf("write side: content changed on a skipped declaration")
	}

	if got := classifyCodexSkillPath(decl); got != codexPathOddlyFormed {
		t.Fatalf("read side: %q classified %s, want %s — the two sides must refuse the same declaration", decl, shapeName(got), shapeName(codexPathOddlyFormed))
	}
}

// TestJudgeCodexSkillEntryBackslashHomeStillExpands pins the design §3.2
// REJECTED: the backslash check runs on the DECLARATION, never on the
// expansion (AC-CHB-007). On Windows the expanded home is legitimately
// C:\Users\x, so a check moved after expansion would refuse every entry there.
//
// The discriminating power of this test is entirely the codexUserHomeDir stub,
// and a stat count of 1 is produced identically when the stub is inert (the
// real home also stats once). So the assertion is on the observed stat
// ARGUMENT, not the count: filepath.Join on a '/'-separator host leaves the
// backslash an ordinary byte, so the expansion stays backslash-bearing and the
// argument is the proof the stub took effect.
func TestJudgeCodexSkillEntryBackslashHomeStillExpands(t *testing.T) {
	const backslashHome = `C:\Users\x`
	const want = backslashHome + "/ok/SKILL.md"

	origHome := codexUserHomeDir
	codexUserHomeDir = func() (string, error) { return backslashHome, nil }
	t.Cleanup(func() { codexUserHomeDir = origHome })

	var stated []string
	origStat := osStatFn
	osStatFn = func(p string) (os.FileInfo, error) {
		stated = append(stated, p)
		return nil, fs.ErrNotExist
	}
	t.Cleanup(func() { osStatFn = origStat })

	v := judgeCodexSkillEntry(codexwiring.SkillEntry{
		Path:                  "~/ok/SKILL.md",
		FirstUnrecognizedLine: -1,
	})

	if v.SkipReason != "" {
		t.Fatalf("declaration refused with %q — a backslash in the expanded HOME must not refuse a clean declaration", v.SkipReason)
	}
	if len(stated) != 1 {
		t.Fatalf("osStatFn called %d time(s) with %v, want exactly 1", len(stated), stated)
	}
	if stated[0] != want {
		t.Fatalf("stat argument = %q, want %q — a different argument means the codexUserHomeDir stub never took effect, and this test's whole discriminating power is that stub", stated[0], want)
	}
}
