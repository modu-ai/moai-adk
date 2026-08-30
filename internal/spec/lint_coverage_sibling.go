package spec

// lint_coverage_sibling.go — the sibling-acceptance.md half of CoverageRule
// (SPEC-COVERAGE-RULE-SCOPE-001 M4, REQ-CRS-001-006..008).
//
// WHAT THIS CLOSES. CoverageRule judged AC→REQ coverage from `doc.Criteria`
// alone, and `SPECDoc` carries spec.md alone. That is a Tier S premise applied
// to every tier: `spec-assembly.md` states inline AC as a property of Tier S,
// while Tier M/L require a sibling `acceptance.md`, and
// `manager-develop-prompt-template.md` names that file the AC SSOT. A
// convention-following Tier M SPEC therefore read as uncovered — and this SPEC
// reproduced the defect on itself during authoring (8 CoverageIncomplete),
// which is why a minimal mirror table sits in its own spec.md as a stopgap.
//
// THE SHAPE OF THE REPAIR is plan.md §C.2 option (ii): CoverageRule reads the
// sibling itself. Option (i) — merging acceptance.md into `doc.Criteria` inside
// parseSPECDoc — would reach every rule consuming `doc.Criteria`, so
// `ACIDInvalid` and friends would begin judging a document they were never
// written against. The blast radius is the whole difference between the two.
//
// A rule reaching past its single parsed document is the ESTABLISHED PATTERN in
// this package, not a new mechanism: HaikuResidualRule, MovingRefUnpinnedRule,
// and ArtifactStatusFieldForbiddenRule all open sibling artifacts, and all keep
// the per-SPEC shape so `lint.skip` and era demotion continue to apply
// unchanged. The path resolution and the missing-file handling below are copied
// from ArtifactStatusFieldForbiddenRule deliberately rather than reinvented.
//
// WHY THE WHOLE FILE, AND WHY ExtractRequirementMappings. The inline path is
// ParseAcceptanceCriteria, which is scoped twice over: findACSectionStart needs
// an `##` heading containing "acceptance", and parseSingleACLine needs the
// `AC-…:` colon form. BOTH scopings exist because spec.md is a mixed document
// in which prose must not be read as AC. acceptance.md is not mixed — the file
// IS the acceptance criteria, by name and by convention — so neither scoping
// carries its justification across, and applying them to the sibling would make
// this repair vacuous rather than conservative. Measured on the live corpus at
// the M4 baseline: 622 acceptance.md files exist, 169 carry an `##` heading
// containing "acceptance", and only 3 carry an inline-parseable `AC-…:` list
// line. This SPEC's own acceptance.md is one of the 619 that would parse to
// nothing — a repair that cannot cover the document that motivated it is a
// check whose non-execution is indistinguishable from its success.
//
// So the covered set is taken with ExtractRequirementMappings over the file's
// full text. That is the SAME function the inline AC parser uses for exactly
// this purpose (parseSingleACLine calls it on each AC line), applied without
// the two spec.md-specific scopings. It needs no heading and no AC-line shape.
//
// THE WIDENING IS ONE-DIRECTIONAL AND ADMITTED. A `maps REQ-…` written in prose
// inside acceptance.md counts as coverage here, where the inline path would not
// count the same text in spec.md. Two things bound the cost: acceptance.md is
// the AC SSOT, so a `maps REQ-…` in it is an AC mapping declaration by
// convention rather than by accident; and CoverageRule emits advisory `warning`
// (M3, plan.md §D option A), so the direction of the error is a report not made,
// never a gate wrongly closed. The opposite direction — a checker stricter than
// the convention it enforces — is the one that reddens a corpus on landing.
//
// NO NEW DUPLICATION OBLIGATION (REQ-CRS-001-008). Nothing here asks a Tier M/L
// spec.md to restate its AC. The inline set and the sibling set are UNIONED, so
// a SPEC that already covers a REQ inline is unaffected and a SPEC that covers
// it only in the sibling now counts. acceptance.md remains the SSOT.
//
// CORPUS DELTA IS ZERO, AND THAT IS MEASURED RATHER THAN HOPED. At the M4
// baseline the 846 CoverageIncomplete findings sit on 47 SPECs. 43 of those 47
// DO carry an acceptance.md, and 0 of the 43 declare a `maps REQ-…` mapping —
// 23 name AC ids with no machine-readable mapping at all, the rest neither.
// Conversely the 14 acceptance.md files that DO carry `maps REQ-…` belong to
// SPECs with 0 findings. The two populations are disjoint, so no corpus finding
// can move whatever this code does; the delta is decided by the corpus, not by
// the repair, and a zero here is not evidence the sibling goes unread.
//
// The residual is a corpus-authoring gap, NOT a narrowness of this predicate.
// Widening further — counting any bare `REQ-…` token in acceptance.md as
// coverage — would silence most of the 846 findings M3 deliberately surfaced,
// and would count a REQ named in prose as excluded as if it were covered. The
// mapping form is the coverage declaration; a file that declares none is
// uncovered, and saying so is the rule working.
//
// The discriminating evidence is therefore the fixture pair and its mutant
// probe, plus this SPEC's own acceptance.md once the spec.md §2.3 mirror table
// is removed: measured on a scratch copy, 8 CoverageIncomplete before this
// change and 0 after.

import (
	"os"
	"path/filepath"
)

// siblingAcceptanceArtifact is the sibling filename that carries a SPEC's
// acceptance criteria under the Tier M/L convention. It is a single fixed name
// rather than a scan: `design.md` and `research.md` are not AC documents, and
// enumerating keeps them out by construction instead of by an exclusion that
// could later be forgotten.
const siblingAcceptanceArtifact = "acceptance.md"

// siblingAcceptanceCoveredREQIDs returns the REQ IDs mapped by the acceptance
// criteria in the SPEC directory's sibling acceptance.md, keyed in the same
// `REQ-`-prefixed form collectAllREQIDs produces so the two sets union directly.
//
// An absent acceptance.md returns an empty set and no error: Tier S SPECs have
// none by design, and that is the dominant corpus shape (622 of 710 SPEC
// directories carry one, so 88 do not). An unreadable file is treated the same
// way — CoverageRule reports coverage, and a rule that fails a whole lint run
// because a sibling artifact could not be opened would convert an advisory
// report into an outage.
func siblingAcceptanceCoveredREQIDs(specPath string) map[string]bool {
	covered := make(map[string]bool)
	if specPath == "" {
		return covered
	}
	path := filepath.Join(filepath.Dir(specPath), siblingAcceptanceArtifact)
	data, err := os.ReadFile(path) // #nosec G304 -- path is derived from the discovered SPEC directory
	if err != nil {
		return covered // absent sibling is the Tier S shape, not a defect
	}
	for _, id := range ExtractRequirementMappings(string(data)) {
		covered["REQ-"+id] = true
	}
	return covered
}
