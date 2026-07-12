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
