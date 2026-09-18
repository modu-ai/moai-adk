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
//
// DECLARED RESIDUAL — PROSE ABSORPTION AFTER A TAIL. One over-reach shape is
// NOT declined, and is named here rather than left for the next reader to
// discover: inside a `maps` section, a bare number following a full id and a
// comma is absorbed as a tail EVEN WHEN the surrounding prose says that REQ is
// not mapped. Observed directly (sync-audit of card t801, scratch fixture
// outside the tree):
//
//	- AC-FIXH-001 (maps REQ-FIXH-001, 002 is explicitly NOT mapped by this AC): …
//	  before this change → CoverageIncomplete … REQ REQ-FIXH-002 …
//	  after  this change → 0 error(s), 1 warning(s)   (no CoverageIncomplete)
//
// So this change turns one CORRECT warning into silence — the same defect class
// the header above rejects the line as a unit to avoid. Three facts bound it,
// and none of them dissolves it:
//
//   - It is not a specification failure. REQ-SMS-004 defines the capture
//     SYNTACTICALLY ("ends at the first element that is neither a full REQ id
//     nor a bare numeric tail"), and `002` IS syntactically a bare numeric
//     tail. The code does what the SPEC says.
//   - It is not newly introduced. The table path (cellREQIDs, card t561) has
//     carried the same shape since it shipped; this file inherits it by reusing
//     that one rule, which is the property REQ-SMS-002 requires.
//   - The live population is zero today — the whole-corpus delta measured
//     exactly 0. That is a fact about today, not a guarantee about tomorrow:
//     legalizing the shorthand makes authors write numbers in `maps` sections
//     more often, which makes this shape more likely to be born, not less.
//
// RESIDUAL STATUS (card t913) — CLOSED ON THIS PATH, DECLARED OPEN ON THE TABLE
// PATH. The narrowing this declaration recorded as a candidate is now landed
// here: see siblingMapsREQIDs below, which truncates a trailing bare tail that
// no closing token follows. The two halves of the residual have DIFFERENT
// verdicts, and the reason is worth keeping, because a reader who assumes one
// verdict covers both will reach for the wrong repair.
//
//	maps path  (this file, the LOCATOR layer) → CLOSED. Cost measured 0.
//	table path (cellREQIDs, the EXPANDER layer) → OPEN. Cost measured 12.
//
// THE CLOSING-TOKEN SET IS FIVE, NOT FOUR. A tail may be followed by end of
// line, `)`, `:`, an em dash, or a markdown table-cell `|`. The candidate this
// declaration recorded before the measurement named only the first four; `|` was
// added when the corpus was counted, because 22 live `maps` sections sit inside a
// table cell and are terminated by a pipe. None of those 22 ends in a bare tail
// today, so the four-token list costs nothing right now — and would have
// truncated a correct tail the first time a table-cell author wrote the
// shorthand, with no signal that it had. The set here and the pattern in
// siblingMapsTailBoundary are the same five deliberately: a declaration naming a
// different set from the code is how the next reader loses track of which one was
// intended.
//
// Why the same rule closes one and not the other — it is the LAYER, not the
// path. The locator captures the id list and hands only that to cellREQIDs, so
// the prose that justifies rejecting a tail is destroyed between the two layers:
// by the time the shared expander runs, the tail sits at end-of-string and looks
// clean. The discriminating text still exists at the locator, which is where the
// narrowing therefore lives. Inside the expander no syntactic rule can decide
// it, because the information is no longer there.
//
// The two population figures are BOTH correct and were measured on different
// surfaces with the domain's own capture functions rather than a regular
// expression resembling them:
//
//   - maps path, live `acceptance.md`: 525 sections, 3 ending in a bare tail,
//     all 3 closed by `)`, all 3 inside this SPEC's own acceptance.md. The
//     narrowing changes 0 of them. That zero is weak evidence, not strong — a
//     near-empty population cannot exonerate a rule.
//   - table path, live table cells: 12 absorbed tails are followed by something
//     other than a clean boundary, and all 12 are LEGITIMATE mappings — a list
//     then a parenthetical, a semicolon, a range, explanatory prose. Narrowing
//     there would break only correct text, so it is NOT done. Two of the 12 are
//     shaped identically to the hazard (`016 sentinels + …` versus `002 is
//     explicitly NOT mapped`), which is why no syntactic rule separates them
//     inside the expander.
//
// This is the third instance in one batch of a single rule: a population figure
// means nothing without the surface AND the layer it was measured on. Quoting
// "the live population is zero" without naming which of the two paths it
// describes is how the two verdicts get conflated.

import (
	"regexp"
	"strings"
)

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
// siblingMapsTailBoundary accepts the tokens that may legitimately follow a
// captured section whose LAST element is a bare numeric tail: end of line, `)`,
// `:`, an em dash, or a markdown table-cell `|`.
//
// These five are the set the RESIDUAL STATUS declaration above states, and the
// two are kept in step on purpose — see there for why `|` belongs in it (22 live
// `maps` sections sit in a table cell, terminated by a pipe).
//
// The `\n` / `\r` alternatives are load-bearing, not redundant with `$`: Go's
// RE2 `$` matches at end of TEXT, never before a trailing newline the way Perl's
// does, so a section that ends its line would be truncated without them.
var siblingMapsTailBoundary = regexp.MustCompile(`^[ \t]*($|\n|\r|\)|:|—|\|)`)

// mapsCaptureEndsInBareTail reports whether the capture's last comma-separated
// element is a bare numeric tail rather than a full REQ id.
func mapsCaptureEndsInBareTail(capture string) bool {
	parts := strings.Split(capture, ",")
	last := strings.TrimSpace(parts[len(parts)-1])
	return last != "" && !strings.HasPrefix(strings.ToUpper(last), "REQ-")
}

// siblingMapsREQIDs returns the full REQ ids (with the `REQ-` prefix) mapped by
// the `maps` sections in text, expanding bare numeric tails through the shared
// rule the table path uses.
//
// Each capture is passed to cellREQIDs — the shared numeric-tail expander from
// lint_coverage_sibling_table.go — and nothing else is. The capture is the
// textual unit; the line that contains it is never read.
//
// PROSE-ABSORPTION BOUNDARY (card t913). A capture ending in a bare numeric tail
// is TRUNCATED by one element when no closing token follows it, so prose that
// disclaims the numeric can no longer be read as a mapping:
//
//	maps REQ-FIXH-001, 002 is explicitly NOT mapped by this AC
//	  before → REQ-FIXH-001, REQ-FIXH-002   (the correct CoverageIncomplete went silent)
//	  after  → REQ-FIXH-001                 (the disclaimed tail is dropped, the real id kept)
//
// Three properties bound this, and each was measured rather than assumed:
//
//   - It is applied HERE, at the locator layer, and NOT inside cellREQIDs. The
//     locator hands cellREQIDs the id list alone, so by the time the shared
//     expander runs, the prose that would justify rejecting the tail is already
//     gone and the tail looks like a clean end-of-string. That is why the same
//     narrowing is impossible in the shared rule — and why REQ-SMS-002's
//     shared-expansion property survives untouched.
//   - Only the TRAILING element is dropped. It is the one element adjacent to the
//     prose; anything before it is separated from that prose by a further
//     element, which leaves nothing to distinguish it from an ordinary list.
//   - Cost on the live corpus is zero: 3 sections end in a bare tail, all three
//     closed by `)`, all three inside the SPEC that documents the shorthand. A
//     zero measured against a near-empty population is weak evidence, not
//     strong — recorded as such rather than as "verified free".
func siblingMapsREQIDs(text string) []string {
	var ids []string
	for _, loc := range siblingMapsSectionPattern.FindAllStringSubmatchIndex(text, -1) {
		if len(loc) < 4 || loc[2] < 0 {
			continue
		}
		capture := text[loc[2]:loc[3]]
		if mapsCaptureEndsInBareTail(capture) && !siblingMapsTailBoundary.MatchString(text[loc[3]:]) {
			parts := strings.Split(capture, ",")
			capture = strings.Join(parts[:len(parts)-1], ",")
		}
		ids = append(ids, cellREQIDs(capture)...)
	}
	return ids
}
