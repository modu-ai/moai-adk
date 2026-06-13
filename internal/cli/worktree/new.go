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
	"github.com/modu-ai/moai-adk/internal/cli/specid"
	"github.com/modu-ai/moai-adk/internal/tmux"
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

// tmuxSessionFactory creates a new tmux.SessionManager.
// Overridable in tests to avoid real tmux CLI calls.
var tmuxSessionFactory = func() tmux.SessionManager {
	return tmux.NewSessionManager()
}

// gitRepoRootFunc returns the absolute path of the git repository root via
// `git rev-parse --show-toplevel`. Overridable in tests to avoid cwd leak:
// BODP audit trail must always anchor on git root, never os.Getwd(), because
// test processes run with cwd=package-dir which differs from the repo root.
//
// @MX:NOTE BODP audit trail path bug fix (SPEC-V3R3-CI-AUTONOMY-001 Round 7):
// os.Getwd() returns the package directory during test execution, causing the
// audit trail to leak into internal/cli/worktree/.moai/... — this var allows
// tests to inject a tempDir.
var gitRepoRootFunc = func() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	return strings.TrimSpace(string(out)), err
}

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
to branch names using the feature/ prefix convention.

Base branch selection (mutually exclusive):
  --base origin/main    Default. Fetches latest from remote before checkout.
                        Team-safe: includes any teammate-merged PRs even if
                        local main has not been pulled yet.
  --base main           Use local main ref. Includes local-only commits that
                        are not yet on the remote. Use this when you have
                        committed directly to main locally and want them
                        in the new worktree's parent.
  --from-current        Use current HEAD as base (any branch). Skips
                        'git fetch origin main' entirely.

For solo development with local-only commits, prefer --base main.
For team workflows where origin is the source of truth, the default
origin/main is safer because it always reflects the latest merged state.`,
		Args: cobra.ExactArgs(1),
		RunE: runNew,
	}
	cmd.Flags().String("path", "", "Custom path for the worktree (default: .moai/worktrees/<SPEC-ID> for SPEC IDs, ../<branch-name> otherwise)")
	cmd.Flags().String("base", bodp.DefaultBase, "Base branch (default origin/main, auto-fetched). Use --base main for local-only commits, --from-current for current HEAD.")
	cmd.Flags().Bool("from-current", false, "Use current HEAD as the worktree base (skips `git fetch origin main`)")
	cmd.Flags().Bool("tmux", false, "Create a tmux session after worktree creation")
	// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 M1: declare --team surface so M2 can
	// wire dispatch and M3's tmux tests can invoke the command consistently.
	// Dispatch logic lands in M2 (handoff_guidance.go + decidePattern wiring).
	cmd.Flags().Bool("team", false, "Spawn a Claude/GLM session in the new worktree (P1 tmux+CG → moai glm window, P2 tmux+CC → moai cc window, P3 no-tmux → in-process, P4 no-flag → handoff guidance)")
	// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 M2 / R9 / OQ-2: --team subsumes tmux
	// launching (P1/P2 paths). Combining --team with --tmux creates ambiguous
	// intent, so cobra rejects them before any worktree-creation side-effect.
	// AC-WTL-007 edge case: cobra reports the mutex error and exits non-zero.
	cmd.MarkFlagsMutuallyExclusive("team", "tmux")
	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	specID := args[0]
	// SPEC-SEC-HARDEN-002 M2a (A-F1): CLI args[0] 경계에서 path-traversal을
	// filepath.Join/os.MkdirAll/WorktreeProvider.Add 도달 전에 거부한다 (HIGH).
	// worktree new는 SPEC-ID 또는 브랜치명(예: "fix/something" — "/" 정상)을 모두
	// 받으므로 "/"는 허용하고 ".."/절대경로(봉쇄 탈출)만 거부하는 ValidateNoTraversal 사용.
	if err := specid.ValidateNoTraversal(specID); err != nil {
		return err
	}
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
	// @MX:REASON Prevent user-intent ambiguity (BODP rationale clarity).
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
	//
	// gitRepoRootFunc is used instead of os.Getwd() to ensure audit trail is
	// always written under the git repository root, not the process cwd (which
	// diverges during test execution to the package directory).
	repoRoot, rootErr := gitRepoRootFunc()
	if rootErr != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: cannot determine repo root for audit trail: %v\n", rootErr)
		repoRoot = "." // non-fatal fallback
	}
	if err := writeWorktreeAuditTrail(repoRoot, specID, branchName, effectiveBase); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: BODP audit trail write failed: %v\n", err)
	}

	_, _ = fmt.Fprintln(out, wtSuccessCard(
		fmt.Sprintf("Created worktree for branch %s", branchName),
		fmt.Sprintf("Path: %s", wtPath),
	))

	// R5: tmux session creation after worktree creation
	tmuxFlag, _ := cmd.Flags().GetBool("tmux")
	if tmuxFlag || isTmuxPreferred() {
		if isTmuxAvailableFunc() {
			projectName := getProjectNameFunc()
			cfg, err := BuildTmuxSessionConfig(projectName, specID, wtPath, ".")
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: tmux config build failed: %v\n", err)
			} else {
				mgr := tmuxSessionFactory()
				if err := CreateTmuxSession(cmd.Context(), cfg, mgr); err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: tmux session creation failed: %v\n", err)
					sessionName := GenerateTmuxSessionName(projectName, specID)
					_, _ = fmt.Fprintf(out, "To manually create session: tmux new-session -s %s -c %s\n", sessionName, wtPath)
				}
			}
		} else {
			// Graceful degradation: print manual instructions
			err := NewTmuxNotAvailableError(specID, wtPath)
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		}
	}

	// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 M2 — --team dispatch wiring.
	//
	// Cobra has already rejected `--team --tmux` via MarkFlagsMutuallyExclusive
	// before reaching runNew, so the two flags can never both be true here.
	//
	// Dispatch decisions:
	//   P4 (default, no --team)  → printHandoff on stdout
	//   P3 (--team, no tmux)     → launchP3 (POSIX syscall.Exec; Windows: handoff)
	//   P1/P2 (--team + in tmux) → M3 territory; for now fall back to handoff
	//                              with an info notice so the user still has a
	//                              paste-ready command while M3 lands.
	if err := dispatchTeamLaunch(cmd, repoRoot, specID, branchName, wtPath); err != nil {
		return err
	}

	return nil
}

// dispatchTeamLaunch is the M2 entry point for --team handling. It is invoked
// after worktree creation + BODP audit trail write + tmux (legacy --tmux)
// side-effects, so an error here does NOT roll back the worktree. P3 callers
// should be aware: on POSIX with --team and a valid moai binary, this function
// REPLACES the current process via syscall.Exec and never returns.
//
// @MX:NOTE Pattern dispatch is purely flag-driven; no user prompts (REQ-WTL-013).
func dispatchTeamLaunch(cmd *cobra.Command, repoRoot, specID, branchName, wtPath string) error {
	out := cmd.OutOrStdout()
	stderr := cmd.ErrOrStderr()

	teamFlag, _ := cmd.Flags().GetBool("team")
	inTmux := tmux.NewDetector().InTmuxSession()
	settingsPath := filepath.Join(repoRoot, ".claude", "settings.local.json")
	cgMode, cgErr := tmux.IsCGMode(settingsPath, stderr)
	if cgErr != nil {
		// settings.local.json corrupt JSON (per acceptance.md §2 edge case):
		// surface a warning, treat as CC mode (cgMode=false already by IsCGMode
		// error path) and continue. Worktree creation already succeeded.
		_, _ = fmt.Fprintf(stderr, "Warning: CG mode detection failed (assuming CC mode): %v\n", cgErr)
	}

	pattern := decidePattern(teamFlag, inTmux, cgMode)
	llm := "cc"
	if cgMode {
		llm = "glm"
	}
	cfg := TeamLaunchConfig{
		Pattern:      pattern,
		WorktreePath: wtPath,
		Branch:       branchName,
		SpecID:       specID,
		LLM:          llm,
		LaunchTime:   time.Now(),
	}

	switch pattern {
	case PatternP4Handoff:
		// Default no-flag path: print paste-ready instructions and exit 0.
		// REQ-WTL-008: no spawn occurred → no swarm registry entry.
		printHandoff(out, cfg)
		return nil

	case PatternP3InProgress:
		// M4 (REQ-WTL-008): syscall.Exec REPLACES the process image on
		// success, so the registry entry MUST be written BEFORE launchP3
		// otherwise the write code path is unreachable. On Windows launchP3
		// returns nil after printing handoff fallback — same ordering keeps
		// the registry consistent across platforms.
		entry := SwarmEntry{
			SpecID:       cfg.SpecID,
			WorktreePath: cfg.WorktreePath,
			Branch:       cfg.Branch,
			PaneID:       "", // P3 has no tmux pane
			Mode:         patternToMode(cfg.Pattern, cfg.LLM),
			CreatedAt:    time.Now().UTC(),
			CreatedByPID: os.Getpid(),
		}
		if regErr := WriteSwarmEntry(repoRoot, entry); regErr != nil {
			// Non-fatal: launch is about to happen / process about to be
			// replaced. Surface a stderr warning so the user knows the
			// registry is incomplete but proceed with the launch.
			_, _ = fmt.Fprintf(stderr, "Warning: failed to write swarm registry: %v\n", regErr)
		}
		// POSIX: launchP3 REPLACES the current process via syscall.Exec on
		// success and never returns. The error path here covers: moai binary
		// missing from PATH, worktree chdir failure, exec call failure.
		// Windows: launchP3 returns nil after printing the handoff fallback
		// (REQ-WTL-012).
		return launchP3(cfg)

	case PatternP1TmuxGLM, PatternP2TmuxCC:
		// M3 implementation — spawn a new tmux window in the caller's current
		// tmux session. On failure (tmux server down, syntax error, etc.) we
		// degrade to the P4 handoff path with a stderr notice (REQ-WTL-007).
		// Worktree creation already succeeded, so exit code remains 0 either
		// way.
		paneID, launchErr := launchP1P2(cfg)
		if launchErr != nil {
			// REQ-WTL-007 + REQ-WTL-008: pane spawn failure → P4 handoff
			// fallback AND no registry write. The "tmux pane spawn failed"
			// substring is the verification anchor for AC-WTL-007's stderr
			// assertion. Registry write is deferred until after the
			// launchP1P2 success branch so the failure path is structurally
			// guaranteed to skip it.
			printHandoffWithError(out, stderr, cfg, fmt.Sprintf("tmux pane spawn failed: %v", launchErr))
			return nil
		}
		// M4 (REQ-WTL-008): write swarm registry with the captured pane_id.
		// Non-fatal failure: tmux window has already spawned, so we cannot
		// undo it; surface a warning and continue.
		entry := SwarmEntry{
			SpecID:       cfg.SpecID,
			WorktreePath: cfg.WorktreePath,
			Branch:       cfg.Branch,
			PaneID:       paneID,
			Mode:         patternToMode(cfg.Pattern, cfg.LLM),
			CreatedAt:    time.Now().UTC(),
			CreatedByPID: os.Getpid(),
		}
		if regErr := WriteSwarmEntry(repoRoot, entry); regErr != nil {
			_, _ = fmt.Fprintf(stderr, "Warning: failed to write swarm registry: %v\n", regErr)
		}
		// Success: report the captured pane_id so the user can switch to the
		// new window (tmux: C-b w, or use `tmux select-window -t <pane>`).
		_, _ = fmt.Fprintf(out, "tmux window spawned in pane %s — running `moai %s` in %s\n", paneID, cfg.LLM, cfg.WorktreePath)
		return nil
	}
	// Defensive: decidePattern returns one of the four constants above; an
	// unreachable default is documented per Go style for completeness.
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
// @MX:NOTE Decision precedence: --from-current > --base > default (origin/main).
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
// repoRoot MUST be the git repository root returned by gitRepoRootFunc, NOT
// os.Getwd(). The two diverge during test execution: Go runs tests with
// cwd=package-dir, which would leak audit trail files into
// internal/cli/worktree/.moai/branches/decisions/ instead of the repo root.
//
// @MX:NOTE CLI path MUST NOT invoke user prompts — agent-common-protocol
// §User Interaction Boundary HARD (orchestrator-only). BODP signal collection
// is for audit-trail purposes only (decision input comes from the user choosing plan/worktree directly).
func writeWorktreeAuditTrail(repoRoot, specID, branchName, base string) error {
	currentBranch, _ := gitWorktreeCmd("rev-parse", "--abbrev-ref", "HEAD")
	currentBranch = strings.TrimSpace(currentBranch)
	decision, _ := bodp.Check(bodp.CheckInput{
		CurrentBranch: currentBranch,
		NewSpecID:     specID,
		RepoRoot:      repoRoot,
		EntryPoint:    bodp.EntryWorktreeCLI,
	})
	return bodp.WriteDecision(repoRoot, bodp.AuditEntry{
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
// Reads worktree.tmux_preferred from .moai/config/sections/workflow.yaml.
//
// @MX:NOTE: SPEC-WORKTREE-002 workflow config integration point
// @MX:SPEC: SPEC-WORKTREE-002
func isTmuxPreferred() bool {
	repoRoot, err := gitRepoRootFunc()
	if err != nil {
		return false
	}
	configPath := filepath.Join(repoRoot, ".moai", "config", "sections", "workflow.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "tmux_preferred:") {
			parts := strings.SplitN(trimmed, "tmux_preferred:", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				return value == "true"
			}
		}
	}
	return false
}
