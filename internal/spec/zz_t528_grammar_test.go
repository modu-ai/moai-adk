// zz_t528_grammar_test.go — t528 / SPEC-AC-COLLECTOR-ANCHOR-001.
//
// Axis tests for the widened AC item grammar in parseSingleACLine.
//
// EVERY fixture line below is copied VERBATIM from the corpus, with its source
// path and line number in the comment. Nothing here is invented: a fixture line
// that does not occur in the corpus proves the grammar accepts a shape, but says
// nothing about whether the corpus uses it — and this card's whole justification
// is corpus reach. The one deliberate exception is the negative case
// AC-FOO-BAR (§D.3), which must NOT parse and therefore has no corpus source by
// construction, and the en-dash separator (§D.9), which acceptance.md records as
// a design decision with no corpus observation. Both are marked at their site.
package spec

import (
	"strings"
	"testing"
)

// t528Section wraps fixture bullets in the minimal document the collector needs:
// an "## ... Acceptance ..." heading, since findACSectionStart gates on it and
// this card does not touch that axis.
func t528Section(bullets ...string) string {
	return "# Fixture\n\n## Acceptance Criteria\n\n" + strings.Join(bullets, "\n") + "\n"
}

// t528RootIDs returns the root-level AC IDs the collector produced.
func t528RootIDs(t *testing.T, md string) []string {
	t.Helper()
	criteria, _ := ParseAcceptanceCriteria(md, false)
	var ids []string
	for _, c := range criteria {
		ids = append(ids, c.ID)
	}
	return ids
}

func t528Has(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

// TestT528SegmentGrammar — REQ-ACA-001-001 / AC-ACA-001-003.
// Variable segment count (2..5) and alphanumeric middle segments, with the last
// segment pinned numeric.
func TestT528SegmentGrammar(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{
			// .moai/specs/SPEC-CC2122-HOOK-002/spec.md:70 — 2 segments
			name: "two_segments",
			line: "- AC-001: `internal/hook/post_tool_duration.go` 가 helper 를 호출한다",
			want: "AC-001",
		},
		{
			// .moai/specs/SPEC-V3R2-RT-001/spec.md:148 — 5 segments, alnum middles
			name: "five_segments_alnum_middle",
			line: "- AC-V3R2-RT-001-01: WHEN a PreToolUse hook wrapper writes stdout, THE system SHALL replace the pending tool input.",
			want: "AC-V3R2-RT-001-01",
		},
		{
			// .moai/specs/SPEC-V3R2-WF-001/spec.md:174 — 3 segments, alnum middle
			name: "three_segments_alnum_middle",
			line: "- **AC-WF001-01**: Given the v2.13.2 tree with 48 skills When the full Stage 1 consolidation is applied Then `.claude/skills/` contains exactly 38 directories (maps REQ-WF001-001).",
			want: "AC-WF001-01",
		},
		{
			// .moai/specs/SPEC-V3R2-ORC-001/spec.md:209 — 4 segments
			name: "four_segments",
			line: "- **AC-ORC-001-01**: Counting active (non-stub) agent files under `.claude/agents/moai/` yields exactly 17, matching the REQ-001 list.",
			want: "AC-ORC-001-01",
		},
		{
			// .moai/specs/SPEC-V3R2-MIG-001/spec.md:165
			name: "alnum_middle_mig",
			line: "- **AC-MIG001-01**: Given a v2 project When `moai migrate v2-to-v3 --dry-run` runs Then dry-run report is produced without file modification (maps REQ-MIG001-009).",
			want: "AC-MIG001-01",
		},
		{
			// .moai/specs/SPEC-AUDIT-SNAPSHOT-001/spec.md:131 — 4 segments, word middles
			name: "word_middle_segments",
			line: "- AC-AUDIT-SNAPSHOT-001 (A1): sticky cache — past-24h unchanged-hash skip still fires.",
			want: "AC-AUDIT-SNAPSHOT-001",
		},
		{
			// .moai/specs/SPEC-SPC-001 family — the shape the CURRENT anchor already
			// accepts. Present so a regression in the accepted set fails here too.
			name: "already_accepted_three_segment",
			line: "- AC-SPC-001-01: Given a SPEC with hierarchical acceptance criteria, When parsed, Then a tree is produced.",
			want: "AC-SPC-001-01",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids := t528RootIDs(t, t528Section(tc.line))
			if !t528Has(ids, tc.want) {
				t.Errorf("declaration not collected\n  line: %s\n  want id: %s\n  got ids: %v", tc.line, tc.want, ids)
			}
		})
	}
}

// TestT528NumericTailRequired — REQ-ACA-001-001 / AC-ACA-001-003, negative case.
//
// The last segment stays pinned numeric. Without this case the axis test above
// is also satisfied by an implementation that accepts any AC-prefixed token, and
// then AC-ACA-001-003 is vacuous. This fixture is DELIBERATELY not a corpus
// line: the claim is about a shape the grammar must refuse.
func TestT528NumericTailRequired(t *testing.T) {
	ids := t528RootIDs(t, t528Section("- AC-FOO-BAR: this is not a declaration id the grammar admits."))
	if t528Has(ids, "AC-FOO-BAR") {
		t.Errorf("non-numeric-tail token was collected as a declaration: got ids %v", ids)
	}
}

// TestT528SubIDSuffixPreserved — REQ-ACA-001-002 / AC-ACA-001-004.
//
// Recognition alone is not the claim. hasIDSuffix / autoWrapSingle branch on the
// dot, so the assertion is on the TREE SHAPE: a suffixed root stays unwrapped, a
// non-suffixed root is still auto-wrapped with a ".a" child. The fixture mixes
// both in one section — unmixed, the wrapping branch is never exercised.
func TestT528SubIDSuffixPreserved(t *testing.T) {
	md := t528Section(
		"- AC-SPC-001-01: Given a SPEC with hierarchical acceptance criteria, When parsed, Then a tree is produced.",
		"- AC-SPC-001-02.a: Given a flat legacy SPEC, When parsed, Then the flat branch is taken.",
		"- AC-SPC-001-03.a.i: Given a nested sub-id, When parsed, Then the deepest suffix is preserved.",
	)
	criteria, _ := ParseAcceptanceCriteria(md, false)

	shapes := map[string]string{}
	for _, c := range criteria {
		shapes[c.ID] = t528TreeShape(c)
	}

	want := map[string]string{
		"AC-SPC-001-01":     "AC-SPC-001-01(AC-SPC-001-01.a)", // no suffix -> auto-wrapped
		"AC-SPC-001-02.a":   "AC-SPC-001-02.a",                // suffix -> left alone
		"AC-SPC-001-03.a.i": "AC-SPC-001-03.a.i",              // nested suffix -> left alone
	}
	for id, wantShape := range want {
		got, ok := shapes[id]
		if !ok {
			t.Errorf("declaration %s not collected; got shapes %v", id, shapes)
			continue
		}
		if got != wantShape {
			t.Errorf("tree shape changed for %s\n  want: %s\n  got:  %s", id, wantShape, got)
		}
	}
}

// TestT528BoldWrapper — REQ-ACA-001-003 / AC-ACA-001-005.
//
// The opening ** is already removed by the existing strings.TrimLeft(trimmed,
// "- *"); the CLOSING ** survives and lands between the id and the separator.
func TestT528BoldWrapper(t *testing.T) {
	cases := []struct{ name, line, want string }{
		{
			// .moai/specs/SPEC-ACHWD-STRIP-EXEMPT-001/spec.md:54
			name: "bold_id_emdash_separator",
			line: "- **AC-ASE-001** — **Given** the run-phase commit on `WT-achwd-strip-exempt`,",
			want: "AC-ASE-001",
		},
		{
			// .moai/specs/SPEC-ACHWD-STRIP-EXEMPT-001/spec.md:61
			name: "bold_id_emdash_separator_2",
			line: "- **AC-ASE-002** — **Given** the amended tree, **When** the §I.4 A4 perl guard runs,",
			want: "AC-ASE-002",
		},
		{
			// .moai/specs/SPEC-V3R2-ORC-001/spec.md:213 — bold id, colon separator
			name: "bold_id_colon_separator",
			line: "- **AC-ORC-001-05**: All 7 stub files exist at deprecated paths with status=retired frontmatter.",
			want: "AC-ORC-001-05",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids := t528RootIDs(t, t528Section(tc.line))
			if !t528Has(ids, tc.want) {
				t.Errorf("bold-wrapped declaration not collected\n  line: %s\n  want id: %s\n  got ids: %v", tc.line, tc.want, ids)
			}
		})
	}
}

// TestT528ParenQualifier — REQ-ACA-001-004 / AC-ACA-001-008.
func TestT528ParenQualifier(t *testing.T) {
	cases := []struct{ name, line, want string }{
		{
			// .moai/specs/SPEC-AUDIT-SNAPSHOT-001/spec.md:131
			name: "short_qualifier",
			line: "- AC-AUDIT-SNAPSHOT-001 (A1): sticky cache — past-24h unchanged-hash skip still fires.",
			want: "AC-AUDIT-SNAPSHOT-001",
		},
		{
			// .moai/specs/SPEC-AUTONOMY-TIERS-001/spec.md:162
			name: "req_qualifier",
			line: "- AC-AUTONOMY-TIERS-001 (REQ-001): `moai init` wizard offers 3-tier selection; `--autonomy-tier` flag validates the closed set.",
			want: "AC-AUTONOMY-TIERS-001",
		},
		{
			// .moai/specs/SPEC-V3R3-RETIRED-AGENT-001/spec.md:303 — bold id + qualifier
			name: "bold_id_with_req_qualifier",
			line: "- **AC-RA-07** (REQ-RA-007): retired-rejection guard returns proper JSON + exit 2",
			want: "AC-RA-07",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids := t528RootIDs(t, t528Section(tc.line))
			if !t528Has(ids, tc.want) {
				t.Errorf("qualified declaration not collected\n  line: %s\n  want id: %s\n  got ids: %v", tc.line, tc.want, ids)
			}
		})
	}
}

// TestT528ParenQualifierLeavesREQMappingAlone — AC-ACA-001-008, second clause.
//
// A "(REQ-001)" qualifier is not the "maps REQ-..." form ExtractRequirementMappings
// reads. Whether the qualifier is consumed or left in the content could move the
// mapping result, so the mapping is asserted explicitly rather than assumed.
func TestT528ParenQualifierLeavesREQMappingAlone(t *testing.T) {
	// .moai/specs/SPEC-V3R2-WF-001/spec.md:174 — a "maps REQ-" line the current
	// anchor already accepts, used as the control for the mapping path.
	control := "- **AC-WF001-01**: Given the v2.13.2 tree When the consolidation is applied Then 38 directories remain (maps REQ-WF001-001)."
	criteria, _ := ParseAcceptanceCriteria(t528Section(control), false)
	if len(criteria) == 0 {
		t.Fatalf("control declaration not collected at all")
	}
	found := false
	for _, c := range criteria {
		if c.ID != "AC-WF001-01" {
			continue
		}
		found = true
		got := c.RequirementIDs
		for _, ch := range c.Children {
			if len(got) == 0 {
				got = ch.RequirementIDs
			}
		}
		// ExtractRequirementMappings strips the "REQ-" prefix (ears.go:130 captures
		// the group AFTER it), so the stored form is "WF001-001". That is
		// pre-existing behaviour, not something this widening introduced: the
		// pre-widening before-image shows the same bare form
		// (.moai/reports/t528/probe/before/reqmap.txt, e.g. "AST-001-001").
		if len(got) != 1 || got[0] != "WF001-001" {
			t.Errorf("maps REQ mapping changed: want [WF001-001], got %v", got)
		}
	}
	if !found {
		t.Errorf("control declaration AC-WF001-01 missing from result")
	}
}

// TestT528SeparatorSet — REQ-ACA-001-005 / AC-ACA-001-009.
//
// Separators are compared by CODE POINT, not by eye: em dash and en dash are
// visually near-identical and mixing them up is a silent fixture defect.
func TestT528SeparatorSet(t *testing.T) {
	const (
		colon  = ":" // :
		emDash = "—" // — observed 93 times (discriminator B)
		enDash = "–" // – NOT observed in the corpus; a design decision only
	)
	cases := []struct{ name, sep, want string }{
		{"colon", colon, "AC-AUDIT-SNAPSHOT-002"},
		{"em_dash_u2014", emDash, "AC-AUDIT-SNAPSHOT-002"},
		{"en_dash_u2013", enDash, "AC-AUDIT-SNAPSHOT-002"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Body copied from .moai/specs/SPEC-AUDIT-SNAPSHOT-001/spec.md:132;
			// only the separator varies, so the separator is the sole variable.
			line := "- AC-AUDIT-SNAPSHOT-002 " + tc.sep + " per-tier skip threshold: a 0.78 Tier M SPEC is skip-eligible."
			ids := t528RootIDs(t, t528Section(line))
			if !t528Has(ids, tc.want) {
				t.Errorf("declaration with separator %q (U+%04X) not collected: got ids %v", tc.sep, []rune(tc.sep)[0], ids)
			}
		})
	}
}

// TestT528SectionScopingInvariant — REQ-ACA-001-006 / AC-ACA-001-006 / M5.
//
// Both clauses in one fixture: an in-section declaration IS collected, and an
// identically-shaped out-of-section declaration is NOT. The in-section control
// has to live in the same document — with only the outside line present, a
// collection of 0 does not separate "scoping held" from "the grammar rejected it".
func TestT528SectionScopingInvariant(t *testing.T) {
	md := "# Fixture\n\n" +
		"## Background\n\n" +
		// identical shape to the in-section line below, but outside the AC section
		"- AC-AUDIT-SNAPSHOT-009 (A9): outside the acceptance section entirely.\n\n" +
		"## Acceptance Criteria\n\n" +
		"- AC-AUDIT-SNAPSHOT-001 (A1): sticky cache — past-24h unchanged-hash skip still fires.\n"

	ids := t528RootIDs(t, md)
	if !t528Has(ids, "AC-AUDIT-SNAPSHOT-001") {
		t.Errorf("in-section declaration was NOT collected; got ids %v", ids)
	}
	if t528Has(ids, "AC-AUDIT-SNAPSHOT-009") {
		t.Errorf("out-of-section declaration WAS collected — section scoping broke; got ids %v", ids)
	}
}

// The over-acceptance guard (REQ-ACA-001-012 / AC-ACA-001-014) lands at M6, in
// zz_t528_overacceptance_test.go. It is deliberately absent here: its fixture is
// a roster of real non-declaration bullets sampled from the newly-accepted set,
// and that set does not exist until the widening has been measured. An empty
// roster written now would be a passing test over an empty set.
