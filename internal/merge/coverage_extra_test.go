package merge

import (
	"context"
	"testing"
)

// Additional tests to push coverage above 85%.

func TestMergeJSON_InvalidBase(t *testing.T) {
	t.Parallel()

	_, err := mergeJSON([]byte("not json"), []byte("{}"), []byte("{}"))
	if err == nil {
		t.Error("expected error for invalid base JSON")
	}
}

func TestMergeJSON_InvalidCurrent(t *testing.T) {
	t.Parallel()

	_, err := mergeJSON([]byte("{}"), []byte("not json"), []byte("{}"))
	if err == nil {
		t.Error("expected error for invalid current JSON")
	}
}

func TestMergeJSON_InvalidUpdated(t *testing.T) {
	t.Parallel()

	_, err := mergeJSON([]byte("{}"), []byte("{}"), []byte("not json"))
	if err == nil {
		t.Error("expected error for invalid updated JSON")
	}
}

func TestMergeYAML_InvalidBase(t *testing.T) {
	t.Parallel()

	_, err := mergeYAML([]byte(":\n  :\n    - [invalid"), []byte("a: 1\n"), []byte("a: 1\n"))
	if err == nil {
		t.Error("expected error for invalid base YAML")
	}
}

func TestMergeYAML_InvalidCurrent(t *testing.T) {
	t.Parallel()

	_, err := mergeYAML([]byte("a: 1\n"), []byte(":\n  :\n    - [invalid"), []byte("a: 1\n"))
	if err == nil {
		t.Error("expected error for invalid current YAML")
	}
}

func TestMergeYAML_InvalidUpdated(t *testing.T) {
	t.Parallel()

	_, err := mergeYAML([]byte("a: 1\n"), []byte("a: 1\n"), []byte(":\n  :\n    - [invalid"))
	if err == nil {
		t.Error("expected error for invalid updated YAML")
	}
}

func TestLineMerge_CurrentOnlyChanged(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB\nC")
	current := []byte("A\nB_user\nC")
	updated := []byte("A\nB\nC")

	result, err := mergeLineBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict when only user changed")
	}
	if string(result.Content) != "A\nB_user\nC" {
		t.Errorf("got %q, want %q", string(result.Content), "A\nB_user\nC")
	}
}

func TestLineMerge_UserAddsLines(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB")
	current := []byte("A\nB\nC_user\nD_user")
	updated := []byte("A\nB")

	result, err := mergeLineBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict when user adds lines")
	}
}

func TestLineMerge_TemplateAddsLines(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB")
	current := []byte("A\nB")
	updated := []byte("A\nB\nC_tpl\nD_tpl")

	result, err := mergeLineBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(result.Content)
	if !containsLine(content, "C_tpl") {
		t.Error("expected template additions to appear")
	}
}

func TestLineMerge_EmptyFiles(t *testing.T) {
	t.Parallel()

	result, err := mergeLineBased([]byte(""), []byte(""), []byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict for empty files")
	}
}

func TestEngine_MergeFile_UnsupportedStrategy(t *testing.T) {
	t.Parallel()

	e := NewEngineWithSelector(&fixedSelector{strategy: MergeStrategy("unknown")})
	_, err := e.MergeFile(context.Background(), "file.xyz", nil, nil, nil)
	if err == nil {
		t.Error("expected error for unsupported strategy")
	}
}

func TestEngine_NewEngineWithSelector(t *testing.T) {
	t.Parallel()

	sel := NewStrategySelector()
	e := NewEngineWithSelector(sel)
	if e == nil {
		t.Error("expected non-nil engine")
	}
}

func TestValuesEqual_MarshalError(t *testing.T) {
	t.Parallel()

	// Channels cannot be marshaled to JSON.
	ch := make(chan int)
	result := valuesEqual(ch, ch)
	if result {
		t.Error("expected false for unmarshalable values")
	}
}

func TestDeepMergeMap_BothAddedSameKey(t *testing.T) {
	t.Parallel()

	base := map[string]any{}
	current := map[string]any{"new_key": "val_a"}
	updated := map[string]any{"new_key": "val_b"}

	_, conflicts := deepMergeMap(base, current, updated, "")
	if len(conflicts) == 0 {
		t.Error("expected conflict when both add same key with different values")
	}
}

func TestDeepMergeMap_UserDeletedKey(t *testing.T) {
	t.Parallel()

	base := map[string]any{"a": 1, "b": 2}
	current := map[string]any{"a": 1}
	updated := map[string]any{"a": 1, "b": 3}

	result, conflicts := deepMergeMap(base, current, updated, "")
	_ = conflicts
	if _, exists := result["b"]; exists {
		t.Error("expected user-deleted key 'b' to not be in result")
	}
}

// fixedSelector always returns the same strategy (for testing).
type fixedSelector struct {
	strategy MergeStrategy
}

func (f *fixedSelector) SelectStrategy(path string) MergeStrategy {
	return f.strategy
}
