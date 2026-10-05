package hook

// protected_zone_path.go — path normalization for the protected-zone guard.
//
// Two layers, split on purpose. zoneLexicalRel is a pure string function that
// touches no filesystem and behaves identically on every platform, so one table
// test covers darwin, linux and windows. resolveZoneTarget adds the filesystem
// steps (symlink resolution of the project root and of the target's deepest
// existing ancestor, with one function for both) and returns every relative form
// the zone is matched against.

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"golang.org/x/text/unicode/norm"
)

// zoneGetwd is the working-directory lookup used to resolve a relative file path
// the way the existing file-access check resolves it (against the hook process's
// working directory). Tests replace it.
var zoneGetwd = os.Getwd

// zoneForm is one relative spelling of a target: Display keeps the original case
// for messages, Folded is NFC plus ASCII-lower for comparison.
type zoneForm struct {
	Display string
	Folded  string
}

// zoneSlash rewrites backslashes to slashes.
func zoneSlash(p string) string {
	return strings.ReplaceAll(filepath.ToSlash(p), "\\", "/")
}

// zoneIsAbs reports whether a slash-normalized path is absolute under the rules
// every platform shares: a leading "/" (which includes a leading "//"), or a
// drive letter followed by ":" and "/". A POSIX host therefore does not call a
// drive-letter path relative.
func zoneIsAbs(p string) bool {
	if strings.HasPrefix(p, "/") {
		return true
	}
	return len(p) >= 3 && p[1] == ':' && p[2] == '/' &&
		((p[0] >= 'a' && p[0] <= 'z') || (p[0] >= 'A' && p[0] <= 'Z'))
}

// zoneLexicalRel expresses p relative to root without consulting the filesystem.
// A relative p is taken as already relative to root. An absolute p is stripped of
// the root prefix (case-insensitively, on a segment boundary); one not under the
// root is outside the zone. Leading "//" input is outside unless the root itself
// starts that way. The returned rel is cleaned and NFC-normalized and keeps its
// original case.
func zoneLexicalRel(root, p string) (rel string, inside bool) {
	p = norm.NFC.String(zoneSlash(p))
	root = norm.NFC.String(zoneSlash(root))
	if p == "" {
		return "", false
	}
	if !zoneIsAbs(p) {
		cp := path.Clean(p)
		if cp == ".." || strings.HasPrefix(cp, "../") {
			return "", false
		}
		return cp, true
	}
	if root == "" || !zoneIsAbs(root) {
		return "", false
	}
	if strings.HasPrefix(p, "//") && !strings.HasPrefix(root, "//") {
		return "", false
	}
	cp, cr := path.Clean(p), path.Clean(root)
	fp, fr := config.FoldZoneText(cp), config.FoldZoneText(cr)
	if fp == fr {
		return ".", true
	}
	prefix := fr
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if !strings.HasPrefix(fp, prefix) {
		return "", false
	}
	// ASCII folding never changes byte length, so the prefix length slices the original-case text.
	return cp[len(prefix):], true
}

// zoneResolve resolves p through symlinks with physical ".." semantics.
// filepath.EvalSymlinks alone is not enough twice over: it fails when the leaf
// does not exist yet, and — measured on go1.26 — it collapses a ".." that
// follows a symlink against the LEXICAL parent, so "deep/../secret.md" with
// deep -> zone_dir/sub resolves to the project root instead of zone_dir. The
// shell walks the same path physically: each existing component is resolved as
// encountered and a ".." pops the resolved prefix. The not-yet-existing tail
// rejoins onto the deepest resolved prefix and is cleaned there (nothing below
// a missing component exists for the shell to walk either).
// (merge-gate round 1 P1-4, observed red first.) p must be absolute — the
// symlink arm of resolveZoneTarget guarantees it.
func zoneResolve(p string) (string, bool) {
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real, true
	}
	parts := strings.Split(p, "/")
	resolved := ""
	if len(parts) > 0 && parts[0] == "" {
		resolved = "/"
	}
	skipped := 0
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		switch part {
		case "", ".":
			// repeated or trailing separators, and the dot itself
		case "..":
			if skipped == 0 {
				// ".." over a prefix that never resolved: the path walks
				// outside anything this function can vouch for
				return "", false
			}
			trimmed := strings.TrimSuffix(resolved, "/")
			if idx := strings.LastIndex(trimmed, "/"); idx >= 0 {
				resolved = trimmed[:idx+1]
			}
			skipped--
		default:
			probe := resolved + part
			real, err := filepath.EvalSymlinks(probe)
			if err != nil {
				// the first missing component starts the unresolved tail
				return filepath.Join(resolved, strings.Join(parts[i:], "/")), true
			}
			resolved = strings.TrimSuffix(real, "/") + "/"
			skipped++
		}
	}
	return filepath.Clean(resolved), true
}

// resolveZoneTarget returns the relative forms of raw under root: the lexical form
// first, then the symlink-resolved form when it differs. An empty result means the
// target is outside the project root on every reading.
func resolveZoneTarget(root, raw string) []zoneForm {
	abs := zoneSlash(raw)
	if !zoneIsAbs(abs) {
		if cwd, err := zoneGetwd(); err == nil && cwd != "" {
			// concatenated, never path.Join: the symlink arm must see the raw
			// segments — Join would collapse "deep/../secret.md" before the
			// filesystem resolves "deep" (merge-gate round 1 P1-4). The lexical
			// arm cleans inside zoneLexicalRel, which is its own semantics.
			abs = zoneSlash(cwd) + "/" + abs
		}
	}
	var forms []zoneForm
	add := func(rel string) {
		f := zoneForm{Display: rel, Folded: config.FoldZoneText(rel)}
		for _, have := range forms {
			if have.Folded == f.Folded {
				return
			}
		}
		forms = append(forms, f)
	}
	if rel, inside := zoneLexicalRel(root, abs); inside {
		add(rel)
	}
	native := filepath.FromSlash(abs)
	if root != "" && filepath.IsAbs(native) {
		realRoot, ok := zoneResolve(filepath.FromSlash(zoneSlash(root)))
		if !ok {
			realRoot = filepath.FromSlash(zoneSlash(root))
		}
		if realTarget, ok := zoneResolve(native); ok {
			if rel, inside := zoneLexicalRel(filepath.ToSlash(realRoot), filepath.ToSlash(realTarget)); inside {
				add(rel)
			}
		}
	}
	return forms
}
