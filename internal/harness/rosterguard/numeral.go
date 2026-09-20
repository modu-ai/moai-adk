package rosterguard

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// This file is the numeral-adjacency layer — the guard's SECOND discovery axis.
//
// # Why one axis was not enough
//
// Sweep() finds a roster listing by ENUMERATION: a file qualifies when it
// carries at least SweepThreshold distinct agent names. A site that states only
// a SIZE — a number — while naming few agents or none at all is structurally
// invisible to it. .moai/project/tech.md sizes the retained roster and names
// ZERO agents, so no enumeration threshold reaches it: not 10, not 1. The axis
// a site ENUMERATES on and the axis it CLAIMS on are different populations.
//
// This layer anchors on a roster NOUN instead and reads the numeral next to it,
// which is what lets it reach a claim the enumeration sweep cannot.
//
// # Two sets, named apart
//
// The BREADTH set is every post-neutralisation match, discharged or not. The
// FINDING set is its undischarged subset — what DischargeNumeralHits reports
// and what fails the guard. A neutralised phrase is in NEITHER: neutralisation
// runs before matching, so it never becomes a match at all.
//
// The breadth set is what the test PRINTS. Printing only the finding set would
// make a green tree print nothing, and nothing would then constrain BREADTH: a
// layer implementing one noun form passes every shell-level criterion while
// missing most sites.
//
// The shape is reused from internal/web/docs_tab_contract_test.go, which
// implements the same layer for a different subject (the settings-tab count).
// Its decisions are adopted here rather than re-derived: an ASCII-word-bounded
// noun class, a measured adjacency window, a digit axis and a word axis as
// separate concerns, and a neutraliser that removes a numeral which is not a
// count.

// ── Scope ────────────────────────────────────────────────────────────────────

// numeralResearchPrefix is the layer's ONE addition to the sweep's exclusions.
//
// .moai/research/ is a dated-record tree: it describes the roster AS IT WAS on
// the day each file was written, and a live digit-axis candidate sits in it
// (anthropic-best-practices-2026-05-24.md — "CLAUDE.md §4 agent catalog name
// mismatch"). The sweep does not exclude it today because an enumeration hit
// there would be a genuine listing; a count claim there is a dated citation.
const numeralResearchPrefix = ".moai/research/"

// numeralSkipPrefixes is the layer's exclusion set, stated in ONE place.
//
// It CITES the sweep's own list rather than re-listing it: two hand-maintained
// copies of an exclusion set drift, and the drift is silent in the direction
// that matters (a tree excluded from one walk and not the other).
func numeralSkipPrefixes() []string {
	out := make([]string, 0, len(sweepSkipPrefixes)+1)
	out = append(out, sweepSkipPrefixes...)
	out = append(out, numeralResearchPrefix)
	return out
}

// InNumeralScope reports whether rel — repo-root-relative, forward-slashed — is
// inside the numeral layer's scope.
func InNumeralScope(rel string) bool {
	if sweepSkipFiles[rel] {
		return false
	}
	for _, skip := range numeralSkipPrefixes() {
		if strings.HasPrefix(rel, skip) {
			return false
		}
	}
	return true
}

// ── Noun class, numerals, adjacency ──────────────────────────────────────────

// rosterNounRe is the adopted noun class (decision D1), bounded by ASCII word
// boundaries so the noun does not match inside a longer word: without the
// boundary, "retained agents-reference" reads as a roster claim.
//
// The class deliberately sits between two measured alternatives. The narrow
// class {retained agent(s)} misses the live "11-agent catalog" stale wording
// outright. The unrestricted class {agent(s)} produced 613 hits, at which point
// the allowlist needed to tame it — not the roster — becomes what a reader
// reviews.
//
// "N-agent catalog" needs no alternative of its own: the numeral is attached to
// the noun, and the nearest-preceding rule below reads it.
var rosterNounRe = regexp.MustCompile(`\b(?:retained agents?|agent catalog|agent roster|retained catalog)\b`)

// numeralAdjacencyRunes is the gap allowed between the numeral's end and the
// noun's start, in runes.
//
// Adopted from the reuse target, where ten characters separate "nine" from
// "tabs" in "The nine settings tabs" — a window narrow enough to look tidy is
// structurally blind to the English sites. Re-measured here rather than
// assumed: the live delegationmap citation "CLAUDE.md §4 (the 13 retained
// agents" carries a 1-rune gap, and the widest live gap observed on this tree
// is "11 retained agents x 3 model tiers" at 1. The window is therefore set by
// the reuse target's measurement, not fitted to this tree's prose — a window
// chosen to make the tree green is a guard fitted to today's writing.
const numeralAdjacencyRunes = 12

// wordNumeralAlt is the spelled-out-numeral axis.
//
// Alternatives are ordered longest-first because Go's regexp is leftmost-FIRST,
// so "ten" placed ahead of "thirteen" would report a truncated numeral.
//
// This axis is a STRUCTURAL hole, not a current catch: on this tree its live
// breadth set is empty, and the only live candidate — "All four are retained
// agents." — is a subset predication the neutraliser removes before matching.
// It is implemented so it is CORRECT WHEN IT TURNS ON LATER, and a positive
// control on synthetic input proves the regexp fires, so an empty live result
// stays distinguishable from a dead regexp. Maintaining the locale forms is the
// standing cost of opening the axis: a spelling not listed here drops out
// silently.
const wordNumeralAlt = `seventeen|thirteen|fourteen|eighteen|nineteen|` +
	`fifteen|sixteen|twelve|twenty|eleven|` +
	`three|seven|eight|four|five|nine|one|two|six|ten|` +
	`열여섯|열다섯|열네|열세|열두|열한|여덟|다섯|여섯|일곱|아홉|하나|` +
	`열|둘|셋|넷|한|두|세|네|` +
	`[一二三四五六七八九十]`

var (
	digitNumeralRe = regexp.MustCompile(`[0-9]+`)
	wordNumeralRe  = regexp.MustCompile(`(?i)(?:` + wordNumeralAlt + `)`)
)

// ── Neutralisation (runs BEFORE matching) ────────────────────────────────────

// neutraliseRes remove a numeral that is not a count, by the mechanism the
// reuse target uses for 第三 in 第三方: the numeral is blanked before the axes
// run, so the phrase enters NEITHER set.
//
// The ORDER is normative (REQ-RNA-007): neutralise -> match -> discharge ->
// report. It decides what the populations ARE. Under this order the live
// word-axis breadth set is empty by construction; under the reverse order the
// same string would be matched and then discharged, and the two criteria that
// describe those populations would contradict each other.
var neutraliseRes = []*regexp.Regexp{
	// Selector — predicates over ONE member of a set rather than stating a
	// size: "manager-spec is one of the 11 retained agents".
	regexp.MustCompile(`(?i)\bone of the +(?:[0-9]+|` + wordNumeralAlt + `)\b`),
	// Subset predication — states that SOME members have a property:
	// "All four are retained agents." (live, at
	// .claude/agents/harness/workflow-specialist.md:52). The numeral must be
	// immediately followed by is/are, so a genuine size claim ("13 retained
	// agents are listed") is untouched — there the numeral is followed by the
	// noun, not by the verb.
	regexp.MustCompile(`(?i)\b(?:all +|only +)?(?:[0-9]+|` + wordNumeralAlt + `) +(?:are|is)\b`),
	// Ordinal — carried from the reuse target for the same reason it exists
	// there: 第三 ("third") is not a count of anything.
	regexp.MustCompile(`第[一二三四五六七八九十]`),
}

// blankNumerals replaces every numeral inside s with "X", preserving everything
// else, so a neutralised span can no longer be read as a count.
func blankNumerals(s string) string {
	s = digitNumeralRe.ReplaceAllString(s, "X")
	return wordNumeralRe.ReplaceAllString(s, "X")
}

// NeutraliseNumerals removes the non-count numerals from one line.
func NeutraliseNumerals(line string) string {
	for _, re := range neutraliseRes {
		line = re.ReplaceAllStringFunc(line, blankNumerals)
	}
	return line
}

// ── The breadth set ──────────────────────────────────────────────────────────

// NumeralHit is one member of the breadth set: a numeral observed adjacent to a
// roster noun, after neutralisation.
type NumeralHit struct {
	// Path is repo-root-relative, forward-slashed. Empty for a line-level scan.
	Path string
	// Line is 1-based. Zero for a line-level scan.
	Line int
	// Phrase is the matched text from the numeral through the noun.
	Phrase string
	// Numeral is the selected numeral AS WRITTEN — "13", "thirteen", "열세".
	// Kept as text rather than parsed: the word axis carries locale forms whose
	// integer value the layer never needs, and a parsed 0 would be
	// indistinguishable from an unparsed one.
	Numeral string
	// Axis is "digit" or "word".
	Axis string
}

// ScanNumeralLine returns the breadth-set hits on one line.
func ScanNumeralLine(line string) []NumeralHit {
	scanned := NeutraliseNumerals(line)
	var out []NumeralHit
	for _, loc := range rosterNounRe.FindAllStringIndex(scanned, -1) {
		if h, ok := numeralBefore(scanned, loc[0], loc[1]); ok {
			out = append(out, h)
		}
	}
	return out
}

// numeralBefore selects the NEAREST PRECEDING numeral for the noun occupying
// [nounStart, nounEnd) in line.
//
// Nearest-preceding, not leftmost, and the difference is not theoretical: a
// leftmost rule reads the LIVE string "CLAUDE.md §4 (the 13 retained agents"
// (internal/harness/delegationmap/types.go:73) and reports 4 — the section
// marker — rather than 13. The registry's own CountPattern doc comment cites
// the same string as the reason that pattern is a regexp rather than a line
// anchor.
func numeralBefore(line string, nounStart, nounEnd int) (NumeralHit, bool) {
	prefix := line[:nounStart]

	best := -1
	var bestText, bestAxis string
	consider := func(axis string, locs [][]int) {
		for _, m := range locs {
			gap := prefix[m[1]:]
			// A sentence terminator inside the gap breaks the adjacency: the
			// numeral then belongs to the previous sentence.
			if strings.ContainsAny(gap, ".。") {
				continue
			}
			if utf8.RuneCountInString(gap) > numeralAdjacencyRunes {
				continue
			}
			if m[1] > best {
				best, bestText, bestAxis = m[0], prefix[m[0]:m[1]], axis
			}
		}
	}
	consider("digit", digitNumeralRe.FindAllStringIndex(prefix, -1))
	consider("word", wordNumeralRe.FindAllStringIndex(prefix, -1))

	if best < 0 {
		return NumeralHit{}, false
	}
	return NumeralHit{
		Phrase:  strings.TrimSpace(line[best:nounEnd]),
		Numeral: bestText,
		Axis:    bestAxis,
	}, true
}

// maxNumeralFileBytes bounds what the layer reads, for the same reason the
// sweep bounds itself: a large binary or vendored blob must not dominate the
// walk.
const maxNumeralFileBytes = 4 << 20

// ScanNumeralAxis walks root and returns the BREADTH set — every
// post-neutralisation match in the layer's scope, discharged or not.
func ScanNumeralAxis(root string) ([]NumeralHit, error) {
	var out []NumeralHit
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			// An unreadable entry is reported, never silently treated as
			// absent: "the layer found nothing here" and "the layer could not
			// look here" are different facts.
			return err
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "." {
				return nil
			}
			for _, skip := range numeralSkipPrefixes() {
				if strings.HasPrefix(rel+"/", skip) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !InNumeralScope(rel) {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil || info.Size() > maxNumeralFileBytes || !info.Mode().IsRegular() {
			return nil
		}
		b, readErr := os.ReadFile(p)
		if readErr != nil {
			return nil
		}
		for i, line := range strings.Split(string(b), "\n") {
			for _, h := range ScanNumeralLine(line) {
				h.Path, h.Line = rel, i+1
				out = append(out, h)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("numeral scan of %s: %w", root, err)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Line < out[j].Line
	})
	return out, nil
}

// ── The finding set ──────────────────────────────────────────────────────────

// NumeralExempt declares, with a mandatory reason, a path the numeral layer
// reaches and must NOT report.
//
// A declaration, never an inference — the same stance Staleness and
// SweepUnreachable take, and for the same reason: an inferred exemption
// silently absorbs the next site that happens to look like the ones already
// excused, which is the failure mode this package exists to prevent. An empty
// reason produces a finding reporting the declaration as INCOMPLETE rather than
// a suppression, so "silence the warning" never becomes cheaper than "state the
// reason".
type NumeralExempt struct {
	// ID is the stable identifier used in messages.
	ID string
	// Path is repo-root-relative, forward-slashed.
	Path string
	// Reason is the mandatory why, stated so a reviewer can disagree with it.
	Reason string
}

// DischargeNumeralHits returns the FINDING set as messages — the undischarged
// subset of hits. An empty slice is a pass.
//
// The discharge rule (decision D2) is strict by choice: a hit discharges only
// on a ClaimCount row for its path, or on a numeral exemption. Reusing the
// sweep's any-row-by-path rule would free exactly one further path on this
// tree, and would let a path registered only for a MEMBERSHIP claim silently
// discharge a stale COUNT claim inside it — precisely the failure class this
// layer exists to close. A count and a membership are independent claims, and
// ClaimKind already models them as such.
func DischargeNumeralHits(hits []NumeralHit, sites []Site, exempts []NumeralExempt) []string {
	counted := map[string]bool{}
	member := map[string]bool{}
	for _, s := range sites {
		if s.Claims.Has(ClaimCount) {
			counted[s.Path] = true
		}
		if s.Claims.Has(ClaimMembership) {
			member[s.Path] = true
		}
	}
	exempt := map[string]NumeralExempt{}
	for _, e := range exempts {
		exempt[e.Path] = e
	}

	var bad []string
	reported := map[string]bool{}
	for _, h := range hits {
		if counted[h.Path] {
			continue
		}
		if e, ok := exempt[h.Path]; ok {
			if strings.TrimSpace(e.Reason) != "" {
				continue
			}
			key := "incomplete:" + h.Path
			if !reported[key] {
				reported[key] = true
				bad = append(bad, fmt.Sprintf(
					"numeral exemption %s (%s): the declaration is incomplete — its reason is empty or whitespace-only, so it excuses nothing and the hit is NOT discharged. State a reason a reviewer can disagree with, or delete the row and register a ClaimCount claim",
					e.ID, h.Path))
			}
			continue
		}
		if member[h.Path] {
			bad = append(bad, fmt.Sprintf(
				"%s:%d states a roster count (%s, in %q) and %s is registered for a MEMBERSHIP claim only — a membership registration does not discharge a count claim; the two are independent, and a stale count inside a correct listing is a real and observed shape. Add a row carrying ClaimCount with its own CountPattern",
				h.Path, h.Line, h.Numeral, h.Phrase, h.Path))
			continue
		}
		bad = append(bad, fmt.Sprintf(
			"undeclared count claim: %s:%d states a roster count (%s, in %q) on the %s axis and is not declared. Discharge it either by registering a Registry() row whose Claims carry ClaimCount with its own CountPattern (and a KnownStale marker if the number disagrees with template.ProfileMatrixAgents()), or by declaring a numeral exemption for this path with a non-empty reason",
			h.Path, h.Line, h.Numeral, h.Phrase, h.Axis))
	}
	return bad
}
