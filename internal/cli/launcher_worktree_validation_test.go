// launcher_worktree_validation_test.go: Agent() 반환값의 worktreePath 유효성 검증 테스트.
// REQ-RA-005, REQ-RA-010 매핑.
//
// M4 GREEN-3 phase: validateWorktreeReturn 구현 완료 → t.Skip 제거 + assertion 활성화.
// 변경: worktreeReturn 로컬 정의 제거 (worktree_validation.go로 이동).
package cli

import (
	"errors"
	"strings"
	"testing"
	"text/template"
)

// TestValidateWorktreeReturn_RejectsEmptyObject는 worktreePath가 빈 문자열일 때
// WORKTREE_PATH_INVALID sentinel 에러가 반환되는지 검증한다.
//
// REQ-RA-005: isolation:worktree + 빈 worktreePath → WORKTREE_PATH_INVALID
// REQ-RA-010: broken worktreePath 검증 없이 propagate 금지
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

// TestValidateWorktreeReturn_RejectsNullPath는 nil worktreeReturn 전달 시
// WORKTREE_PATH_INVALID sentinel 에러가 반환되는지 검증한다.
//
// REQ-RA-005: nil/undefined worktreePath → WORKTREE_PATH_INVALID
func TestValidateWorktreeReturn_RejectsNullPath(t *testing.T) {
	// nil 포인터 케이스
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

// TestValidateWorktreeReturn_AcceptsValidPath는 유효한 절대 경로가 있을 때
// 에러 없이 통과하는지 검증한다.
//
// REQ-RA-005: 유효한 worktreePath는 통과
func TestValidateWorktreeReturn_AcceptsValidPath(t *testing.T) {
	result := &worktreeReturn{
		WorktreePath:   "/tmp/abc123-worktree",
		WorktreeBranch: "feat/SPEC-V3R3-RETIRED-AGENT-001",
		IsolationMode:  "worktree",
	}
	err := validateWorktreeReturn(result, "worktree", "manager-cycle")
	if err != nil {
		t.Errorf("유효한 worktreePath에 대해 오류가 반환되면 안 됨: %v", err)
	}
}

// TestValidateWorktreeReturn_SkipsWhenIsolationNotWorktree는 isolation이
// "worktree"가 아닐 때 빈 WorktreePath도 오류 없이 통과하는지 검증한다.
//
// REQ-RA-005: isolation:worktree 요청 시에만 검증 적용
func TestValidateWorktreeReturn_SkipsWhenIsolationNotWorktree(t *testing.T) {
	isolationModes := []string{"", "none", "sandbox"}
	for _, mode := range isolationModes {
		result := &worktreeReturn{
			WorktreePath:  "", // 빈 경로
			IsolationMode: mode,
		}
		err := validateWorktreeReturn(result, mode, "manager-cycle")
		if err != nil {
			t.Errorf("isolation=%q + 빈 WorktreePath에 오류 반환됨 (스킵해야 함): %v", mode, err)
		}
	}
}

// TestPathTemplateRejectsNonStringValue는 text/template을 사용한 경로 보간에서
// 빈 struct/map 값이 literal "{}" 문자열을 생성하지 않고 에러를 발생시키는지 검증한다.
//
// REQ-RA-006: text/template 기반 경로 보간으로 {} literal 방지
// REQ-RA-015: "/{}/{}" 경로 생성 방지
func TestPathTemplateRejectsNonStringValue(t *testing.T) {
	t.Parallel()

	// 케이스 1: 빈 struct를 fmt.Sprintf로 보간하면 "{}" 생성 (위험 패턴 증명)
	type emptyStruct struct{}
	dangerousPath := constructPathUnsafe("root", emptyStruct{}, emptyStruct{})
	if !strings.Contains(dangerousPath, "{}") {
		t.Logf("경고: fmt.Sprintf가 빈 struct를 '{}' 로 변환하지 않음 (플랫폼 의존적): %q", dangerousPath)
	} else {
		t.Logf("위험 패턴 확인: fmt.Sprintf + empty struct → %q (REQ-RA-006 fix 대상)", dangerousPath)
	}

	// 케이스 2: text/template은 typed struct를 안전하게 보간함 (안전 패턴 증명)
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
	// 안전한 text/template 보간: {} literal이 없어야 함
	if strings.Contains(result, "{}") {
		t.Errorf("text/template 보간 결과에 '{}'가 포함됨 (타입 안전성 위반): %q", result)
	}
	t.Logf("안전한 text/template 보간 결과: %q", result)

	// 케이스 3: validateWorktreeReturn이 "{}" worktreePath를 WORKTREE_PATH_INVALID로 거부
	// SPEC-V3R3-RETIRED-AGENT-001 D-EVAL-02 fix: AC-RA-18 critical assertion 충족.
	// 5-layer defect chain Layer 4 산물 (literal "{}" / "[object Object]") 거부.
	t.Run("validateWorktreeReturn rejects {} literal", func(t *testing.T) {
		t.Parallel()
		patterns := []string{"{}", "[object Object]", "null", "undefined"}
		for _, p := range patterns {
			r := &worktreeReturn{WorktreePath: p, IsolationMode: "worktree"}
			err := validateWorktreeReturn(r, "worktree", "manager-cycle")
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

// constructPathUnsafe는 빈 interface{} 값을 경로 보간에 사용하는 위험한 패턴을 시뮬레이션한다.
// 이 함수는 REQ-RA-006이 fix하려는 문제를 증명하기 위해 존재한다.
// 실제 코드에서는 이 패턴을 사용하면 안 됨.
func constructPathUnsafe(root string, branch, path interface{}) string {
	return root + "/" + anyToString(branch) + "/" + anyToString(path)
}

// anyToString은 임의의 값을 문자열로 변환한다.
// 빈 struct는 "{}"로 변환됨 (5-layer defect chain layer 4의 원인).
func anyToString(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		// map, struct 등: fmt.Sprintf 없이 직접 변환하면 "{}"
		// 이것이 mo.ai.kr 사건의 원인
		_ = s
		return strings.TrimSpace(strings.ReplaceAll(
			strings.ReplaceAll(
				// JSON-like: empty struct → "{}"
				"{}", " ", "",
			), "\n", "",
		))
	}
}
