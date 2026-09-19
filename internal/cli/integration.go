package cli

// integration.go — `moai integration`, the lane-facing surface of the
// release-integration holder lock (card t194).
//
// The doctrine (`kanban-dispatch.md` § Integration into the release branch is
// self-served) serializes lanes by announcement. Card t181 wrote that rule and
// named its gap: announcement is a social protocol with nothing behind it.
// These three verbs are what a lane runs so the announcement leaves a record
// the PreToolUse guard can read.
//
// The lock record lives in the PRIMARY checkout's .moai/state, not in the
// caller's worktree: a serialization point visible from only one of the
// serialized trees is not one. See integrationLockRoot for how that directory
// is resolved.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/spf13/cobra"
)

// integrationLockRoot resolves the directory whose .moai/state holds the lock
// record — the primary checkout, shared by every linked worktree.
//
// CLAUDE_PROJECT_DIR is consulted FIRST, following this package's B7
// convention. It also happens to be exactly right here: Claude Code sets it to
// the PROJECT root even for a session working inside a worktree, which is the
// measured behaviour SPEC-MCP-WORKTREE-ROOT-001 had to work around and is the
// property this lock wants.
//
// The git fallback covers a plain shell with no Claude Code environment:
// `git rev-parse --git-common-dir` answers `<primary>/.git` from inside a
// worktree and `.git` from the primary itself, so the parent of its absolute
// form is the root in both cases. A git failure degrades to cwd rather than
// erroring — a lane that cannot resolve the shared root still gets a usable
// local answer, and the guard fails open on the same uncertainty.
func integrationLockRoot() string {
	if dir := strings.TrimSpace(os.Getenv("CLAUDE_PROJECT_DIR")); dir != "" {
		return canonicalTreeSpelling(dir)
	}
	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err == nil {
		p := strings.TrimSpace(string(out))
		if p != "" {
			if !filepath.IsAbs(p) {
				if abs, absErr := filepath.Abs(p); absErr == nil {
					p = abs
				}
			}
			if filepath.Base(p) == ".git" {
				return filepath.Dir(p)
			}
		}
	}
	return resolveProjectDir()
}

// canonicalTreeSpelling returns git's own spelling of dir's top level, and dir
// verbatim whenever git cannot answer for it (card t766).
//
// The spelling matters because everything downstream JOINS onto this string —
// the preserved copy's directory, and the `preserved_path` recorded beside it
// — while the sibling fields `source_path` and `worktree` come from
// `git rev-parse --show-toplevel`, which answers the on-disk spelling. One row
// could therefore name a tree two different ways, and the real ledger does:
// three rows from 2026-09-08 carry a `preserved_path` under `/Users/goos/moai/`
// against a `source_path` under `/Users/goos/MoAI/`. On a case-insensitive
// volume that is a record defect; on a case-sensitive one it is a second
// directory tree, and a preserved copy filed where nobody will look for it.
//
// Two properties bound the repair, and neither is incidental:
//
//   - git is the source, not `filepath.EvalSymlinks`. EvalSymlinks does NOT
//     correct case on macOS — it resolves links and returns whatever spelling
//     it was handed — so it cannot see this defect at all.
//
//   - The answer is accepted only when it names the SAME DIRECTORY, compared
//     by identity (os.SameFile) rather than by string. `--show-toplevel` walks
//     UP: handed a plain subdirectory of some repository, it answers that
//     repository's root, and relocating the lock root there would be a
//     different and much larger change than fixing a spelling. A path that is
//     not a repository, or is not its top level, keeps behaving exactly as it
//     does today.
func canonicalTreeSpelling(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return dir
	}
	top := strings.TrimSpace(string(out))
	if top == "" || top == dir {
		return dir
	}
	given, givenErr := os.Stat(dir)
	found, foundErr := os.Stat(top)
	if givenErr != nil || foundErr != nil || !os.SameFile(given, found) {
		return dir
	}
	return top
}

// integrationSessionID resolves the caller's own session id: the --session
// flag wins, then the per-process env var Claude Code stamps into every
// subprocess, then the side-channel file the SessionStart hook writes.
//
// The env var precedes the file because the file is one slot per project and
// names whichever session started last — a lock taken under a foreign id is
// held by nobody who can release it. The file remains the degraded path for
// runtimes that do not stamp the var.
//
// An empty result is returned as such rather than substituted with a
// placeholder. A lock whose holder is a made-up identity cannot be released by
// its holder and cannot be recognized by the guard, so a missing id is a
// blocker to report, never a value to invent.
func integrationSessionID(explicit string) string {
	if s := strings.TrimSpace(explicit); s != "" {
		return s
	}
	if id := strings.TrimSpace(os.Getenv(config.EnvClaudeCodeSessionID)); id != "" {
		return id
	}
	data, err := os.ReadFile(filepath.Join(integrationLockRoot(), session.CurrentSideChannelFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// currentBranch reports the checked-out branch of the caller's tree, for the
// record's `branch` field. Best-effort: the field is for a human reading a
// refusal, not a decision input.
func currentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// resolveIntegrationTarget returns the (branch, worktree) the window actually
// locks. Resolution order: the explicit --branch flag, then the project's
// configured git-flow develop branch (LoadGitFlowDevelopBranch), then the
// caller's own tree.
//
// The ordering exists because acquire used to record the CALLER's cwd and the
// CALLER's checked-out branch (card t449), while the window it serializes is
// the integration worktree — a lane sitting in its own card worktree recorded
// that card tree against a window whose whole purpose was the develop tree.
// The caller fallback is deliberate and NOT that defect: when no integration
// branch is configured, the caller's tree genuinely is the tree being
// integrated, so the caller's branch and cwd stay correct there.
//
// source names the tier that won (card t637). It is decided here, from the
// same trimmed values that decide the branch, so the recorded provenance can
// never disagree with the recorded branch.
func resolveIntegrationTarget(explicitBranch, configuredBranch string) (branch, worktree, source string) {
	target := strings.TrimSpace(explicitBranch)
	source = kanban.BranchSourceFlag
	if target == "" {
		target, source = strings.TrimSpace(configuredBranch), kanban.BranchSourceConfig
	}
	if target != "" {
		// An honest unknown beats a confidently wrong path: no worktree has
		// the branch checked out, so the record carries an empty worktree for
		// a human to read as "not provisioned yet" rather than a path that
		// names some unrelated tree.
		return target, worktreeForBranch(target), source
	}
	wt, _ := os.Getwd()
	return currentBranch(), wt, kanban.BranchSourceCaller
}

// integrationFallbackWarning is the one-line standard-error warning for a
// git-flow project whose develop branch is empty (card t637): the window was
// recorded against the caller's own branch, which is the integration target
// only if the caller happens to be standing in it. The prefix is the
// integration-lock family's, shared with the guard's advisory lines.
func integrationFallbackWarning(branch string) string {
	return fmt.Sprintf("[moai:integration-lock] warning: git-flow project with no develop branch configured; the window was recorded against the caller's branch %q. Set git_strategy.manual.develop_branch, or pass --branch <integration-target>.", branch)
}

// integrationInvalidWorkflowWarning is the one-line standard-error warning
// for a git_strategy workflow value outside the allowed set (card t656,
// REQ-GWS-003): it names the offending value and the allowed set, and names
// the same caller-fallback consequence as the t637 warning. Fail-open — the
// window stands; diagnosis is the whole job of this warning.
func integrationInvalidWorkflowWarning(value string) string {
	return fmt.Sprintf("[moai:integration-lock] warning: git_strategy workflow %q is not one of the allowed flows (%s); the window was recorded against the caller's branch. Repair the value in .moai/config/sections/git-strategy.yaml.",
		value, strings.Join(config.AllowedWorkflows(), ", "))
}

// integrationUnwiredTargetWarning is the one-line standard-error warning for
// a project whose git_strategy names an integration target that acquire does
// not read (card t886). acquire resolves through DevelopBranch alone, which
// is git-flow-gated by contract, so a github-flow / gitlab-flow / release-flow
// project falls back to the caller's branch even though the D2 interpretation
// table (config.WorkflowIntegrationTarget, REQ-GWS-004) already answered the
// question. REQ-GWS-008 left that seam unwired and named the follow-up; until
// it is wired, the disagreement is at least named. Fail-open, exactly like the
// t637 and t656 warnings: the record and the exit code do not move.
func integrationUnwiredTargetWarning(workflow, target, branch string) string {
	return fmt.Sprintf("[moai:integration-lock] warning: the %s integration target is %q, but the window was recorded against the caller's branch %q (acquire does not read that target yet). Pass --branch %s if the window is for the integration branch.",
		workflow, target, branch, target)
}

// worktreeForBranch returns the path of the worktree with branch checked out,
// or the empty string when none does (a failed git call included).
func worktreeForBranch(branch string) string {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return ""
	}
	return worktreeForBranchFromList(string(out), branch)
}

// worktreeForBranchFromList parses `git worktree list --porcelain` output and
// returns the path of the worktree holding branch, or "" when none does.
// (Extracted from worktreeForBranch so a caller — the SPEC-WORKTREE-KEY-WIRING-001
// tests — can resolve against a fixture repo other than the process cwd.)
//
// Records arrive in blocks — `worktree <path>` / `HEAD <sha>` / `branch
// refs/heads/<name>` / `bare` / `detached` — so each block's path is
// remembered until its branch line answers the question. Paths are taken
// whole from git's output and converted with filepath semantics, never split
// on a separator: git prints forward slashes even on Windows, and the
// recorded path should read like every other path this CLI writes.
func worktreeForBranchFromList(porcelain, branch string) string {
	wtPath := ""
	for _, line := range strings.Split(porcelain, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			wtPath = filepath.FromSlash(strings.TrimPrefix(line, "worktree "))
		case strings.HasPrefix(line, "branch "):
			if strings.TrimPrefix(line, "branch ") == "refs/heads/"+branch && wtPath != "" {
				return wtPath
			}
		}
	}
	return ""
}

func newIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration [command]",
		Short: "Hold and release the release-integration window (lane serialization)",
		Long: `Hold and release the release-integration window.

A lane announces its integration to the lead, then records the hold here so
the PreToolUse guard can refuse a second lane's ` + "`git merge`" + ` in the release
worktree. The record lives in the primary checkout's .moai/state, visible
from every linked worktree.

The deny layer is opt-in (workflow.integration_lock.enabled, default false);
these verbs work regardless, so a project may keep the record as a
coordination signal without enabling refusal.`,
	}
	cmd.AddCommand(newIntegrationStatusCmd(), newIntegrationAcquireCmd(), newIntegrationReleaseCmd(), newIntegrationPreflightCmd())
	return cmd
}

func newIntegrationStatusCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report who holds the release-integration window",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := integrationLockRoot()
			lock, err := kanban.ReadIntegrationLock(root)
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"held":  lock.Held(),
					"stale": lock.Stale(),
					"lock":  lock,
					"root":  root,
				})
			}
			if !lock.Held() {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "release-integration window: free")
				return nil
			}
			state := "held"
			if lock.Stale() {
				state = "held by a session that is gone (reclaimable)"
			}
			// The name is what a lane recognizes its own queue position by; the
			// bare id is opaque to the human deciding whether to reclaim. When
			// no name was recorded, today's shape stands.
			holder := fmt.Sprintf("%s (pid %d)", lock.SessionID, lock.PID)
			if lock.SessionName != "" {
				holder = fmt.Sprintf("%s (%s, pid %d)", lock.SessionName, lock.SessionID, lock.PID)
			}
			// Card t637: the card and the branch's provenance are printed only
			// when recorded, so a record without them — every record written
			// before they existed — keeps today's exact text shape.
			card := ""
			if lock.Card != "" {
				card = fmt.Sprintf("  card:     %s\n", lock.Card)
			}
			branch := lock.Branch
			if lock.BranchSource != "" {
				branch = fmt.Sprintf("%s (source: %s)", lock.Branch, lock.BranchSource)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "release-integration window: %s\n  holder:   %s\n%s  branch:   %s\n  worktree: %s\n  since:    %s\n",
				state, holder, card, branch, lock.Worktree, lock.AcquiredAt)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newIntegrationAcquireCmd() *cobra.Command {
	var sessionFlag, nameFlag, branchFlag, cardFlag string
	var force, jsonOut, allowSettingsDrift bool
	cmd := &cobra.Command{
		Use:   "acquire",
		Short: "Record this session as the holder of the release-integration window",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := integrationSessionID(sessionFlag)
			if sessionID == "" {
				return fmt.Errorf("cannot resolve this session's id; pass --session <id> (a lock with an invented holder can be neither released by its holder nor recognized by the guard)")
			}
			root := integrationLockRoot()

			// The settings-drift precondition runs BEFORE the record is
			// written (card t488). Its detection, preservation and ledger row
			// are unconditional; only its refusal is gated on
			// workflow.settings_drift_gate.enabled. A refusal returns here, so
			// no window is taken — a gate that refuses and takes the window
			// anyway would be the worst of both.
			drift, driftErr := acquireSettingsDriftPrecondition(cmd, root, cardFlag, allowSettingsDrift)
			if driftErr != nil {
				return driftErr
			}

			// One read of the git strategy answers both questions: the develop
			// branch that drives resolution, and whether the project is
			// git-flow at all — the latter only decides the warning below.
			gitFlow := config.LoadGitFlowIntegrationConfig(root)
			branch, wt, source := resolveIntegrationTarget(branchFlag, gitFlow.DevelopBranch)
			// The pid recorded is the OWNING SESSION's, never this process's.
			// This command exits the moment it returns, so its own pid is dead
			// before any reader probes it — recording it made every window read
			// as abandoned the instant it was taken. An unresolvable owner is
			// recorded as pid 0 rather than guessed at: pid 0 reads LIVE, so the
			// failure direction is "an operator must ask the holder to release"
			// rather than "two lanes merge at once".
			ownerPID, _ := session.ResolveOwnerPID()
			replaced, err := kanban.AcquireIntegrationLock(root, kanban.IntegrationLock{
				SessionID:    sessionID,
				SessionName:  nameFlag,
				PID:          ownerPID,
				PIDSource:    kanban.PIDSourceSessionOwner,
				Branch:       branch,
				BranchSource: source,
				Worktree:     wt,
				Card:         cardFlag,
				// Recorded only when a refusal was actually bypassed; the
				// precondition resolves that, so the flag alone does not stamp
				// the record.
				SettingsDriftBypass:    drift.Bypassed,
				SettingsDriftPreserved: settingsDriftBypassPreservedPath(drift),
			}, force)
			if err != nil {
				// A refused acquire recorded nothing, so there is no
				// fallback to warn about.
				return err
			}
			// Warn-only (card t637): a git-flow project whose develop branch
			// is empty fell back to the caller's tree silently. The window is
			// already recorded; the warning neither refuses nor changes
			// stdout, and it goes to the error writer so --json stays one
			// parseable object.
			if source == kanban.BranchSourceCaller && gitFlow.IsGitFlow() {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), integrationFallbackWarning(branch))
			}
			// Warn-only (card t656, REQ-GWS-003): an invalid workflow value is
			// diagnosed — the offending value and the allowed set named — but
			// the fallback, the record, and the exit code are exactly the
			// pre-change non-git-flow path. A valid non-git-flow choice is
			// never warned about.
			if gitFlow.Disposition == config.DispositionInvalid {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), integrationInvalidWorkflowWarning(gitFlow.Workflow))
			}
			// Warn-only (card t886): the config DID name an integration target
			// and acquire still fell back to the caller, because it reads the
			// git-flow-gated DevelopBranch rather than the flow-scoped target.
			// Silent only when there is nothing to say: the git-flow cell
			// resolves an empty target (the t637 warning owns it), and a caller
			// already standing in the named target is not a disagreement.
			if source == kanban.BranchSourceCaller && gitFlow.IntegrationTarget != "" && gitFlow.IntegrationTarget != branch {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), integrationUnwiredTargetWarning(gitFlow.Workflow, gitFlow.IntegrationTarget, branch))
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"acquired":       true,
					"session":        sessionID,
					"branch":         branch,
					"replaced":       replaced,
					"settings_drift": settingsDriftJSON(drift),
				})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "release-integration window acquired by %s on %s\n", sessionID, branch)
			if replaced != nil {
				// Never silent: the next lane must be able to say what it
				// cleared, and the displaced lane may still be alive.
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  displaced: %s (pid %d), held since %s\n", replaced.SessionID, replaced.PID, replaced.AcquiredAt)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session id to record as holder (default: this session)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Human-facing lane name recorded alongside the id")
	cmd.Flags().StringVar(&branchFlag, "branch", "", "The integration target branch the merge lands on, not the card branch being merged (default: the configured git-flow develop branch, else the current branch)")
	cmd.Flags().StringVar(&cardFlag, "card", "", "Card id this integration belongs to")
	cmd.Flags().BoolVar(&force, "force", false, "Take the window over from a live holder (recorded, never silent)")
	cmd.Flags().BoolVar(&allowSettingsDrift, "allow-settings-drift", false, "Record the window despite a refused settings-drift verdict (recorded in the lock, never silent). Deliberately separate from --force, which is a different decision")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

// settingsDriftBypassPreservedPath returns the preserved copy's path only when
// a refusal was actually bypassed. Recording it otherwise would put a bypass
// artefact on a record that bypassed nothing.
func settingsDriftBypassPreservedPath(r kanban.SettingsDriftResult) string {
	if !r.Bypassed {
		return ""
	}
	return r.PreservedPath
}

func newIntegrationReleaseCmd() *cobra.Command {
	var sessionFlag string
	var force, jsonOut bool
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release the release-integration window this session holds",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := integrationSessionID(sessionFlag)
			if sessionID == "" && !force {
				return fmt.Errorf("cannot resolve this session's id; pass --session <id> or --force")
			}
			released, err := kanban.ReleaseIntegrationLock(integrationLockRoot(), sessionID, force)
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"released": true,
					"lock":     released,
				})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "release-integration window released (was %s on %s)\n", released.SessionID, released.Branch)
			return nil
		},
	}
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session id whose hold to release (default: this session)")
	cmd.Flags().BoolVar(&force, "force", false, "Release a window held by a different session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func init() {
	rootCmd.AddCommand(newIntegrationCmd())
}
