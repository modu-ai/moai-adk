package evolution_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

const sampleSkillWithStructure = `# Skill Header

Introduction content that must be preserved.

<!-- moai:evolvable-start id="best-practices" -->
Initial best practice content.
<!-- moai:evolvable-end -->

Footer content that must also be preserved.
`

// TestApplyProposal_PreservesSurroundingContent is an integration test verifying that
// ApplyProposal preserves header/footer content surrounding the zone.
// CRITICAL 3: verifies the fix for MergeEvolvableZones misuse (file destruction)
func TestApplyProposal_PreservesSurroundingContent(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	relPath := ".claude/skills/moai-lang-go/SKILL.md"
	fullPath := filepath.Join(projectRoot, relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(sampleSkillWithStructure), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	proposal := &evolution.ProposedChange{
		TargetFile: relPath,
		ZoneID:     "best-practices",
		Addition:   "- Always pass context.Context as first argument.\n",
	}

	if err := evolution.ApplyProposal(projectRoot, proposal); err != nil {
		t.Fatalf("ApplyProposal: %v", err)
	}

	updated, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read updated file: %v", err)
	}

	updatedStr := string(updated)

	// Header must be preserved
	if !strings.Contains(updatedStr, "# Skill Header") {
		t.Error("ApplyProposal: file header disappeared (CRITICAL 3 regression)")
	}

	// Introduction content must be preserved
	if !strings.Contains(updatedStr, "Introduction content that must be preserved.") {
		t.Error("ApplyProposal: introduction content disappeared (CRITICAL 3 regression)")
	}

	// Footer must be preserved
	if !strings.Contains(updatedStr, "Footer content that must also be preserved.") {
		t.Error("ApplyProposal: footer disappeared (CRITICAL 3 regression)")
	}

	// Original zone content must be preserved
	if !strings.Contains(updatedStr, "Initial best practice content.") {
		t.Error("ApplyProposal: original zone content disappeared")
	}

	// Added content must be present
	if !strings.Contains(updatedStr, "Always pass context.Context as first argument.") {
		t.Error("ApplyProposal: added content not found")
	}
}
