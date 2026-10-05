package cli

// worktree_branch_flag.go — card t295: the launcher's existing-branch worktree
// creation path.
//
// Before this file the launcher could only create worktrees with NEW branches:
// `-w <name>` passes `--worktree <name>` to the backend, which cuts a fresh
// branch, and an absolute `-w <path>` only re-enters a tree that already
// exists. A worktree checked out at an EXISTING branch — the develop
// integration worktree the gitflow chain provisions every batch — had no
// sanctioned creation path, so the operator's only option was a raw
// `git worktree add`, which is a [HARD] launcher-route violation and leaves
// the tree invisible to moai's worktree tooling.
//
// The surface: `moai cc -w <name> --branch <existing>` (same flag accepted by
// cg and glm, which share this launcher seam). moai creates the tree itself
// via the hardened core primitive (WorktreeManager.Add), registers it in the
// same state file the WorktreeCreate hook uses (.moai/state/worktrees.json, so
// `moai worktree clean`'s anchor check sees it), then strips the --branch
// tokens so the backend only ever sees `--worktree <name>` and re-enters the
// tree that now exists.
//
// Flag semantics are deliberately narrow: the branch must EXIST (a missing
// branch is an error, never a silent `git worktree add -b` — WorktreeManager.
// Add would happily create one on a typo, which is the opposite of what this
// flag promises), and `-w` must be a short name (an absolute path already
// names an existing tree — there is nothing to create).

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// launcherWorktreeAgentName marks registry entries created by the launcher's
// --branch path, distinguishable from the hook's per-agent entries.
const launcherWorktreeAgentName = "launcher"

// launcherWorktreeMaterialize creates the worktree for the --branch flag.
// Seam: tests override to record the call without touching git or disk.
var launcherWorktreeMaterialize = launcherWorktreeMaterializeReal

// resolveWorktreeExistingBranch scans args for a --branch flag accompanying a
// short-name -w, materializes the worktree at the existing branch, and returns
// the args with the --branch tokens removed. When no --branch flag is present
// the args are returned unchanged and nothing runs — the flag is purely
// opt-in. Notices (idempotent re-entry) go to warn; errors abort the launch.
func resolveWorktreeExistingBranch(args []string, warn io.Writer) ([]string, error) {
	rest, branch, has, err := splitWorktreeBranchFlag(args)
	if err != nil {
		return nil, err
	}
	if !has {
		return args, nil
	}

	name, err := requireShortWorktreeName(rest)
	if err != nil {
		return nil, err
	}
	projectRoot, err := findProjectRootFn()
	if err != nil {
		return nil, fmt.Errorf("--branch: resolve project root: %w", err)
	}
	if err := launcherWorktreeMaterialize(projectRoot, name, branch, warn); err != nil {
		return nil, err
	}
	return rest, nil
}

// splitWorktreeBranchFlag finds and removes the --branch flag from args,
// scanning only tokens before the "--" pass-through marker (tokens after it
// are verbatim backend input and are never touched). Accepts both the
// two-token `--branch <value>` and the joined `--branch=<value>` forms. A
// missing or option-shaped value is a parse error at this boundary — an
// unconsumed option-shaped token must never leak into the backend argv.
func splitWorktreeBranchFlag(args []string) (rest []string, branch string, has bool, err error) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			rest = append(rest, args[i:]...)
			return rest, branch, has, nil
		}
		switch {
		case arg == "--branch":
			has = true
			if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "-") {
				return nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
			branch = args[i+1]
			i++
		case strings.HasPrefix(arg, "--branch="):
			has = true
			branch = strings.TrimPrefix(arg, "--branch=")
			if branch == "" {
				return nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
		default:
			rest = append(rest, arg)
		}
	}
	return rest, branch, has, nil
}

// requireShortWorktreeName returns the -w flag's value and enforces the two
// preconditions the --branch path depends on: a value is present, and it is a
// short name rather than an absolute path (an absolute -w names an existing
// tree; creating is not meaningful there). The name must also be a single
// safe path segment — it becomes a directory under .claude/worktrees/.
func requireShortWorktreeName(args []string) (string, error) {
	name, ok := worktreeFlagValue(args)
	if !ok || name == "" {
		return "", fmt.Errorf("--branch requires -w <name>: name the worktree to create")
	}
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("--branch does not combine with an absolute -w path: %q already names an existing worktree, re-enter it without --branch", name)
	}
	if name == "." || name == ".." || strings.ContainsRune(name, filepath.Separator) || strings.ContainsRune(name, '/') {
		return "", fmt.Errorf("--branch: worktree name %q must be a single path segment", name)
	}
	return name, nil
}

// launcherWorktreeMaterializeReal creates .claude/worktrees/<name> checked out
// at the existing branch and registers the tree in the shared worktree state
// file. Idempotent: an existing directory at the target path is reported and
// reused (the backend re-enters it), mirroring the WorktreeCreate hook's
// reuse contract.
func launcherWorktreeMaterializeReal(projectRoot, name, branch string, warn io.Writer) error {
	if branch == "" || strings.HasPrefix(branch, "-") {
		return fmt.Errorf("--branch: invalid branch name %q", branch)
	}

	// The branch must already exist. Check BEFORE WorktreeManager.Add, which
	// auto-creates missing branches with -b — a silent typo branch is exactly
	// what this flag must never do.
	exists, err := branchRefExists(projectRoot, branch)
	if err != nil {
		return fmt.Errorf("--branch: check branch %q: %w", branch, err)
	}
	if !exists {
		return fmt.Errorf("--branch: branch %q does not exist; this flag checks out an existing branch and never creates one", branch)
	}

	treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)
	if _, statErr := os.Stat(treePath); statErr == nil {
		// Idempotent re-entry: report and reuse. A failed notice write must
		// not abort an otherwise valid launch, so it degrades to the log.
		if _, writeErr := fmt.Fprintf(warn, "moai: worktree %s already exists; re-entering it\n", treePath); writeErr != nil {
			slog.Debug("worktree --branch: notice write failed", "error", writeErr)
		}
	} else if err := gitcore.NewWorktreeManager(projectRoot).Add(treePath, branch); err != nil {
		return fmt.Errorf("--branch: create worktree %s at branch %q: %w", treePath, branch, err)
	}

	// Register in the same state file the WorktreeCreate hook maintains so
	// `moai worktree clean`'s anchor check sees launcher-created trees.
	hook.RegisterLauncherWorktree(projectRoot, treePath, branch, launcherWorktreeAgentName)
	return nil
}

// branchRefExists reports whether refs/heads/<branch> resolves, via
// `git show-ref --verify --quiet`. Exit code 1 means the ref is absent (a
// normal answer); any other failure is reported as an error with its git
// stderr, not chained through *exec.ExitError — the same execerr treatment
// session_worktree.go applies at the cmd/moai seam (t130).
func branchRefExists(projectRoot, branch string) (bool, error) {
	cmd := exec.Command("git", "-C", projectRoot, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("git show-ref: %s", strings.TrimSpace(stderr.String()))
	}
	return false, fmt.Errorf("git show-ref: %s", execerr.StatusDetail(err))
}
