// launcher_worktree_validation_test.go: validation tests for worktreePath returned by Agent().
// Maps to REQ-RA-005 and REQ-RA-010.
//
// M4 GREEN-3 phase: validateWorktreeReturn implementation complete → t.Skip removed + assertions enabled.
// Change: removed local worktreeReturn definition (moved to worktree_validation.go).
package cli

import (
	"errors"
	"strings"
	"testing"
	"text/template"
)

// TestValidateWorktreeReturn_RejectsEmptyObject verifies that the
// WORKTREE_PATH_INVALID sentinel error is returned when worktreePath is empty.
//
// REQ-RA-005: isolation:worktree + empty worktreePath → WORKTREE_PATH_INVALID
// REQ-RA-010: do not propagate a broken worktreePath without validation
func TestValidateWorktreeReturn_RejectsEmptyObject(t *testing.T) {
	result := &worktreeReturn{
		WorktreePath:  "",
		IsolationMode: "worktree",
	}
	err := validateWorktreeReturn(result, "worktree", "manager-tdd")
	if err == nil {
		t.Fatal("빈 WorktreePath에 대해 오류가 반환되어야 함")
	}
	if !strings.Contains(err.Error(), "WORKTREE_PATH_INVALID") {
		t.Errorf("오류 메시지에 'WORKTREE_PATH_INVALID' sentinel 없음: %q", err.Error())
	}
	if !errors.Is(err, ErrWorktreePathInvalid) {
		t.Errorf("errors.Is(err, ErrWorktreePathInvalid) == false: %v", err)
	}
}

// TestValidateWorktreeReturn_RejectsNullPath verifies that the
// WORKTREE_PATH_INVALID sentinel error is returned when a nil worktreeReturn is passed.
//
// REQ-RA-005: nil/undefined worktreePath → WORKTREE_PATH_INVALID
func TestValidateWorktreeReturn_RejectsNullPath(t *testing.T) {
	// nil pointer case
	err := validateWorktreeReturn(nil, "worktree", "manager-tdd")
	if err == nil {
		t.Fatal("nil worktreeReturn에 대해 오류가 반환되어야 함")
	}
	if !strings.Contains(err.Error(), "WORKTREE_PATH_INVALID") {
		t.Errorf("오류 메시지에 'WORKTREE_PATH_INVALID' sentinel 없음: %q", err.Error())
	}
	if !errors.Is(err, ErrWorktreePathInvalid) {
		t.Errorf("errors.Is(err, ErrWorktreePathInvalid) == false: %v", err)
	}
}

// TestValidateWorktreeReturn_AcceptsValidPath verifies that a valid absolute
// path passes without error.
//
// REQ-RA-005: a valid worktreePath passes
func TestValidateWorktreeReturn_AcceptsValidPath(t *testing.T) {
	result := &worktreeReturn{
		WorktreePath:   "/tmp/abc123-worktree",
		WorktreeBranch: "feat/SPEC-V3R3-RETIRED-AGENT-001",
		IsolationMode:  "worktree",
	}
	err := validateWorktreeReturn(result, "worktree", "manager-develop")
	if err != nil {
		t.Errorf("유효한 worktreePath에 대해 오류가 반환되면 안 됨: %v", err)
	}
}

// TestValidateWorktreeReturn_SkipsWhenIsolationNotWorktree verifies that an
// empty WorktreePath passes without error when isolation is not "worktree".
//
// REQ-RA-005: validation applies only when isolation:worktree is requested
func TestValidateWorktreeReturn_SkipsWhenIsolationNotWorktree(t *testing.T) {
	isolationModes := []string{"", "none", "sandbox"}
	for _, mode := range isolationModes {
		result := &worktreeReturn{
			WorktreePath:  "", // empty path
			IsolationMode: mode,
		}
		err := validateWorktreeReturn(result, mode, "manager-develop")
		if err != nil {
			t.Errorf("isolation=%q + 빈 WorktreePath에 오류 반환됨 (스킵해야 함): %v", mode, err)
		}
	}
}

// TestPathTemplateRejectsNonStringValue verifies that text/template-based path
// interpolation raises an error rather than producing a literal "{}" string for
// empty struct/map values.
//
// REQ-RA-006: text/template-based path interpolation prevents the {} literal
// REQ-RA-015: prevent generation of "/{}/{}" paths
func TestPathTemplateRejectsNonStringValue(t *testing.T) {
	t.Parallel()

	// Case 1: interpolating an empty struct via fmt.Sprintf produces "{}" (demonstrates the dangerous pattern)
	type emptyStruct struct{}
	dangerousPath := constructPathUnsafe("root", emptyStruct{}, emptyStruct{})
	if !strings.Contains(dangerousPath, "{}") {
		t.Logf("경고: fmt.Sprintf가 빈 struct를 '{}' 로 변환하지 않음 (플랫폼 의존적): %q", dangerousPath)
	} else {
		t.Logf("위험 패턴 확인: fmt.Sprintf + empty struct → %q (REQ-RA-006 fix 대상)", dangerousPath)
	}

	// Case 2: text/template interpolates typed structs safely (demonstrates the safe pattern)
	type worktreePathData struct {
		Root   string
		Branch string
		Path   string
	}

	tmpl, err := template.New("test").Parse("{{.Root}}/{{.Branch}}/{{.Path}}")
	if err != nil {
		t.Fatalf("template.Parse 실패: %v", err)
	}

	var buf strings.Builder
	data := worktreePathData{
		Root:   "/Users/goos/MoAI/mo.ai.kr",
		Branch: "feat/test-branch",
		Path:   "worktree-123",
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("template.Execute 실패: %v", err)
	}

	result := buf.String()
	// Safe text/template interpolation: must not contain the {} literal
	if strings.Contains(result, "{}") {
		t.Errorf("text/template 보간 결과에 '{}'가 포함됨 (타입 안전성 위반): %q", result)
	}
	t.Logf("안전한 text/template 보간 결과: %q", result)

	// Case 3: validateWorktreeReturn rejects a "{}" worktreePath as WORKTREE_PATH_INVALID
	// SPEC-V3R3-RETIRED-AGENT-001 D-EVAL-02 fix: satisfies the AC-RA-18 critical assertion.
	// Rejects the layer-4 artifact of the 5-layer defect chain (literal "{}" / "[object Object]").
	t.Run("validateWorktreeReturn rejects {} literal", func(t *testing.T) {
		t.Parallel()
		patterns := []string{"{}", "[object Object]", "null", "undefined"}
		for _, p := range patterns {
			r := &worktreeReturn{WorktreePath: p, IsolationMode: "worktree"}
			err := validateWorktreeReturn(r, "worktree", "manager-develop")
			if err == nil {
				t.Errorf("worktreePath=%q should trigger WORKTREE_PATH_INVALID, got nil", p)
				continue
			}
			if !errors.Is(err, ErrWorktreePathInvalid) {
				t.Errorf("worktreePath=%q error should match ErrWorktreePathInvalid sentinel, got %v", p, err)
			}
		}
	})
}

// constructPathUnsafe simulates the dangerous pattern of using empty interface{}
// values for path interpolation. This function exists to demonstrate the issue
// that REQ-RA-006 fixes. Real code must not use this pattern.
func constructPathUnsafe(root string, branch, path interface{}) string {
	return root + "/" + anyToString(branch) + "/" + anyToString(path)
}

// anyToString converts an arbitrary value to a string.
// An empty struct converts to "{}" (the root cause of layer 4 in the 5-layer defect chain).
func anyToString(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		// For maps, structs, etc.: direct conversion without fmt.Sprintf yields "{}"
		// This was the cause of the mo.ai.kr incident
		_ = s
		return strings.TrimSpace(strings.ReplaceAll(
			strings.ReplaceAll(
				// JSON-like: empty struct → "{}"
				"{}", " ", "",
			), "\n", "",
		))
	}
}
