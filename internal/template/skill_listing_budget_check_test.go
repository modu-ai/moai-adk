package template_test

import (
	"fmt"
	"strings"
	"unicode/utf16"

	"gopkg.in/yaml.v3"
)

const (
	// skillEntryCapChars is Claude Code's default skillListingMaxDescChars: the
	// cap on one entry, counted in UTF-16 units of
	// "description - when_to_use". A user can raise the host setting; the guard
	// pins the default.
	skillEntryCapChars = 1536
	// skillListingCeilingChars is the deployed-total ceiling in the same host
	// count: 862 units (3.4% of the ceiling) above the 24,138 measured on card
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

// chars is the entry's length in UTF-16 units of "description - when_to_use"
// (the separator only when when_to_use is non-empty) with each "<" written as
// "&lt;" and each ">" as "&gt;". An astral rune such as an emoji counts 2.
//
// Invariant: for the deployed set, chars() is at least the host's count, because
// it applies every step the host performs that lengthens the string and skips
// those that only shorten it. Read from Claude Code 2.1.302: the "<" / ">" escape
// (rB; "&" is not escaped), the " - " join (bPt), and the when_to_use newline.
// The host does not trim when_to_use and its sanitizer turns the trailing newline
// of a block scalar into one space, so each block-scalar when_to_use counts +1
// (measured: 35 deployed skills, total 24,138, largest entry 1,441); chars()
// counts that newline as 1 unit and adds no collapse logic. Skipped: invisible
// character run collapse, ANSI stripping, description trim (the guard overcounts
// for them).
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
			out = append(out, fmt.Sprintf("skill %q: description - when_to_use is %d UTF-16 units, over the %d cap (skillListingMaxDescChars)", e.Name, n, skillEntryCapChars))
		}
	}
	if total > skillListingCeilingChars {
		out = append(out, fmt.Sprintf("deployed skill listing text is %d UTF-16 units of `description - when_to_use`, over the %d ceiling; raising the ceiling needs a recorded decision referencing card t1796 (skill-authoring.md §Listing budget)", total, skillListingCeilingChars))
	}
	return out
}

// parseSkillListingEntry reads description and when_to_use from a frontmatter
// block. The one edit is dropping the trailing newline of the description: the
// host trims it. when_to_use keeps its trailing newline, which the host keeps.
func parseSkillListingEntry(name, frontmatter string) (skillListingEntry, error) {
	var fm asideFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return skillListingEntry{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	return skillListingEntry{
		Name:        name,
		Description: strings.TrimRight(fm.Description, "\n"),
		WhenToUse:   fm.WhenToUse,
	}, nil
}

// parseSkillMD extracts the frontmatter of a SKILL.md and parses it. It is the
// one extraction the guard uses, so only this guard is affected.
//
// asideSplitFrontmatter cuts at "\n---", which removes the clip newline of a
// block scalar that is the last frontmatter key. Assumption, not an observed
// host fact: the host cuts after the closing delimiter and keeps that newline.
// Restoring it here can only overcount, the safe direction.
func parseSkillMD(name, text string) (skillListingEntry, error) {
	fm, ok := asideSplitFrontmatter(text)
	if !ok {
		return skillListingEntry{}, fmt.Errorf("skill %q: no frontmatter", name)
	}
	if !strings.HasSuffix(fm, "\n") {
		fm += "\n"
	}
	return parseSkillListingEntry(name, fm)
}
