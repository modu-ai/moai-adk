package ledger

import (
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
// (a path ending in "/" is a directory itself). Paths are cleaned and are taken relative
// to the repository root; blank ones are skipped.
func KeysFor(paths []string) []string {
	set := map[string]bool{}
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
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
		set[keyPkgPrefix+dir] = true
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
