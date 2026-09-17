package spec

// lint_coverage_sibling_maps.go — bare numeric-tail shorthand in a sibling
// acceptance.md `maps` list (card t801, SPEC-SIBLING-MAPS-SHORTHAND-001).
//
// WHAT THIS CLOSES. siblingAcceptanceCoveredREQIDs collects `maps`-list
// coverage with ExtractRequirementMappings, whose locator requires the `REQ-`
// prefix on every comma-separated element:
//
//	reqSectionPattern = (?i)maps\s+(REQ-[A-Z0-9-]+(?:\s*,\s*REQ-[A-Z0-9-]+)*)
//
// So on `- AC-FIXA-001 (maps REQ-FIXA-001, 002): …` the captured section is
// `REQ-FIXA-001` and `, 002` is dropped silently. The enumerator never sees the
// tail, REQ-FIXA-002 is reported uncovered, and the author DID map it. The
// warning is FALSE. The same shorthand inside a TABLE CELL is already expanded
// by cellREQIDs (card t561); only the `maps`-list path was blind.
//
// THE TEXTUAL UNIT, AND WHY THE TWO OBVIOUS CANDIDATES ARE BOTH WRONG.
// cellREQIDs is sound because it is handed ONE TABLE CELL: fullREQIDPattern
// matches every `REQ-…` token in whatever string it receives, so reusing it
// here needs a unit supplied from outside. Two candidates present themselves
// and both are rejected:
//
//   - THE LINE is wrong. On
//     `- AC-X-001 (maps REQ-X-001) — verifies REQ-X-009, 010 are unreachable.`
//     a line-scoped unit counts REQ-X-009 and an expanded REQ-X-010 as covered,
//     suppressing two GENUINE findings. That manufactured silence is the exact
//     defect class this file exists to close — over-reaching here would trade
//     one false warning for two missing ones.
//   - reqSectionPattern's CAPTURE is wrong. Its truncation at the first
//     non-`REQ-` element IS the defect being repaired. ExtractRequirementMappings
//     and reqSectionPattern are also immutable (REQ-SMS-003, card t561's
//     standing decision): they are called by the inline spec.md AC path
//     (parseSingleACLine), which this repair must not reach.
//
// THE UNIT IS THEREFORE THE CAPTURE OF THE SIBLING-LOCAL WIDENED LOCATOR below,
// which holds four properties. They are the SPEC's (REQ-SMS-004..006), not this
// file's, and none of them is negotiable:
//
//  1. ANCHORED AT `maps`. Tokens before `maps` on the line are outside the unit.
//  2. A CONTIGUOUS COMMA-SEPARATED RUN. The capture ends at the first element
//     that is neither a full `REQ-…` id nor a bare numeric tail, so trailing
//     prose (`— verifies REQ-X-009, 010 …`) falls outside it.
//  3. HORIZONTAL WHITESPACE ONLY between elements. `\s` would admit `\n` and
//     let a section ending in a comma absorb the next list item; `[ \t]` bounds
//     the unit to one line (REQ-SMS-005).
//  4. ENUMERATION READS THE CAPTURE ALONE. It never receives the line, the
//     paragraph, or the file. This is what makes the reuse of cellREQIDs sound
//     rather than merely convenient.
//
// Property 3 also makes reusing numericTailPattern (`^\s*,\s*([0-9]+)\b`, whose
// `\s` DOES admit `\n`) safe without editing it: the capture handed to the
// enumerator contains no newline, so the enumerator cannot cross one. The
// line-bounding is enforced once, at the locator, and inherited. Each `maps`
// section is its own unit, so a tail in one never takes a prefix from a
// previous one.
//
// ONE RULE, NOT TWO (REQ-SMS-002). The expansion is not re-implemented here.
// cellREQIDs is CALLED — the same function the table path calls — so the two
// paths cannot drift. A second, independently-written expansion rule is
// forbidden; TestSiblingMapsExpansionSharesTheTableRule exercises both paths on
// the same element run through that one helper and fails if they disagree.
//
// DECLINED FORMS, named rather than left implicit (mirrored from spec.md §D):
// a tail with a non-numeric body (`, 00b`); a tail separated by anything other
// than a comma; a tail preceding the first full id in its section; and a full
// id written without the `REQ-` prefix. None is expanded. The wrapped `maps`
// list spanning two physical lines is declined for the same reason and by the
// same mechanism (property 3): admitting the line crossing is an unbounded
// widening bought for a measured population of zero.

import "regexp"

// siblingMapsSectionPattern is the sibling-local WIDENED `maps`-section
// locator. It differs from ears.go's reqSectionPattern — which it deliberately
// does not touch — in exactly two ways, each carrying one of the properties
// above:
//
//   - after the first full id, a comma-separated element may be a full REQ id
//     OR a bare numeric tail (`\d+`), which is the widening itself; the trailing
//     `\b` declines a non-numeric body such as `00b`;
//   - the inter-element separator is `[ \t]` rather than `\s`, so the capture
//     cannot cross a line boundary (REQ-SMS-005).
//
// The leading element still REQUIRES the `REQ-` prefix, which is what makes a
// bare tail with no full id before it in its own capture unexpandable: there is
// no capture at all for the enumerator to read (REQ-SMS-006).
//
// THE TAIL IS WRITTEN `\d+`, NOT `[0-9]+`, AND THE SPELLING IS DELIBERATE. This
// is a section LOCATOR, not the expansion rule: it recognises where a tail sits
// and never forms an id from one. AC-SMS-005's second command is a shape
// tripwire for a duplicated numeric-tail RULE — a compiled pattern whose body
// carries a comma and a literal `[0-9]` class — and this locator would trip it
// while duplicating nothing, so the two are spelled apart. `\d` is ASCII-exact
// in RE2, so the spelling changes no behaviour; it keeps the tripwire counting
// the one real expansion rule (numericTailPattern) and nothing else.
var siblingMapsSectionPattern = regexp.MustCompile(`(?i)maps[ \t]+(REQ-[A-Z0-9-]+(?:[ \t]*,[ \t]*(?:REQ-[A-Z0-9-]+|\d+\b))*)`)

// siblingMapsREQIDs returns the full REQ ids (with the `REQ-` prefix) mapped by
// the `maps` sections in text, expanding bare numeric tails through the shared
// rule the table path uses.
//
// Each capture is passed to cellREQIDs — the shared numeric-tail expander from
// lint_coverage_sibling_table.go — and nothing else is. The capture is the
// textual unit; the line that contains it is never read.
func siblingMapsREQIDs(text string) []string {
	var ids []string
	for _, section := range siblingMapsSectionPattern.FindAllStringSubmatch(text, -1) {
		if len(section) < 2 {
			continue
		}
		ids = append(ids, cellREQIDs(section[1])...)
	}
	return ids
}
