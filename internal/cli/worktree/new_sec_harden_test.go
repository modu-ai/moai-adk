// SPEC-SEC-HARDEN-002 M2a (A-F1, HIGH) — worktree new path-traversal 거부 테스트.
//
// 재현 우선(reproduction-first): 가드 추가 전에는 traversal SPEC-ID가
// filepath.Join(homeDir, ".moai", "worktrees", projectName, specID)에 도달해
// 봉쇄 디렉터리(~/.moai/worktrees/<project>/) 밖에 worktree 디렉터리를
// 생성했다(ALLOW). 가드 추가 후 CLI args[0] 경계에서 specid.ValidateSpecID로
// 거부된다(DENY). 가드가 WorktreeProvider nil 체크보다 먼저 발동함을 검증한다.
package worktree

import (
	"strings"
	"testing"
)

func TestRunNew_PathTraversalSpecIDRejected(t *testing.T) {
	tests := []struct {
		name   string
		specID string
	}{
		{name: "dotdot traversal", specID: "../../../../tmp/evil"},
		{name: "absolute path unix", specID: "/tmp/evil"},
		{name: "forward slash separator", specID: "foo/bar"},
		{name: "backslash separator", specID: "foo\\bar"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd := newNewCmd()
			cmd.SetArgs([]string{tt.specID})
			out := &strings.Builder{}
			cmd.SetOut(out)
			cmd.SetErr(out)

			err := cmd.RunE(cmd, []string{tt.specID})
			if err == nil {
				t.Fatalf("worktree new %q = nil error, want path-traversal rejection (M2a A-F1 HIGH)", tt.specID)
			}
			// 가드는 WorktreeProvider nil 체크보다 먼저 발동해야 한다 — 에러가
			// SPEC-ID 검증 에러여야 하며 "worktree manager not initialized"가 아니어야 함.
			if strings.Contains(err.Error(), "worktree manager not initialized") {
				t.Fatalf("M2a guard did not fire before WorktreeProvider check for %q: %v", tt.specID, err)
			}
		})
	}
}
