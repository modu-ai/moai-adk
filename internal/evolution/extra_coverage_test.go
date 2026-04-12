// extra_coverage_test.go covers error paths and edge cases not tested in the
// primary test files.  All tests use t.TempDir() for isolation.
package evolution_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

// TestLoadLearning_MissingFile verifies that LoadLearning returns nil, nil for
// a non-existent file.
func TestLoadLearning_MissingFile(t *testing.T) {
	entry, err := evolution.LoadLearning("/tmp/no-such-file-ever.md")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if entry != nil {
		t.Fatalf("expected nil entry, got %+v", entry)
	}
}

// TestUpdateLearning_MissingID verifies that UpdateLearning returns an error
// when the ID does not exist.
func TestUpdateLearning_MissingID(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	err := evolution.UpdateLearning(projectRoot, "LEARN-NONEXISTENT", func(e *evolution.LearningEntry) {
		e.Observations++
	})
	if err == nil {
		t.Fatal("expected error for missing learning ID, got nil")
	}
}

// TestListLearnings_FilterByStatus verifies that ListLearnings correctly
// filters entries by status.
func TestListLearnings_FilterByStatus(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	obs := sampleEntry("LEARN-20260411-F01")
	obs.Status = evolution.StatusObservation
	if err := evolution.CreateLearning(projectRoot, obs); err != nil {
		t.Fatalf("create obs: %v", err)
	}

	rule := sampleEntry("LEARN-20260411-F02")
	rule.Status = evolution.StatusRule
	rule.Confidence = 0.9
	rule.Observations = 5
	if err := evolution.CreateLearning(projectRoot, rule); err != nil {
		t.Fatalf("create rule: %v", err)
	}

	rules, err := evolution.ListLearnings(projectRoot, evolution.LearningFilter{Status: evolution.StatusRule})
	if err != nil {
		t.Fatalf("ListLearnings: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule entry, got %d", len(rules))
	}
	if rules[0].ID != rule.ID {
		t.Errorf("expected %q, got %q", rule.ID, rules[0].ID)
	}
}

// TestListLearnings_FilterBySkillID verifies skill-based filtering.
func TestListLearnings_FilterBySkillID(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	for _, sk := range []string{"skill-a", "skill-b", "skill-b"} {
		e := sampleEntry(fmt.Sprintf("LEARN-%s-%d", sk, time.Now().Nanosecond()%9999))
		e.SkillID = sk
		if err := evolution.CreateLearning(projectRoot, e); err != nil {
			t.Fatalf("create %s: %v", sk, err)
		}
	}

	entries, err := evolution.ListLearnings(projectRoot, evolution.LearningFilter{SkillID: "skill-b"})
	if err != nil {
		t.Fatalf("ListLearnings: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 skill-b entries, got %d", len(entries))
	}
}

// TestWriteRateState_AtomicWriteAndRead verifies the atomic write and read
// of rate state.
func TestWriteRateState_AtomicWriteAndRead(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	state := &evolution.RateState{
		WeekStart:         "2026-04-07",
		ProposalsThisWeek: 2,
		LastProposalTimes: map[string]string{
			".claude/skills/foo/SKILL.md": time.Now().UTC().Format(time.RFC3339),
		},
	}

	if err := evolution.WriteRateState(projectRoot, state); err != nil {
		t.Fatalf("WriteRateState: %v", err)
	}

	loaded, err := evolution.ReadRateState(projectRoot)
	if err != nil {
		t.Fatalf("ReadRateState: %v", err)
	}
	if loaded.ProposalsThisWeek != state.ProposalsThisWeek {
		t.Errorf("ProposalsThisWeek: want %d, got %d",
			state.ProposalsThisWeek, loaded.ProposalsThisWeek)
	}
	if loaded.WeekStart != state.WeekStart {
		t.Errorf("WeekStart: want %q, got %q", state.WeekStart, loaded.WeekStart)
	}
}

// TestCheckFileCooldown_NoCooldownForNewFile verifies that a file with no
// prior proposal passes the cooldown check.
func TestCheckFileCooldown_NoCooldownForNewFile(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	err := evolution.CheckFileCooldown(projectRoot, ".claude/skills/new/SKILL.md")
	if err != nil {
		t.Fatalf("expected no cooldown error for new file, got %v", err)
	}
}

// TestCheckFileCooldown_EmptyTarget verifies that an empty target file passes
// cooldown (no file-level limit applicable).
func TestCheckFileCooldown_EmptyTarget(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	err := evolution.CheckFileCooldown(projectRoot, "")
	if err != nil {
		t.Fatalf("expected nil for empty target, got %v", err)
	}
}

// TestCheckFrozenGuard_SkillInSubdir verifies that skill files in a nested
// subdirectory are permitted.
func TestCheckFrozenGuard_SkillInSubdir(t *testing.T) {
	target := ".claude/skills/moai-lang-go/modules/examples.md"
	err := evolution.CheckFrozenGuard(target)
	if err != nil {
		t.Fatalf("expected nil for nested skill file %q, got %v", target, err)
	}
}

// TestLoadLearning_MalformedFile verifies that a file without an ID header
// returns an error.
func TestLoadLearning_MalformedFile(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(badFile, []byte("not a valid learning file\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := evolution.LoadLearning(badFile)
	if err == nil {
		t.Fatal("expected error for malformed learning file, got nil")
	}
}

// TestRevertProposal_NilProposal verifies that a nil proposal returns an error.
func TestRevertProposal_NilProposal(t *testing.T) {
	projectRoot := t.TempDir()
	err := evolution.RevertProposal(projectRoot, nil)
	if err == nil {
		t.Fatal("expected error for nil proposal, got nil")
	}
}

// TestRevertProposal_MissingBak verifies that RevertProposal returns an error
// when the .bak file does not exist.
func TestRevertProposal_MissingBak(t *testing.T) {
	projectRoot := t.TempDir()
	proposal := &evolution.ProposedChange{
		TargetFile: ".claude/skills/no-exist/SKILL.md",
		ZoneID:     "zone",
		Addition:   "x",
	}
	err := evolution.RevertProposal(projectRoot, proposal)
	if err == nil {
		t.Fatal("expected error when .bak missing, got nil")
	}
}

// TestEvaluateGraduation_GraduatedPreserved verifies that a graduated entry
// stays graduated.
func TestEvaluateGraduation_GraduatedPreserved(t *testing.T) {
	entry := &evolution.LearningEntry{
		Status:       evolution.StatusGraduated,
		Observations: 20,
		Confidence:   1.0,
	}
	got := evolution.EvaluateGraduation(entry)
	if got != evolution.StatusGraduated {
		t.Fatalf("expected StatusGraduated, got %q", got)
	}
}

// TestDetectDuplicate_EmptyProjectDir verifies that an empty learnings dir
// returns nil without error.
func TestDetectDuplicate_EmptyProjectDir(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	dup, err := evolution.DetectDuplicate(projectRoot, "some unique observation")
	if err != nil {
		t.Fatalf("DetectDuplicate: %v", err)
	}
	if dup != nil {
		t.Fatal("expected nil for empty learnings dir")
	}
}

// TestCheckFrozenGuard_CoreRuleFile verifies that any file under .claude/rules/moai/core/
// is blocked.
func TestCheckFrozenGuard_CoreRuleFile(t *testing.T) {
	target := ".claude/rules/moai/core/some-new-rule.md"
	err := evolution.CheckFrozenGuard(target)
	if err != evolution.ErrFrozenPath {
		t.Fatalf("expected ErrFrozenPath for core rule file, got %v", err)
	}
}

// TestArchiveOldLearnings_NoExcess verifies that ArchiveOldLearnings is a
// no-op when count is within limit.
func TestArchiveOldLearnings_NoExcess(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("LEARN-20260411-AO%d", i)
		if err := evolution.CreateLearning(projectRoot, sampleEntry(id)); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	if err := evolution.ArchiveOldLearnings(projectRoot, 5); err != nil {
		t.Fatalf("ArchiveOldLearnings: %v", err)
	}

	active, err := evolution.ListLearnings(projectRoot, evolution.LearningFilter{ExcludeArchived: true})
	if err != nil {
		t.Fatalf("ListLearnings: %v", err)
	}
	if len(active) != 3 {
		t.Fatalf("want 3 active (no archiving needed), got %d", len(active))
	}
}

// TestCreateLearning_ErrorOnExistingID verifies that writing the same entry
// twice succeeds (overwrite is permitted for simplicity).
func TestCreateLearning_DuplicateIDOverwrite(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	e := sampleEntry("LEARN-20260411-DUP")
	if err := evolution.CreateLearning(projectRoot, e); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Overwriting should not return an error.
	e.Observations = 2
	if err := evolution.CreateLearning(projectRoot, e); err != nil {
		t.Fatalf("overwrite create: %v", err)
	}

	loaded, _ := evolution.LoadLearningByID(projectRoot, e.ID)
	if loaded == nil || loaded.Observations != 2 {
		t.Fatal("expected overwritten value to be 2")
	}
}
