package template_test

import (
	"fmt"
	"regexp"
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
//     units), which the guard does not measure; parseSkillListingEntry rejects
//     such a skill by name instead of undercounting it. The same parser rejects
//     keep-chomped block scalars (|+ / >+) and non-string scalars (a float, a
//     bool or an int) in description or when_to_use. No deployed skill uses any
//     of the three.
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

// keepChompedScalar matches a description or when_to_use key line whose block
// scalar carries the keep chomping indicator (+), with or without an indent digit.
var keepChompedScalar = regexp.MustCompile(`(?m)^(description|when_to_use):[ \t]*[|>](?:\+[1-9]?|[1-9]\+)[ \t]*(?:#.*)?$`)

// parseSkillListingEntry reads description and when_to_use from a frontmatter
// block. The one edit is dropping the trailing newline of the description: the
// host trims it. when_to_use keeps its trailing newline, which the host keeps.
// Shapes whose host count the guard cannot measure are rejected by name: a
// keep-chomped block scalar (the host counts its trailing blank lines), a
// description or when_to_use that is not a string, and an empty description (the
// host then falls back to the first body line).
func parseSkillListingEntry(name, frontmatter string) (skillListingEntry, error) {
	if m := keepChompedScalar.FindStringSubmatch(frontmatter); m != nil {
		return skillListingEntry{}, fmt.Errorf("skill %q: %s is a keep-chomped block scalar (|+ or >+); keep-chomped block scalars are not supported by the listing guard because their trailing blank lines are counted by the host", name, m[1])
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(frontmatter), &doc); err != nil {
		return skillListingEntry{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) == 1 && doc.Content[0].Kind == yaml.MappingNode {
		kv := doc.Content[0].Content
		for i := 0; i+1 < len(kv); i += 2 {
			key, val := kv[i].Value, kv[i+1]
			if key != "description" && key != "when_to_use" {
				continue
			}
			if tag := val.ShortTag(); val.Kind == yaml.ScalarNode && tag != "!!str" && tag != "!!null" {
				return skillListingEntry{}, fmt.Errorf("skill %q: %s resolves to %s, not a string; quote it so the guard measures the text the host reads", name, key, tag)
			}
		}
	}
	var fm asideFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return skillListingEntry{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	if strings.TrimSpace(fm.Description) == "" {
		return skillListingEntry{}, fmt.Errorf("skill %q: description is empty; the host then falls back to the first body line, which the guard does not measure, so the skill is rejected rather than undercounted", name)
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
