// Resolution: NEW — EnterWorktree/ExitWorktree re-stamp (t236 / issue #1640).
package hook

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/modu-ai/moai-adk/internal/config"
)

// @MX:NOTE: [AUTO] t236/#1640 — new integration point on the PostToolUse
// surface: Claude Code emits no CwdChanged for tool-based tree moves (live
// measurement 2026-09-02: registry cwd stayed at the launch tree after an
// EnterWorktree with intact CwdChanged wiring — .moai/reports/t236/
// reproduction-evidence.md L1), so the env-file stamp and the session-registry
// relocation ride PostToolUse instead.
//
// handleWorktreeMove reacts to the EnterWorktree/ExitWorktree PostToolUse
// events: it re-stamps MOAI_PROJECT_DIR into CLAUDE_ENV_FILE for the new tree
// and keeps the session registry's cwd in step with the move, then warns that
// the already-running moai mcp-server still resolves its project root from
// spawn-frozen state and that catalog tools take an explicit project_root.
func handleWorktreeMove(input *HookInput) *HookOutput {
	newCwd := input.CWD
	if newCwd == "" {
		newCwd = input.NewCwd
	}
	if newCwd == "" {
		slog.Warn("worktree move: hook input carried no cwd; nothing re-stamped",
			"session_id", input.SessionID,
		)
		return &HookOutput{}
	}

	// Keep the session registry's cwd in step with the move — the same
	// fail-open call the CwdChanged handler makes — so anchor detection sees
	// sessions that entered a worktree mid-session.
	relocateSessionCwd(input, newCwd)

	stamped := false
	if envFile := os.Getenv(config.EnvClaudeEnvFile); envFile != "" {
		stampProjectDirEnv(envFile, newCwd)
		stamped = true
	}

	return &HookOutput{
		SystemMessage: fmt.Sprintf(
			"Session tree moved to %s. The running moai mcp-server still resolves its project "+
				"root from spawn-frozen state (CLAUDE_PROJECT_DIR env / server cwd), so a session that "+
				"moved worktrees is reading another tree: pass project_root = `git rev-parse --show-toplevel` "+
				"on the spec/verify/codex/glm/audit catalog tools that accept it. "+
				"(MOAI_PROJECT_DIR re-stamped in the env file: %t.)",
			newCwd, stamped),
	}
}
