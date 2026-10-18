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
		// Trim step: description is trimmed, when_to_use is not.
		// "  " + 1536 + "  " trims to 1,536 and passes.
		{"padded description trims to 1536", skillListingEntry{Name: "e", Description: "  " + strings.Repeat("a", 1536) + "  "}, false},
		// 1000 + 3 + (" " + 532 + " " = 534) = 1,537: the padding of when_to_use counts.
		{"padded when_to_use counts its padding", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: " " + strings.Repeat("b", 532) + " "}, true},
		// Collapse step: a run of control characters is one space.
		// 1534 a + "\n\n\n" (one space) + "b" = 1,536 passes (raw count 1,538).
		{"newline run collapses to one unit", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1534) + "\n\n\nb"}, false},
		// Same run in when_to_use: 1000 + 3 + 531 + 1 + 1 = 1,536 passes (raw 1,538).
		{"control run in when_to_use collapses", skillListingEntry{Name: "e", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 531) + "\t\r\nc"}, false},
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
		// A run of newlines is one space: "a b" = 3.
		{"newline run", skillListingEntry{Name: "x", Description: "a\n\nb"}, 3},
		// Zero-width spaces (format characters) collapse as one run: "a b" = 3.
		{"zero-width run", skillListingEntry{Name: "x", Description: "a\u200b\u200bb"}, 3},
		// Literal spaces are not part of the collapsed class: "a  b" = 4.
		{"literal spaces stay", skillListingEntry{Name: "x", Description: "a  b"}, 4},
		// description is trimmed: "ab" = 2.
		{"description trimmed", skillListingEntry{Name: "x", Description: "  ab  "}, 2},
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
	// The host pipeline (trim, collapse, escape) runs in chars(), so parsing hands
	// it the YAML values unfolded. A literal block scalar keeps its interior
	// newline; the single trailing newline of the block is dropped (a stated
	// limit: the host YAML reader's treatment of it is not mirrored).
	fm := "name: x\ndescription: >-\n  alpha\n  beta\nwhen_to_use: |\n  one\n  two\n"
	got, err := parseSkillListingEntry("x", fm)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Description != "alpha beta" {
		t.Errorf("Description = %q, want %q", got.Description, "alpha beta")
	}
	if got.WhenToUse != "one\ntwo" {
		t.Errorf("WhenToUse = %q, want %q", got.WhenToUse, "one\ntwo")
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
		fm, ok := asideSplitFrontmatter(string(raw))
		if !ok {
			t.Fatalf("skill %q: no frontmatter", d.Name())
		}
		entry, err := parseSkillListingEntry(d.Name(), fm)
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
