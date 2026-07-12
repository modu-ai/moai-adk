package curator

import (
	"regexp"
	"strings"
	"testing"
)

// --- AC-HEV2-007: LEARNED-WORKFLOW marker regex matches heading + start/end
// markers as one atomic group (consistent with layer3.go markerBlockPattern) ---

func TestLearnedWorkflowMarker_RegexMatchesHeadingPlusMarkers(t *testing.T) {
	// A canonical populated block.
	block := strings.Join([]string{
		"## MOAI:LEARNED-WORKFLOW",
		"<!-- moai:learned-start -->",
		"- distilled rule one <!-- key: k1 -->",
		"- distilled rule two <!-- key: k2 -->",
		"<!-- moai:learned-end -->",
		"",
	}, "\n")

	re, ok := compiledPatterns[BlockTypeLearnedWorkflow]
	if !ok {
		t.Fatal("BlockTypeLearnedWorkflow missing from compiledPatterns registry")
	}
	if !re.MatchString(block) {
		t.Errorf("compiled pattern did not match canonical block:\n%s", block)
	}

	// The match MUST include the heading, the start marker, the body, AND the
	// end marker (atomic group) — replacing the match removes the whole block.
	match := re.FindString(block)
	for _, frag := range []string{
		"## MOAI:LEARNED-WORKFLOW",
		"<!-- moai:learned-start -->",
		"distilled rule one",
		"<!-- moai:learned-end -->",
	} {
		if !strings.Contains(match, frag) {
			t.Errorf("atomic match missing %q; match=\n%s", frag, match)
		}
	}
}

// --- AC-HEV2-008: marker atomic match group — start marker may carry attrs,
// body is non-greedy up to the FIRST end marker ---

func TestLearnedWorkflowMarker_AtomicMatchGroup(t *testing.T) {
	re := compiledPatterns[BlockTypeLearnedWorkflow]

	t.Run("start_marker_with_attributes", func(t *testing.T) {
		// The start marker may carry attributes (e.g. a tier or provenance
		// token) inside the HTML comment before the closing -->. The regex
		// [^>]*--> absorbs them.
		block := "## MOAI:LEARNED-WORKFLOW\n" +
			"<!-- moai:learned-start tier=\"4\" -->\n" +
			"- rule\n" +
			"<!-- moai:learned-end -->\n"
		if !re.MatchString(block) {
			t.Errorf("pattern did not match start marker with attrs:\n%s", block)
		}
	})

	t.Run("non_greedy_body_stops_at_first_end_marker", func(t *testing.T) {
		// Two consecutive blocks: the non-greedy body (.*?) must stop at the
		// FIRST end marker, so the second block remains intact after a replace.
		two := "## MOAI:LEARNED-WORKFLOW\n" +
			"<!-- moai:learned-start -->\n" +
			"- first block rule\n" +
			"<!-- moai:learned-end -->\n" +
			"\n" +
			"## MOAI:LEARNED-WORKFLOW\n" +
			"<!-- moai:learned-start -->\n" +
			"- second block rule\n" +
			"<!-- moai:learned-end -->\n"
		// Replace only the first match with a sentinel; the second block must survive.
		replaced := re.ReplaceAllString(two, "@@REPLACED@@")
		if c := strings.Count(replaced, "@@REPLACED@@"); c != 2 {
			t.Errorf("expected 2 atomic matches (ReplaceAll), got %d", c)
		}
		if strings.Contains(replaced, "first block rule") {
			t.Error("non-greedy body over-matched: first block body survived replace")
		}
	})

	t.Run("heading_without_markers_not_matched", func(t *testing.T) {
		// A bare heading with no marker pair MUST NOT match (prevents treating
		// prose as a managed block).
		bare := "## MOAI:LEARNED-WORKFLOW\n\ndescriptive prose\n"
		if re.MatchString(bare) {
			t.Error("pattern matched a heading with no start/end marker pair")
		}
	})
}

// TestLearnedWorkflowMarker_PatternCompiles is a defensive guard: the marker
// registry pattern must be a valid regexp (panics at init if not, but this
// makes the contract explicit and grep-visible).
func TestLearnedWorkflowMarker_PatternCompiles(t *testing.T) {
	spec, ok := markerRegistry[BlockTypeLearnedWorkflow]
	if !ok {
		t.Fatal("BlockTypeLearnedWorkflow missing from markerRegistry")
	}
	pattern := "(?s)" +
		regexp.QuoteMeta(spec.heading) + `\n` +
		regexp.QuoteMeta(spec.startPrefix) + `[^>]*-->` +
		`.*?` +
		regexp.QuoteMeta(spec.endMarker)
	if _, err := regexp.Compile(pattern); err != nil {
		t.Fatalf("marker pattern does not compile: %v", err)
	}
}

// --- AC-HEV2-019: LEARNED-WORKFLOW-LOCAL marker regex matches heading +
// start/end markers as one atomic group (REQ-HEV2-013 / spec.md §C.3). ---
//
// The Tier-3 append-only surface in CLAUDE.local.md carries a DISTINCT marker
// contract from the digest block: the heading is suffixed `-LOCAL` and the
// start/end markers use the `learned-local-*` token. The compiled pattern must
// match the LOCAL block atomically and MUST NOT cross-match the digest block.

func TestLearnedLocalMarker_RegexMatchesHeadingPlusMarkers(t *testing.T) {
	// A canonical populated LOCAL block.
	block := strings.Join([]string{
		"## MOAI:LEARNED-WORKFLOW-LOCAL",
		"<!-- moai:learned-local-start -->",
		"- observed pattern alpha <!-- key: local-1 -->",
		"- observed pattern beta <!-- key: local-2 -->",
		"<!-- moai:learned-local-end -->",
		"",
	}, "\n")

	re, ok := compiledPatterns[BlockTypeLearnedLocal]
	if !ok {
		t.Fatal("BlockTypeLearnedLocal missing from compiledPatterns registry")
	}
	if !re.MatchString(block) {
		t.Errorf("compiled pattern did not match canonical LOCAL block:\n%s", block)
	}

	// The atomic match MUST include the heading, start marker, body, AND end marker.
	match := re.FindString(block)
	for _, frag := range []string{
		"## MOAI:LEARNED-WORKFLOW-LOCAL",
		"<!-- moai:learned-local-start -->",
		"observed pattern alpha",
		"<!-- moai:learned-local-end -->",
	} {
		if !strings.Contains(match, frag) {
			t.Errorf("atomic LOCAL match missing %q; match=\n%s", frag, match)
		}
	}
}

func TestLearnedLocalMarker_DoesNotMatchDigestBlock(t *testing.T) {
	// The LOCAL pattern MUST NOT match the digest (non-LOCAL) block — the two
	// marker contracts are disjoint and a cross-match would let the append-only
	// writer corrupt the digest surface.
	localRe := compiledPatterns[BlockTypeLearnedLocal]
	digestBlock := strings.Join([]string{
		"## MOAI:LEARNED-WORKFLOW",
		"<!-- moai:learned-start -->",
		"- distilled rule <!-- key: k1 -->",
		"<!-- moai:learned-end -->",
		"",
	}, "\n")
	if localRe.MatchString(digestBlock) {
		t.Error("LOCAL pattern matched the digest (non-LOCAL) block — marker contracts are not disjoint")
	}
}

func TestLearnedLocalMarker_PatternCompiles(t *testing.T) {
	spec, ok := markerRegistry[BlockTypeLearnedLocal]
	if !ok {
		t.Fatal("BlockTypeLearnedLocal missing from markerRegistry")
	}
	pattern := "(?s)" +
		regexp.QuoteMeta(spec.heading) + `\n` +
		regexp.QuoteMeta(spec.startPrefix) + `[^>]*-->` +
		`.*?` +
		regexp.QuoteMeta(spec.endMarker)
	if _, err := regexp.Compile(pattern); err != nil {
		t.Fatalf("LOCAL marker pattern does not compile: %v", err)
	}
}
