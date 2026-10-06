package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// postToolTargetOutsideProject reports whether the Write/Edit/MultiEdit target
// lies outside the project, so the lint (Quality Gate) and security scans skip
// it: a throwaway script under a session scratchpad or /tmp is not project
// code and its findings are noise (card t1507, re-landing card t1499's
// reverted M4).
//
// "Inside" means inside the project root (CLAUDE_PROJECT_DIR) or inside the
// root of the worktree or repository the session cwd belongs to — the nearest
// ancestor of the cwd carrying a .git entry. The cwd directory itself is NOT a
// root: a session parked in a scratchpad must not re-scope scratchpad files to
// inside, and a lane working in a worktree is covered by that worktree's own
// root, which spans its parent and sibling directories too (the hole that
// reverted the first attempt).
//
// Both sides are compared after symlink resolution, so a spelling (macOS /tmp
// versus /private/tmp) never decides scope, and on case-insensitive platforms
// a case-only spelling difference does not either (the launcher prints
// /Users/goos/moai while the session reports /Users/goos/MoAI — the same
// directory).
//
// It fails closed toward scanning: with no root known, no file_path, or an
// ambiguous judgment it reports false and the scans run exactly as before.
func postToolTargetOutsideProject(input *HookInput) bool {
	if input == nil {
		return false
	}
	var ti struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(input.ToolInput, &ti); err != nil || ti.FilePath == "" {
		return false
	}

	// The project root, plus the root of the worktree the session cwd belongs
	// to. A duplicate (a worktree inside the project, or cwd at the project
	// root) costs one redundant comparison and is not worth deduplicating.
	var roots []string
	if projectDir := os.Getenv(config.EnvClaudeProjectDir); projectDir != "" {
		roots = append(roots, projectDir)
	}
	wtRoot, wtUncertain := containingWorktreeRoot(input.CWD)
	if wtUncertain {
		// Whether the cwd belongs to a worktree could not be determined: a
		// root may exist that we could not see, so the target cannot be
		// proven outside and the scans run.
		return false
	}
	if wtRoot != "" {
		roots = append(roots, wtRoot)
	}
	if len(roots) == 0 {
		return false
	}

	// A relative file_path is relative to the session cwd; the project root is
	// only the base when the payload carries no cwd. The spelling is
	// concatenated, not Joined: a Clean would normalise ".." and inner path
	// components away before symlinks resolve, and the judgment would follow
	// the spelling instead of the link (the second round-2 revert hole).
	target := ti.FilePath
	if !filepath.IsAbs(target) {
		base := input.CWD
		if base == "" {
			base = roots[0]
		}
		target = concatPath(base, target)
	}
	realTarget, targetUncertain := evalSymlinksCertainty(target)
	if targetUncertain {
		// An existing path component refused to resolve (a dangling alias, a
		// permission or loop error): where the target really lies is unknown,
		// and the scans run.
		return false
	}
	for _, r := range roots {
		realRoot, rootUncertain := evalSymlinksCertainty(r)
		if rootUncertain {
			return false
		}
		if pathWithin(realRoot, realTarget) {
			return false
		}
	}
	return true
}

// containingWorktreeRoot returns the root of the worktree or repository dir
// belongs to: the nearest ancestor (dir itself included) carrying a .git
// entry, which is a directory in a normal checkout and a file in a linked
// worktree. The walk runs on the resolved dir, so a cwd reached through a
// symlink still finds its worktree. An empty result with uncertain=false
// means no git root was found — a scratchpad directory, for example — and
// contributes no scope root; uncertain=true means the walk could not tell
// (the dir refused to resolve, or a .git probe failed other than by absence)
// and the caller must not prove the target outside.
func containingWorktreeRoot(dir string) (string, bool) {
	if dir == "" {
		return "", false
	}
	p, uncertain := evalSymlinksCertainty(dir)
	if uncertain {
		return "", true
	}
	for {
		if fi, err := os.Stat(filepath.Join(p, ".git")); err == nil && (fi.IsDir() || fi.Mode().IsRegular()) {
			return p, false
		} else if err != nil && !os.IsNotExist(err) {
			return "", true
		}
		parent := filepath.Dir(p)
		if parent == p {
			return "", false
		}
		p = parent
	}
}

// concatPath concatenates base and rel with a single separator, deliberately
// without filepath.Join: Join (and the Clean under it) resolves ".." lexically
// before symlinks are ever looked at, which is exactly the hole the alias
// chain tests guard.
func concatPath(base, rel string) string {
	if strings.HasSuffix(base, string(filepath.Separator)) {
		return base + rel
	}
	return base + string(filepath.Separator) + rel
}

// caseInsensitivePathFS marks the platforms whose filesystems treat two
// spellings of a path differing only in letter case as the same directory.
var caseInsensitivePathFS = runtime.GOOS == "darwin" || runtime.GOOS == "windows"

// evalSymlinksCertainty resolves symlinks in path and reports whether the
// result is certain. A path that does not exist yet (a file being created) is
// resolved through its deepest existing ancestor, with the missing tail
// appended — a certain result, because components that do not exist carry no
// links. uncertain=true means an EXISTING component refused to resolve (a
// dangling alias, a permission or loop error): the caller cannot prove where
// the path really lies and must scan instead of judging.
//
// The spelling is deliberately NOT cleaned first: a ".." that follows a symlink
// component climbs from the link's real target, which EvalSymlinks handles but
// a lexical Clean would get wrong. Ancestors are therefore peeled off by string
// slicing (filepath.Dir would Clean) and the tail is joined only afterwards.
// Both the platform separator and "/" are accepted as component boundaries —
// Windows tooling emits forward slashes in otherwise backslash paths.
func evalSymlinksCertainty(path string) (string, bool) {
	sep := string(filepath.Separator)
	var tail []string
	for p := path; ; {
		if real, err := filepath.EvalSymlinks(p); err == nil {
			for i := len(tail) - 1; i >= 0; i-- {
				real = filepath.Join(real, tail[i])
			}
			return real, false
		}
		trimmed := strings.TrimRight(p, sep)
		idx := strings.LastIndex(trimmed, sep)
		if i := strings.LastIndex(trimmed, "/"); i > idx {
			idx = i
		}
		if idx < 0 || trimmed == "" {
			// Every component is absent: the spelling is the only truth and
			// the lexical judgment over it is certain.
			return filepath.Clean(path), false
		}
		if _, err := os.Lstat(trimmed); err == nil {
			// The component exists but refused to resolve — uncertainty, not
			// absence.
			return filepath.Clean(path), true
		} else if !os.IsNotExist(err) {
			return filepath.Clean(path), true
		}
		tail = append(tail, trimmed[idx+1:])
		p = trimmed[:idx]
		if p == "" {
			p = sep
		}
	}
}

// pathWithin reports whether target equals root or sits beneath it. Both are
// resolved absolute paths. On a case-insensitive filesystem a case-only
// spelling difference between root and target is the same directory, so the
// comparison is retried case-folded before the target is called outside.
func pathWithin(root, target string) bool {
	if pathBeneath(root, target) {
		return true
	}
	return caseInsensitivePathFS && pathBeneath(strings.ToLower(root), strings.ToLower(target))
}

// pathBeneath is the lexical containment test: target equals root or lies
// beneath it. An unresolvable pair (an error from Rel) is an ambiguous
// judgment and reports true, so the scans run — card t1507's fail-closed
// condition, the reverse of the first attempt's fail-outside.
func pathBeneath(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return true
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}
