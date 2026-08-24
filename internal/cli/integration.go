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
		return dir
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
	cmd.AddCommand(newIntegrationStatusCmd(), newIntegrationAcquireCmd(), newIntegrationReleaseCmd())
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "release-integration window: %s\n  holder:   %s (pid %d)\n  branch:   %s\n  worktree: %s\n  since:    %s\n",
				state, lock.SessionID, lock.PID, lock.Branch, lock.Worktree, lock.AcquiredAt)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newIntegrationAcquireCmd() *cobra.Command {
	var sessionFlag, nameFlag, branchFlag, cardFlag string
	var force, jsonOut bool
	cmd := &cobra.Command{
		Use:   "acquire",
		Short: "Record this session as the holder of the release-integration window",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := integrationSessionID(sessionFlag)
			if sessionID == "" {
				return fmt.Errorf("cannot resolve this session's id; pass --session <id> (a lock with an invented holder can be neither released by its holder nor recognized by the guard)")
			}
			root := integrationLockRoot()
			wt, _ := os.Getwd()
			branch := branchFlag
			if branch == "" {
				branch = currentBranch()
			}
			replaced, err := kanban.AcquireIntegrationLock(root, kanban.IntegrationLock{
				SessionID:   sessionID,
				SessionName: nameFlag,
				Branch:      branch,
				Worktree:    wt,
				Card:        cardFlag,
			}, force)
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"acquired": true,
					"session":  sessionID,
					"branch":   branch,
					"replaced": replaced,
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
	cmd.Flags().StringVar(&branchFlag, "branch", "", "Branch being integrated (default: current branch)")
	cmd.Flags().StringVar(&cardFlag, "card", "", "Card id this integration belongs to")
	cmd.Flags().BoolVar(&force, "force", false, "Take the window over from a live holder (recorded, never silent)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
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
