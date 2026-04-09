package eval

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewEngine NewEngine가 nil이 아닌 EvalEngine을 반환하는지 검증한다.
func TestNewEngine(t *testing.T) {
	t.Parallel()

	e := NewEngine()
	if e == nil {
		t.Fatal("NewEngine() returned nil")
	}
}

// TestEngine_Integration 엔진을 통한 전체 워크플로우 통합 테스트.
func TestEngine_Integration(t *testing.T) {
	t.Parallel()

	// YAML 파일 준비
	dir := t.TempDir()
	p := filepath.Join(dir, "suite.yaml")
	if err := os.WriteFile(p, []byte(validSuiteYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	e := NewEngine()

	// LoadSuite 테스트
	suite, err := e.LoadSuite(p)
	if err != nil {
		t.Fatalf("Engine.LoadSuite() error = %v", err)
	}
	if len(suite.Criteria) != 2 {
		t.Fatalf("Criteria 개수 = %d, want 2", len(suite.Criteria))
	}

	// Evaluate 테스트 — 전부 통과
	results := map[string]bool{
		"정확성": true,
		"가독성": true,
	}
	evalResult, err := e.Evaluate(suite, results)
	if err != nil {
		t.Fatalf("Engine.Evaluate() error = %v", err)
	}
	if evalResult.Overall != 1.0 {
		t.Errorf("Overall = %f, want 1.0", evalResult.Overall)
	}
	if !evalResult.MustPassOK {
		t.Error("MustPassOK = false, want true")
	}

	// Evaluate 테스트 — must_pass 실패
	results2 := map[string]bool{
		"정확성": false,
		"가독성": true,
	}
	evalResult2, err := e.Evaluate(suite, results2)
	if err != nil {
		t.Fatalf("Engine.Evaluate() error = %v", err)
	}
	if evalResult2.MustPassOK {
		t.Error("MustPassOK = true, want false (must_pass 기준 실패)")
	}
}

// TestEngine_LoadSuite_Error 엔진의 LoadSuite가 잘못된 경로에서 에러를 반환하는지 검증한다.
func TestEngine_LoadSuite_Error(t *testing.T) {
	t.Parallel()

	e := NewEngine()
	_, err := e.LoadSuite("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("Engine.LoadSuite() 존재하지 않는 파일에서 에러를 반환해야 함")
	}
}

// TestEngine_Evaluate_NilSuite nil 스위트에서 Evaluate가 에러를 반환하는지 검증한다.
func TestEngine_Evaluate_NilSuite(t *testing.T) {
	t.Parallel()

	e := NewEngine()
	_, err := e.Evaluate(nil, map[string]bool{})
	if err == nil {
		t.Fatal("Engine.Evaluate(nil, ...) 에러를 반환해야 함")
	}
}
