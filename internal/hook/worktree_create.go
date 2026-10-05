// Resolution: KEEP — worktree registry update at .moai/state/worktrees.json.
package hook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/gitenv"
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

	if err := validateWorktreeName(input.WorktreeName); err != nil {
		return nil, err
	}

	repoRoot, err := resolveWorktreeRepoRoot(cwd)
	if err != nil {
		return nil, err
	}

	parent := filepath.Join(repoRoot, agentWorktreeParentDir)
	path := filepath.Join(parent, input.WorktreeName)
	if err := ensureWithinWorktreeParent(parent, path); err != nil {
		return nil, err
	}
	branch := agentWorktreeBranchPrefix + sanitizeWorktreeBranchSuffix(input.WorktreeName)

	// Idempotent reuse: an existing REGISTERED worktree at the target path
	// satisfies the contract without a second git invocation (Claude Code
	// validates that the echoed path is a directory before handing it to the
	// agent). Lstat rather than Stat, and a registry lookup rather than a
	// bare IsDir: a plain directory or a symlink that merely happens to sit
	// at the target path is NOT an isolated worktree, and handing one back
	// silently de-isolates the agent that asked for isolation.
	if info, statErr := os.Lstat(path); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("worktree create: %s is a symlink, not a worktree directory", path)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("worktree create: %s exists and is not a directory", path)
		}
		if err := ensureRegisteredWorktree(repoRoot, parent, path); err != nil {
			return nil, err
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
	// Write-side resolver, never input.CWD: a subdirectory cwd would otherwise
	// grow a stray <subdir>/.moai/state/worktrees.json (card t1165).
	if root := resolveProjectRoot(input); root != "" {
		registerWorktree(root, path, branch, input.AgentName)
	}
}

// resolveWorktreeRepoRoot returns the absolute checkout root of the git
// repository containing dir. The hook may run with cwd deep inside the tree,
// so the root must be resolved rather than assumed.
func resolveWorktreeRepoRoot(dir string) (string, error) {
	full := []string{"-C", dir, "rev-parse", "--show-toplevel"}
	cmd := exec.Command("git", full...)
	// `git -C dir` only changes the directory; an inherited GIT_DIR outranks it
	// and would resolve the toplevel of the caller's repository instead of the
	// one being created.
	cmd.Env = gitenv.Env()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if runErr := cmd.Run(); runErr != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			// Deliberately NOT %w-chained through *exec.ExitError. Its
			// ExitCode() method structurally satisfies the CLI's ExitCoder
			// interface, which both seams match with errors.As
			// (cmd/moai/main.go for the process exit code, internal/cli/fang.go
			// for whether the error is printed at all). A raw ExitError in the
			// chain therefore silenced this failure completely and exited with
			// the subprocess's own code: measured on a cwd outside any
			// repository as rc 128 with zero bytes of stderr, which is the
			// worst possible shape here because this hook gates every
			// worktree-isolated agent spawn. CommandError keeps the stderr text
			// printable and the exit status readable as data — the same
			// treatment execGit already applies in internal/core/git.
			return "", fmt.Errorf("worktree create: resolve repository root under %s: %w",
				dir, &gitcore.CommandError{Op: "rev-parse", Stderr: stderrStr, ExitStatus: exitErr.ExitCode()})
		}
		return "", fmt.Errorf("worktree create: resolve repository root under %s: %w (%s)",
			dir, runErr, stderrStr)
	}
	root := strings.TrimSpace(stdout.String())
	if root == "" || !filepath.IsAbs(root) {
		return "", fmt.Errorf("worktree create: repository root %q under %s is not an absolute path", root, dir)
	}
	return root, nil
}

// validateWorktreeName rejects a suggested slug that cannot name a directory
// strictly under .claude/worktrees/.
//
// Claude Code's own name validation admits '/'-separated segments of letters,
// digits, dots, underscores and dashes, so the separator itself stays legal —
// rejecting it outright would refuse an input the contract declares valid and
// leave sanitizeWorktreeBranchSuffix (which exists to support it) unreachable.
// What is rejected is every segment that cannot be a directory name: empty
// (a leading, trailing, or doubled separator), "." and ".." (which fold the
// joined path back out of the parent), and any character outside the contract's
// alphabet — '\' among them, so a Windows separator cannot slip through the
// '/'-only split.
func validateWorktreeName(name string) error {
	for _, segment := range strings.Split(name, "/") {
		switch segment {
		case "":
			return fmt.Errorf("worktree create: name %q has an empty path segment", name)
		case ".", "..":
			return fmt.Errorf("worktree create: name %q has a relative path segment %q", name, segment)
		}
		for _, r := range segment {
			if !isWorktreeNameRune(r) {
				return fmt.Errorf("worktree create: name %q contains disallowed character %q", name, r)
			}
		}
	}
	return nil
}

// isWorktreeNameRune reports whether r is admitted by Claude Code's worktree
// name alphabet (letters, digits, dots, underscores, dashes).
func isWorktreeNameRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	case r == '.', r == '_', r == '-':
		return true
	}
	return false
}

// ensureWithinWorktreeParent confirms the joined path is strictly inside the
// worktree parent directory. It is the containment backstop behind
// validateWorktreeName: the name check decides what may be joined, this one
// measures what the join actually produced, so a future change to either
// cannot silently reopen the escape.
func ensureWithinWorktreeParent(parent, path string) error {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return fmt.Errorf("worktree create: resolve %s against %s: %w", path, parent, err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("worktree create: path %s escapes the worktree directory %s", path, parent)
	}
	return nil
}

// ensureRegisteredWorktree confirms the directory already present at path is a
// worktree git knows about for this repository — not the primary checkout, not
// a plain directory, and not a link to a tree outside the parent.
//
// Reuse keyed on "a directory exists here" was the actual damage path: the
// name "../.." folded the joined path back onto the repository root, which is
// always a directory, so the hook handed the primary checkout to an agent that
// had asked for an isolated worktree.
func ensureRegisteredWorktree(repoRoot, parent, path string) error {
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("worktree create: resolve %s: %w", path, err)
	}
	if resolvedParent, err := filepath.EvalSymlinks(parent); err == nil {
		if err := ensureWithinWorktreeParent(resolvedParent, resolvedPath); err != nil {
			return err
		}
	}

	entries, err := gitcore.NewWorktreeManager(repoRoot).List()
	if err != nil {
		return fmt.Errorf("worktree create: %w", err)
	}
	for _, entry := range entries {
		registered := entry.Path
		if resolved, resolveErr := filepath.EvalSymlinks(registered); resolveErr == nil {
			registered = resolved
		}
		if registered == resolvedPath {
			return nil
		}
	}
	return fmt.Errorf("worktree create: %s is not a registered worktree of %s", path, repoRoot)
}

// sanitizeWorktreeBranchSuffix converts a worktree slug (letters, digits,
// dots, underscores, dashes, and '/'-separated segments per Claude Code's
// name validation) into a single git-branch-safe path segment. The
// worktree- prefix added by the caller keeps the result from ever starting
// with '-' or '.'.
func sanitizeWorktreeBranchSuffix(name string) string {
	return strings.ReplaceAll(name, "/", "-")
}
