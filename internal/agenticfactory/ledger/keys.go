package ledger

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// The shared conflict keys (contract §3): files many cards touch that no package
// directory stands for.
const (
	KeyGoMod     = "shared:go.mod"
	KeyTemplates = "shared:templates"
	KeyBuild     = "shared:build"
	KeyRules     = "shared:rules"
	keyPkgPrefix = "pkg:"
)

// KeysFor turns the paths a card edits into its conflict keys, sorted and without
// duplicates. A path under a shared area gives that area's key and no package key; any
// other path gives the key of its nearest package directory, which is its own directory
// (a path ending in "/" is a directory itself), and a "testdata" or "fixtures" segment
// belongs to the package that owns it, so it gives that package's key. Paths are taken
// relative to the repository root and cleaned of "." segments; blank ones are skipped. A
// path that is absolute, has a ".." segment, or has a backslash makes the whole call
// ErrInvalid: it leaves the repository or is not written as the repository's paths are.
func KeysFor(paths []string) ([]string, error) {
	set := map[string]bool{}
	for _, raw := range paths {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		if strings.ContainsRune(p, '\\') || strings.HasPrefix(p, "/") {
			return nil, fmt.Errorf("%w: path %q is absolute or uses a backslash", ErrInvalid, raw)
		}
		for _, seg := range strings.Split(p, "/") {
			if seg == ".." {
				return nil, fmt.Errorf("%w: path %q leaves the repository", ErrInvalid, raw)
			}
		}
		isDir := strings.HasSuffix(p, "/")
		p = path.Clean(strings.TrimPrefix(path.Clean(p), "./"))
		if k := sharedKey(p); k != "" {
			set[k] = true
			continue
		}
		dir := p
		if !isDir {
			dir = path.Dir(p)
		}
		set[keyPkgPrefix+owningDir(p, dir)] = true
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// owningDir is dir, or the directory that owns it when p runs through a testdata or
// fixtures segment: the part of p before the first such segment.
func owningDir(p, dir string) string {
	segs := strings.Split(p, "/")
	for i, seg := range segs {
		if seg == "testdata" || seg == "fixtures" {
			if i == 0 {
				return "."
			}
			return strings.Join(segs[:i], "/")
		}
	}
	return dir
}

func sharedKey(p string) string {
	switch {
	case p == "go.mod" || p == "go.sum":
		return KeyGoMod
	case strings.HasPrefix(p, "internal/template/templates/") || p == "internal/template/templates":
		return KeyTemplates
	case p == "Makefile" || p == ".github" || strings.HasPrefix(p, ".github/"):
		return KeyBuild
	case p == ".claude/rules" || strings.HasPrefix(p, ".claude/rules/"):
		return KeyRules
	}
	return ""
}
