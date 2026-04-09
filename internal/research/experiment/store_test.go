package experiment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// TestResultStore는 ResultStore의 모든 기능을 테이블 기반으로 검증한다.
func TestResultStore(t *testing.T) {
	t.Parallel()

	makeExperiment := func(id string, score float64) *Experiment {
		return &Experiment{
			ID:         id,
			Target:     "skills/moai-lang-go",
			Hypothesis: "테스트 가설",
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

	t.Run("Save_3개_실험_파일명_검증", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/moai-lang-go"

		for i := 1; i <= 3; i++ {
			exp := makeExperiment("exp-"+strings.Repeat("0", 3-len(string(rune('0'+i))))+string(rune('0'+i)), float64(i)*0.1)
			if err := store.SaveExperiment(target, exp); err != nil {
				t.Fatalf("SaveExperiment %d 실패: %v", i, err)
			}
		}

		// 파일명 검증
		sanitized := sanitizeTarget(target)
		targetDir := filepath.Join(dir, sanitized)
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			t.Fatalf("디렉토리 읽기 실패: %v", err)
		}

		wantFiles := []string{"exp-001.json", "exp-002.json", "exp-003.json"}
		got := make([]string, 0, len(entries))
		for _, e := range entries {
			if !e.IsDir() {
				got = append(got, e.Name())
			}
		}

		if len(got) != len(wantFiles) {
			t.Fatalf("파일 수: got %d, want %d\n파일: %v", len(got), len(wantFiles), got)
		}
		for i, want := range wantFiles {
			if got[i] != want {
				t.Errorf("파일[%d]: got %q, want %q", i, got[i], want)
			}
		}
	})

	t.Run("LoadExperiments_순서_보장", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/alpha"

		scores := []float64{0.5, 0.7, 0.9}
		for _, s := range scores {
			if err := store.SaveExperiment(target, makeExperiment("", s)); err != nil {
				t.Fatalf("SaveExperiment 실패: %v", err)
			}
		}

		loaded, err := store.LoadExperiments(target)
		if err != nil {
			t.Fatalf("LoadExperiments 실패: %v", err)
		}

		if len(loaded) != 3 {
			t.Fatalf("로드된 실험 수: got %d, want 3", len(loaded))
		}
		for i, want := range scores {
			if loaded[i].Result.Overall != want {
				t.Errorf("실험[%d] Overall: got %f, want %f", i, loaded[i].Result.Overall, want)
			}
		}
	})

	t.Run("ExperimentCount_정확성", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/beta"

		if got := store.ExperimentCount(target); got != 0 {
			t.Errorf("초기 카운트: got %d, want 0", got)
		}

		for i := 0; i < 5; i++ {
			if err := store.SaveExperiment(target, makeExperiment("", 0.5)); err != nil {
				t.Fatalf("SaveExperiment 실패: %v", err)
			}
		}

		if got := store.ExperimentCount(target); got != 5 {
			t.Errorf("5개 저장 후 카운트: got %d, want 5", got)
		}
	})

	t.Run("AppendChangelog_생성_및_추가", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)
		target := "skills/gamma"

		entries := []ChangelogEntry{
			{ExperimentID: "exp-001", Score: 0.75, Change: "프롬프트 수정", Reasoning: "정확도 향상", Decision: DecisionKeep},
			{ExperimentID: "exp-002", Score: 0.60, Change: "예제 제거", Reasoning: "성능 저하", Decision: DecisionDiscard},
		}

		for _, entry := range entries {
			if err := store.AppendChangelog(target, entry); err != nil {
				t.Fatalf("AppendChangelog 실패: %v", err)
			}
		}

		// changelog.md 파일 확인
		sanitized := sanitizeTarget(target)
		changelogPath := filepath.Join(dir, sanitized, "changelog.md")
		data, err := os.ReadFile(changelogPath)
		if err != nil {
			t.Fatalf("changelog.md 읽기 실패: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "exp-001") {
			t.Error("changelog에 exp-001이 포함되어야 함")
		}
		if !strings.Contains(content, "exp-002") {
			t.Error("changelog에 exp-002가 포함되어야 함")
		}
		if !strings.Contains(content, "keep") {
			t.Error("changelog에 keep 결정이 포함되어야 함")
		}
		if !strings.Contains(content, "discard") {
			t.Error("changelog에 discard 결정이 포함되어야 함")
		}
	})

	t.Run("빈_디렉토리_LoadExperiments_빈_슬라이스", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		store := NewResultStore(dir)

		loaded, err := store.LoadExperiments("nonexistent/target")
		if err != nil {
			t.Fatalf("빈 디렉토리에서 에러 발생하면 안됨: %v", err)
		}
		if len(loaded) != 0 {
			t.Errorf("빈 디렉토리에서 실험 수: got %d, want 0", len(loaded))
		}
	})
}
