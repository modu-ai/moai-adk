// SPEC-SEC-HARDEN-002 M2a (A-F1, HIGH) — worktree new path-traversal 거부 테스트.
//
// 재현 우선(reproduction-first): 가드 추가 전에는 traversal arg가
// filepath.Join(homeDir, ".moai", "worktrees", projectName, specID)에 도달해
// 봉쇄 디렉터리(~/.moai/worktrees/<project>/) 밖에 worktree 디렉터리를
// 생성했다(ALLOW). 가드 추가 후 CLI args[0] 경계에서 specid.ValidateNoTraversal로
// 거부된다(DENY).
//
// worktree new는 SPEC-ID 또는 브랜치명(예: "fix/something")을 모두 받는
// polymorphic arg이므로 "/"는 정상이며 NO-REG로 허용을 검증한다 — ".." 와
// 절대 경로(봉쇄 탈출)만 거부한다. 가드는 WorktreeProvider nil 체크보다 먼저 발동한다.
package worktree

import (
	"strings"
	"testing"
)

// guardRejection은 에러가 M2a traversal 가드의 거부인지 판별한다.
func guardRejection(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "path traversal") || strings.Contains(msg, "absolute path")
}

func TestRunNew_PathTraversalSpecIDRejected(t *testing.T) {
	// RED: traversal arg는 M2a 가드가 거부해야 한다 (현행 코드에선 ALLOW였음).
	traversal := []string{
		"../../../../tmp/evil", // 봉쇄 디렉터리 탈출
		"/tmp/evil",            // 절대 경로
		"fix/../../../etc",     // 브랜치명에 임베드된 ".."
	}
	for _, arg := range traversal {
		arg := arg
		t.Run("reject "+arg, func(t *testing.T) {
			cmd := newNewCmd()
			cmd.SetArgs([]string{arg})
			out := &strings.Builder{}
			cmd.SetOut(out)
			cmd.SetErr(out)

			err := cmd.RunE(cmd, []string{arg})
			if !guardRejection(err) {
				t.Fatalf("worktree new %q: want M2a traversal-guard rejection, got %v", arg, err)
			}
		})
	}

	// NO-REG: 브랜치명에 "/"가 포함되는 정상 입력은 가드를 통과해야 한다.
	// (A-F1 fix가 worktree new의 브랜치명 인자를 깨뜨리면 안 됨 —
	//  characterization 테스트 TestRunNew_DefaultPath/branch_with_slash_uses_global_path도 동일.)
	noReg := []string{"fix/something", "SPEC-SEC-HARDEN-002"}
	for _, arg := range noReg {
		arg := arg
		t.Run("allow "+arg, func(t *testing.T) {
			cmd := newNewCmd()
			cmd.SetArgs([]string{arg})
			out := &strings.Builder{}
			cmd.SetOut(out)
			cmd.SetErr(out)

			err := cmd.RunE(cmd, []string{arg})
			// 가드는 통과해야 한다 — 하류(WorktreeProvider nil 등) 에러는 무방하나
			// traversal-가드 거부 에러는 아니어야 한다.
			if guardRejection(err) {
				t.Fatalf("worktree new %q: M2a guard wrongly rejected a legitimate branch/SPEC arg: %v", arg, err)
			}
		})
	}
}
