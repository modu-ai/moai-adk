package merge

import (
	"strings"
	"testing"
)

// TestMergeRoundTrip_PopulatedLearnedBlockSurvivesTemplateSync
// (REQ-HEV2-019/020, M7 template-merge round-trip): a full `moai update`-style
// round-trip. The local CLAUDE.md carries a MULTI-bullet populated
// MOAI:LEARNED-WORKFLOW block plus surrounding real content; the upstream
// template ships the EMPTY marker plus an updated unrelated section. After
// mergeSectionBased the populated local block survives verbatim (every bullet
// preserved, no clobber, exactly one marker pair) AND the unrelated template
// update lands. This complements the minimal M4 preservation cases with a
// realistic full-file round-trip.
func TestMergeRoundTrip_PopulatedLearnedBlockSurvivesTemplateSync(t *testing.T) {
	t.Parallel()

	// base = the template CLAUDE.md the project was last synced from (empty marker).
	base := []byte(
		"# Project Guide\n\n" +
			"## Workflow\nold workflow guidance\n\n" +
			"## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n<!-- moai:learned-end -->\n")

	// current = the LOCAL CLAUDE.md, whose learned block the Curator populated.
	current := []byte(
		"# Project Guide\n\n" +
			"## Workflow\nold workflow guidance\n\n" +
			"## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n" +
			"- Run make build after editing templates. <!-- ledger_key: lw-build-001 -->\n" +
			"- Prefer sequential sub-agents for coding work. <!-- ledger_key: lw-seq-002 -->\n" +
			"- Validate cross-file reachability before declaring PASS. <!-- ledger_key: lw-reach-003 -->\n" +
			"<!-- moai:learned-end -->\n")

	// updated = the NEW upstream template (still ships the empty marker; updates
	// the unrelated Workflow section).
	updated := []byte(
		"# Project Guide\n\n" +
			"## Workflow\nnew upstream workflow guidance\n\n" +
			"## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n<!-- moai:learned-end -->\n")

	result, err := mergeSectionBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Fatalf("expected no conflict for empty-upstream populated-local round-trip; conflicts=%+v", result.Conflicts)
	}

	got := string(result.Content)

	// Every populated local bullet MUST survive verbatim (no clobber).
	for _, want := range []string{
		"Run make build after editing templates.", "ledger_key: lw-build-001",
		"Prefer sequential sub-agents for coding work.", "ledger_key: lw-seq-002",
		"Validate cross-file reachability before declaring PASS.", "ledger_key: lw-reach-003",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("populated local learned bullet not preserved: %q missing from:\n%s", want, got)
		}
	}
	// The unrelated upstream section update MUST land.
	if !strings.Contains(got, "new upstream workflow guidance") {
		t.Errorf("unrelated upstream template update not reflected")
	}
	if strings.Contains(got, "old workflow guidance") {
		t.Errorf("stale local workflow guidance not updated from upstream")
	}
	// Exactly one marker pair (no duplication / shadowing by the empty upstream).
	if c := strings.Count(got, "moai:learned-start"); c != 1 {
		t.Errorf("expected exactly 1 learned-start marker, got %d (clobber/duplication suspected)", c)
	}
	if c := strings.Count(got, "moai:learned-end"); c != 1 {
		t.Errorf("expected exactly 1 learned-end marker, got %d", c)
	}
	if result.Strategy != SectionMerge {
		t.Errorf("Strategy = %q, want %q", result.Strategy, SectionMerge)
	}
}
