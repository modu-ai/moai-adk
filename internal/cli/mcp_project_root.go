package cli

// SPEC-MCP-WORKTREE-ROOT-001 — let an MCP caller name the tree it means.
//
// resolveProjectDir() prefers CLAUDE_PROJECT_DIR, and Claude Code sets that to
// the PROJECT root even for a session working inside a worktree. Measured across
// four live moai mcp-server processes in two repositories: every server whose
// working directory was a worktree carried the primary checkout in that
// variable. The env branch wins, so the working directory is never consulted and
// the outcome is deterministic — an MCP audit issued from a worktree audits the
// primary checkout, and a SPEC that exists only on the card's branch is absent
// from the catalogue the auditor reads without anything reporting it missing.
//
// The server cannot infer the right answer: its own working directory is stale
// by construction (a long-lived subprocess cannot follow a worktree switch) and
// its environment names the project. The caller CAN — an agent with a shell runs
// `git rev-parse --show-toplevel`. So the tool takes it as an input.
//
// The repair sits here, at the handler boundary, rather than inside
// resolveProjectDir(): that function has other consumers whose correct answer is
// undecided (goal state, verification snapshots, the convergence state
// directory), and changing it would move them all at once.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// projectRootArg is the optional tool input naming the tree to act on.
const projectRootArg = "project_root"

// projectRootDesc is the shared input description. One string, so the tools
// cannot drift into describing the same parameter differently.
const projectRootDesc = "Optional project or worktree root to act on. Supply your own `git rev-parse --show-toplevel`. " +
	"Absent or empty resolves as before (CLAUDE_PROJECT_DIR, then the server's working directory), which in a " +
	"worktree session is the PRIMARY checkout rather than the worktree. An unusable path is rejected, never " +
	"silently replaced by the default."

// projectRootOption declares the optional input on a tool.
func projectRootOption() mcp.ToolOption {
	return mcp.WithString(projectRootArg, mcp.Description(projectRootDesc))
}

// resolveToolProjectRoot returns the project root a tool call should act on.
//
// An absent or empty project_root resolves exactly as the tool resolved it
// before this parameter existed, so a caller that never learns about it sees no
// change.
//
// A non-empty project_root that cannot be a project root is REJECTED rather than
// replaced by the default. The ergonomic alternative — ignore the bad path, use
// the default — is the one choice that reintroduces the very defect this
// parameter exists to fix: a caller who mistyped its own worktree path would be
// silently returned to acting on the primary checkout, and told it succeeded.
func resolveToolProjectRoot(req mcp.CallToolRequest) (string, error) {
	raw := strings.TrimSpace(req.GetString(projectRootArg, ""))
	if raw == "" {
		return resolveProjectDir(), nil
	}

	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("project_root %q cannot be resolved to an absolute path: %w", raw, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("project_root %q does not exist", raw)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("project_root %q is not a directory", raw)
	}

	moaiDir := filepath.Join(abs, ".moai")
	moaiInfo, err := os.Stat(moaiDir)
	if err != nil || !moaiInfo.IsDir() {
		return "", fmt.Errorf("project_root %q has no .moai directory, so it is not a MoAI project root", raw)
	}

	return abs, nil
}
