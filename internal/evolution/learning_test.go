package evolution_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

// TestCreateLearning_CreatesFile verifies that CreateLearning writes a
// markdown file under .moai/evolution/learnings/.
func TestCreateLearning_CreatesFile(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	entry := sampleEntry("LEARN-20260411-001")
	if err := evolution.CreateLearning(projectRoot, entry); err != nil {
		t.Fatalf("CreateLearning: %v", err)
	}

	// Loading should succeed.
	loaded, err := evolution.LoadLearningByID(projectRoot, entry.ID)
	if err != nil {
		t.Fatalf("LoadLearningByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected loaded entry, got nil")
	}
	if loaded.ID != entry.ID {
		t.Fatalf("ID mismatch: want %q, got %q", entry.ID, loaded.ID)
	}
	if loaded.SkillID != entry.SkillID {
		t.Fatalf("SkillID mismatch: want %q, got %q", entry.SkillID, loaded.SkillID)
	}
}

// TestLoadLearning_ParsesMarkdown verifies that the learning file is stored
// and parsed as markdown with YAML-like headers.
func TestLoadLearning_ParsesMarkdown(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	entry := sampleEntry("LEARN-20260411-002")
	entry.Observation = "missing context.Context parameter causes build failure"
	if err := evolution.CreateLearning(projectRoot, entry); err != nil {
		t.Fatalf("CreateLearning: %v", err)
	}

	loaded, err := evolution.LoadLearningByID(projectRoot, entry.ID)
	if err != nil {
		t.Fatalf("LoadLearningByID: %v", err)
	}
	if loaded.Observation != entry.Observation {
		t.Fatalf("Observation mismatch:\nwant: %q\ngot:  %q", entry.Observation, loaded.Observation)
	}
}

// TestUpdateLearning_IncrementsObservationCount verifies that UpdateLearning
// can modify the observation count via a callback.
func TestUpdateLearning_IncrementsObservationCount(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	entry := sampleEntry("LEARN-20260411-003")
	if err := evolution.CreateLearning(projectRoot, entry); err != nil {
		t.Fatalf("CreateLearning: %v", err)
	}

	err := evolution.UpdateLearning(projectRoot, entry.ID, func(e *evolution.LearningEntry) {
		e.Observations++
		e.Updated = time.Now()
	})
	if err != nil {
		t.Fatalf("UpdateLearning: %v", err)
	}

	loaded, err := evolution.LoadLearningByID(projectRoot, entry.ID)
	if err != nil {
		t.Fatalf("LoadLearningByID after update: %v", err)
	}
	if loaded.Observations != entry.Observations+1 {
		t.Fatalf("Observations: want %d, got %d", entry.Observations+1, loaded.Observations)
	}
}

// TestListLearnings_ReturnsAllActive verifies that ListLearnings returns all
// non-archived entries.
func TestListLearnings_ReturnsAllActive(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	const count = 5
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("LEARN-20260411-%03d", i+1)
		if err := evolution.CreateLearning(projectRoot, sampleEntry(id)); err != nil {
			t.Fatalf("CreateLearning %q: %v", id, err)
		}
	}

	entries, err := evolution.ListLearnings(projectRoot, evolution.LearningFilter{ExcludeArchived: true})
	if err != nil {
		t.Fatalf("ListLearnings: %v", err)
	}
	if len(entries) != count {
		t.Fatalf("want %d entries, got %d", count, len(entries))
	}
}

// TestMaxActiveLearnings_Enforced verifies that creating more than
// MaxActiveLearnings returns ErrMaxLearningsReached.
func TestMaxActiveLearnings_Enforced(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	// Fill up to the limit.
	for i := 0; i < evolution.MaxActiveLearnings; i++ {
		id := fmt.Sprintf("LEARN-20260411-%03d", i+1)
		if err := evolution.CreateLearning(projectRoot, sampleEntry(id)); err != nil {
			t.Fatalf("setup: CreateLearning %q: %v", id, err)
		}
	}

	// One more should fail.
	extra := sampleEntry("LEARN-20260411-EXTRA")
	err := evolution.CreateLearning(projectRoot, extra)
	if err == nil {
		t.Fatal("expected ErrMaxLearningsReached, got nil")
	}
	if err != evolution.ErrMaxLearningsReached {
		t.Fatalf("expected ErrMaxLearningsReached, got %v", err)
	}
}

// --- helpers ---

func sampleEntry(id string) *evolution.LearningEntry {
	return &evolution.LearningEntry{
		ID:           id,
		SkillID:      "moai-lang-go",
		ZoneID:       "best-practices",
		Observation:  "sample observation for testing",
		Status:       evolution.StatusObservation,
		Observations: 1,
		Created:      time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
		Updated:      time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
	}
}
