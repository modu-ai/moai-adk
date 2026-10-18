package template_test

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const (
	// skillEntryCapChars is Claude Code's maxSkillDescriptionChars: the cap on
	// one entry's combined description + when_to_use text.
	skillEntryCapChars = 1536
	// skillListingCeilingChars is the deployed-total ceiling, about 4% above the
	// 23,986 measured on card t1796 against main 03faddb705.
	skillListingCeilingChars = 25000
)

// skillListingEntry is one skill's listing text.
type skillListingEntry struct {
	Name        string
	Description string
	WhenToUse   string
}

// chars counts runes (Unicode code points), the same unit the card t1796
// measurement script used, not bytes and not UTF-16 units.
func (e skillListingEntry) chars() int {
	return utf8.RuneCountInString(e.Description) + utf8.RuneCountInString(e.WhenToUse)
}

// checkSkillListing returns one message per violated limit: each entry over the
// per-entry cap, then the total when it exceeds the ceiling.
func checkSkillListing(entries []skillListingEntry) []string {
	var out []string
	total := 0
	for _, e := range entries {
		n := e.chars()
		total += n
		if n > skillEntryCapChars {
			out = append(out, fmt.Sprintf("skill %q: description + when_to_use is %d chars, over the %d cap (maxSkillDescriptionChars)", e.Name, n, skillEntryCapChars))
		}
	}
	if total > skillListingCeilingChars {
		out = append(out, fmt.Sprintf("deployed skill listing text is %d chars, over the %d ceiling; raising the ceiling needs a recorded decision referencing card t1796 (skill-authoring.md §Listing budget)", total, skillListingCeilingChars))
	}
	return out
}

// parseSkillListingEntry reads description and when_to_use from a frontmatter
// block, folding line breaks to single spaces like the measurement script.
func parseSkillListingEntry(name, frontmatter string) (skillListingEntry, error) {
	var fm asideFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return skillListingEntry{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	fold := func(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\n", " ")) }
	return skillListingEntry{Name: name, Description: fold(fm.Description), WhenToUse: fold(fm.WhenToUse)}, nil
}
