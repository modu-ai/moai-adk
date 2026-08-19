// checkout.go — git-directory resolution for primary-checkout discrimination
// and board-root resolution (SPEC-KANBAN-BOARD-001 REQ-KB-005, extracting the
// resolution internal/hook/branch_guard.go carried in isPrimaryCheckout per
// the SPEC-KANBAN-WORKTREE-001 REQ-KW-018 extraction disposition).
//
// The resolution is a function of the REPOSITORY, never of the caller's
// working tree: from any worktree and from the primary checkout alike it
// yields the same absolute common git directory, whose parent is the primary
// checkout's root.
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/execerr"
)

// ExecCommand is the package-level indirection over exec.Command. Tests inject
// a mock runner here to force the older-git fallback branch through the
// dispatcher inside ResolveGitDirs (an older-git host rejecting
// --path-format=absolute). Direct invocation of the fallback is INSUFFICIENT —
// a vacuous pass that bypasses the dispatcher. Exported so out-of-package
// consumers (the hook's fallback suite, the kanban board's fallback-forced
// criterion) can force the same branch.
var ExecCommand = exec.Command

// GitDirs holds the absolute paths of a repository's git directory and git
// common directory, as resolved by ResolveGitDirs.
type GitDirs struct {
	// GitDir is the absolute git directory of the queried checkout: the
	// repository's .git in a primary checkout, .git/worktrees/<name> in a
	// linked worktree.
	GitDir string

	// CommonDir is the absolute git COMMON directory, shared by every checkout
	// of one repository. Its parent is the primary checkout's root.
	CommonDir string
}

// ResolveGitDirs resolves the absolute git-dir and git-common-dir of the
// repository containing dir. The caller's location inside the repository does
// not affect the result.
//
// Primary path (git 2.31+, March 2021): a SINGLE rev-parse call emits both
// absolute paths — --path-format=absolute applies to every following path
// flag, so --git-dir and --git-common-dir each print one absolute line, in
// argument order. Fallback (older git / Apple Git rejecting the flag):
// --absolute-git-dir plus a --git-common-dir result normalized against dir
// when it is not already absolute. The bare --git-common-dir form is never
// used alone: it returns a repository-relative path (.git) in the primary
// checkout, so "the parent of it" would not be a path at all.
func ResolveGitDirs(dir string) (*GitDirs, error) {
	if err := validateDirArg(dir); err != nil {
		return nil, err
	}

	out, err := runGitRevParse(dir, "--path-format=absolute", "--git-dir", "--git-common-dir")
	if err == nil {
		lines := strings.Split(out, "\n")
		if len(lines) >= 2 {
			return &GitDirs{
				GitDir:    strings.TrimSpace(lines[0]),
				CommonDir: strings.TrimSpace(lines[1]),
			}, nil
		}
		// Unexpected output shape; fall through to the fallback below.
	}

	absGitDir, err := runGitRevParse(dir, "--absolute-git-dir")
	if err != nil {
		return nil, fmt.Errorf("resolve git dirs: --absolute-git-dir failed: %w", err)
	}
	relCommon, err := runGitRevParse(dir, "--git-common-dir")
	if err != nil {
		return nil, fmt.Errorf("resolve git dirs: --git-common-dir failed: %w", err)
	}
	absCommon := relCommon
	if !filepath.IsAbs(absCommon) {
		absCommon = filepath.Join(dir, relCommon)
	}
	return &GitDirs{
		GitDir:    filepath.Clean(absGitDir),
		CommonDir: filepath.Clean(absCommon),
	}, nil
}

// IsPrimaryCheckout reports whether dir is the primary git checkout (absolute
// git-dir == absolute git-common-dir). Returns an error on any uncertainty
// (non-git dir, missing git binary, unresolvable rev-parse) so callers can
// fail in their own safe direction.
func IsPrimaryCheckout(dir string) (bool, error) {
	dirs, err := ResolveGitDirs(dir)
	if err != nil {
		return false, err
	}
	return dirs.GitDir == dirs.CommonDir, nil
}

// validateDirArg rejects an empty query directory before any process spawn.
func validateDirArg(dir string) error {
	if dir == "" {
		return fmt.Errorf("resolve git dirs: empty directory")
	}
	return nil
}

// runGitRevParse runs `git -C dir rev-parse <args...>` via the package-level
// ExecCommand indirection and returns the trimmed stdout. Returns an error on
// non-zero exit (covers missing git binary, non-git directory, and
// unknown-flag rejection by older git).
func runGitRevParse(dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir, "rev-parse"}, args...)
	cmd := ExecCommand("git", full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// execerr.StatusDetail, not %w: a raw *exec.ExitError chain would be
		// mistaken for an intentional ExitCoder at the cmd/moai seam (t130).
		return "", fmt.Errorf("git rev-parse %s: %s (%s)",
			strings.Join(args, " "), execerr.StatusDetail(err), strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}
