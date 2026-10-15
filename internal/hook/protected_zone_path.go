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
	"runtime"
	"strings"
	"unicode"

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

// zoneNativeSlash converts p for the filesystem-facing steps of the zone target
// resolution: on Windows a "\" is a separator and is normalized exactly like
// zoneSlash; on every POSIX platform it is an ordinary filename character and
// rides through untouched. Rewriting it before the walk validates a fictional
// path the OS never names — a literal `lnk\dir` symlink into the zone resolved
// as `lnk/dir` misses the real link and the guard allowed a Write that the OS
// landed inside the zone. The lexical comparison arm keeps zoneSlash's
// slash-normalized matching semantics.
//
// @MX:SPEC:SPEC-HOOK-ZONE-BACKSLASH-001
func zoneNativeSlash(p string) string {
	if runtime.GOOS == "windows" {
		return zoneSlash(p)
	}
	return p
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
	return zoneResolveDepth(p, 0)
}

// zoneResolveDepth carries the symlink-following depth; a chain deeper than
// the bound reads as outside (fail closed), standing in for the kernel's
// ELOOP.
func zoneResolveDepth(p string, depth int) (string, bool) {
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real, true
	}
	return zoneResolveWalk(p, depth, runtime.GOOS == "windows")
}

// zoneResolveWalk is the physical component walk behind zoneResolveDepth,
// split out so a caller that must NOT take the EvalSymlinks fast path can
// force it (card t1712): the fast path collapses a ".." that follows a
// symlink against the LEXICAL parent, which a parent-only resolution cannot
// afford — a fully-existing `dirlink/..` would rejoin the final name against
// the lexical parent while the shell walks it physically.
func zoneResolveWalk(p string, depth int, windows bool) (string, bool) {
	volume := filepath.VolumeName(p)
	parts := pathSegments(strings.TrimPrefix(p, volume), windows)
	resolved := filepath.ToSlash(volume) + "/"
	skipped := 0
	absolute := volume != "" || strings.HasPrefix(p, "/") ||
		(windows && strings.HasPrefix(p, "\\"))
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		switch part {
		case "", ".":
			// repeated or trailing separators, and the dot itself
		case "..":
			if absolute {
				// POSIX root semantics (card t1712 repair round): ".." at the
				// filesystem root stays at the root — the shell resolves
				// /../r/alias as r/alias and deletes through it. The walk
				// continues from the root instead of failing closed, matching
				// the EvalSymlinks fast path and the lexical arm's Clean on
				// the same spelling.
				// Root-reach is decided from the RESOLVED PREFIX STRING, never
				// the skipped counter (card t1712 repair round 6): one
				// resolved dir symlink adds many real components but a single
				// counter unit, so a counter-driven root rule runs out
				// mid-walk and ignores the remaining ".." while real bash is
				// still climbing INSIDE the zone. ".." is a no-op only when
				// the resolved prefix IS the filesystem root; otherwise it
				// pops one real component — the same climb the shell performs.
				if resolved != filepath.ToSlash(volume)+"/" {
					trimmed := strings.TrimSuffix(resolved, "/")
					if idx := strings.LastIndex(trimmed, "/"); idx >= 0 {
						resolved = trimmed[:idx+1]
					}
				}
				continue
			}
			if skipped == 0 {
				// RELATIVE walk keeps the fail-closed answer: nothing vouches
				// for a ".." above its base
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
				// EvalSymlinks fails for a MISSING component and equally for
				// an EXISTING symlink whose target is missing — only the
				// first is an unresolved tail. The second is a real link the
				// shell would still follow (a Write through it creates the
				// destination), so its destination rejoins the walk (round
				// 11 P1: `innocent.md -> zone_dir/new.md` let a Write create
				// the file inside the zone).
				if _, lerr := os.Lstat(probe); lerr != nil {
					// the first genuinely missing component starts the
					// unresolved tail
					return filepath.Join(resolved, strings.Join(parts[i:], "/")), true
				}
				dest, rerr := os.Readlink(probe)
				if rerr != nil {
					return "", false // an unreadable existing entry: fail closed
				}
				if !filepath.IsAbs(dest) {
					dest = resolved + dest
				}
				if depth >= zoneSymlinkDepthBound {
					return "", false // chain too deep: fail closed
				}
				sub, ok := zoneResolveDepth(dest, depth+1)
				if !ok {
					return "", false
				}
				// the link's resolution takes the component's place; what
				// remains of the original path rejoins onto it
				resolved = strings.TrimSuffix(filepath.ToSlash(sub), "/") + "/"
				skipped++
				continue
			}
			resolved = strings.TrimSuffix(filepath.ToSlash(real), "/") + "/"
			skipped++
		}
	}
	return filepath.Clean(resolved), true
}

// resolveZoneTarget returns the relative forms of raw under root: the lexical form
// first, then the symlink-resolved form when it differs. An empty result means the
// target is outside the project root on every reading.
func resolveZoneTarget(root, raw string, listings ...map[string]*zoneNameIndex) []zoneForm {
	abs := zoneSlash(raw)
	// walk carries the same input with its POSIX component identity preserved
	// for the symlink arm: the filesystem steps must follow the components the
	// OS actually names, and on POSIX a "\" inside a component is an ordinary
	// filename character (zoneSlash would destroy a literal `lnk\dir` symlink
	// into the zone — SPEC-HOOK-ZONE-BACKSLASH-001). On Windows both spellings
	// coincide. Relative input joins against the cwd the same way abs does,
	// without the rewrite.
	walk := zoneNativeSlash(raw)
	if !zoneIsAbs(abs) {
		if cwd, err := zoneGetwd(); err == nil && cwd != "" {
			// concatenated, never path.Join: the symlink arm must see the raw
			// segments — Join would collapse "deep/../secret.md" before the
			// filesystem resolves "deep" (merge-gate round 1 P1-4). The lexical
			// arm cleans inside zoneLexicalRel, which is its own semantics.
			abs = zoneSlash(cwd) + "/" + abs
			walk = zoneNativeSlash(cwd) + "/" + walk
		}
	}
	// The prepend decision above reads the CONVERTED spelling: on POSIX a raw
	// `\alias/x` converts to a "/"-leading form and `C:\alias/x` to a
	// drive-letter form — both wrongly judged absolute, so the walk never
	// received the cwd and no arm followed the literal backslash component into
	// the zone (gate round 8 vector). When the conversion changed the spelling,
	// give the walk its own platform-correct prepend; when it did not
	// (walk == abs, every backslash-free input), this is a no-op and the
	// pre-existing absoluteness semantics are byte-identical.
	if walk != abs && !filepath.IsAbs(walk) {
		if cwd, err := zoneGetwd(); err == nil && cwd != "" {
			walk = zoneNativeSlash(cwd) + "/" + walk
		}
	}
	var forms []zoneForm
	add := func(rel string) {
		forms = appendZoneForm(forms, rel)
	}
	if rel, inside := zoneLexicalRel(root, abs); inside {
		add(rel)
	}
	native := filepath.FromSlash(walk)
	if root != "" && filepath.IsAbs(native) {
		realRoot, ok := zoneResolve(filepath.FromSlash(zoneNativeSlash(root)))
		if !ok {
			realRoot = filepath.FromSlash(zoneNativeSlash(root))
		}
		if realTarget, ok := zoneResolve(native); ok {
			if rel, inside := zoneLexicalRel(filepath.ToSlash(realRoot), filepath.ToSlash(realTarget)); inside {
				add(rel)
			}
		}
	}
	// Backslash-literal arm (card t1570): on a POSIX host a backslash is an
	// ordinary filename character, but zoneSlash has already folded the
	// candidate to slashes — `lnk\dir` became `lnk/dir`, and the symlink the
	// command writes through is invisible to both arms above. Resolve the
	// raw candidate too, with every backslash intact. On Windows the raw
	// form carries separators already, so the arm re-derives the native
	// forms and the dedup drops the duplicates.
	if root != "" {
		absRaw := raw
		if !zoneIsAbs(zoneSlash(raw)) {
			if cwd, err := zoneGetwd(); err == nil && cwd != "" {
				absRaw = zoneSlash(cwd) + "/" + raw
			}
		}
		if zoneIsAbs(zoneSlash(absRaw)) {
			realRoot, ok := zoneResolve(filepath.FromSlash(zoneSlash(root)))
			if !ok {
				realRoot = filepath.FromSlash(zoneSlash(root))
			}
			if realTarget, ok := zoneResolve(absRaw); ok {
				if rel, inside := zoneLexicalRel(filepath.ToSlash(realRoot), filepath.ToSlash(realTarget)); inside {
					add(rel)
				}
			}
		}
	}
	return forms
}

// appendZoneForm adds one relative spelling to forms unless its folded form
// is already present; the dedup keeps the first spelling's Display text.
// Shared by both zone target resolvers.
func appendZoneForm(forms []zoneForm, rel string) []zoneForm {
	f := zoneForm{Display: rel, Folded: config.FoldZoneText(rel)}
	for _, have := range forms {
		if have.Folded == f.Folded {
			return forms
		}
	}
	return append(forms, f)
}

// resolveZoneLinkTarget returns the relative forms of raw under root for an
// unlink-family operand judged by its FINAL component itself (card t1712):
// rm/unlink remove the link an operand names, never the target behind it, so
// the final component is NOT followed. Parent components resolve through
// symlinks exactly as before — through the physical walk, never the
// EvalSymlinks fast path, whose lexical ".." collapse would rejoin the final
// name outside the zone while the shell deletes inside it (zoneResolveWalk).
// The lexical arm is shared with resolveZoneTarget; only the filesystem arms
// differ (parent-only resolution).
func resolveZoneLinkTarget(root, raw string, listings ...map[string]*zoneNameIndex) []zoneForm {
	abs := zoneSlash(raw)
	walk := zoneNativeSlash(raw)
	if !zoneIsAbs(abs) {
		if cwd, err := zoneGetwd(); err == nil && cwd != "" {
			abs = zoneSlash(cwd) + "/" + abs
			walk = zoneNativeSlash(cwd) + "/" + walk
		}
	}
	if walk != abs && !filepath.IsAbs(walk) {
		if cwd, err := zoneGetwd(); err == nil && cwd != "" {
			walk = zoneNativeSlash(cwd) + "/" + walk
		}
	}
	var forms []zoneForm
	add := func(rel string) {
		forms = appendZoneForm(forms, rel)
	}
	if rel, inside := zoneLexicalRel(root, abs); inside {
		add(rel)
	}
	native := filepath.FromSlash(walk)
	if root != "" && filepath.IsAbs(native) {
		realRoot, ok := zoneResolve(filepath.FromSlash(zoneNativeSlash(root)))
		if !ok {
			realRoot = filepath.FromSlash(zoneNativeSlash(root))
		}
		if rejoined, ok := zoneResolveParent(native, listingCache(listings)); ok {
			for _, realTarget := range rejoined {
				if rel, inside := zoneLexicalRel(filepath.ToSlash(realRoot), filepath.ToSlash(realTarget)); inside {
					add(rel)
				}
			}
		}
	}
	// Backslash-literal arm (card t1570), the parent-only reading: the raw
	// candidate resolves with every backslash intact, split at the slash the
	// candidate's final component starts after. On POSIX a "\" inside the
	// final component is an ordinary filename character and rides through;
	// on Windows the native arm above already carried the raw separators, so
	// the slash split finds no component here and the arm is skipped.
	if root != "" {
		absRaw := raw
		if !zoneIsAbs(zoneSlash(raw)) {
			if cwd, err := zoneGetwd(); err == nil && cwd != "" {
				absRaw = zoneSlash(cwd) + "/" + raw
			}
		}
		if zoneIsAbs(zoneSlash(absRaw)) {
			realRoot, ok := zoneResolve(filepath.FromSlash(zoneSlash(root)))
			if !ok {
				realRoot = filepath.FromSlash(zoneSlash(root))
			}
			if rejoined, ok := zoneResolveParent(absRaw, listingCache(listings)); ok {
				for _, realTarget := range rejoined {
					if rel, inside := zoneLexicalRel(filepath.ToSlash(realRoot), filepath.ToSlash(realTarget)); inside {
						add(rel)
					}
				}
			}
		}
	}
	return forms
}

// zoneResolveParent resolves everything before the FINAL component of p and
// rejoins that component as the link's own path (card t1712): the parent
// prefix goes through the physical walk — never the EvalSymlinks fast path,
// whose ".." handling is lexical — and the final name lands on the resolved
// parent. The final NAME is canonicalized first from the parent's real
// directory listing on ALL platforms (repair round 7 introduced it on
// windows; the landing repair removed the gate — case-insensitive aliasing
// is not a windows property: macOS default APFS behaves the same): an
// operand spelled `ä.txt` meets the entry `Ä.txt` the zone manifest pins,
// and FoldZoneText folds ASCII only, so the raw name missed the exact-entry
// zone form. The match is Unicode case-insensitive (strings.EqualFold) and
// NAME-only — the entry is never opened or followed; a listing failure
// keeps the raw name. This also absorbs the 8.3 short-name case for entries
// that exist, superseding the classifier's "~" guard for real listings
// (which stays as a harmless extra gate). A "." or ".." final segment is not
// a name the shell unlinks as itself; the caller routes those spellings to
// the target-following judgment and this function refuses them.
// zoneParentLister reads a resolved parent's directory entries — a package
// var so tests can count listings (repair round 12 P2) and inject synthetic
// listings a case-sensitive-shape darwin host cannot host (repair round 12
// P1).
var zoneParentLister = func(dir string) ([]os.DirEntry, error) {
	return os.ReadDir(dir)
}

// listingCache unwraps the resolvers' optional listing-cache argument: the
// guard's candidate loop threads ONE per-call map through every resolution
// (repair round 12 P2), while direct callers (tests, the single-shot guard
// path) omit it and run uncached.
func listingCache(listings []map[string]*zoneNameIndex) map[string]*zoneNameIndex {
	if len(listings) > 0 {
		return listings[0]
	}
	return nil
}

// zoneResolveParent resolves everything before the FINAL component of p and
// rejoins that component as the link's own path (card t1712): the parent
// prefix goes through the physical walk — never the EvalSymlinks fast path,
// whose ".." handling is lexical — and the final name lands on the resolved
// parent. The final NAME is canonicalized first from the parent's real
// directory listing on ALL platforms (repair round 7 introduced it on
// windows; the landing repair removed the gate — case-insensitive aliasing
// is not a windows property: macOS default APFS behaves the same): an
// operand spelled `ä.txt` meets the entry `Ä.txt` the zone manifest pins,
// and FoldZoneText folds ASCII only, so the raw name missed the exact-entry
// zone form. The match is Unicode case-insensitive (strings.EqualFold) and
// NAME-only — the entry is never opened or followed; a listing failure
// keeps the raw name. This also absorbs the 8.3 short-name case for entries
// that exist, superseding the classifier's "~" guard for real listings
// (which stays as a harmless extra gate). A "." or ".." final segment is not
// a name the shell unlinks as itself; the caller routes those spellings to
// the target-following judgment and this function refuses them.
//
// The caller threads a per-guard-call LISTING CACHE (keyed by the resolved
// parent path; nil disables caching) so N candidates sharing one parent read
// the listing once (repair round 12 P2) — never carried across guard calls.
// The returned rejoined paths carry the CANONICALIZED spelling FIRST and the
// RAW spelling alongside when the two differ (repair round 12 P1): both
// spellings judge, and the zone-form dedup absorbs the identity.
func zoneResolveParent(p string, listings map[string]*zoneNameIndex) ([]string, bool) {
	dir, final := path.Split(filepath.ToSlash(p))
	if dir == "" || final == "" || final == "." || final == ".." {
		return nil, false
	}
	windows := runtime.GOOS == "windows"
	resolved, ok := zoneResolveWalk(filepath.FromSlash(dir), 0, windows)
	if !ok {
		return nil, false
	}
	// Canonicalize the final NAME from the resolved parent's real listing on
	// ALL platforms (card t1712 landing repair): the windows-only round-7
	// gate missed the POSIX sibling of the same aliasing — a case-insensitive
	// volume (macOS default APFS among them) answers an operand spelled
	// `ä.txt` with the entry `Ä.txt` the manifest pins, and FoldZoneText
	// folds ASCII only, so the exact-entry zone form missed the alias and
	// the removal read allow. (1) On a CASE-SENSITIVE volume a fold-pair
	// query may canonicalize to its existing differently-cased sibling — a
	// deny-side over-block, accepted and documented (the safe direction);
	// exact-only matches resolve to themselves and change nothing. (2) This
	// subsumes the windows round-7 gate — 8.3-in-listing absorption now
	// applies on windows through the same path. (3) A listing failure keeps
	// the raw judgment, today's pre-canonicalization behavior.
	var idx *zoneNameIndex
	if cached, hit := listings[resolved]; hit {
		idx = cached
	} else if entries, err := zoneParentLister(resolved); err == nil {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		idx = newZoneNameIndex(names)
		if listings != nil {
			listings[resolved] = idx
		}
	}
	canonical := final
	if idx != nil {
		canonical = idx.lookup(final)
	}
	rejoined := []string{filepath.Join(resolved, filepath.FromSlash(canonical))}
	if canonical != final {
		// both spellings judge (repair round 12 P1): the raw spelling stays
		// a form alongside the canonicalized one — the zone-form dedup
		// absorbs the identity when the two coincide
		rejoined = append(rejoined, filepath.Join(resolved, filepath.FromSlash(final)))
	}
	return rejoined, true
}

// zoneNameIndex is the per-parent O(1) lookup index over a directory
// listing (repair round 13 P2): a PreToolUse hook cannot afford 4,096
// candidates × a 100,000-entry linear scan, so the canonical-name match is
// indexed — the EXACT entry name and its Unicode simple-fold orbit key both
// point at the entry — built once per resolved parent on first listing, in
// the same zoneParentLister seam as the listing cache. Semantics are the
// round-12 matcher's, exactly: the exact match wins before any fold (a
// case-sensitive volume can hold both spellings, and a fold-first lookup
// would judge whichever sibling inserted first); a fold-orbit match
// canonicalizes; no match keeps the raw name.
type zoneNameIndex struct {
	exact  map[string]string
	folded map[string]string
}

// newZoneNameIndex builds the index over one parent's entry names. The
// folded map keeps the FIRST entry per orbit key — the same first-hit
// semantics the round-12 listing-order scan had.
func newZoneNameIndex(names []string) *zoneNameIndex {
	idx := &zoneNameIndex{
		exact:  make(map[string]string, len(names)),
		folded: make(map[string]string, len(names)),
	}
	for _, name := range names {
		idx.exact[name] = name
		key := zoneFoldKey(name)
		if _, ok := idx.folded[key]; !ok {
			idx.folded[key] = name
		}
	}
	return idx
}

// lookup returns the canonical entry spelling for final: the exact entry
// when one exists, the first fold-orbit match otherwise, and final itself
// when nothing matches. NAME-only — the entry is never opened or followed.
func (idx *zoneNameIndex) lookup(final string) string {
	if name, ok := idx.exact[final]; ok {
		return name
	}
	if name, ok := idx.folded[zoneFoldKey(final)]; ok {
		return name
	}
	return final
}

// zoneFoldKey maps every rune to the minimum rune of its Unicode
// simple-fold orbit — exactly the relation strings.EqualFold accepts, so
// index lookups match the reference scan without walking the listing.
// Boundary kept from the matcher: locale case tables are not simple folds
// (an ı/I pair NTFS equates stays unmatched here), the gap the windows
// narrowing deferral (repair round 8) closes structurally on that platform.
func zoneFoldKey(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		min := r
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			if f < min {
				min = f
			}
		}
		b.WriteRune(min)
	}
	return b.String()
}
