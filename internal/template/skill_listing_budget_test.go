package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Card t1796 regression guard for the skill-listing budget. The skill-authoring
// rule (§Listing budget) records why a 200K session overflows the listing by
// design; these tests keep the deployed total and the per-entry cap visible.

// minSweptSkills keeps an empty or partial sweep from passing green: the
// deployed template set held 38 skills when the guard was written.
const minSweptSkills = 30

// englandFlag is U+1F3F4 followed by the tag characters g, b, e, n, g and the
// cancel tag (U+E0067 U+E0062 U+E0065 U+E006E U+E0067 U+E007F).
const englandFlag = "\U0001F3F4\U000E0067\U000E0062\U000E0065\U000E006E\U000E0067\U000E007F"

// Host-unit arithmetic (Claude Code 2.1.302): an entry is the string
// "<description> - <when_to_use>" (the " - " separator, 3 chars, only when
// when_to_use is set), measured in UTF-16 code units against the 1,536 cap.

func TestCheckSkillListing_EntryOverCapFails(t *testing.T) {
	entries := []skillListingEntry{
		// description only: 1,536 units, exactly at the cap.
		{Name: "fits", Description: strings.Repeat("a", skillEntryCapChars)},
		// 1000 + 3 (separator) + 534 = 1,537, one over the cap.
		{Name: "too-long", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 534)},
	}
	got := checkSkillListing(entries)
	if len(got) != 1 {
		t.Fatalf("want exactly one violation, got %d: %v", len(got), got)
	}
	for _, want := range []string{"too-long", "1537", "UTF-16"} {
		if !strings.Contains(got[0], want) {
			t.Errorf("violation must mention %q (offender, host-form length, unit), got %q", want, got[0])
		}
	}
}

func TestCheckSkillListing_HostFormBoundary(t *testing.T) {
	cases := []struct {
		name     string
		entry    skillListingEntry
		wantFail bool
	}{
		// 1000 + 3 + 533 = 1,536: at the cap, passes.
		{"description+when_to_use at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 533)}, false},
		// 1000 + 3 + 534 = 1,537: over the cap, fails (the pre-fix guard counted 1,534).
		{"description+when_to_use at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 534)}, true},
		// description only, no separator: 1,536 passes, 1,537 fails.
		{"description only at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1536)}, false},
		{"description only at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1537)}, true},
		// 1535 + 1 astral rune (2 UTF-16 units) = 1,537 units although only 1,536 runes.
		{"astral rune pushes units over the cap", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1535) + "\U0001F600"}, true},
		// 1534 + 1 astral rune = 1,536 units: at the cap, passes.
		{"astral rune at the cap", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1534) + "\U0001F600"}, false},
		// Sanitizer step: "<" becomes "&lt;" and ">" becomes "&gt;" (4 units each).
		// description: 1532 + 4 = 1,536 passes; 1533 + 4 = 1,537 fails.
		{"description lt at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1532) + "<"}, false},
		{"description lt at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1533) + "<"}, true},
		{"description gt at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1532) + ">"}, false},
		{"description gt at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1533) + ">"}, true},
		// when_to_use: 1000 + 3 + 529 + 4 = 1,536 passes; 1000 + 3 + 530 + 4 = 1,537 fails.
		{"when_to_use lt at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 529) + "<"}, false},
		{"when_to_use lt at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 530) + "<"}, true},
		{"when_to_use gt at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 529) + ">"}, false},
		{"when_to_use gt at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 530) + ">"}, true},
		// The host's escape step leaves "&" alone: "&" is 1 unit and a literal "&lt;" is 4.
		// 1535 + 1 = 1,536 passes; 1536 + 1 = 1,537 fails; 1532 + 4 = 1,536 passes; 1533 + 4 = 1,537 fails.
		{"ampersand at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1535) + "&"}, false},
		{"ampersand at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1536) + "&"}, true},
		{"literal entity text at 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1532) + "&lt;"}, false},
		{"literal entity text at 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1533) + "&lt;"}, true},
		// Upper bound: the guard applies no trim, so padding counts in both fields.
		// "  " + 1536 + "  " = 1,540 > 1,536 fails (the host trims the description to 1,536
		// and would pass it; the guard intentionally overcounts).
		{"padded description counts its padding", skillListingEntry{Name: "e", Description: "  " + strings.Repeat("a", 1536) + "  "}, true},
		// 1000 + 3 + (" " + 532 + " " = 534) = 1,537: the padding of when_to_use counts.
		{"padded when_to_use counts its padding", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: " " + strings.Repeat("b", 532) + " "}, true},
		// Upper bound: runs of invisible characters are counted raw. The host collapses
		// each run to one space, so it would count fewer; the guard intentionally overcounts.
		// 1534 a + 3 newlines + "b" = 1,538 > 1,536 fails (host: 1,536).
		{"newline run counts raw", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1534) + "\n\n\nb"}, true},
		// 1000 + 3 + 531 + 3 (\t\r\n) + 1 = 1,538 fails (host: 1,536).
		{"control run in when_to_use counts raw", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 531) + "\t\r\nc"}, true},
		// 1534 a + U+200B + U+FEFF + "b" = 1,537 fails: format characters count 1 unit each
		// (the host collapses the run to one space and would pass, 1,536).
		{"zero-width and BOM count raw", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1534) + "\u200b\ufeffb"}, true},
		// 1535 a + U+0001 + "b" = 1,537 fails: a control character counts 1 unit (host: 1,536).
		{"control character counts raw", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1535) + "\x01b"}, true},
		// Sync-audit finding F8: the England flag is U+1F3F4 followed by six tag
		// characters (U+E0067 U+E0062 U+E0065 U+E006E U+E0067 U+E007F), seven astral
		// runes = 14 UTF-16 units. The host keeps them inside the RGI emoji sequence:
		// 1,523 + 14 = 1,537 > 1,536 fails. A guard that collapsed the tag characters
		// counted 1,526 and passed an entry the host truncates.
		{"England flag after 1523 chars is 1537", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1523) + englandFlag}, true},
		// 1,522 + 14 = 1,536 passes.
		{"England flag after 1522 chars is 1536", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1522) + englandFlag}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkSkillListing([]skillListingEntry{tc.entry})
			if tc.wantFail && len(got) == 0 {
				t.Errorf("host-form length %d is over the %d cap, want a violation", tc.entry.chars(), skillEntryCapChars)
			}
			if !tc.wantFail && len(got) != 0 {
				t.Errorf("host-form length %d is within the cap, want no violation, got %v", tc.entry.chars(), got)
			}
		})
	}
}

func TestCheckSkillListing_TotalOverCeilingFails(t *testing.T) {
	// Host form per entry: 1000 + 3 + 496 = 1,499. Sixteen of them = 23,984.
	var entries []skillListingEntry
	for len(entries) < 16 {
		entries = append(entries, skillListingEntry{Name: "pad", Description: strings.Repeat("p", 1000), WhenToUse: strings.Repeat("w", 496)})
	}
	// A 17th entry of 1000 + 3 + 14 = 1,017 makes the sum exactly 25,001.
	entries = append(entries, skillListingEntry{Name: "last", Description: strings.Repeat("p", 1000), WhenToUse: strings.Repeat("w", 14)})
	got := checkSkillListing(entries)
	if len(got) != 1 {
		t.Fatalf("want exactly one violation, got %d: %v", len(got), got)
	}
	for _, want := range []string{"25001", "25000", "t1796", "UTF-16"} {
		if !strings.Contains(got[0], want) {
			t.Errorf("total violation must mention %q, got %q", want, got[0])
		}
	}
}

func TestCheckSkillListing_TotalCountsEscapedBrackets(t *testing.T) {
	// Sixteen entries of 1000 + 3 + 496 = 1,499 give 23,984. The 17th is
	// 1000 + 3 + (10 + "<" -> 4) = 1,017 in the host form (1,014 raw), so the
	// host-form sum is exactly 25,001 while the raw sum is 24,998.
	var entries []skillListingEntry
	for len(entries) < 16 {
		entries = append(entries, skillListingEntry{Name: "pad", Description: strings.Repeat("p", 1000), WhenToUse: strings.Repeat("w", 496)})
	}
	entries = append(entries, skillListingEntry{Name: "last", Description: strings.Repeat("p", 1000), WhenToUse: strings.Repeat("w", 10) + "<"})
	got := checkSkillListing(entries)
	if len(got) != 1 {
		t.Fatalf("want exactly one violation (the total), got %d: %v", len(got), got)
	}
	if !strings.Contains(got[0], "25001") {
		t.Errorf("total violation must report the host-form sum 25001, got %q", got[0])
	}
}

func TestCheckSkillListing_AtLimitsPasses(t *testing.T) {
	// One entry exactly at the cap: 1000 + 3 + 533 = 1,536.
	entries := []skillListingEntry{
		{Name: "cap", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 533)},
	}
	// Fifteen padding entries of 1000 + 3 + 496 = 1,499 each = 22,485.
	for len(entries) < 16 {
		entries = append(entries, skillListingEntry{Name: "pad", Description: strings.Repeat("p", 1000), WhenToUse: strings.Repeat("w", 496)})
	}
	// 1,536 + 22,485 = 24,021 <= 25,000.
	if got := checkSkillListing(entries); len(got) != 0 {
		t.Fatalf("within limits must pass, got %v", got)
	}
}

func TestCountSkillChars_UTF16Semantics(t *testing.T) {
	cases := []struct {
		name string
		e    skillListingEntry
		want int
	}{
		// 3 Hangul (BMP, 1 unit each) + " - " (3) + 1 astral rune (2 units) = 8.
		{"hangul and astral with when_to_use", skillListingEntry{Name: "x", Description: "한글가", WhenToUse: "\U0001F600"}, 8},
		// description only: no separator; one astral rune = 2 units.
		{"astral description only", skillListingEntry{Name: "x", Description: "\U0001F600"}, 2},
		// ASCII with when_to_use: 3 + 3 + 2 = 8.
		{"ascii with when_to_use", skillListingEntry{Name: "x", Description: "abc", WhenToUse: "de"}, 8},
		// ASCII description only: 3.
		{"ascii description only", skillListingEntry{Name: "x", Description: "abc"}, 3},
		// "<a>" escapes to "&lt;a&gt;": 4 + 1 + 4 = 9.
		{"angle brackets escape", skillListingEntry{Name: "x", Description: "<a>"}, 9},
		// description 1 + separator 3 + "&lt;" 4 = 8.
		{"lt in when_to_use", skillListingEntry{Name: "x", Description: "d", WhenToUse: "<"}, 8},
		// "&" is not escaped by the host: 3 units.
		{"ampersand is one unit", skillListingEntry{Name: "x", Description: "a&b"}, 3},
		// A literal "&lt;" is already four characters and stays four: 4.
		{"literal entity is not re-escaped", skillListingEntry{Name: "x", Description: "&lt;"}, 4},
		// Upper bound: no collapse. Two newlines count 2, so "a\n\nb" = 4 (host: 3).
		{"newline run counts raw", skillListingEntry{Name: "x", Description: "a\n\nb"}, 4},
		// Two zero-width spaces count 2: 1 + 2 + 1 = 4 (host: 3).
		{"zero-width run counts raw", skillListingEntry{Name: "x", Description: "a\u200b\u200bb"}, 4},
		// Literal spaces count 1 each: "a  b" = 4.
		{"literal spaces stay", skillListingEntry{Name: "x", Description: "a  b"}, 4},
		// Upper bound: no trim. "  ab  " = 6 (host: 2).
		{"description padding counts", skillListingEntry{Name: "x", Description: "  ab  "}, 6},
		// England flag: seven astral runes, 7 x 2 = 14 units.
		{"england flag is fourteen units", skillListingEntry{Name: "x", Description: englandFlag}, 14},
		// when_to_use is not trimmed: 1 + 3 + 3 (" x ") = 7.
		{"when_to_use untrimmed", skillListingEntry{Name: "x", Description: "d", WhenToUse: " x "}, 7},
		// A whitespace-only when_to_use is still set, so the separator counts: 1 + 3 + 3 = 7.
		{"blank when_to_use keeps the separator", skillListingEntry{Name: "x", Description: "d", WhenToUse: "   "}, 7},
		// description-only entry with a bracket: "&lt;" = 4, no separator.
		{"lt description only", skillListingEntry{Name: "x", Description: "<"}, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e.chars(); got != tc.want {
				t.Errorf("chars() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestParseSkillListingEntry_KeepsRawFields(t *testing.T) {
	// chars() applies only the lengthening steps, so parsing hands it the YAML
	// values unfolded. A literal block scalar keeps its interior newline, and
	// when_to_use keeps the single trailing newline of the block as well: the host
	// does not trim when_to_use and its sanitizer turns that newline into one
	// space, a unit the host counts. The description loses it (the host trims).
	fm := "name: x\ndescription: >-\n  alpha\n  beta\nwhen_to_use: |\n  one\n  two\n"
	got, err := parseSkillListingEntry("x", fm)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Description != "alpha beta" {
		t.Errorf("Description = %q, want %q", got.Description, "alpha beta")
	}
	if got.WhenToUse != "one\ntwo\n" {
		t.Errorf("WhenToUse = %q, want %q", got.WhenToUse, "one\ntwo\n")
	}
	// A folded block scalar (`>`) keeps one trailing newline; parsing drops it
	// from the description too (the host trims the description).
	got, err = parseSkillListingEntry("x", "name: x\ndescription: >\n  alpha\n  beta\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Description != "alpha beta" {
		t.Errorf("folded Description = %q, want %q", got.Description, "alpha beta")
	}
}

// Sync-audit finding F9: real frontmatter text goes through the parser into the
// checker, so a parse-level edit that changes a count shows up at the boundary.
func TestCheckSkillListing_ParsedFrontmatterBoundary(t *testing.T) {
	desc1000 := strings.Repeat("a", 1000)
	cases := []struct {
		name     string
		fm       string
		wantUnit int
		wantFail bool
	}{
		// The block scalar `|` ends in a newline the host keeps and counts as one
		// space: 1000 + 3 (separator) + 533 + 1 (newline) = 1,537 fails.
		{"block when_to_use 533 is 1537", "name: x\ndescription: " + desc1000 + "\nwhen_to_use: |\n  " + strings.Repeat("b", 533) + "\n", 1537, true},
		// 1000 + 3 + 532 + 1 = 1,536 passes.
		{"block when_to_use 532 is 1536", "name: x\ndescription: " + desc1000 + "\nwhen_to_use: |\n  " + strings.Repeat("b", 532) + "\n", 1536, false},
		// A folded `>` description keeps a trailing newline in the parsed value, the
		// host trims it: 1,536 passes, 1,537 fails.
		{"folded description 1536", "name: x\ndescription: >\n  " + strings.Repeat("a", 1536) + "\n", 1536, false},
		{"folded description 1537", "name: x\ndescription: >\n  " + strings.Repeat("a", 1537) + "\n", 1537, true},
		// A single-line when_to_use has no trailing newline: 1000 + 3 + 533 = 1,536
		// passes, 534 gives 1,537 and fails.
		{"plain when_to_use 533 is 1536", "name: x\ndescription: " + desc1000 + "\nwhen_to_use: " + strings.Repeat("b", 533) + "\n", 1536, false},
		{"plain when_to_use 534 is 1537", "name: x\ndescription: " + desc1000 + "\nwhen_to_use: " + strings.Repeat("b", 534) + "\n", 1537, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, err := parseSkillListingEntry("x", tc.fm)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := e.chars(); got != tc.wantUnit {
				t.Errorf("chars() = %d, want %d", got, tc.wantUnit)
			}
			if failed := len(checkSkillListing([]skillListingEntry{e})) != 0; failed != tc.wantFail {
				t.Errorf("violation = %v, want %v", failed, tc.wantFail)
			}
		})
	}
}

// A SKILL.md-shaped string goes through the extraction the real-tree test uses.
// asideSplitFrontmatter cuts at "\n---", which removes the clip newline of a
// block scalar that is the last frontmatter key; the guard restores it.
func TestCheckSkillListing_SkillMDLastKeyWhenToUse(t *testing.T) {
	desc := "description: " + strings.Repeat("a", 1000) + "\n"
	cases := []struct {
		name     string
		text     string
		wantUnit int
		wantFail bool
	}{
		// 1000 + 3 (separator) + 533 + 1 (clip newline) = 1,537 fails.
		{"last key 533 is 1537", "---\nname: x\n" + desc + "when_to_use: |\n  " + strings.Repeat("b", 533) + "\n---\nbody\n", 1537, true},
		// 1000 + 3 + 532 + 1 = 1,536 passes.
		{"last key 532 is 1536", "---\nname: x\n" + desc + "when_to_use: |\n  " + strings.Repeat("b", 532) + "\n---\nbody\n", 1536, false},
		// Not last: the same numbers.
		{"middle key 533 is 1537", "---\nname: x\n" + desc + "when_to_use: |\n  " + strings.Repeat("b", 533) + "\nmodel: m\n---\nbody\n", 1537, true},
		{"middle key 532 is 1536", "---\nname: x\n" + desc + "when_to_use: |\n  " + strings.Repeat("b", 532) + "\nmodel: m\n---\nbody\n", 1536, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, err := parseSkillMD("x", tc.text)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := e.chars(); got != tc.wantUnit {
				t.Errorf("chars() = %d, want %d", got, tc.wantUnit)
			}
			if failed := len(checkSkillListing([]skillListingEntry{e})) != 0; failed != tc.wantFail {
				t.Errorf("violation = %v, want %v", failed, tc.wantFail)
			}
		})
	}
}

// SKILL.md-shaped text goes through parseSkillMD, the path the real-tree test
// uses. Shapes whose host count the guard cannot measure are rejected by name
// instead of being undercounted; valid shapes must still parse.
func TestParseSkillMD_RejectsUnsupportedShapes(t *testing.T) {
	md := func(fm string) string { return "---\nname: x\n" + fm + "---\nfirst body line\n" }
	cases := []struct {
		name    string
		text    string
		wantErr []string // all substrings must appear in the error; nil means valid
	}{
		// Keep-chomped block scalars: the host counts their trailing blank lines.
		{"keep when_to_use last and empty", md("description: d\nwhen_to_use: |+\n\n\n"), []string{"when_to_use", "keep-chomped"}},
		{"keep when_to_use in the middle", md("description: d\nwhen_to_use: |+\n  text\n\n\nmodel: m\n"), []string{"when_to_use", "keep-chomped"}},
		{"keep folded description", md("description: >+\n  text\n\n\n"), []string{"description", "keep-chomped"}},
		{"keep with indent digit", md("description: |2+\n    text\n\n"), []string{"description", "keep-chomped"}},
		{"keep with digit after plus", md("description: >+2\n    text\n\n"), []string{"description", "keep-chomped"}},
		// Non-string scalars.
		{"float description", md("description: 1e20\n"), []string{"description", "not a string"}},
		{"bool description", md("description: true\n"), []string{"description", "not a string"}},
		{"int when_to_use", md("description: d\nwhen_to_use: 12\n"), []string{"when_to_use", "not a string"}},
		// Empty description: the host falls back to the first body line.
		{"missing description", md(""), []string{"description", "empty", "first body line"}},
		{"null description", md("description:\n"), []string{"description", "empty", "first body line"}},
		{"quoted empty description", md("description: \"\"\n"), []string{"description", "empty", "first body line"}},
		{"whitespace block description", md("description: |\n   \n"), []string{"description", "empty", "first body line"}},
		// Valid shapes must not be swallowed by the rejections.
		{"plain", md("description: plain text\nwhen_to_use: more\n"), nil},
		{"quoted number string", md("description: \"1e20\"\n"), nil},
		{"literal block", md("description: |\n  one\n  two\n"), nil},
		{"folded block", md("description: >\n  one\n  two\n"), nil},
		{"literal strip", md("description: |-\n  one\n"), nil},
		{"folded strip", md("description: >-\n  one\n"), nil},
		{"null when_to_use is unset", md("description: d\nwhen_to_use:\n"), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseSkillMD("x", tc.text)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("valid shape rejected: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("want a rejection mentioning %v, got none", tc.wantErr)
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q must mention %q", err.Error(), want)
				}
			}
		})
	}
}

func TestTemplateSkillListingBudget(t *testing.T) {
	skillsDir := filepath.Join("templates", ".claude", "skills")
	dirs, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read %q: %v", skillsDir, err)
	}
	var entries []skillListingEntry
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(skillsDir, d.Name(), "SKILL.md"))
		if err != nil {
			t.Fatalf("skill %q: %v", d.Name(), err)
		}
		entry, err := parseSkillMD(d.Name(), string(raw))
		if err != nil {
			t.Fatalf("skill %q: %v", d.Name(), err)
		}
		entries = append(entries, entry)
	}
	if len(entries) < minSweptSkills {
		t.Fatalf("swept %d skills, want at least %d: an empty sweep must not pass", len(entries), minSweptSkills)
	}
	total := 0
	for _, e := range entries {
		total += e.chars()
	}
	t.Logf("swept %d skills, %d UTF-16 units of description - when_to_use (ceiling %d)", len(entries), total, skillListingCeilingChars)
	for _, v := range checkSkillListing(entries) {
		t.Error(v)
	}
}
