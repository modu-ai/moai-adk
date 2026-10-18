package template_test

import (
	"fmt"
	"strings"
	"unicode/utf16"

	"gopkg.in/yaml.v3"
)

const (
	// skillEntryCapChars is Claude Code's default skillListingMaxDescChars: the
	// cap on one entry, measured in UTF-16 units of "description - when_to_use".
	// A user can raise the host setting; the guard pins the default.
	skillEntryCapChars = 1536
	// skillListingCeilingChars is the deployed-total ceiling in the same host
	// unit: 909 units (3.6% of the ceiling) above the 24,091 measured on card
	// t1796 against main 03faddb705.
	skillListingCeilingChars = 25000
)

// skillListingEntry is one skill's listing text.
type skillListingEntry struct {
	Name        string
	Description string
	WhenToUse   string
}

// chars is the entry's length the way Claude Code counts it: UTF-16 code units
// of "description - when_to_use" (the " - " separator only when when_to_use is
// set). An astral rune such as an emoji counts 2.
func (e skillListingEntry) chars() int {
	text := e.Description
	if e.WhenToUse != "" {
		text += " - " + e.WhenToUse
	}
	return len(utf16.Encode([]rune(text)))
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
			out = append(out, fmt.Sprintf("skill %q: description - when_to_use is %d UTF-16 units, as Claude Code counts it, over the %d cap (skillListingMaxDescChars)", e.Name, n, skillEntryCapChars))
		}
	}
	if total > skillListingCeilingChars {
		out = append(out, fmt.Sprintf("deployed skill listing text is %d UTF-16 units of `description - when_to_use`, as Claude Code counts it, over the %d ceiling; raising the ceiling needs a recorded decision referencing card t1796 (skill-authoring.md §Listing budget)", total, skillListingCeilingChars))
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
