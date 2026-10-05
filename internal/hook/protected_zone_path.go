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

// zoneResolve resolves p through symlinks at its deepest existing ancestor and
// rejoins the not-yet-existing remainder, so a target that does not exist yet is
// judged by where it would land. The project root and the target both go through
// this one function: resolving only one of them is the asymmetry that makes an
// in-project path look outside it when the root is reached through a symlink.
func zoneResolve(p string) (string, bool) {
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real, true
	}
	return resolveThroughExistingParent(p)
}

// resolveZoneTarget returns the relative forms of raw under root: the lexical form
// first, then the symlink-resolved form when it differs. An empty result means the
// target is outside the project root on every reading.
func resolveZoneTarget(root, raw string) []zoneForm {
	abs := zoneSlash(raw)
	if !zoneIsAbs(abs) {
		if cwd, err := zoneGetwd(); err == nil && cwd != "" {
			abs = path.Join(zoneSlash(cwd), abs)
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
