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

func TestCheckSkillListing_EntryOverCapFails(t *testing.T) {
	entries := []skillListingEntry{
		{Name: "fits", Description: strings.Repeat("a", skillEntryCapChars)},
		{Name: "too-long", Description: strings.Repeat("a", 1000), WhenToUse: strings.Repeat("b", 537)},
	}
	got := checkSkillListing(entries)
	if len(got) != 1 {
		t.Fatalf("want exactly one violation, got %d: %v", len(got), got)
	}
	if !strings.Contains(got[0], "too-long") || !strings.Contains(got[0], "1537") {
		t.Errorf("violation must name the offender and its length, got %q", got[0])
	}
}

func TestCheckSkillListing_TotalOverCeilingFails(t *testing.T) {
	entries := []skillListingEntry{
		{Name: "a", Description: strings.Repeat("a", 1500)},
	}
	for i := 0; len(entries) < 17; i++ {
		entries = append(entries, skillListingEntry{Name: "pad", Description: strings.Repeat("p", 1500)})
	}
	// 17 x 1500 = 25,500; trim one entry so the sum is exactly 25,001.
	entries[16].Description = strings.Repeat("p", 1001)
	got := checkSkillListing(entries)
	if len(got) != 1 {
		t.Fatalf("want exactly one violation, got %d: %v", len(got), got)
	}
	for _, want := range []string{"25001", "25000", "t1796"} {
		if !strings.Contains(got[0], want) {
			t.Errorf("total violation must mention %q, got %q", want, got[0])
		}
	}
}

func TestCheckSkillListing_AtLimitsPasses(t *testing.T) {
	entries := []skillListingEntry{
		{Name: "cap", Description: strings.Repeat("a", skillEntryCapChars)},
	}
	for len(entries) < 16 {
		entries = append(entries, skillListingEntry{Name: "pad", Description: strings.Repeat("p", 1500)})
	}
	// 1536 + 15 x 1500 = 24,036 <= 25,000.
	if got := checkSkillListing(entries); len(got) != 0 {
		t.Fatalf("within limits must pass, got %v", got)
	}
}

func TestCountSkillChars_RuneSemantics(t *testing.T) {
	// Runes, not bytes and not UTF-16 units: three Hangul syllables and one
	// astral rune count as four.
	e := skillListingEntry{Name: "x", Description: "한글가", WhenToUse: "😀"}
	if got := e.chars(); got != 4 {
		t.Errorf("chars() = %d, want 4", got)
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
	t.Logf("swept %d skills, %d chars of description + when_to_use (ceiling %d)", len(entries), total, skillListingCeilingChars)
	for _, v := range checkSkillListing(entries) {
		t.Error(v)
	}
}
