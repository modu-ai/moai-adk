// SPEC-SEC-HARDEN-002 M2b — git worktree-add argv option-smuggling guard (A-F3).
//
// 재현 우선: 헬퍼가 존재하기 전에는 미정의 심볼 참조로 컴파일 실패(RED).
// `worktree add` argv가 사용자 유래 operand(worktree path, branch) 앞에
// `--` end-of-options 구분자를 삽입함을 검증한다. 이로써 `-`로 시작하는
// operand(예: `--upload-pack=x`)가 git 옵션이 아닌 positional 인자로 처리된다
// (CWE-88 argv option smuggling 차단).
package git

import "testing"

func TestBuildWorktreeAddArgs_DashDashBeforeOperands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		branchExists bool
		path         string
		branch       string
		// wantDashDashBefore: argv 토큰 중 `--`가 등장하고, 그 다음 토큰부터가
		// 사용자 유래 operand여야 한다.
		userOperand string // 첫 사용자 유래 operand (이 토큰 직전에 `--`가 있어야 함)
	}{
		{
			name:         "existing branch: add -- path branch",
			branchExists: true,
			path:         "--upload-pack=x", // `-` 선두 operand (옵션 스머글링 시도)
			branch:       "feature/SPEC-AUTH-001",
			userOperand:  "--upload-pack=x",
		},
		{
			name:         "new branch: add -b branch -- path",
			branchExists: false,
			path:         "--upload-pack=x",
			branch:       "feature/SPEC-AUTH-001",
			userOperand:  "--upload-pack=x",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			args := buildWorktreeAddArgs(tt.branchExists, tt.path, tt.branch)

			// argv는 반드시 "worktree", "add"로 시작한다 (동작 보존).
			if len(args) < 3 || args[0] != "worktree" || args[1] != "add" {
				t.Fatalf("argv must start with [worktree add ...]; got %v", args)
			}

			// `--` 구분자가 argv에 존재해야 한다.
			dashIdx := -1
			for i, a := range args {
				if a == "--" {
					dashIdx = i
					break
				}
			}
			if dashIdx == -1 {
				t.Fatalf("argv missing `--` end-of-options separator; got %v", args)
			}

			// `--` 바로 다음 토큰이 첫 사용자 유래 operand여야 한다
			// (그래야 `-` 선두 operand가 positional로 처리됨).
			if dashIdx+1 >= len(args) {
				t.Fatalf("`--` must precede a user operand; got %v", args)
			}
			if args[dashIdx+1] != tt.userOperand {
				t.Errorf("token after `--` = %q, want first user operand %q; argv=%v",
					args[dashIdx+1], tt.userOperand, args)
			}
		})
	}
}

// TestBuildWorktreeAddArgs_NoRegressionLegitimate verifies NO-REG: legitimate
// path/branch still assemble a valid `worktree add` argv with the `--` guard
// and the operands intact, preserving worktree creation behavior.
func TestBuildWorktreeAddArgs_NoRegressionLegitimate(t *testing.T) {
	t.Parallel()

	// existing branch: worktree add -- <path> <branch>
	existing := buildWorktreeAddArgs(true, "/home/user/.moai/worktrees/proj/SPEC-A-001", "feature/SPEC-A-001")
	wantExisting := []string{"worktree", "add", "--", "/home/user/.moai/worktrees/proj/SPEC-A-001", "feature/SPEC-A-001"}
	if !equalArgs(existing, wantExisting) {
		t.Errorf("existing-branch argv = %v, want %v", existing, wantExisting)
	}

	// new branch: worktree add -b <branch> -- <path>
	newBranch := buildWorktreeAddArgs(false, "/home/user/.moai/worktrees/proj/SPEC-A-001", "feature/SPEC-A-001")
	wantNew := []string{"worktree", "add", "-b", "feature/SPEC-A-001", "--", "/home/user/.moai/worktrees/proj/SPEC-A-001"}
	if !equalArgs(newBranch, wantNew) {
		t.Errorf("new-branch argv = %v, want %v", newBranch, wantNew)
	}
}

func equalArgs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
