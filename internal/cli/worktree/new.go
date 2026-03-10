package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// userHomeDirFunc resolves the user's home directory.
// Overridable in tests.
var userHomeDirFunc = os.UserHomeDir

// getProjectNameFunc resolves the project name for worktree path construction.
// Overridable in tests.
var getProjectNameFunc = func() string { return detectProjectName(".") }

// legacyWorktreeDir is the project-local worktrees path checked for migration warnings.
// Overridable in tests.
var legacyWorktreeDir = filepath.Join(".", ".moai", "worktrees")

func newNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [branch-name]",
		Short: "Create a new worktree",
		Long: `Create a new Git worktree for the given branch name.
If the branch does not exist, it is created automatically.

SPEC-ID patterns (e.g., SPEC-AUTH-001) are automatically converted
to branch names using the feature/ prefix convention.`,
		Args: cobra.ExactArgs(1),
		RunE: runNew,
	}
	cmd.Flags().String("path", "", "Custom path for the worktree (default: .moai/worktrees/<SPEC-ID> for SPEC IDs, ../<branch-name> otherwise)")
	cmd.Flags().String("base", "main", "Base branch to create the worktree from")
	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	specID := args[0]
	branchName := resolveSpecBranch(specID)

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	wtPath, err := cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("get path flag: %w", err)
	}
	if wtPath == "" {
		homeDir, err := userHomeDirFunc()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		projectName := getProjectNameFunc()
		wtPath = filepath.Join(homeDir, ".moai", "worktrees", projectName, specID)
		if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
			return fmt.Errorf("create worktree directory: %w", err)
		}
	}

	// R5: warn when legacy project-local worktrees exist.
	if dirHasEntries(legacyWorktreeDir) {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: Legacy worktrees detected in .moai/worktrees/. Consider moving to ~/.moai/worktrees/{Project}/.")
	}

	if err := WorktreeProvider.Add(wtPath, branchName); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard(
		fmt.Sprintf("Created worktree for branch %s", branchName),
		fmt.Sprintf("Path: %s", wtPath),
	))
	return nil
}

// dirHasEntries returns true when dir exists and contains at least one entry.
func dirHasEntries(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) > 0
}

// resolveSpecBranch converts SPEC-ID patterns to branch names.
// e.g., "SPEC-AUTH-001" -> "feature/SPEC-AUTH-001"
// Regular branch names pass through unchanged.
func resolveSpecBranch(name string) string {
	if isSpecID(name) {
		return "feature/" + name
	}
	return name
}

// isSpecID checks if the given name matches the SPEC-ID pattern.
// Pattern: SPEC-<CATEGORY>-<NUMBER> (e.g., SPEC-AUTH-001, SPEC-UI-042)
func isSpecID(name string) bool {
	if !strings.HasPrefix(name, "SPEC-") {
		return false
	}
	parts := strings.SplitN(name, "-", 3)
	return len(parts) >= 3 && parts[2] != ""
}

// GenerateTmuxSessionName은 R5.1: tmux 세션 이름 생성 패턴을 구현합니다.
// SPEC-WORKTREE-002 요구사항: moai-{ProjectName}-{SPEC-ID}
//
// @MX:NOTE: SPEC-WORKTREE-002 R5.1 구현 - tmux 세션 이름 패턴
// @MX:SPEC: SPEC-WORKTREE-002
//
// Args:
//   - projectName: 프로젝트 이름 (예: "moai-adk-go")
//   - specID: SPEC 식별자 (예: "SPEC-WORKTREE-002")
//
// Returns:
//   - tmux 세션 이름 (예: "moai-moai-adk-go-SPEC-WORKTREE-002")
//
// Example:
//   sessionName := GenerateTmuxSessionName("moai-adk-go", "SPEC-WORKTREE-002")
//   fmt.Println(sessionName) // Output: moai-moai-adk-go-SPEC-WORKTREE-002
func GenerateTmuxSessionName(projectName, specID string) string {
	return fmt.Sprintf("moai-%s-%s", projectName, specID)
}

// ShouldAutoMerge는 R3: auto-merge 기본 동작을 구현합니다.
// SPEC-WORKTREE-002 요구사항: worktree 흐름에서 auto-merge가 기본값
//
// @MX:NOTE: SPEC-WORKTREE-002 R3 구현 - auto-merge 기본값
// @MX:SPEC: SPEC-WORKTREE-002
//
// Args:
//   - noMergeFlag: --no-merge 플래그가 설정되었는지 여부
//
// Returns:
//   - auto-merge를 수행해야 하면 true, 아니면 false
//
// Behavior:
//   - 기본값(true): 플래그가 없으면 auto-merge 수행
//   - --no-merge 플래그: false 반환하여 auto-merge 건너뜀
//
// Example:
//   shouldMerge := ShouldAutoMerge(false) // true
//   shouldMerge := ShouldAutoMerge(true)  // false
func ShouldAutoMerge(noMergeFlag bool) bool {
	// R3: auto-merge는 기본값(true)
	// --no-merge 플래그로만 비활성화 가능
	return !noMergeFlag
}
