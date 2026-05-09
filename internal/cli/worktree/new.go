package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/bodp"
)

// userHomeDirFunc resolves the user's home directory.
// Overridable in tests.
var userHomeDirFunc = os.UserHomeDir

// gitWorktreeCmd executes a git subcommand for the worktree CLI.
// Overridable in tests.
//
// @MX:NOTE Used by W7-T03 BODP integration (audit trail + git fetch). Default
// implementation shells out to git; tests inject fakes to avoid network/IO.
var gitWorktreeCmd = func(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

// getProjectNameFunc resolves the project name for worktree path construction.
// Overridable in tests.
var getProjectNameFunc = func() string { return detectProjectName(".") }

// isTmuxAvailableFunc checks if tmux is available.
// Overridable in tests.
var isTmuxAvailableFunc = IsTmuxAvailable

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
	cmd.Flags().String("base", bodp.DefaultBase, "Base branch to create the worktree from (default origin/main per BODP)")
	cmd.Flags().Bool("from-current", false, "Use current HEAD as the worktree base (skips `git fetch origin main`)")
	cmd.Flags().Bool("tmux", false, "Create a tmux session after worktree creation")
	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	specID := args[0]
	branchName := resolveSpecBranch(specID)

	if WorktreeProvider == nil {
		return fmt.Errorf("worktree manager not initialized (git module not available)")
	}

	// W7-T03: BODP flag handling — must precede any worktree side-effect.
	baseFlag, err := cmd.Flags().GetString("base")
	if err != nil {
		return fmt.Errorf("get base flag: %w", err)
	}
	fromCurrent, err := cmd.Flags().GetBool("from-current")
	if err != nil {
		return fmt.Errorf("get from-current flag: %w", err)
	}
	// @MX:WARN --base and --from-current must be mutually exclusive.
	// @MX:REASON 사용자 의도 모호성 방지 (BODP rationale clarity).
	if baseFlag != bodp.DefaultBase && fromCurrent {
		return fmt.Errorf("--base and --from-current are mutually exclusive")
	}
	effectiveBase := determineBase(baseFlag, fromCurrent)

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

	// BODP gate: fetch origin/main when defaulting; skip on --from-current.
	if effectiveBase == bodp.DefaultBase {
		if _, err := gitWorktreeCmd("fetch", "origin", "main"); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: git fetch origin main failed (continuing): %v\n", err)
		}
	}

	if err := WorktreeProvider.Add(wtPath, branchName); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}

	// BODP audit trail (W7-T04 reused). Failures are non-fatal — surface a
	// warning but keep the worktree creation result.
	if err := writeWorktreeAuditTrail(specID, branchName, effectiveBase, wtPath); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: BODP audit trail write failed: %v\n", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard(
		fmt.Sprintf("Created worktree for branch %s", branchName),
		fmt.Sprintf("Path: %s", wtPath),
	))

	// R5: tmux session creation after worktree creation
	// Check if --tmux flag is set
	tmuxFlag, _ := cmd.Flags().GetBool("tmux")
	if tmuxFlag || isTmuxPreferred() {
		if isTmuxAvailableFunc() {
			// Create tmux session with environment isolation
			projectName := getProjectNameFunc()
			_, err := BuildTmuxSessionConfig(projectName, specID, wtPath, ".")
			if err != nil {
				return fmt.Errorf("build tmux config: %w", err)
			}

			// Note: Using a nil tmux manager for now - will need proper initialization
			// This is a simplified implementation for the TDD cycle
			_, _ = fmt.Fprintln(out, "Tmux session creation requested but tmux manager not yet initialized.")
			_, _ = fmt.Fprintf(out, "To manually create session: tmux new-session -s %s -c %s\n", GenerateTmuxSessionName(projectName, specID), wtPath)
		} else {
			// Graceful degradation: print manual instructions
			err := NewTmuxNotAvailableError(specID, wtPath)
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		}
	}

	return nil
}

// dirHasEntries returns true when dir exists and contains at least one entry.
func dirHasEntries(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) > 0
}

// determineBase resolves the effective base branch from CLI flags. The caller
// is responsible for verifying mutual exclusion of --base and --from-current.
//
// @MX:NOTE 결정 우선순위: --from-current > --base > default (origin/main).
func determineBase(baseFlag string, fromCurrent bool) string {
	if fromCurrent {
		return "HEAD"
	}
	if baseFlag != "" {
		return baseFlag
	}
	return bodp.DefaultBase
}

// writeWorktreeAuditTrail records the BODP decision for a CLI-driven worktree
// creation. The CLI path never prompts the user (orchestrator-only HARD), so
// UserChoice mirrors the recommendation.
//
// @MX:NOTE CLI path은 사용자 프롬프트 호출 절대 금지 — agent-common-protocol
// §User Interaction Boundary HARD (orchestrator-only). BODP signal collection
// 은 audit trail 목적 (decision input은 사용자가 plan/worktree 직접 선택).
func writeWorktreeAuditTrail(specID, branchName, base, _ string) error {
	currentBranch, _ := gitWorktreeCmd("rev-parse", "--abbrev-ref", "HEAD")
	currentBranch = strings.TrimSpace(currentBranch)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	decision, _ := bodp.Check(bodp.CheckInput{
		CurrentBranch: currentBranch,
		NewSpecID:     specID,
		RepoRoot:      cwd,
		EntryPoint:    bodp.EntryWorktreeCLI,
	})
	return bodp.WriteDecision(cwd, bodp.AuditEntry{
		Timestamp:     time.Now().UTC(),
		EntryPoint:    bodp.EntryWorktreeCLI,
		CurrentBranch: currentBranch,
		NewBranch:     branchName,
		Decision:      decision,
		UserChoice:    decision.Recommended,
		ExecutedCmd:   fmt.Sprintf("git worktree add %s %s", branchName, base),
	})
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

// GenerateTmuxSessionName implements R5.1: tmux session name generation pattern.
// SPEC-WORKTREE-002 requirement: moai-{ProjectName}-{SPEC-ID}
//
// @MX:NOTE: SPEC-WORKTREE-002 R5.1 implementation - tmux session name pattern
// @MX:SPEC: SPEC-WORKTREE-002
//
// Args:
//   - projectName: the project name (e.g., "moai-adk-go")
//   - specID: the SPEC identifier (e.g., "SPEC-WORKTREE-002")
//
// Returns:
//   - the tmux session name (e.g., "moai-moai-adk-go-SPEC-WORKTREE-002")
//
// Example:
//
//	sessionName := GenerateTmuxSessionName("moai-adk-go", "SPEC-WORKTREE-002")
//	fmt.Println(sessionName) // Output: moai-moai-adk-go-SPEC-WORKTREE-002
func GenerateTmuxSessionName(projectName, specID string) string {
	return fmt.Sprintf("moai-%s-%s", projectName, specID)
}

// ShouldAutoMerge implements R3: default auto-merge behavior.
// SPEC-WORKTREE-002 requirement: auto-merge is the default in the worktree flow.
//
// @MX:NOTE: SPEC-WORKTREE-002 R3 implementation - auto-merge default
// @MX:SPEC: SPEC-WORKTREE-002
//
// Args:
//   - noMergeFlag: whether the --no-merge flag is set
//
// Returns:
//   - true if auto-merge should be performed, false otherwise
//
// Behavior:
//   - Default (true): performs auto-merge when the flag is absent
//   - --no-merge flag: returns false, skipping auto-merge
//
// Example:
//
//	shouldMerge := ShouldAutoMerge(false) // true
//	shouldMerge := ShouldAutoMerge(true)  // false
func ShouldAutoMerge(noMergeFlag bool) bool {
	// R3: auto-merge is the default (true)
	// Can only be disabled via the --no-merge flag
	return !noMergeFlag
}

// isTmuxPreferred checks if tmux session creation is preferred in workflow config.
// This is a placeholder for future workflow.yaml integration.
//
// @MX:NOTE: SPEC-WORKTREE-002 workflow config integration point
// @MX:SPEC: SPEC-WORKTREE-002
func isTmuxPreferred() bool {
	// TODO: Read from .moai/config/sections/workflow.yaml
	// For now, return false to require explicit --tmux flag
	return false
}
