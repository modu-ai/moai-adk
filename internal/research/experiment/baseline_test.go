package experiment

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// TestBaselineManager는 BaselineManager의 Save/Load/Exists를 테이블 기반으로 검증한다.
func TestBaselineManager(t *testing.T) {
	t.Parallel()

	// 테스트 데이터 준비
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

	t.Run("Save_후_Load_라운드트립", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)
		result := makeResult(0.85, true)

		if err := mgr.Save("skills/moai-lang-go", result); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		loaded, err := mgr.Load("skills/moai-lang-go")
		if err != nil {
			t.Fatalf("Load 실패: %v", err)
		}

		if loaded.Overall != result.Overall {
			t.Errorf("Overall: got %f, want %f", loaded.Overall, result.Overall)
		}
		if loaded.MustPassOK != result.MustPassOK {
			t.Errorf("MustPassOK: got %v, want %v", loaded.MustPassOK, result.MustPassOK)
		}
	})

	t.Run("Load_존재하지_않는_파일", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		_, err := mgr.Load("nonexistent/target")
		if err == nil {
			t.Fatal("존재하지 않는 파일 Load에서 에러가 발생해야 함")
		}
	})

	t.Run("Exists_Save_후_true", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		if mgr.Exists("skills/moai-lang-go") {
			t.Fatal("Save 전에 Exists가 false여야 함")
		}

		if err := mgr.Save("skills/moai-lang-go", makeResult(0.75, true)); err != nil {
			t.Fatalf("Save 실패: %v", err)
		}

		if !mgr.Exists("skills/moai-lang-go") {
			t.Fatal("Save 후에 Exists가 true여야 함")
		}
	})

	t.Run("Exists_Save_전_false", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		if mgr.Exists("any/target") {
			t.Fatal("Save 전에 Exists가 false여야 함")
		}
	})

	t.Run("여러_타겟_간섭_없음", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		mgr := NewBaselineManager(dir)

		result1 := makeResult(0.80, true)
		result2 := makeResult(0.60, false)

		if err := mgr.Save("target/alpha", result1); err != nil {
			t.Fatalf("Save target/alpha 실패: %v", err)
		}
		if err := mgr.Save("target/beta", result2); err != nil {
			t.Fatalf("Save target/beta 실패: %v", err)
		}

		loaded1, err := mgr.Load("target/alpha")
		if err != nil {
			t.Fatalf("Load target/alpha 실패: %v", err)
		}
		loaded2, err := mgr.Load("target/beta")
		if err != nil {
			t.Fatalf("Load target/beta 실패: %v", err)
		}

		if loaded1.Overall != 0.80 {
			t.Errorf("target/alpha Overall: got %f, want 0.80", loaded1.Overall)
		}
		if loaded2.Overall != 0.60 {
			t.Errorf("target/beta Overall: got %f, want 0.60", loaded2.Overall)
		}
	})
}
