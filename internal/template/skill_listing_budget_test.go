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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e.chars(); got != tc.want {
				t.Errorf("chars() = %d, want %d", got, tc.want)
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
