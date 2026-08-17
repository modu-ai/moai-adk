// Resolution: KEEP — worktree registry update at .moai/state/worktrees.json.
package hook

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
)

// worktreeCreateHandler processes WorktreeCreate events.
// Fired when Claude Code creates an isolated git worktree for an agent
// with isolation: worktree in its frontmatter (v2.1.49+).
type worktreeCreateHandler struct{}

// NewWorktreeCreateHandler creates a new WorktreeCreate event handler.
func NewWorktreeCreateHandler() Handler {
	return &worktreeCreateHandler{}
}

// EventType returns EventWorktreeCreate.
func (h *worktreeCreateHandler) EventType() EventType {
	return EventWorktreeCreate
}

// agentWorktreeBranchPrefix namespaces the branches this handler creates so
// they never collide with user branches.
const agentWorktreeBranchPrefix = "worktree-"

// agentWorktreeParentDir is the directory, relative to the repository root,
// under which hook-created worktrees live — mirroring Claude Code's native
// agent-worktree layout (.claude/worktrees/<name>/).
var agentWorktreeParentDir = filepath.Join(".claude", "worktrees")

// Handle processes a WorktreeCreate event as an ACTIVE CREATOR (issue #1570).
//
// Per the Claude Code contract (v2.1.49+; verified against the embedded hook
// schema of the 2.1.233 runtime), the stdin payload carries the suggested
// worktree slug in the official `name` field — NOT a `worktree_path` — and the
// hook MUST create the worktree directory itself and echo its absolute path
// to stdout as plain text:
//
//	Input to command is JSON with name (suggested worktree slug).
//	Stdout should contain the absolute path to the created worktree directory.
//	Exit code 0 - worktree created successfully
//	Other exit codes - worktree creation failed
//
// The previous passthrough-observer implementation echoed input.WorktreePath,
// which never arrives on this event, so stdout stayed empty and Claude Code
// aborted every isolation: worktree agent spawn with "hook succeeded but
// returned no worktree path".
//
// Creation mirrors the native agent-worktree layout (.claude/worktrees/<name>,
// branch worktree-<name>) with one deliberate simplification: the branch is
// cut from the current HEAD rather than origin/HEAD, so the isolated agent
// starts from the exact tree the spawning session sees.
//
// Failure contract: a non-nil error aborts creation (the CLI dispatcher
// exits non-zero) — the honest "worktree creation failed" signal. An existing
// directory at the target path is reused idempotently.
func (h *worktreeCreateHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("worktree create requested for isolated agent",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"worktree_name", input.WorktreeName,
	)

	if input.WorktreeName == "" {
		return nil, fmt.Errorf("worktree create: missing required input field %q — Claude Code sends the suggested worktree slug in name", "name")
	}

	cwd := input.CWD
	if cwd == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("worktree create: resolve working directory: %w", err)
		}
		cwd = wd
	}

	repoRoot, err := resolveWorktreeRepoRoot(cwd)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(repoRoot, agentWorktreeParentDir, input.WorktreeName)
	branch := agentWorktreeBranchPrefix + sanitizeWorktreeBranchSuffix(input.WorktreeName)

	// Idempotent reuse: an existing directory at the target path satisfies
	// the contract without a second git invocation (Claude Code validates
	// that the echoed path is a directory before handing it to the agent).
	if info, statErr := os.Stat(path); statErr == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("worktree create: %s exists and is not a directory", path)
		}
		slog.Info("reusing existing worktree for isolated agent",
			"worktree_path", path,
			"worktree_branch", branch,
		)
		h.registerEntry(input, path, branch)
		return &HookOutput{WorktreePath: path}, nil
	}

	// Add creates the branch when it does not exist and reuses it when it
	// does (both cases must stay spawnable across repeated creations of the
	// same slug).
	if err := gitcore.NewWorktreeManager(repoRoot).Add(path, branch); err != nil {
		return nil, fmt.Errorf("worktree create: %w", err)
	}

	slog.Info("worktree created for isolated agent",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"worktree_path", path,
		"worktree_branch", branch,
	)

	h.registerEntry(input, path, branch)

	return &HookOutput{WorktreePath: path}, nil
}

// registerEntry persists the worktree to the registry so other sessions can
// inspect active worktrees. Non-blocking on error (failures are logged).
func (h *worktreeCreateHandler) registerEntry(input *HookInput, path, branch string) {
	if input.CWD != "" {
		registerWorktree(input.CWD, path, branch, input.AgentName)
	}
}

// resolveWorktreeRepoRoot returns the absolute checkout root of the git
// repository containing dir. The hook may run with cwd deep inside the tree,
// so the root must be resolved rather than assumed.
func resolveWorktreeRepoRoot(dir string) (string, error) {
	full := []string{"-C", dir, "rev-parse", "--show-toplevel"}
	cmd := exec.Command("git", full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("worktree create: resolve repository root under %s: %w (%s)",
			dir, err, strings.TrimSpace(stderr.String()))
	}
	root := strings.TrimSpace(stdout.String())
	if root == "" || !filepath.IsAbs(root) {
		return "", fmt.Errorf("worktree create: repository root %q under %s is not an absolute path", root, dir)
	}
	return root, nil
}

// sanitizeWorktreeBranchSuffix converts a worktree slug (letters, digits,
// dots, underscores, dashes, and '/'-separated segments per Claude Code's
// name validation) into a single git-branch-safe path segment. The
// worktree- prefix added by the caller keeps the result from ever starting
// with '-' or '.'.
func sanitizeWorktreeBranchSuffix(name string) string {
	return strings.ReplaceAll(name, "/", "-")
}
