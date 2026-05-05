// worktree_validation.go: isolation:worktree Agent() 반환값 검증 헬퍼.
// REQ-RA-005 (검증 헬퍼), REQ-RA-010 (WORKTREE_PATH_INVALID sentinel).
//
// 이 파일은 향후 SubagentStart hook re-delegation 플로우와 외부 통합을 위해
// standalone helper 형태로 존재한다. 현 시점에는 직접 callsite가 없으며
// (plan-stage 가정: 3-5 callsite → 실제 grep 결과: 0개),
// 테스트로만 동작이 검증된다. 실제 wire-up은 후속 SPEC에서 수행한다.
package cli

import (
	"errors"
	"fmt"
)

// worktreeReturn은 Agent() 호출 반환값에서 worktree 관련 필드를 모델링한다.
// isolation 필드: "worktree", "", "none" 등.
//
// 이 타입은 test 파일(launcher_worktree_validation_test.go)에도 동일하게 정의되어 있다.
// M4 완료 후 test 파일의 로컬 정의를 이 타입 참조로 대체한다.
//
// 주의: test 파일과 동일한 패키지(cli)이므로 test 파일에 동일한 이름의 타입이 있으면
// 컴파일 오류가 발생한다. test 파일의 로컬 worktreeReturn 정의를 제거했음.
type worktreeReturn struct {
	WorktreePath   string
	WorktreeBranch string
	IsolationMode  string
}

// WorktreePathInvalidError는 isolation:worktree 호출이 broken worktreePath를
// 반환했을 때 발생하는 typed 에러이다.
// errors.Is(err, ErrWorktreePathInvalid) 또는 errors.As로 식별 가능.
//
// AC-RA-10: 에러 메시지에 agentName + reason 컨텍스트 포함.
type WorktreePathInvalidError struct {
	AgentName string
	Reason    string
}

// Error는 "WORKTREE_PATH_INVALID" sentinel 문자열을 포함한 에러 메시지를 반환한다.
func (e *WorktreePathInvalidError) Error() string {
	return fmt.Sprintf("WORKTREE_PATH_INVALID: agent=%q reason=%q", e.AgentName, e.Reason)
}

// Is는 target이 ErrWorktreePathInvalid이면 true를 반환한다.
// errors.Is(err, ErrWorktreePathInvalid) 패턴을 지원한다.
func (e *WorktreePathInvalidError) Is(target error) bool {
	return errors.Is(target, ErrWorktreePathInvalid)
}

// ErrWorktreePathInvalid는 isolation:worktree 호출이 broken worktreePath를
// 반환했을 때 사용하는 sentinel 에러이다.
// errors.Is(err, ErrWorktreePathInvalid)로 식별 가능 (REQ-RA-005, REQ-RA-010).
var ErrWorktreePathInvalid = errors.New("WORKTREE_PATH_INVALID")

// @MX:NOTE: [AUTO] validateWorktreeReturn은 SPEC-V3R3-RETIRED-AGENT-001 5-layer defect chain
// Layer 2-4 (worktree allocation broken state propagation + path interpolation literal "{}")
// 차단의 defense-in-depth 헬퍼다. Layer 1 (SubagentStart guard, internal/hook/subagent_start.go의
// agentStartHandler)이 차단되더라도 in-depth defense로 작동한다.
// 현재 callsite 0개 (standalone helper); 후속 SPEC에서 wire-up 예정.
//
// validateWorktreeReturn은 isolation:"worktree" Agent() 호출의 반환값 worktreePath가
// 유효한지 검증한다.
//
// 동작:
//   - isolationMode가 "worktree"가 아니면 검증을 건너뛰고 nil을 반환한다.
//   - result가 nil이거나 WorktreePath가 빈 문자열이면 WORKTREE_PATH_INVALID 에러를 반환한다.
//   - WorktreeBranch는 optional — 없거나 빈 문자열이어도 통과한다.
//
// REQ-RA-005: 검증 헬퍼 신규 추가.
// REQ-RA-010: broken worktreePath를 검증 없이 propagate 금지.
//
// 현재 callsite: 없음 (standalone helper, 후속 SPEC에서 wire-up 예정).
func validateWorktreeReturn(result *worktreeReturn, isolationMode string, agentName string) error {
	// isolation이 "worktree"가 아닌 경우 검증 스킵 (REQ-RA-005 edge case)
	if isolationMode != "worktree" {
		return nil
	}

	// nil 반환값 검사
	if result == nil {
		return &WorktreePathInvalidError{
			AgentName: agentName,
			Reason:    "nil worktreeReturn",
		}
	}

	// 빈 worktreePath 검사
	if result.WorktreePath == "" {
		return &WorktreePathInvalidError{
			AgentName: agentName,
			Reason:    "empty WorktreePath",
		}
	}

	// SPEC-V3R3-RETIRED-AGENT-001 D-EVAL-02 fix:
	// 5-layer defect chain의 Layer 4 산물인 literal "{}" 또는 "[object Object]"
	// 패턴 거부. AC-RA-18 critical assertion 충족 (validation layer raises
	// WORKTREE_PATH_INVALID before any path interpolation).
	switch result.WorktreePath {
	case "{}", "[object Object]", "null", "undefined":
		return &WorktreePathInvalidError{
			AgentName: agentName,
			Reason:    "literal stringified-object pattern in WorktreePath: " + result.WorktreePath,
		}
	}

	return nil
}
