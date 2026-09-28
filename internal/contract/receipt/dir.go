package receipt

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// storeSubdir is the store's directory under the project's home-state
// directory, the same directory the escalation detector keeps its arming
// state in.
const storeSubdir = "contract"

// StoreDir returns $MOAI_HOME/db/<project-key>/contract for the project the
// worktree belongs to.
//
// It reproduces escalation.StoreDir (and the project key of
// homestate.ProjectKey) instead of importing them: both import
// internal/config, whose tests sign fixture contracts through this package,
// so importing them here would close an import cycle. A parity test pins the
// two resolutions to the same directory. Like the escalation resolver it
// reads the repository's .git files and never starts git, and it refuses a
// layout whose common directory is not named ".git" rather than guess.
//
// @MX:NOTE: [AUTO] Deliberate reproduction of the escalation store resolver;
// the receipt parity test (TestStoreDirMatchesEscalation) fails when the two
// drift apart.
func StoreDir(worktreeRoot string) (string, error) {
	root, err := canonicalRoot(worktreeRoot)
	if err != nil {
		return "", err
	}
	home, err := paths.MoaiHome()
	if err != nil {
		return "", fmt.Errorf("contract store: resolve MOAI_HOME: %w", err)
	}
	return filepath.Join(home, "db", projectKey(root), storeSubdir), nil
}

// projectKey is the readable project key: the sanitized base name plus the
// first four bytes of the SHA-256 of the canonical root.
func projectKey(root string) string {
	sum := sha256.Sum256([]byte(root))
	base := filepath.Base(root)
	if base == "." || base == string(filepath.Separator) || base == "" {
		base = "project"
	}
	base = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		default:
			return '-'
		}
	}, base)
	return fmt.Sprintf("%s-%x", base, sum[:4])
}

// canonicalRoot is the primary checkout itself, or the parent of a linked
// worktree's common ".git" directory, symlinks resolved.
func canonicalRoot(worktreeRoot string) (string, error) {
	abs, err := filepath.Abs(worktreeRoot)
	if err != nil {
		return "", err
	}
	dotGit := filepath.Join(abs, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return "", fmt.Errorf("contract store: not a git worktree root: %s", abs)
	}
	root := abs
	if !info.IsDir() {
		data, err := os.ReadFile(dotGit)
		if err != nil {
			return "", fmt.Errorf("contract store: read %s: %w", dotGit, err)
		}
		rest, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "gitdir:")
		if !ok {
			return "", errors.New("contract store: " + dotGit + " is not a gitdir file")
		}
		gitDir := strings.TrimSpace(rest)
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(abs, gitDir)
		}
		commonDir := gitDir
		if c, err := os.ReadFile(filepath.Join(gitDir, "commondir")); err == nil {
			commonDir = strings.TrimSpace(string(c))
			if !filepath.IsAbs(commonDir) {
				commonDir = filepath.Join(gitDir, commonDir)
			}
		}
		gitDir, commonDir = filepath.Clean(gitDir), filepath.Clean(commonDir)
		if gitDir != commonDir {
			if filepath.Base(commonDir) != ".git" {
				return "", fmt.Errorf("contract store: unsupported git layout (common dir %s)", commonDir)
			}
			root = filepath.Dir(commonDir)
		}
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	return filepath.Clean(root), nil
}
