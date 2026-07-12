package curator

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeLocalFixture writes a CLAUDE.local.md carrying an EMPTY
// MOAI:LEARNED-WORKFLOW-LOCAL block (the Tier-3 append surface) under a
// t.TempDir()-isolated directory and returns its absolute path.
func writeLocalFixture(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.local.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write local fixture: %v", err)
	}
	return p
}

// emptyLocalBlock is a CLAUDE.local.md fixture with the empty Tier-3 append
// surface markers present (AppendLearnedLocal requires the block to exist).
const emptyLocalBlock = "# Local\n\n## MOAI:LEARNED-WORKFLOW-LOCAL\n" +
	"<!-- moai:learned-local-start -->\n<!-- moai:learned-local-end -->\n"

// TestTier4Qualified_ClaudeMdWrite (AC-HEV2-032, Scenario 7): a Tier-4-qualified
// pattern (P4 = 10 observations) writes to the CLAUDE.md managed-block surface.
func TestTier4Qualified_ClaudeMdWrite(t *testing.T) {
	path := writeFixture(t, "# Project\n")
	err := TierGatedWrite(path, BlockTypeLearnedWorkflow, 10, BlockContent{
		Tier:    4,
		Bullets: []Bullet{{LedgerKey: "p4", Text: "prefer sequential sub-agents for coding work"}},
	})
	if err != nil {
		t.Fatalf("Tier-4 qualified write should succeed, got %v", err)
	}
	data := string(mustRead(t, path))
	if !strings.Contains(data, "## MOAI:LEARNED-WORKFLOW") {
		t.Errorf("CLAUDE.md missing the LEARNED-WORKFLOW block after the Tier-4 write")
	}
	if !strings.Contains(data, "prefer sequential sub-agents for coding work") {
		t.Errorf("CLAUDE.md missing the Tier-4 bullet text")
	}
}

// TestTier3Qualified_ClaudeLocalMdAppend (AC-HEV2-033, Scenario 7): a
// Tier-3-qualified pattern (P3 = 5 observations) appends to the CLAUDE.local.md
// append surface, landing the bullet inside the LOCAL block markers.
func TestTier3Qualified_ClaudeLocalMdAppend(t *testing.T) {
	path := writeLocalFixture(t, emptyLocalBlock)
	err := TierGatedWrite(path, BlockTypeLearnedLocal, 5, BlockContent{
		Tier:    3,
		Bullets: []Bullet{{LedgerKey: "p3", Text: "record the project-local deploy quirk"}},
	})
	if err != nil {
		t.Fatalf("Tier-3 qualified append should succeed, got %v", err)
	}
	data := string(mustRead(t, path))
	startIdx := strings.Index(data, "moai:learned-local-start")
	bulletIdx := strings.Index(data, "record the project-local deploy quirk")
	endIdx := strings.Index(data, "moai:learned-local-end")
	if bulletIdx < 0 {
		t.Fatalf("CLAUDE.local.md missing the Tier-3 appended bullet")
	}
	if startIdx >= bulletIdx || bulletIdx >= endIdx {
		t.Errorf("appended bullet not within the LOCAL block markers (start=%d bullet=%d end=%d)",
			startIdx, bulletIdx, endIdx)
	}
}

// TestUnderTierWrite_ErrTierNotQualified (AC-HEV2-034, Scenario 7): a
// 6-observation pattern (P_under) qualifies for Tier 3 only. A direct Tier-4
// write attempt returns ErrTierNotQualified WITHOUT touching the file (no
// self-tier-escalation, REQ-HEV2-027); the same pattern IS accepted at Tier 3.
func TestUnderTierWrite_ErrTierNotQualified(t *testing.T) {
	// Direct Tier-4 write attempt for a 6-observation pattern → rejected.
	claudeMd := writeFixture(t, "# Project\n")
	before := mustRead(t, claudeMd)
	err := TierGatedWrite(claudeMd, BlockTypeLearnedWorkflow, 6, BlockContent{
		Tier:    4,
		Bullets: []Bullet{{LedgerKey: "under", Text: "not yet Tier-4 qualified"}},
	})
	if !errors.Is(err, ErrTierNotQualified) {
		t.Fatalf("6-obs Tier-4 write should return ErrTierNotQualified, got %v", err)
	}
	if string(before) != string(mustRead(t, claudeMd)) {
		t.Errorf("CLAUDE.md modified despite tier-not-qualified rejection")
	}

	// The SAME 6-observation pattern IS qualified for Tier 3 — the append surface
	// accepts it (Scenario 7: "neither reaches Tier 4 nor triggers
	// ErrTierNotQualified when written to Tier 3").
	local := writeLocalFixture(t, emptyLocalBlock)
	if err := TierGatedWrite(local, BlockTypeLearnedLocal, 6, BlockContent{
		Tier:    3,
		Bullets: []Bullet{{LedgerKey: "under", Text: "qualified for Tier 3"}},
	}); err != nil {
		t.Fatalf("6-obs Tier-3 append should succeed, got %v", err)
	}
	if !strings.Contains(string(mustRead(t, local)), "qualified for Tier 3") {
		t.Errorf("CLAUDE.local.md missing the Tier-3 bullet for the 6-obs pattern")
	}
}
