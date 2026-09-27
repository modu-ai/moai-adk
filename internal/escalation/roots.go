package escalation

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Exemption root names used in not-observed lines (design.md §C.9).
const (
	rootScratchpad = "session scratchpad"
	rootMemory     = "auto-memory store"
)

// MemorySlug is the project directory name Claude Code's auto-memory store
// uses for a project path: every '/', '\\', '.', and ':' becomes '-'.
//
// @MX:NOTE: [AUTO] mirrors memoryProjectSlug in internal/cli/memory.go (unexported there); the two must stay equal or the auto-memory exemption of REQ-AE-013 stops matching
func MemorySlug(absPath string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', '.', ':':
			return '-'
		}
		return r
	}, filepath.Clean(absPath))
}

// memoryRoots returns every candidate auto-memory store for the worktree
// (its own path and its primary checkout, under $CLAUDE_CONFIG_DIR and
// ~/.claude), the set `moai memory doctor` audits. ok is false when neither
// base directory can be determined.
func memoryRoots(worktreeRoot string) (roots []string, ok bool) {
	var bases []string
	if cfg := os.Getenv(config.EnvClaudeConfigDir); cfg != "" {
		bases = append(bases, cfg)
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		bases = append(bases, filepath.Join(home, ".claude"))
	}
	if len(bases) == 0 {
		return nil, false
	}
	keys := []string{worktreeRoot}
	if primary, err := canonicalProjectRoot(worktreeRoot); err == nil && primary != worktreeRoot {
		keys = append(keys, primary)
	}
	for _, b := range bases {
		for _, k := range keys {
			roots = append(roots, filepath.Join(b, "projects", MemorySlug(k), "memory"))
		}
	}
	return roots, true
}

// tempRoot is the OS temporary directory; os.TempDir already honors $TMPDIR,
// and within compares its resolved spelling (design.md §C.9: never
// undeterminable).
func tempRoot() string {
	return os.TempDir()
}

// canonicalPath resolves symlinks in p, or in its nearest existing ancestor
// when p does not exist yet, so paths compare in one spelling (on macOS
// /var and /private/var name the same directory).
func canonicalPath(p string) string {
	p = filepath.Clean(p)
	var tail []string
	for cur := p; ; {
		if resolved, err := filepath.EvalSymlinks(cur); err == nil {
			for i := len(tail) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, tail[i])
			}
			return resolved
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return p
		}
		tail = append(tail, filepath.Base(cur))
		cur = parent
	}
}

// within reports whether path lies inside (or is) dir; both are compared in
// canonical spelling.
func within(path, dir string) bool {
	rel, err := filepath.Rel(canonicalPath(dir), canonicalPath(path))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
