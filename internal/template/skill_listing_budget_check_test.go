package template_test

import (
	"fmt"
	"strings"
	"unicode/utf16"

	"gopkg.in/yaml.v3"
)

const (
	// skillEntryCapChars is Claude Code's default skillListingMaxDescChars: the
	// cap on one entry. The guard measures an upper bound of it in UTF-16
	// units of "description - when_to_use". A user can raise the host setting; the
	// guard pins the default.
	skillEntryCapChars = 1536
	// skillListingCeilingChars is the deployed-total ceiling in the same host
	// count: 897 units (3.6% of the ceiling) above the 24,103 measured on card
	// t1796 against main 03faddb705.
	skillListingCeilingChars = 25000
)

// skillListingEntry is one skill's listing text, as the YAML frontmatter
// carries it: chars() counts it as is.
type skillListingEntry struct {
	Name        string
	Description string
	WhenToUse   string
}

// chars is an UPPER BOUND of the entry's length in Claude Code's per-entry
// count: the UTF-16 units of "description - when_to_use" (the separator only
// when when_to_use is non-empty) with each "<" written as "&lt;" and each ">" as
// "&gt;". An astral rune such as an emoji counts 2.
//
// Invariant: for every skill with a non-empty description, chars() is at least
// the host's count. The guard applies only the steps that lengthen the string.
// The two it applies, read from Claude Code 2.1.302, are the "<" / ">" escape
// (rB) and the " - " join (bPt); "&" is not escaped. The steps that only shorten
// or keep the length are deliberately NOT applied (invisible-character run
// collapse, ANSI stripping, description trim), so the guard overcounts for them
// and cannot pass an entry the host truncates.
//
// Stated limits:
//
//  1. An empty description makes the host use the first body line (up to 100
//     units); chars() counts 0, an undercount. No deployed skill is affected, and
//     a test row pins it.
//  2. The 1,536 cap is the host default skillListingMaxDescChars; a user can
//     raise it.
//  3. The reading is of one Claude Code version; a later release can change the
//     pipeline and this guard would not notice.
//  4. Whether the host YAML reader keeps the trailing newline of a block scalar
//     is unobserved; parseSkillListingEntry drops it from both fields. The host
//     trims the description, so only when_to_use matters: if the host keeps its
//     newline, the host counts 1 more unit per such when_to_use, at most 35 on
//     the deployed total.
func (e skillListingEntry) chars() int {
	text := hostEscaper.Replace(e.Description)
	if e.WhenToUse != "" {
		text += " - " + hostEscaper.Replace(e.WhenToUse)
	}
	return len(utf16.Encode([]rune(text)))
}

// hostEscaper is the host's rB: only angle brackets are escaped.
var hostEscaper = strings.NewReplacer("<", "&lt;", ">", "&gt;")

// checkSkillListing returns one message per violated limit: each entry over the
// per-entry cap, then the total when it exceeds the ceiling.
func checkSkillListing(entries []skillListingEntry) []string {
	var out []string
	total := 0
	for _, e := range entries {
		n := e.chars()
		total += n
		if n > skillEntryCapChars {
			out = append(out, fmt.Sprintf("skill %q: description - when_to_use is %d UTF-16 units (upper bound of the host count), over the %d cap (skillListingMaxDescChars)", e.Name, n, skillEntryCapChars))
		}
	}
	if total > skillListingCeilingChars {
		out = append(out, fmt.Sprintf("deployed skill listing text is %d UTF-16 units of `description - when_to_use` (upper bound of the host count), over the %d ceiling; raising the ceiling needs a recorded decision referencing card t1796 (skill-authoring.md §Listing budget)", total, skillListingCeilingChars))
	}
	return out
}

// parseSkillListingEntry reads description and when_to_use from a frontmatter
// block. The one edit is dropping the trailing newline of a block scalar in
// either field (limit 4), so a `|` or `>` value is measured as its text. The
// host trims the description, so its trailing newline never counts there.
func parseSkillListingEntry(name, frontmatter string) (skillListingEntry, error) {
	var fm asideFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return skillListingEntry{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	return skillListingEntry{
		Name:        name,
		Description: strings.TrimRight(fm.Description, "\n"),
		WhenToUse:   strings.TrimRight(fm.WhenToUse, "\n"),
	}, nil
}
