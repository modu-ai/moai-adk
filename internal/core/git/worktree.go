package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Compile-time interface compliance check.
var _ WorktreeManager = (*worktreeManager)(nil)

// worktreeManager implements the WorktreeManager interface using system git.
type worktreeManager struct {
	root   string
	logger *slog.Logger
}

// NewWorktreeManager creates a new WorktreeManager for the repository at root.
func NewWorktreeManager(root string) *worktreeManager {
	return &worktreeManager{
		root:   root,
		logger: slog.Default().With("module", "git.worktree"),
	}
}

// Add creates a new worktree at the given path for the given branch.
// If the branch does not exist, it is created automatically with -b.
func (w *worktreeManager) Add(path, branch string) error {
	w.logger.Debug("system git fallback", "operation", "worktree add", "reason", "go-git lacks worktree support")
	w.logger.Debug("adding worktree", "path", path, "branch", branch)

	// Check if path already exists.
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("add worktree at %q: %w", path, ErrWorktreePathExists)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the branch already exists.
	exists := branchExists(ctx, w.root, branch)
	args := buildWorktreeAddArgs(exists, path, branch)
	if _, err := execGit(ctx, w.root, args...); err != nil {
		if exists {
			return fmt.Errorf("add worktree for existing branch %q: %w", branch, err)
		}
		return fmt.Errorf("add worktree with new branch %q: %w", branch, err)
	}

	w.logger.Debug("worktree added", "path", path, "branch", branch)
	return nil
}

// buildWorktreeAddArgs는 `git worktree add` argv를 조립한다. SPEC-SEC-HARDEN-002
// M2b: 사용자 유래 operand(worktree path, branch) 앞에 `--` end-of-options
// 구분자를 삽입해, `-`로 시작하는 operand가 git 옵션으로 해석되는 argv option
// smuggling(CWE-88)을 차단한다. `-b <branch>`는 git 옵션 쌍이므로 `--` 앞에 둔다.
//
//   - 기존 브랜치: worktree add -- <path> <branch>
//   - 신규 브랜치: worktree add -b <branch> -- <path>
//
// @MX:NOTE: [AUTO] SPEC-SEC-HARDEN-002 M2b — argv `--` 구분자로 worktree-add operand option smuggling 차단. 정상 path/branch 동작 보존.
func buildWorktreeAddArgs(branchExists bool, path, branch string) []string {
	if branchExists {
		return []string{"worktree", "add", "--", path, branch}
	}
	// 신규 브랜치: -b <branch>는 옵션 쌍 → `--` 이전에 유지하고 path만 positional로.
	return []string{"worktree", "add", "-b", branch, "--", path}
}

// List returns all active worktrees including the main worktree.
func (w *worktreeManager) List() ([]Worktree, error) {
	w.logger.Debug("system git fallback", "operation", "worktree list", "reason", "go-git lacks worktree support")
	w.logger.Debug("listing worktrees")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := execGit(ctx, w.root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}

	worktrees := parsePorcelainWorktreeList(out)
	w.logger.Debug("worktrees listed", "count", len(worktrees))
	return worktrees, nil
}

// Remove deletes a worktree at the given path.
// If force is true, the worktree is removed even with uncommitted changes.
func (w *worktreeManager) Remove(path string, force bool) error {
	w.logger.Debug("system git fallback", "operation", "worktree remove", "reason", "go-git lacks worktree support")
	w.logger.Debug("removing worktree", "path", path, "force", force)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	_, err := execGit(ctx, w.root, args...)
	if err != nil {
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "is not a working tree"):
			return fmt.Errorf("remove worktree at %q: %w", path, ErrWorktreeNotFound)
		case strings.Contains(errStr, "contains modified or untracked files"):
			return fmt.Errorf("remove worktree at %q: %w", path, ErrWorktreeDirty)
		default:
			return fmt.Errorf("remove worktree at %q: %w", path, err)
		}
	}

	w.logger.Debug("worktree removed", "path", path)
	return nil
}

// Prune removes stale worktree references for deleted directories.
func (w *worktreeManager) Prune() error {
	w.logger.Debug("system git fallback", "operation", "worktree prune", "reason", "go-git lacks worktree support")
	w.logger.Debug("pruning worktrees")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := execGit(ctx, w.root, "worktree", "prune")
	if err != nil {
		return fmt.Errorf("prune worktrees: %w", err)
	}

	w.logger.Debug("worktrees pruned")
	return nil
}

// Repair repairs worktree administrative files if they have become corrupted.
func (w *worktreeManager) Repair() error {
	w.logger.Debug("system git fallback", "operation", "worktree repair", "reason", "go-git lacks worktree support")
	w.logger.Debug("repairing worktrees")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := execGit(ctx, w.root, "worktree", "repair")
	if err != nil {
		return fmt.Errorf("repair worktrees: %w", err)
	}

	w.logger.Debug("worktrees repaired")
	return nil
}

// Root returns the repository root path.
func (w *worktreeManager) Root() string {
	return w.root
}

// Sync fetches the latest changes from origin and merges or rebases the
// base branch into the worktree at wtPath.
func (w *worktreeManager) Sync(wtPath, baseBranch, strategy string) error {
	w.logger.Debug("syncing worktree", "path", wtPath, "base", baseBranch, "strategy", strategy)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch latest from origin.
	if _, err := execGit(ctx, wtPath, "fetch", "origin"); err != nil {
		return fmt.Errorf("fetch origin in %q: %w", wtPath, err)
	}

	// Apply sync strategy.
	ref := "origin/" + baseBranch
	switch strategy {
	case "rebase":
		if _, err := execGit(ctx, wtPath, "rebase", ref); err != nil {
			return fmt.Errorf("rebase %s in %q: %w", ref, wtPath, err)
		}
	default: // merge
		if _, err := execGit(ctx, wtPath, "merge", ref, "--no-edit"); err != nil {
			return fmt.Errorf("merge %s in %q: %w", ref, wtPath, err)
		}
	}

	w.logger.Debug("worktree synced", "path", wtPath, "base", baseBranch)
	return nil
}

// DeleteBranch deletes a local branch by name.
func (w *worktreeManager) DeleteBranch(name string) error {
	w.logger.Debug("deleting branch", "name", name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := execGit(ctx, w.root, "branch", "-d", name); err != nil {
		return fmt.Errorf("delete branch %q: %w", name, err)
	}

	w.logger.Debug("branch deleted", "name", name)
	return nil
}

// IsBranchMerged reports whether a branch's changes relative to its merge-base
// with base are present in base's HEAD tree, irrespective of the merge strategy
// that placed them there (SPEC-WORKTREE-SQUASH-MERGE-001 REQ-WSM-001).
//
// The predicate is an ordered OR over five signals, cheapest first:
//
//	S1  reachability     git branch --merged <base> lists <branch>
//	S2  empty diff       git diff --quiet <merge-base> <branch> reports none
//	S3  rebase-merge     git cherry <base> <branch>, non-empty & all '-'
//	S4  squash-merge     synthetic-commit git cherry, non-empty & all '-'
//	S5  state check      every branch-touched path is identical in base HEAD
//
// S1 and S2 are standalone; S3 and S4 are each CONJOINED with S5, because a
// history signal answers "did an equivalent patch-id ever exist in base" while
// S5 answers "does the content survive in base HEAD" — a later revert leaves
// the first true and the second false (SC-8). S5 is a conjunct only: it may
// withhold a verdict, never grant one (REQ-WSM-014).
//
// This is a safety-critical predicate: it gates `moai worktree clean --stale`,
// and a false positive destroys unmerged user work with no undo. Eight guards
// carry the no-false-positive argument; see spec.md §2 and plan.md §D. Each
// guard's removal flips a measured-disjoint scenario row (acceptance.md §C.1).
//
// @MX:ANCHOR: [AUTO] reachability (S1) is retained alongside the patch-id probes (S3/S4) and conjoined with the state check (S5)
// @MX:REASON: S1 is the only signal that recognises a true merge commit and a strictly-behind branch; replacing it with the patch-id probes regresses SC-3 and SC-4. S5 is the only signal that withholds a verdict when a history signal fires on content base later reverted/removed; dropping it regresses SC-8 through SC-15 and P3c/P4c (acceptance.md §C.1 `no-state`). The two are independent invariants a future edit is most likely to break, so they are pinned here.
func (w *worktreeManager) IsBranchMerged(branch, base string) (bool, error) {
	w.logger.Debug("checking if branch merged", "branch", branch, "base", base)

	// S1: reachability (retained verbatim, including the marker-stripping
	// helper). This is the only signal that recognises a true merge commit
	// (SC-3) and a strictly-behind branch (SC-4); replacing it with the
	// patch-id probes would regress both.
	s1ctx, s1cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer s1cancel()
	out, err := execGit(s1ctx, w.root, "branch", "--merged", base)
	if err != nil {
		return false, fmt.Errorf("check merged branches: %w", err)
	}
	for line := range strings.SplitSeq(out, "\n") {
		name := strings.TrimSpace(trimBranchListMarker(line))
		if name == branch {
			return true, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// merge-base: computed once, shared by S2..S5. Per REQ-WSM-007, rc 1 with
	// empty stdout means the two refs share no common ancestor (unrelated
	// histories): the conservative outcome is (false, nil) — keep the worktree.
	mbResult, err := execGitExit(ctx, w.root, nil, "merge-base", base, branch)
	if err != nil {
		return false, fmt.Errorf("merge-base %s %s: %w", base, branch, err)
	}
	switch mbResult.exitCode {
	case 0:
		// found; continue
	case 1:
		if strings.TrimSpace(mbResult.stdout) == "" {
			return false, nil
		}
		return false, fmt.Errorf("merge-base %s %s: unexpected rc=1 with output", base, branch)
	default:
		return false, fmt.Errorf("merge-base %s %s: %s: exit %d",
			base, branch, strings.TrimSpace(mbResult.stderr), mbResult.exitCode)
	}
	mb := strings.TrimSpace(mbResult.stdout)

	// S2: empty diff between merge-base and branch. rc 0 = no difference =
	// merged; rc 1 = a difference exists = continue. Both rc 0 and rc 1 are
	// verdicts, NOT failures (REQ-WSM-007) — treating rc 1 as an error would
	// make every non-merged branch surface as a failure.
	s2, err := execGitExit(ctx, w.root, nil, "diff", "--quiet", mb, branch)
	if err != nil {
		return false, fmt.Errorf("diff --quiet %s %s: %w", mb, branch, err)
	}
	switch s2.exitCode {
	case 0:
		return true, nil
	case 1:
		// difference exists; fall through to history signals
	default:
		return false, fmt.Errorf("diff --quiet %s %s: %s: exit %d",
			mb, branch, strings.TrimSpace(s2.stderr), s2.exitCode)
	}

	// names: every path the branch changed since its merge-base. Computed once
	// and reused by S5 for both S3 and S4. Three enumeration guards apply:
	//   - --no-renames          keeps a renamed source path in the list (SC-10/11)
	//   - -z + NUL split        keeps the bytes verbatim, not C-quoted (SC-12)
	//   - --ignore-submodules=none keeps a moved gitlink in the list (SC-15)
	// --ignore-submodules=none appears here AND on the comparison below; neither
	// occurrence is redundant (spec.md §2 finding 8; AC-WSM-006 fifth
	// falsification — removing it from either stage alone reproduces SC-15).
	namesResult, err := execGitExit(ctx, w.root, nil,
		"diff", "--ignore-submodules=none", "--no-renames", "--name-only", "-z", mb, branch)
	if err != nil {
		return false, fmt.Errorf("enumerate changed paths: %w", err)
	}
	if namesResult.exitCode != 0 {
		return false, fmt.Errorf("enumerate changed paths: %s: exit %d",
			strings.TrimSpace(namesResult.stderr), namesResult.exitCode)
	}
	names := splitNul(namesResult.stdout)

	// S5: state check (conjunct only). Computed once, reused for S3 and S4.
	// Never a merged verdict on its own — it may withhold, never grant.
	stateHolds, err := stateCheck(ctx, w.root, names, branch, base)
	if err != nil {
		return false, err
	}

	// S3: rebase-merge history signal. Merged only when cherry output is
	// non-empty AND every line is prefixed '-' AND S5 holds. Empty cherry
	// output means "no commits compared" and is never a merged verdict; a
	// single '+' line means part of the branch's work never landed.
	s3, err := execGitExit(ctx, w.root, nil, "cherry", base, branch)
	if err != nil {
		return false, fmt.Errorf("cherry %s %s: %w", base, branch, err)
	}
	if s3.exitCode != 0 {
		return false, fmt.Errorf("cherry %s %s: %s: exit %d",
			base, branch, strings.TrimSpace(s3.stderr), s3.exitCode)
	}
	if stateHolds && cherryAllDash(s3.stdout) {
		return true, nil
	}

	// S4: squash-merge history signal. Collapse the branch's entire
	// diff-versus-merge-base into one synthetic commit so it has the same shape
	// as the single squashed commit on base, then ask git cherry whether base
	// already contains an equivalent patch.
	treeOut, err := execGit(ctx, w.root, "rev-parse", branch+"^{tree}")
	if err != nil {
		return false, fmt.Errorf("resolve branch tree for synthetic commit: %w", err)
	}
	tree := strings.TrimSpace(treeOut)
	synth, err := execGitExit(ctx, w.root, syntheticCommitEnv(),
		"commit-tree", tree, "-p", mb, "-m", syntheticCommitMessage)
	if err != nil {
		return false, fmt.Errorf("synthetic commit-tree: %w", err)
	}
	if synth.exitCode != 0 {
		return false, fmt.Errorf("synthetic commit-tree: %s: exit %d",
			strings.TrimSpace(synth.stderr), synth.exitCode)
	}
	synthCommit := strings.TrimSpace(synth.stdout)
	s4, err := execGitExit(ctx, w.root, nil, "cherry", base, synthCommit)
	if err != nil {
		return false, fmt.Errorf("cherry %s %s: %w", base, synthCommit, err)
	}
	if s4.exitCode != 0 {
		return false, fmt.Errorf("cherry %s %s: %s: exit %d",
			base, synthCommit, strings.TrimSpace(s4.stderr), s4.exitCode)
	}
	if stateHolds && cherryAllDash(s4.stdout) {
		return true, nil
	}

	return false, nil
}

// syntheticCommitMessage is the fixed commit message used by the S4
// synthetic-commit probe. Combined with the pinned identity values
// (syntheticCommitEnv, REQ-WSM-008), it makes the object deterministic across
// repeated evaluations of an unchanged branch, bounding dangling objects to
// distinct branch states rather than to the number of sweeps.
const syntheticCommitMessage = "moai-squash-probe"

// syntheticCommitIdentity pins the author and committer identity so a synthetic
// commit's object hash depends only on (tree, parent, message), not on
// wall-clock time or the invoking user. The date is fixed to the epoch; the
// name/email are fixed arbitrary values. Without this, repeated evaluation of
// one unchanged branch yields a fresh object each call (measured), and the
// dangling-commit count grows with the sweep count rather than the distinct
// branch-state count (REQ-WSM-008; spec.md §5 decision 2).
const (
	syntheticAuthorName  = "moai"
	syntheticAuthorEmail = "moai@localhost"
	syntheticCommitDate  = "@0 +0000"
)

// syntheticCommitEnv returns the per-call environment overlay that pins the
// synthetic commit's identity. It is passed to execGitExit's extraEnv, the
// mechanism M1 introduced; this function supplies the M2 values. Existing
// execGit callers are unaffected — they never pass through extraEnv.
func syntheticCommitEnv() []string {
	return []string{
		"GIT_AUTHOR_NAME=" + syntheticAuthorName,
		"GIT_AUTHOR_EMAIL=" + syntheticAuthorEmail,
		"GIT_AUTHOR_DATE=" + syntheticCommitDate,
		"GIT_COMMITTER_NAME=" + syntheticAuthorName,
		"GIT_COMMITTER_EMAIL=" + syntheticAuthorEmail,
		"GIT_COMMITTER_DATE=" + syntheticCommitDate,
	}
}

// splitNul splits NUL-delimited `git diff -z` output into a []string, dropping
// the empty trailing element a NUL terminator leaves behind. The result is one
// element per path, suitable for variadic expansion. NEVER split on newline:
// `git diff --name-only` C-quotes any path containing a backslash, tab, double
// quote, or control character — and, under git's documented default
// core.quotePath=true, any non-ASCII byte — and the quoted rendering matches
// nothing as a pathspec, which by the fail-open property reads as merged
// (spec.md §2 findings 5-6; SC-12). core.quotePath=false is NOT a substitute:
// it leaves backslash and control-character paths quoted regardless.
func splitNul(out string) []string {
	if out == "" {
		return nil
	}
	parts := strings.Split(out, "\x00")
	if n := len(parts); n > 0 && parts[n-1] == "" {
		parts = parts[:n-1]
	}
	return parts
}

// cherryAllDash reports whether git cherry output is non-empty AND every
// non-blank line is prefixed '-'. Empty output is NEVER a merged verdict (it
// means "no commits compared"); a single '+' line means part of the branch's
// work is absent from base. This implements two of the three S3/S4 guards; the
// third (S5 holds) is applied by the caller as a conjunct.
func cherryAllDash(cherryOut string) bool {
	trimmed := strings.TrimSpace(cherryOut)
	if trimmed == "" {
		return false
	}
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "-") {
			return false
		}
	}
	return true
}

// stateCheck evaluates S5: whether every path the branch changed since its
// merge-base has identical content (and mode) in base HEAD. It is a CONJUNCT
// ONLY — it may withhold a verdict, never grant one (REQ-WSM-014).
//
// When names is empty the branch touched nothing since its merge-base: S5
// holds trivially, and scoping the later comparison to an empty pathspec would
// instead compare the whole trees (the wrong question).
//
// Five guards govern S5; each has a silent fail-open failure mode when omitted
// (spec.md §2 findings 5-8; plan.md §D):
//   - literal matching: --literal-pathspecs is a GIT-LEVEL option and MUST
//     precede the subcommand. It matches each path's bytes as a literal
//     filename rather than parsing them as pathspec magic, so a repo-root path
//     whose first byte is ':' is checked instead of naming no file (SC-13).
//   - observed comparison: --no-textconv and --ignore-submodules=none are
//     DIFF-LEVEL options and MUST follow the subcommand. They stop a textconv
//     diff driver (SC-14) and a submodule ignore directive (SC-15) from
//     collapsing a real difference.
//   - mode sensitivity: `git diff` compares modes, so a symlink/file blob-OID
//     collision (P4c) and a chmod-only divergence (P3c) are detected. A
//     blob-OID substitute would not detect either.
//
// `names` MUST be passed as a []string (variadic to execGitExit), NEVER as a
// joined shell string: a single pathspec matching nothing exits 0 and reads as
// "no difference" (spec.md §2 finding 5; plan.md §D).
func stateCheck(ctx context.Context, root string, names []string, branch, base string) (bool, error) {
	if len(names) == 0 {
		return true, nil
	}
	args := make([]string, 0, 6+len(names))
	args = append(args, "--literal-pathspecs") // git-level: precedes subcommand
	args = append(args, "diff", "--quiet")
	args = append(args, "--no-textconv", "--ignore-submodules=none") // diff-level: follows subcommand
	args = append(args, branch, base, "--")
	args = append(args, names...) // one argument per path; never joined
	r, err := execGitExit(ctx, root, nil, args...)
	if err != nil {
		return false, fmt.Errorf("state check: %w", err)
	}
	switch r.exitCode {
	case 0:
		return true, nil
	case 1:
		return false, nil
	default:
		return false, fmt.Errorf("state check: %s: exit %d", strings.TrimSpace(r.stderr), r.exitCode)
	}
}

// trimBranchListMarker strips the leading status marker `git branch` puts on a
// listing line: "* " for the branch checked out here, and "+ " for a branch
// checked out in another worktree.
//
// The "+ " case is load-bearing for every worktree-scoped caller: a worktree's
// branch is by definition checked out somewhere, so it always carries "+ " in
// `git branch --merged` output. Stripping only "* " made IsBranchMerged report
// every worktree branch as unmerged, which silently turned worktree cleanup
// paths into no-ops.
//
// @MX:ANCHOR: [AUTO] both "* " and "+ " markers must be stripped
// @MX:REASON: worktree branches always carry "+ "; dropping that case makes
// IsBranchMerged return false for every worktree branch, disabling the cleanup
// commands that gate on it without any error surfacing.
func trimBranchListMarker(line string) string {
	for _, marker := range []string{"* ", "+ "} {
		if after, found := strings.CutPrefix(line, marker); found {
			return after
		}
	}
	return line
}

// parsePorcelainWorktreeList parses the output of git worktree list --porcelain.
//
// The porcelain format consists of stanzas separated by blank lines:
//
//	worktree /path/to/worktree
//	HEAD abc123def456
//	branch refs/heads/main
//
//	worktree /tmp/wt-feature
//	HEAD def789abc012
//	branch refs/heads/feature
func parsePorcelainWorktreeList(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		switch {
		case strings.HasPrefix(line, "worktree "):
			// Save the previous entry if present.
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			// Normalize the path from git output to use OS-native separators
			current = Worktree{Path: filepath.Clean(strings.TrimPrefix(line, "worktree "))}
		case strings.HasPrefix(line, "HEAD "):
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "detached":
			current.Branch = ""
		}
	}

	// Append the last entry.
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}
