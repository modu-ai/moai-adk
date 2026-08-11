package cli

// @MX:NOTE: [AUTO] cc command switches LLM backend to Claude-only mode
// @MX:NOTE: [AUTO] Removes GLM env vars and resets team mode before launching Claude Code
// @MX:NOTE: [AUTO] Supports profile switching via CLAUDE_CONFIG_DIR
// @MX:NOTE: [AUTO] M6-S1 DDD: cc is a thin delegate-only entry point; print sites live in launcher.go::launchClaudeDefault

import (
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// findProjectRootFn is the function used to locate the project root.
// Tests may override this to return a temporary directory, preventing
// any file mutations from reaching the real project or home directory.
var findProjectRootFn = findProjectRoot

var ccCmd = &cobra.Command{
	Use:   "cc [-p profile] [-- claude-args...]",
	Short: "Launch Claude Code with Claude backend",
	Long: `Launch Claude Code with Claude backend.

This command:
  1. Removes GLM-specific environment variables from .claude/settings.local.json
  2. Resets team mode if it was enabled (glm or cg)
  3. Optionally sets a profile via -p flag (CLAUDE_CONFIG_DIR)
  4. Reads DO_CLAUDE_* settings and converts them to CLI flags
  5. Launches Claude Code via exec (replaces current process)

Flags:
  -p, --profile <name>          Use a named Claude profile (~/.moai/claude-profiles/<name>/)
  --permission-mode <mode>      Set permission mode (default, acceptEdits, plan, auto, bypassPermissions, dontAsk)
  -b, --bypass                  Shorthand for --permission-mode bypassPermissions
  -c, --continue                Continue previous session
  -m, --model <model>           Override model selection
  -w, --worktree [name]         Launch in an isolated git worktree (.claude/worktrees/<name>/);
                                name omitted = auto-generated (same as claude --worktree)
      --spawn                   Run this command in a new tmux window instead of
                                replacing the current session (requires tmux)
  --chrome / --no-chrome        Toggle Chrome MCP

Permission Modes:
  default            Ask permissions for file edits and commands
  acceptEdits        Auto-accept file edits, ask for commands (project default)
  plan               Read-only exploration and planning
  auto               Background classifier checks actions (requires Team plan + Sonnet/Opus 4.6)
  bypassPermissions  Skip all checks (isolated environments only)
  dontAsk            Only pre-approved tools

Examples:
  moai cc                              # Default profile, launch Claude
  moai cc -p work                      # Use 'work' profile
  moai cc --permission-mode auto       # Launch with auto mode
  moai cc -p work -- --print           # Profile + pass-through args to Claude
  moai cc -w feat-login                # Launch in isolated worktree 'feat-login'
  moai cc -w                           # Launch in auto-named isolated worktree
  moai cc -w feat-login --spawn        # Teammate session in a new tmux window`,
	GroupID:            "launch",
	DisableFlagParsing: true,
	RunE:               runCC,
}

func init() {
	rootCmd.AddCommand(ccCmd)
}

// runCC switches the LLM backend to Claude, then launches Claude Code.
func runCC(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
		if arg == "--" {
			break
		}
	}

	// --spawn re-issues this same command in a new tmux window instead of
	// replacing the current process. It runs before any settings mutation so a
	// failed spawn leaves the environment untouched; the spawned `moai cc`
	// performs the mutations itself.
	if spawnArgs, spawn := stripSpawnFlag(args); spawn {
		return spawnLaunch(cmd.OutOrStdout(), "cc", spawnArgs)
	}

	profileName, filteredArgs, err := parseProfileFlag(args)
	if err != nil {
		return err
	}
	// SPEC-FACTORY-MODE-001: --factory / -f seeds a plan -> run -> verify -> sync
	// chain in the launched session. Parsed after --spawn is stripped (a spawned
	// session re-issues this command and must carry the token through) and before
	// worktree handling (so a factory token can never be mistaken for a -w value).
	// The environment mutation is restored on every return path, including error.
	if specID, factoryEnabled, factoryArgs := parseFactoryFlag(filteredArgs); factoryEnabled {
		filteredArgs = factoryArgs
		defer enterFactoryMode(specID)()
		recordFactorySession(specID, factory.BackendClaude)
	} else if label, isCompanion := parseCompanionLabel(filteredArgs); isCompanion {
		// A companion of a factory run: same raised Stop-hook block cap as the
		// lead, no chain seed. --name belongs to claude, so it is recognized
		// here and left in filteredArgs.
		defer enterFactoryCompanionMode(label)()
	}
	// SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a: validate absolute-path -w values
	// BEFORE normalizeWorktreeFlag so out-of-prefix paths are rejected with a
	// clear error (AC-WES-010c) and L2 (~/.moai/worktrees/) paths are accepted
	// (AC-WES-010a). normalizeWorktreeFlag remains the owner of short-name
	// token normalization (AC-WES-010b).
	if err := resolveWorktreeL2Path(filteredArgs); err != nil {
		return err
	}
	filteredArgs = normalizeWorktreeFlag(filteredArgs)
	return unifiedLaunch(profileName, "claude", filteredArgs)
}
