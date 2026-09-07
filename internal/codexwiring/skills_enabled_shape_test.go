package codexwiring

// t508 (SPEC-CODEX-ENABLED-FATAL-001) M2 — the `enabled` reading distinguishes
// THREE cases that were previously collapsed onto two.
//
// Before this SPEC, SkillEnabledUnspecified's own doc comment read "an entry
// declaring no `enabled` key, OR one whose value this parser does not
// recognise" — one state for two observations. The parser therefore could not
// tell "nothing was said" from "something was said that codex rejects", which
// is why `enabled = 1` was exactly as silent as an absent key.
//
// Measured on codex-cli 0.153.4 in an isolated CODEX_HOME (probe:
// `codex mcp list`, a config-loading verb — `codex --version` does not load the
// config and is not a usable probe):
//
//	enabled = true      rc=0
//	(line absent)       rc=1  missing field `enabled` in `skills.config`
//	enabled = 1         rc=1  invalid type: integer `1`, expected a boolean
//	enabled = "true"    rc=1  invalid type: string "true", expected a boolean
//	enabled = 'false'   rc=1  invalid type: string "false", expected a boolean
//
// Codex reports the absent key and the wrong-typed value with DIFFERENT errors,
// which is the ground for keeping them separate states rather than widening
// Unspecified.

import "testing"

// TestParseSkillEntriesEnabledThreeWayReading (AC-CEF-010, parser half) pins
// all three readings in one table so no row can be changed without the others
// being looked at.
func TestParseSkillEntriesEnabledThreeWayReading(t *testing.T) {
	cases := []struct {
		name string
		line string // the `enabled` line, or "" for an absent key
		want SkillEnabled
	}{
		{"absent", "", SkillEnabledUnspecified},
		{"bare_true", "enabled = true\n", SkillEnabledTrue},
		{"bare_false", "enabled = false\n", SkillEnabledFalse},
		{"bare_true_no_spaces", "enabled=true\n", SkillEnabledTrue},
		{"trailing_comment", "enabled = false # off\n", SkillEnabledFalse},
		{"integer", "enabled = 1\n", SkillEnabledNonBoolean},
		{"double_quoted_true", "enabled = \"true\"\n", SkillEnabledNonBoolean},
		{"double_quoted_false", "enabled = \"false\"\n", SkillEnabledNonBoolean},
		{"single_quoted_true", "enabled = 'true'\n", SkillEnabledNonBoolean},
		{"single_quoted_false", "enabled = 'false'\n", SkillEnabledNonBoolean},
		{"bareword_yes", "enabled = yes\n", SkillEnabledNonBoolean},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseSkillEntries([]byte("[[skills.config]]\npath = \"/a/SKILL.md\"\n" + c.line))
			if len(got) != 1 {
				t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(got))
			}
			if got[0].Enabled != c.want {
				t.Errorf("Enabled for %q = %v, want %v", c.line, got[0].Enabled, c.want)
			}
			if got[0].Path != "/a/SKILL.md" {
				t.Errorf("the `enabled` reading disturbed the path: %+v", got[0])
			}
		})
	}
}

// TestParseSkillEntriesNonBooleanEnabledIsUnrecognisedLine records the
// consequence this change has in ANOTHER card's code, so it is stated here
// rather than discovered later.
//
// A non-boolean `enabled` line is not consumed through a recognised branch, so
// FirstUnrecognizedLine goes non-negative for its entry. The prune verb landed
// by card t506 (SPEC-CODEX-GHOST-SKILLS-PRUNE-001) skips any entry with
// FirstUnrecognizedLine >= 0, so such an entry's disposition moves from
// ELIGIBLE to PRESERVED. That direction is safe — it preserves a registration
// rather than deleting one — and t506's code is read here, never modified.
//
// A bare boolean stays recognised, which is the control: without it this test
// would pass equally well against a parser that recognised nothing at all.
func TestParseSkillEntriesNonBooleanEnabledIsUnrecognisedLine(t *testing.T) {
	cases := []struct {
		name        string
		line        string
		wantUnrecog bool
		wantAtLine  int // index of the `enabled` line when unrecognised
	}{
		{"bare_true_recognised", "enabled = true\n", false, -1},
		{"bare_false_recognised", "enabled = false\n", false, -1},
		{"double_quoted_unrecognised", "enabled = \"true\"\n", true, 2},
		{"single_quoted_unrecognised", "enabled = 'false'\n", true, 2},
		{"integer_unrecognised", "enabled = 1\n", true, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseSkillEntries([]byte("[[skills.config]]\npath = \"/a/SKILL.md\"\n" + c.line))
			if len(got) != 1 {
				t.Fatalf("ParseSkillEntries returned %d entries, want 1", len(got))
			}
			unrecognised := got[0].FirstUnrecognizedLine >= 0
			if unrecognised != c.wantUnrecog {
				t.Errorf("FirstUnrecognizedLine = %d (unrecognised=%v), want unrecognised=%v",
					got[0].FirstUnrecognizedLine, unrecognised, c.wantUnrecog)
			}
			if c.wantUnrecog && got[0].FirstUnrecognizedLine != c.wantAtLine {
				t.Errorf("FirstUnrecognizedLine = %d, want %d (the `enabled` line)",
					got[0].FirstUnrecognizedLine, c.wantAtLine)
			}
		})
	}
}
