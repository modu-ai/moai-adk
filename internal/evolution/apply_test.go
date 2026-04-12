package evolution_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

const sampleSkillContent = `# Test Skill

Introduction paragraph.

<!-- moai:evolvable-start id="best-practices" -->
Initial best practice content.
<!-- moai:evolvable-end -->

Footer content.
`

// TestApplyProposal_AddsContentToZone verifies that ApplyProposal appends the
// addition to the evolvable zone and writes the file atomically.
func TestApplyProposal_AddsContentToZone(t *testing.T) {
	projectRoot := t.TempDir()
	relPath := ".claude/skills/moai-lang-go/SKILL.md"
	fullPath := filepath.Join(projectRoot, relPath)

	// Write the skill file.
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(sampleSkillContent), 0o644); err != nil {
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

	if !strings.Contains(string(updated), "Always pass context.Context") {
		t.Fatal("expected updated file to contain the added content")
	}
	if !strings.Contains(string(updated), "Initial best practice content.") {
		t.Fatal("expected updated file to preserve original zone content")
	}
}

// TestApplyProposal_CreatesBak verifies that a .bak backup is written.
func TestApplyProposal_CreatesBak(t *testing.T) {
	projectRoot := t.TempDir()
	relPath := ".claude/skills/moai-lang-go/SKILL.md"
	fullPath := filepath.Join(projectRoot, relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(sampleSkillContent), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	proposal := &evolution.ProposedChange{
		TargetFile: relPath,
		ZoneID:     "best-practices",
		Addition:   "addition\n",
	}

	if err := evolution.ApplyProposal(projectRoot, proposal); err != nil {
		t.Fatalf("ApplyProposal: %v", err)
	}

	if _, err := os.Stat(fullPath + ".bak"); err != nil {
		t.Fatalf("expected .bak file to exist: %v", err)
	}
}

// TestRevertProposal_RestoresOriginal verifies that RevertProposal undoes the
// change and removes the .bak file.
func TestRevertProposal_RestoresOriginal(t *testing.T) {
	projectRoot := t.TempDir()
	relPath := ".claude/skills/moai-lang-go/SKILL.md"
	fullPath := filepath.Join(projectRoot, relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(sampleSkillContent), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	proposal := &evolution.ProposedChange{
		TargetFile: relPath,
		ZoneID:     "best-practices",
		Addition:   "addition\n",
	}

	if err := evolution.ApplyProposal(projectRoot, proposal); err != nil {
		t.Fatalf("ApplyProposal: %v", err)
	}

	if err := evolution.RevertProposal(projectRoot, proposal); err != nil {
		t.Fatalf("RevertProposal: %v", err)
	}

	restored, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(restored) != sampleSkillContent {
		t.Fatalf("restored content mismatch:\nwant: %q\ngot:  %q", sampleSkillContent, string(restored))
	}

	// .bak should be removed after revert.
	if _, err := os.Stat(fullPath + ".bak"); !os.IsNotExist(err) {
		t.Fatal("expected .bak file to be removed after revert")
	}
}

// TestApplyProposal_BlocksFrozenFile verifies that frozen files are rejected.
func TestApplyProposal_BlocksFrozenFile(t *testing.T) {
	projectRoot := t.TempDir()
	proposal := &evolution.ProposedChange{
		TargetFile: "CLAUDE.md",
		ZoneID:     "some-zone",
		Addition:   "add this\n",
	}
	err := evolution.ApplyProposal(projectRoot, proposal)
	if err != evolution.ErrFrozenPath {
		t.Fatalf("expected ErrFrozenPath, got %v", err)
	}
}

// TestApplyProposal_NilProposal verifies that a nil proposal returns an error.
func TestApplyProposal_NilProposal(t *testing.T) {
	projectRoot := t.TempDir()
	err := evolution.ApplyProposal(projectRoot, nil)
	if err == nil {
		t.Fatal("expected error for nil proposal, got nil")
	}
}

// TestApplyProposal_ZoneNotFound verifies that ErrZoneNotFound is returned
// when the specified zone does not exist in the skill file.
func TestApplyProposal_ZoneNotFound(t *testing.T) {
	projectRoot := t.TempDir()
	relPath := ".claude/skills/moai-lang-go/SKILL.md"
	fullPath := filepath.Join(projectRoot, relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// File without any evolvable zone.
	if err := os.WriteFile(fullPath, []byte("# Skill\n\nNo zones here.\n"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	proposal := &evolution.ProposedChange{
		TargetFile: relPath,
		ZoneID:     "nonexistent-zone",
		Addition:   "add this\n",
	}
	err := evolution.ApplyProposal(projectRoot, proposal)
	if err != evolution.ErrZoneNotFound {
		t.Fatalf("expected ErrZoneNotFound, got %v", err)
	}
}

// TestArchiveOldLearnings verifies that excess entries are archived.
func TestArchiveOldLearnings(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	const total = 5
	const maxActive = 3

	for i := 0; i < total; i++ {
		id := fmt.Sprintf("LEARN-20260411-%03d", i+1)
		if err := evolution.CreateLearning(projectRoot, sampleEntry(id)); err != nil {
			t.Fatalf("setup: CreateLearning %q: %v", id, err)
		}
	}

	if err := evolution.ArchiveOldLearnings(projectRoot, maxActive); err != nil {
		t.Fatalf("ArchiveOldLearnings: %v", err)
	}

	active, err := evolution.ListLearnings(projectRoot, evolution.LearningFilter{ExcludeArchived: true})
	if err != nil {
		t.Fatalf("ListLearnings: %v", err)
	}
	if len(active) != maxActive {
		t.Fatalf("want %d active entries after archive, got %d", maxActive, len(active))
	}

	all, err := evolution.ListLearnings(projectRoot, evolution.LearningFilter{})
	if err != nil {
		t.Fatalf("ListLearnings all: %v", err)
	}
	if len(all) != total {
		t.Fatalf("want %d total entries, got %d", total, len(all))
	}
}

// TestMarshalUnmarshal_RoundTrip verifies that a learning entry survives a
// full serialisation round-trip with all fields populated.
func TestMarshalUnmarshal_RoundTrip(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	entry := sampleEntry("LEARN-20260411-RT1")
	entry.Observation = "test observation with special chars: <>&"
	entry.Evidence = []evolution.EvidenceEntry{
		{SessionID: "s1", Date: "2026-04-11", Context: "success: all good"},
		{SessionID: "s2", Date: "2026-04-11", Context: "error: build failed"},
	}
	entry.ProposedChange = &evolution.ProposedChange{
		TargetFile: ".claude/skills/foo/SKILL.md",
		ZoneID:     "best-practices",
		Addition:   "line one\nline two\n",
	}

	if err := evolution.CreateLearning(projectRoot, entry); err != nil {
		t.Fatalf("CreateLearning: %v", err)
	}

	loaded, err := evolution.LoadLearningByID(projectRoot, entry.ID)
	if err != nil {
		t.Fatalf("LoadLearningByID: %v", err)
	}

	if loaded.SkillID != entry.SkillID {
		t.Errorf("SkillID: want %q, got %q", entry.SkillID, loaded.SkillID)
	}
	if loaded.ZoneID != entry.ZoneID {
		t.Errorf("ZoneID: want %q, got %q", entry.ZoneID, loaded.ZoneID)
	}
	if loaded.Observations != entry.Observations {
		t.Errorf("Observations: want %d, got %d", entry.Observations, loaded.Observations)
	}
	if len(loaded.Evidence) != len(entry.Evidence) {
		t.Errorf("Evidence length: want %d, got %d", len(entry.Evidence), len(loaded.Evidence))
	}
	if loaded.ProposedChange == nil {
		t.Fatal("expected ProposedChange to be loaded")
	}
	if loaded.ProposedChange.TargetFile != entry.ProposedChange.TargetFile {
		t.Errorf("ProposedChange.TargetFile: want %q, got %q",
			entry.ProposedChange.TargetFile, loaded.ProposedChange.TargetFile)
	}
}
