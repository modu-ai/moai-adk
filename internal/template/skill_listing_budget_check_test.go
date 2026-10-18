package template_test

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"

	"gopkg.in/yaml.v3"
)

const (
	// skillEntryCapChars is Claude Code's default skillListingMaxDescChars: the
	// cap on one entry, measured in UTF-16 units of the sanitized
	// "description - when_to_use". A user can raise the host setting; the guard
	// pins the default.
	skillEntryCapChars = 1536
	// skillListingCeilingChars is the deployed-total ceiling in the same host
	// unit: 897 units (3.6% of the ceiling) above the 24,103 measured on card
	// t1796 against main 03faddb705.
	skillListingCeilingChars = 25000
)

// skillListingEntry is one skill's listing text, as the YAML frontmatter
// carries it (unsanitized): chars() applies the host's pipeline.
type skillListingEntry struct {
	Name        string
	Description string
	WhenToUse   string
}

// Host pipeline mirrored by chars() (Claude Code 2.1.302, read from the
// installed bundle; every step before `.length` is listed):
//
//	step                           host                                   this checker
//	read description               xAe: jee(fm.description) -> trim()     jsTrim
//	read when_to_use               xAe: String(fm.when_to_use), no trim   used as is
//	sanitize (both fields)         Msn: i(x) = rB(t(x)),                  hostSanitize
//	                               t(x) = _O(x, " ", {keepSpelledJoiners})
//	  1. strip ANSI, lone surrogates  _O: Bun.stripANSI(...)              not mirrored (limit 2)
//	  2. collapse invisible runs      _O: each run of [\p{Cc}\p{Cf}        collapseInvisible: one
//	                                  \p{Cs}\p{Co}\p{Cn}          space per run; Cn, Cs and
//	                                  Default_Ignorable ⠀] -> one      emoji joiners see limit 3
//	                                  space; literal spaces untouched
//	  3. escape                       rB: "<" -> "&lt;", ">" -> "&gt;"      strings.NewReplacer
//	                                  ("&" is NOT escaped)
//	join                           bPt: whenToUse ? d + " - " + w : d     same, truthy = non-empty
//	measure                        Ge.length (UTF-16 units) > 1536        len(utf16.Encode)
//
// Stated limits (measured on the 38 deployed skills at card t1796: none sets an
// empty description or carries an invisible character; 35 end a when_to_use
// block scalar with a newline):
//
//  1. An empty description makes the host substitute the first line of the body
//     (up to 100 units); this checker counts 0, an undercount of up to 100 units
//     for such a skill. No deployed skill is affected.
//  2. ANSI escape sequences are not stripped (overcount).
//  3. Cn (unassigned) and lone surrogates are not collapsed, and the joiners
//     U+200D, U+FE0E, U+FE0F are never collapsed (the host keeps them only
//     inside emoji sequences); both overcount, never undercount.
//  4. Whether the host YAML reader keeps the trailing newline of a block scalar
//     is not known. parseSkillListingEntry drops it; if the host keeps it as a
//     space, the host counts 1 more unit per such when_to_use: 35 units on the
//     deployed total and 1 on the largest entry, both inside the margins.
//  5. A whitespace-only when_to_use is set in the host (the separator counts);
//     chars() counts it, but a blank block scalar read from a file arrives
//     as unset because limit 4's trim removes its newlines.
//  6. JS trim and the Go unicode tables follow their own Unicode versions.

// chars is the entry's length the way Claude Code counts it: UTF-16 code units
// of the sanitized "description - when_to_use" (the " - " separator only when
// the sanitized when_to_use is non-empty). An astral rune such as an emoji
// counts 2.
func (e skillListingEntry) chars() int {
	text := hostSanitize(strings.TrimFunc(e.Description, isJSSpace))
	if w := hostSanitize(e.WhenToUse); w != "" {
		text += " - " + w
	}
	return len(utf16.Encode([]rune(text)))
}

// hostEscaper is the host's rB: only angle brackets are escaped.
var hostEscaper = strings.NewReplacer("<", "&lt;", ">", "&gt;")

// hostSanitize applies the host's t() then rB(): collapse invisible runs, then
// escape angle brackets.
func hostSanitize(s string) string {
	return hostEscaper.Replace(collapseInvisible(s))
}

// collapseInvisible replaces each maximal run of host-invisible characters with
// one space. The joiners U+200D, U+FE0E and U+FE0F stay verbatim.
func collapseInvisible(s string) string {
	var b strings.Builder
	inRun := false
	for _, r := range s {
		if isHostInvisible(r) {
			if !inRun {
				b.WriteByte(' ')
				inRun = true
			}
			continue
		}
		inRun = false
		b.WriteRune(r)
	}
	return b.String()
}

// isHostInvisible mirrors the host class [\p{Cc}\p{Cf}\p{Co}
// Default_Ignorable_Code_Point ⠀] (Cn and Cs are limit 3).
func isHostInvisible(r rune) bool {
	switch r {
	case 0x200D, 0xFE0E, 0xFE0F:
		return false
	case 0x2028, 0x2029, 0x2800:
		return true
	}
	return unicode.In(r, unicode.Cc, unicode.Cf, unicode.Co, unicode.Other_Default_Ignorable_Code_Point, unicode.Variation_Selector)
}

// isJSSpace is the set JavaScript's String.prototype.trim strips.
func isJSSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
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
// block, unsanitized. The one edit is dropping the trailing newline of a block
// scalar (limit 4), so a `|` or `>` value is measured as its text.
func parseSkillListingEntry(name, frontmatter string) (skillListingEntry, error) {
	var fm asideFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return skillListingEntry{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	return skillListingEntry{
		Name:        name,
		Description: fm.Description,
		WhenToUse:   strings.TrimRight(fm.WhenToUse, "\n"),
	}, nil
}
