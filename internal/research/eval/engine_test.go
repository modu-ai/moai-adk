package eval

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewEngine verifies that NewEngine returns a non-nil EvalEngine.
func TestNewEngine(t *testing.T) {
	t.Parallel()

	e := NewEngine()
	if e == nil {
		t.Fatal("NewEngine() returned nil")
	}
}

// TestEngine_Integration tests the complete workflow through the engine.
func TestEngine_Integration(t *testing.T) {
	t.Parallel()

	// Prepare YAML file
	dir := t.TempDir()
	p := filepath.Join(dir, "suite.yaml")
	if err := os.WriteFile(p, []byte(validSuiteYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	e := NewEngine()

	// Test LoadSuite
	suite, err := e.LoadSuite(p)
	if err != nil {
		t.Fatalf("Engine.LoadSuite() error = %v", err)
	}
	if len(suite.Criteria) != 2 {
		t.Fatalf("Criteria count = %d, want 2", len(suite.Criteria))
	}

	// Test Evaluate — all pass
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

	// Test Evaluate — must_pass failure
	results2 := map[string]bool{
		"정확성": false,
		"가독성": true,
	}
	evalResult2, err := e.Evaluate(suite, results2)
	if err != nil {
		t.Fatalf("Engine.Evaluate() error = %v", err)
	}
	if evalResult2.MustPassOK {
		t.Error("MustPassOK = true, want false (must_pass criterion failed)")
	}
}

// TestEngine_LoadSuite_Error verifies that Engine.LoadSuite returns an error for an invalid path.
func TestEngine_LoadSuite_Error(t *testing.T) {
	t.Parallel()

	e := NewEngine()
	_, err := e.LoadSuite("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("Engine.LoadSuite() should return error for nonexistent file")
	}
}

// TestEngine_Evaluate_NilSuite verifies that Evaluate returns an error for a nil suite.
func TestEngine_Evaluate_NilSuite(t *testing.T) {
	t.Parallel()

	e := NewEngine()
	_, err := e.Evaluate(nil, map[string]bool{})
	if err == nil {
		t.Fatal("Engine.Evaluate(nil, ...) should return error")
	}
}
