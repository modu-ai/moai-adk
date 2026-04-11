package experiment

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// TestBaselineManager verifies BaselineManager Save/Load/Exists using table-driven tests.
func TestBaselineManager(t *testing.T) {
	t.Parallel()

	// Prepare test data
	makeResult := func(overall float64, mustPassOK bool) *eval.EvalResult {
		return &eval.EvalResult{
			Overall: overall,
			PerCriterion: map[string]eval.CriterionResult{
				"accuracy": {Name: "accuracy", Passed: mustPassOK, Weight: eval.MustPass},
			},
			MustPassOK: mustPassOK,
			Timestamp:  time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		}
	}

	t.Run("Save_then_Load_round_trip", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)
		result := makeResult(0.85, true)

		if err := mgr.Save("skills/moai-lang-go", result); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := mgr.Load("skills/moai-lang-go")
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded.Overall != result.Overall {
			t.Errorf("Overall: got %f, want %f", loaded.Overall, result.Overall)
		}
		if loaded.MustPassOK != result.MustPassOK {
			t.Errorf("MustPassOK: got %v, want %v", loaded.MustPassOK, result.MustPassOK)
		}
	})

	t.Run("Load_nonexistent_file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		_, err := mgr.Load("nonexistent/target")
		if err == nil {
			t.Fatal("Load on nonexistent file should return error")
		}
	})

	t.Run("Exists_true_after_Save", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		if mgr.Exists("skills/moai-lang-go") {
			t.Fatal("Exists should be false before Save")
		}

		if err := mgr.Save("skills/moai-lang-go", makeResult(0.75, true)); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		if !mgr.Exists("skills/moai-lang-go") {
			t.Fatal("Exists should be true after Save")
		}
	})

	t.Run("Exists_false_before_Save", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		if mgr.Exists("any/target") {
			t.Fatal("Exists should be false before Save")
		}
	})

	t.Run("multiple_targets_no_interference", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		result1 := makeResult(0.80, true)
		result2 := makeResult(0.60, false)

		if err := mgr.Save("target/alpha", result1); err != nil {
			t.Fatalf("Save target/alpha failed: %v", err)
		}
		if err := mgr.Save("target/beta", result2); err != nil {
			t.Fatalf("Save target/beta failed: %v", err)
		}

		loaded1, err := mgr.Load("target/alpha")
		if err != nil {
			t.Fatalf("Load target/alpha failed: %v", err)
		}
		loaded2, err := mgr.Load("target/beta")
		if err != nil {
			t.Fatalf("Load target/beta failed: %v", err)
		}

		if loaded1.Overall != 0.80 {
			t.Errorf("target/alpha Overall: got %f, want 0.80", loaded1.Overall)
		}
		if loaded2.Overall != 0.60 {
			t.Errorf("target/beta Overall: got %f, want 0.60", loaded2.Overall)
		}
	})
}
