package cli

// codex_skills_dotdot_test.go — the ".." segment guard for
// classifyCodexSkillPath and upsertCodexSkillDisable (card t582).
//
// The defect this file exists to keep closed: "~/../../etc/x" carries no
// backslash, so it passed the t571 refusal, was classified home-relative, and
// expanded through filepath.Join — whose internal Clean resolved it to /etc/x,
// OUTSIDE the user's home. That expansion was stat'ed, and an absent target
// reached Eligible=true in the prune verb: a deletion verdict on a
// registration the user wrote by hand.
//
// The chosen repair refuses ANY ".." segment in a non-absolute declaration,
// not only the ones whose cleaned form leaves home. One predicate is shared by
// the read side and the write side, so the two can never disagree about the
// same declaration at a boundary input.
//
// [HARD] Nothing in this file opts into parallel execution: the tests
// reassign package-level seams (osStatFn, codexUserHomeDir,
// configPathSeparator), and every override restores through t.Cleanup.

import (
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// TestCodexSkillPathDotDotSegmentRefused pins every cell the ".." guard moves
// and the neighbouring cells it must not move.
func TestCodexSkillPathDotDotSegmentRefused(t *testing.T) {
	cases := []struct {
		name string
		path string
		want codexSkillPathShape
	}{
		// The defect arms: the cleaned expansion leaves the user's home.
		{"home_escape_to_etc", "~/../../etc/x", codexPathOddlyFormed},
		{"home_escape_to_parent", "~/..", codexPathOddlyFormed},
		// INTENDED CHANGE, not a regression: this declaration stays inside home
		// once cleaned, but refusing every ".." segment keeps the read and write
		// predicates identical, and only a hand-written declaration can carry one.
		{"home_inner_dotdot_intended_change", "~/a/../b/SKILL.md", codexPathOddlyFormed},
		// INTENDED CHANGE, not a regression: a relative declaration was already
		// never stat'ed, and the guard sits where the t571 backslash check sits —
		// ahead of every non-absolute branch — so it now reports as oddly-formed.
		{"relative_dotdot_intended_change", "rel/../x/SKILL.md", codexPathOddlyFormed},

		// Controls the guard must not move.
		{"clean_home_relative", "~/ok/SKILL.md", codexPathHomeRelative},
		{"dotdot_prefixed_name_is_not_a_segment", "~/..ok/SKILL.md", codexPathHomeRelative},
		{"plain_relative", "rel/SKILL.md", codexPathRelative},
		{"absolute", "/etc/x", codexPathAbsolute},
		// IsAbs stays FIRST, exactly as it does for a backslash-bearing
		// absolute: an absolute declaration means what it literally says.
		{"absolute_with_dotdot", "/tmp/a/../b", codexPathAbsolute},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyCodexSkillPath(tc.path); got != tc.want {
				t.Fatalf("%q classified %s, want %s", tc.path, shapeName(got), shapeName(tc.want))
			}
		})
	}
}

// TestJudgeCodexSkillEntryRefusesDotDotEscapeBeforeStat checks the deletion
// verdict and that the refusal happens BEFORE stat. The positive control is
// part of the same test: without a stat it can see, the refusal case's zero
// would be an absence rather than a measurement.
func TestJudgeCodexSkillEntryRefusesDotDotEscapeBeforeStat(t *testing.T) {
	withCountingStat := func(t *testing.T) *[]string {
		t.Helper()

		origHome := codexUserHomeDir
		codexUserHomeDir = func() (string, error) { return "/Users/example", nil }
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
		stated := withCountingStat(t)

		v := judgeCodexSkillEntry(codexwiring.SkillEntry{
			Path:                  "~/../../etc/x",
			FirstUnrecognizedLine: -1,
		})

		if v.Eligible {
			t.Fatalf("Eligible = true for a home-escaping declaration (stat saw %v); a deletion verdict must not be reachable here", *stated)
		}
		if !strings.Contains(v.SkipReason, "oddly-formed") {
			t.Fatalf("SkipReason = %q, want it to name the oddly-formed refusal", v.SkipReason)
		}
		if len(*stated) != 0 {
			t.Fatalf("osStatFn called %d time(s) with %v, want exactly 0 — the refusal must happen before any stat", len(*stated), *stated)
		}
	})

	t.Run("positive_control_clean_path_does_stat_once", func(t *testing.T) {
		stated := withCountingStat(t)

		v := judgeCodexSkillEntry(codexwiring.SkillEntry{
			Path:                  "~/ok/SKILL.md",
			FirstUnrecognizedLine: -1,
		})

		if len(*stated) != 1 {
			t.Fatalf("osStatFn called %d time(s) with %v, want exactly 1", len(*stated), *stated)
		}
		if !v.Eligible {
			t.Fatalf("clean home-relative declaration was not eligible: %q", v.SkipReason)
		}
	})
}

// TestCodexDotDotRefusalIsSymmetricAcrossReadAndWrite checks that both sides
// refuse the same declaration. The write side is exercised through its
// PRODUCTION entry point, never by re-implementing its predicate here.
func TestCodexDotDotRefusalIsSymmetricAcrossReadAndWrite(t *testing.T) {
	origSep := configPathSeparator
	configPathSeparator = '/'
	t.Cleanup(func() { configPathSeparator = origSep })

	const decl = "~/../../etc/x"
	fixture := []byte("[[skills.config]]\npath = \"/Users/example/other/SKILL.md\"\nenabled = true\n")

	out, verdict := upsertCodexSkillDisable(fixture, decl)

	if verdict.Action != codexSkillDisableSkipped {
		t.Fatalf("write side: Action = %v, want codexSkillDisableSkipped", verdict.Action)
	}
	if !strings.Contains(verdict.Reason, `".."`) {
		t.Fatalf("write side: Reason = %q, want it to name the \"..\" refusal", verdict.Reason)
	}
	if string(out) != string(fixture) {
		t.Fatalf("write side: content changed on a skipped declaration")
	}

	if got := classifyCodexSkillPath(decl); got != codexPathOddlyFormed {
		t.Fatalf("read side: %q classified %s, want %s — the two sides must refuse the same declaration", decl, shapeName(got), shapeName(codexPathOddlyFormed))
	}
}
