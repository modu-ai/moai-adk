package auditreceipt

// storeroot.go — the one store-root answer for a config-orphaned linked
// worktree (SPEC-WORKTREE-STATE-ROOT-001).
//
// A repository that keeps .moai untracked gives every linked worktree a tree
// with no .moai of its own. Writers and readers of .moai/state run in three
// places — the MCP server, hook processes, and the `moai verify` CLI — and each
// starts from whatever root it was handed. If each kept state under that root,
// a hook that resolved the primary checkout would never find what a tool wrote
// under the worktree. StoreRoot maps every root to ONE store root instead:
// the primary checkout for a config-orphaned worktree, the root itself for
// everything else. The state therefore lives in one place whichever of the two
// a reader starts from.
//
// The predicate and the primary identification moved here from internal/cli
// (SPEC-MCP-WORKTREE-UNTRACKED-001) so the hook package can reach the same
// answer without importing internal/cli. Their behaviour is unchanged: every
// git inspection runs scrubbed (no inherited GIT_DIR / GIT_WORK_TREE /
// GIT_COMMON_DIR / GIT_INDEX_FILE / GIT_CEILING_DIRECTORIES, LC_ALL=C), is
// decided by exit status and output shape, and is never retried.

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GateAssumedRequiredNote is the wording for a config-orphaned root whose
// primary checkout could not be identified, where the codex audit gate is
// treated as `required`. It is distinguishable from a primary that declares
// `required` itself.
const GateAssumedRequiredNote = "workflow.audit.gates.codex assumed `required` because the primary checkout of this worktree could not be identified"

// scrubbedGitVars are removed from every git inspection's environment, so a
// variable inherited from the calling process cannot redirect git at another
// repository.
var scrubbedGitVars = []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE", "GIT_CEILING_DIRECTORIES"}

// scrubbedGitEnv returns the process environment minus scrubbedGitVars and any
// locale override, with LC_ALL=C set.
func scrubbedGitEnv() []string {
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		drop := name == "LC_ALL"
		for _, v := range scrubbedGitVars {
			if name == v {
				drop = true
			}
		}
		if !drop {
			env = append(env, kv)
		}
	}
	return append(env, "LC_ALL=C")
}

// RunScrubbedGit runs `git -C dir args...` in the scrubbed environment and
// returns stdout. Any failure — git missing, non-zero exit — is an error; the
// caller never inspects git's message text.
func RunScrubbedGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = scrubbedGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// SingleGitPath parses output that must be exactly one non-empty absolute path
// line, and returns it canonicalized. Any other shape is an error.
func SingleGitPath(out string) (string, error) {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 || strings.TrimSpace(lines[0]) == "" || !filepath.IsAbs(lines[0]) {
		return "", errors.New("unexpected git output shape")
	}
	return filepath.EvalSymlinks(lines[0])
}

// WorktreeEntry is one `git worktree list --porcelain` record.
type WorktreeEntry struct {
	Path     string // canonical; "" when the listed path cannot be canonicalized
	Prunable bool
}

// parseWorktreePorcelain parses `git worktree list --porcelain`. Every record
// must open with a `worktree <path>` line; any other shape is an error. A
// listed path that cannot be canonicalized (a deleted directory) is kept with
// an empty canonical path so it simply never matches.
func parseWorktreePorcelain(out string) ([]WorktreeEntry, error) {
	var entries []WorktreeEntry
	inRecord := false
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			inRecord = false
			continue
		}
		if !inRecord {
			path, ok := strings.CutPrefix(line, "worktree ")
			if !ok || !filepath.IsAbs(path) {
				return nil, errors.New("unexpected git worktree list shape")
			}
			canon, err := filepath.EvalSymlinks(path)
			if err != nil {
				canon = ""
			}
			entries = append(entries, WorktreeEntry{Path: canon})
			inRecord = true
			continue
		}
		if line == "prunable" || strings.HasPrefix(line, "prunable ") {
			entries[len(entries)-1].Prunable = true
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("git worktree list returned no entries")
	}
	return entries, nil
}

// ErrAmbiguousLayout marks a repository whose primary checkout cannot be
// identified unambiguously (--separate-git-dir, a submodule-internal git dir,
// a bare repository).
var ErrAmbiguousLayout = errors.New("ambiguous repository layout (separate git dir, submodule, or bare repository)")

// IdentifyPrimaryCheckout returns the primary checkout of the repository that
// contains dir, plus the porcelain worktree listing: the parent of the git
// common dir, accepted only when the common dir is named `.git`, the parent's
// own `.git` resolves to that same common dir, and the parent is the first
// listed worktree.
func IdentifyPrimaryCheckout(dir string) (string, []WorktreeEntry, error) {
	out, err := RunScrubbedGit(dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", nil, fmt.Errorf("git could not inspect the repository: %w", err)
	}
	common, err := SingleGitPath(out)
	if err != nil {
		return "", nil, fmt.Errorf("git could not inspect the repository: %w", err)
	}
	if filepath.Base(common) != ".git" {
		return "", nil, ErrAmbiguousLayout
	}
	primary := filepath.Dir(common)
	primaryGit, err := filepath.EvalSymlinks(filepath.Join(primary, ".git"))
	if err != nil || primaryGit != common {
		return "", nil, ErrAmbiguousLayout
	}
	listOut, err := RunScrubbedGit(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return "", nil, fmt.Errorf("git could not list worktrees: %w", err)
	}
	entries, err := parseWorktreePorcelain(listOut)
	if err != nil {
		return "", nil, err
	}
	if entries[0].Path != primary {
		return "", nil, ErrAmbiguousLayout
	}
	return primary, entries, nil
}

// IsConfigOrphanedRoot reports whether root has no workflow config of its own
// AND carries positive, git-free evidence of being a linked worktree top level:
// `<root>/.git` is a regular file whose `gitdir:` path — resolved against root
// when relative, never against the process working directory — names an
// existing directory whose parent is named `worktrees` and which holds a
// `commondir` file. Two file reads, no subprocess.
func IsConfigOrphanedRoot(root string) bool {
	if strings.TrimSpace(root) == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml")); err == nil {
		return false
	}
	dotGit := filepath.Join(root, ".git")
	info, err := os.Lstat(dotGit)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	data, err := os.ReadFile(dotGit)
	if err != nil {
		return false
	}
	gitdir := ""
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "gitdir:"); ok {
			gitdir = strings.TrimSpace(v)
			break
		}
	}
	if gitdir == "" {
		return false
	}
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(root, gitdir)
	}
	adminInfo, err := os.Stat(gitdir)
	if err != nil || !adminInfo.IsDir() {
		return false
	}
	if filepath.Base(filepath.Dir(filepath.Clean(gitdir))) != "worktrees" {
		return false
	}
	cdInfo, err := os.Stat(filepath.Join(gitdir, "commondir"))
	return err == nil && cdInfo.Mode().IsRegular()
}

// UnresolvedStoreError reports a config-orphaned root whose primary checkout
// could not be identified, so it has no store root at all.
type UnresolvedStoreError struct {
	Root  string
	Cause error
}

func (e *UnresolvedStoreError) Error() string {
	return fmt.Sprintf("store root unresolved: the primary checkout of worktree %s could not be identified (%v); "+
		"its state is kept in the primary checkout's .moai/state and is not written anywhere else", e.Root, e.Cause)
}

func (e *UnresolvedStoreError) Unwrap() error { return e.Cause }

// StoreRoot returns the directory whose .moai/state holds root's state: the
// primary checkout for a config-orphaned root whose primary is identifiable,
// root itself for every root that is not config-orphaned (a primary checkout
// therefore maps to itself), and an *UnresolvedStoreError — never a substituted
// root — when root is config-orphaned and its primary cannot be identified.
// A root that is not config-orphaned costs two file reads and runs no git.
//
// @MX:ANCHOR: [AUTO] the one store-root answer shared by MCP writers, hook guards, review gates and the verify CLI
// @MX:REASON: fan_in >= 3 across internal/cli and internal/hook; two call sites answering differently reopens the reader-disagreement defect
func StoreRoot(root string) (string, error) {
	if !IsConfigOrphanedRoot(root) {
		return root, nil
	}
	primary, _, err := IdentifyPrimaryCheckout(root)
	if err != nil {
		return "", &UnresolvedStoreError{Root: root, Cause: err}
	}
	return primary, nil
}
