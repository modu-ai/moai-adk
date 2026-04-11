package experiment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// TestResultStore verifies all ResultStore functionality using table-driven tests.
func TestResultStore(t *testing.T) {
	t.Parallel()

	makeExperiment := func(id string, score float64) *Experiment {
		return &Experiment{
			ID:         id,
			Target:     "skills/moai-lang-go",
			Hypothesis: "test hypothesis",
			Change: ChangeRecord{
				Type:    "modification",
				Section: "prompt",
				Diff:    "- old\n+ new",
			},
			Result: &eval.EvalResult{
				Overall:      score,
				PerCriterion: map[string]eval.CriterionResult{},
				MustPassOK:   true,
				Timestamp:    time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
			},
			Decision:  DecisionKeep,
			Timestamp: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		}
	}

	t.Run("Save_3_experiments_filename_check", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/moai-lang-go"

		for i := 1; i <= 3; i++ {
			exp := makeExperiment("exp-"+strings.Repeat("0", 3-len(string(rune('0'+i))))+string(rune('0'+i)), float64(i)*0.1)
			if err := store.SaveExperiment(target, exp); err != nil {
				t.Fatalf("SaveExperiment %d failed: %v", i, err)
			}
		}

		// Verify filenames
		sanitized := sanitizeTarget(target)
		targetDir := filepath.Join(dir, sanitized)
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			t.Fatalf("failed to read directory: %v", err)
		}

		wantFiles := []string{"exp-001.json", "exp-002.json", "exp-003.json"}
		got := make([]string, 0, len(entries))
		for _, e := range entries {
			if !e.IsDir() {
				got = append(got, e.Name())
			}
		}

		if len(got) != len(wantFiles) {
			t.Fatalf("file count: got %d, want %d\nfiles: %v", len(got), len(wantFiles), got)
		}
		for i, want := range wantFiles {
			if got[i] != want {
				t.Errorf("file[%d]: got %q, want %q", i, got[i], want)
			}
		}
	})

	t.Run("LoadExperiments_order_guaranteed", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/alpha"

		scores := []float64{0.5, 0.7, 0.9}
		for _, s := range scores {
			if err := store.SaveExperiment(target, makeExperiment("", s)); err != nil {
				t.Fatalf("SaveExperiment failed: %v", err)
			}
		}

		loaded, err := store.LoadExperiments(target)
		if err != nil {
			t.Fatalf("LoadExperiments failed: %v", err)
		}

		if len(loaded) != 3 {
			t.Fatalf("loaded experiment count: got %d, want 3", len(loaded))
		}
		for i, want := range scores {
			if loaded[i].Result.Overall != want {
				t.Errorf("experiment[%d] Overall: got %f, want %f", i, loaded[i].Result.Overall, want)
			}
		}
	})

	t.Run("ExperimentCount_accuracy", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/beta"

		if got := store.ExperimentCount(target); got != 0 {
			t.Errorf("initial count: got %d, want 0", got)
		}

		for i := 0; i < 5; i++ {
			if err := store.SaveExperiment(target, makeExperiment("", 0.5)); err != nil {
				t.Fatalf("SaveExperiment failed: %v", err)
			}
		}

		if got := store.ExperimentCount(target); got != 5 {
			t.Errorf("count after 5 saves: got %d, want 5", got)
		}
	})

	t.Run("AppendChangelog_create_and_append", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/gamma"

		entries := []ChangelogEntry{
			{ExperimentID: "exp-001", Score: 0.75, Change: "prompt modification", Reasoning: "improved accuracy", Decision: DecisionKeep},
			{ExperimentID: "exp-002", Score: 0.60, Change: "example removal", Reasoning: "performance degraded", Decision: DecisionDiscard},
		}

		for _, entry := range entries {
			if err := store.AppendChangelog(target, entry); err != nil {
				t.Fatalf("AppendChangelog failed: %v", err)
			}
		}

		// Verify changelog.md file
		sanitized := sanitizeTarget(target)
		changelogPath := filepath.Join(dir, sanitized, "changelog.md")
		data, err := os.ReadFile(changelogPath)
		if err != nil {
			t.Fatalf("failed to read changelog.md: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "exp-001") {
			t.Error("changelog should contain exp-001")
		}
		if !strings.Contains(content, "exp-002") {
			t.Error("changelog should contain exp-002")
		}
		if !strings.Contains(content, "keep") {
			t.Error("changelog should contain keep decision")
		}
		if !strings.Contains(content, "discard") {
			t.Error("changelog should contain discard decision")
		}
	})

	t.Run("empty_directory_LoadExperiments_empty_slice", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)

		loaded, err := store.LoadExperiments("nonexistent/target")
		if err != nil {
			t.Fatalf("should not return error for empty directory: %v", err)
		}
		if len(loaded) != 0 {
			t.Errorf("experiment count for empty directory: got %d, want 0", len(loaded))
		}
	})
}
