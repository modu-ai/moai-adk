package escalation

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// storeSubdir is the contract store's directory under the project's home-state
// directory: $MOAI_HOME/db/<project-key>/contract (spec.md §D definitions).
const storeSubdir = "contract"

// escalationSubdir holds the per-card state files, logs, and locks.
const escalationSubdir = "escalation"

// errNotGitRoot reports a directory that carries no .git entry.
var errNotGitRoot = errors.New("escalation: not a git worktree root")

// StoreDir returns the contract store for the project the worktree belongs
// to: $MOAI_HOME/db/<project-key>/contract, with $MOAI_HOME and the project
// key resolved as the queue database resolves them (research.md P17). It
// reads the repository's .git files and never starts git, so a PreToolUse
// hook may call it (spec.md C4). A layout it cannot resolve that way —
// no home directory, a separate or bare git directory — is an error, which
// the detector reports as a REQ-AE-004 fault; nothing arms without a store.
func StoreDir(worktreeRoot string) (string, error) {
	canonical, err := canonicalProjectRoot(worktreeRoot)
	if err != nil {
		return "", err
	}
	home, err := paths.MoaiHome()
	if err != nil {
		return "", fmt.Errorf("escalation: resolve MOAI_HOME: %w", err)
	}
	return filepath.Join(home, "db", homestate.ProjectKeyForCanonicalRoot(canonical), storeSubdir), nil
}

// CardFiles names the three per-card files in the contract store.
type CardFiles struct {
	// Log is the card audit log, <card>.log.jsonl: authoritative for arming.
	Log string
	// State is the card state file, <card>.json: a cache of the arming
	// snapshot, verify cache, and counters.
	State string
	// Lock serializes every read-modify-write of Log and State.
	Lock string
}

// CardFilesFor returns the card files of card in the worktree's store. The
// card id is used byte for byte as the file name stem (design.md §C.6).
func CardFilesFor(worktreeRoot, card string) (CardFiles, error) {
	store, err := StoreDir(worktreeRoot)
	if err != nil {
		return CardFiles{}, err
	}
	dir := filepath.Join(store, escalationSubdir)
	return CardFiles{
		Log:   filepath.Join(dir, card+".log.jsonl"),
		State: filepath.Join(dir, card+".json"),
		Lock:  filepath.Join(dir, card+".lock"),
	}, nil
}

// gitDirs resolves the git directory and common directory of the worktree
// from its .git entry: a directory in a primary checkout, a "gitdir:" file in
// a linked worktree (whose git directory holds a "commondir" file).
func gitDirs(worktreeRoot string) (gitDir, commonDir string, err error) {
	dotGit := filepath.Join(worktreeRoot, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return "", "", fmt.Errorf("%w: %s", errNotGitRoot, worktreeRoot)
	}
	if info.IsDir() {
		return dotGit, dotGit, nil
	}
	data, err := os.ReadFile(dotGit)
	if err != nil {
		return "", "", fmt.Errorf("escalation: read %s: %w", dotGit, err)
	}
	line := strings.TrimSpace(string(data))
	rest, ok := strings.CutPrefix(line, "gitdir:")
	if !ok {
		return "", "", fmt.Errorf("escalation: %s is not a gitdir file", dotGit)
	}
	gitDir = strings.TrimSpace(rest)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(worktreeRoot, gitDir)
	}
	commonDir = gitDir
	if c, err := os.ReadFile(filepath.Join(gitDir, "commondir")); err == nil {
		commonDir = strings.TrimSpace(string(c))
		if !filepath.IsAbs(commonDir) {
			commonDir = filepath.Join(gitDir, commonDir)
		}
	}
	return filepath.Clean(gitDir), filepath.Clean(commonDir), nil
}

// canonicalProjectRoot reproduces homestate.CanonicalProjectRoot for the
// ordinary layouts without starting git: the primary checkout itself, or the
// parent of a linked worktree's common ".git" directory, symlinks resolved.
// A common directory not named ".git" (separate, bare, submodule) is refused
// rather than guessed.
func canonicalProjectRoot(worktreeRoot string) (string, error) {
	abs, err := filepath.Abs(worktreeRoot)
	if err != nil {
		return "", err
	}
	gitDir, commonDir, err := gitDirs(abs)
	if err != nil {
		return "", err
	}
	root := abs
	if gitDir != commonDir {
		if filepath.Base(commonDir) != ".git" {
			return "", fmt.Errorf("escalation: unsupported git layout (common dir %s)", commonDir)
		}
		root = filepath.Dir(commonDir)
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	return filepath.Clean(root), nil
}

// ReadHead returns the commit HEAD names, read from the worktree's HEAD file
// and the ref it points at (loose ref, then packed-refs), never through git.
func ReadHead(worktreeRoot string) (string, error) {
	gitDir, commonDir, err := gitDirs(worktreeRoot)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", fmt.Errorf("escalation: read HEAD: %w", err)
	}
	head := strings.TrimSpace(string(data))
	ref, ok := strings.CutPrefix(head, "ref:")
	if !ok {
		return head, nil
	}
	ref = strings.TrimSpace(ref)
	for _, dir := range []string{gitDir, commonDir} {
		if b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(ref))); err == nil {
			return strings.TrimSpace(string(b)), nil
		}
	}
	f, err := os.Open(filepath.Join(commonDir, "packed-refs"))
	if err != nil {
		return "", fmt.Errorf("escalation: ref %s not found", ref)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		sha, name, ok := strings.Cut(sc.Text(), " ")
		if ok && name == ref {
			return sha, nil
		}
	}
	return "", fmt.Errorf("escalation: ref %s not found", ref)
}
