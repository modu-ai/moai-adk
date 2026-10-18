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
	"sort"
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
		// the operand's own spelling, named by the entries the OS holds (card
		// t1804 F1): a final that is a link out of the zone leaves the physical
		// arm below with no zone form at all
		if canon := zoneLexicalCanon(root, rel, listingCache(listings)); canon != rel {
			add(canon)
		}
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
				if canon := zoneCanonicalRel(filepath.ToSlash(realRoot), rel, listingCache(listings)); canon != rel {
					add(canon)
				}
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
					if canon := zoneCanonicalRel(filepath.ToSlash(realRoot), rel, listingCache(listings)); canon != rel {
						add(canon)
					}
				}
			}
		}
	}
	// Gate round 26 #3: the user-root forms live in their OWN stream — the
	// project/baseline matching above must never see them (a basename entry
	// like **/AGENTS.md would otherwise deny an untracked user file on name
	// alone). resolveZoneTarget stays project-only; the deny sites call
	// userRootZoneForms separately and route the results exclusively to
	// ZoneUserRoot entries.
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
//
// REPAIR ROUND 14 — the narrowing is EXACT-MATCH-ONLY: it applies only when
// the operand's final component is byte-equal to an entry of the resolved
// parent's real listing (the precondition probe below). EVERYTHING else —
// fold/normalization/short-name variants, nonexistent finals, LISTING
// FAILURES — defers WHOLESALE to the FULL resolver (resolveZoneTarget for
// the same operand): the two semantics never mix for one operand, and the
// full resolver's EvalSymlinks canonicalization judges the real target.
func resolveZoneLinkTarget(root, raw string, listings ...map[string]*zoneNameIndex) []zoneForm {
	cache := listingCache(listings)
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
	native := filepath.FromSlash(walk)
	// the EXACT-MATCH precondition gates the whole narrowing: no exact entry
	// (or an unresolvable parent) means no link-own judgment at all
	if root != "" && filepath.IsAbs(native) {
		_, exact, ok := zoneResolveParent(native, cache)
		if !ok {
			return nil
		}
		if !exact {
			return resolveZoneTarget(root, raw, cache)
		}
	} else {
		return resolveZoneTarget(root, raw, cache)
	}
	var forms []zoneForm
	add := func(rel string) {
		forms = appendZoneForm(forms, rel)
	}
	if rel, inside := zoneLexicalRel(root, abs); inside {
		add(rel)
		if canon := zoneLexicalCanon(root, rel, cache); canon != rel {
			add(canon)
		}
	}
	if root != "" && filepath.IsAbs(native) {
		realRoot, ok := zoneResolve(filepath.FromSlash(zoneNativeSlash(root)))
		if !ok {
			realRoot = filepath.FromSlash(zoneNativeSlash(root))
		}
		if realTarget, _, ok := zoneResolveParent(native, cache); ok {
			if rel, inside := zoneLexicalRel(filepath.ToSlash(realRoot), filepath.ToSlash(realTarget)); inside {
				add(rel)
				if canon := zoneCanonicalRel(filepath.ToSlash(realRoot), rel, cache); canon != rel {
					add(canon)
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
			realTarget, exact, ok2 := zoneResolveParent(absRaw, cache)
			if ok2 && !exact {
				// the backslash-literal reading names no real entry either —
				// defer the operand wholesale (never mix semantics)
				return resolveZoneTarget(root, raw, cache)
			}
			if ok2 {
				if rel, inside := zoneLexicalRel(filepath.ToSlash(realRoot), filepath.ToSlash(realTarget)); inside {
					add(rel)
					if canon := zoneCanonicalRel(filepath.ToSlash(realRoot), rel, cache); canon != rel {
						add(canon)
					}
				}
			}
		}
	}
	return forms
}

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
// rejoins that component (card t1712): the parent prefix goes through the
// physical walk — never the EvalSymlinks fast path, whose ".." handling is
// lexical — and the final name lands on the resolved parent.
//
// REPAIR ROUND 14 — EXACT-MATCH-ONLY narrowing, the generalization of the
// round-8 windows answer to every aliasing mechanism at once (fold pairs,
// NFC/NFD, 8.3 short names, locale case tables): the rejoined link-own
// judgment is COMPLETE only when the final component is byte-equal to an
// entry of the resolved parent's real listing — then the shell really
// removes that entry and nothing behind it, and `exact` is true. EVERYTHING
// else — no exact entry, a LISTING FAILURE (a readable=no writable=yes
// parent must never keep a raw judgment), any listing error — reports
// exact=false and the CALLER defers the operand to the FULL resolver
// (resolveZoneTarget), whose EvalSymlinks canonicalization judges the real
// target; the rounds-11..13 fold/normalization canonicalization machinery
// is deleted. The rejoined path is still returned when exact=false because
// the full resolver's own parent-resolution arm shares this helper (it
// ignores the flag). A "." or ".." final segment is not a name the shell
// unlinks as itself; the caller routes those spellings to the
// target-following judgment and this function refuses them.
//
// The caller threads a per-guard-call LISTING CACHE (keyed by the resolved
// parent path; nil disables caching) so N candidates sharing one parent read
// the listing once (repair round 12 P2) — never carried across guard calls.
func zoneResolveParent(p string, listings map[string]*zoneNameIndex) (string, bool, bool) {
	dir, final := path.Split(filepath.ToSlash(p))
	if dir == "" || final == "" || final == "." || final == ".." {
		return "", false, false
	}
	windows := runtime.GOOS == "windows"
	resolved, ok := zoneResolveWalk(filepath.FromSlash(dir), 0, windows)
	if !ok {
		return "", false, false
	}
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
	exact := idx != nil && idx.has(final)
	return filepath.Join(resolved, filepath.FromSlash(final)), exact, true
}

// zoneNameIndex is the per-parent O(1) EXACT-membership index over a
// directory listing (repair round 13 P2, narrowed by repair round 14): the
// link-own narrowing applies only when the operand's final component is
// byte-equal to an entry of the resolved parent — the index answers that
// membership in O(1), built once per resolved parent on first listing in
// the same zoneParentLister seam as the listing cache. Every non-exact
// spelling defers to the full resolver; no fold, normalization, or
// short-name machinery exists here anymore (rounds 11-13 deleted).
type zoneNameIndex struct {
	exact map[string]bool
}

// newZoneNameIndex builds the exact-membership set over one parent's entry
// names.
func newZoneNameIndex(names []string) *zoneNameIndex {
	idx := &zoneNameIndex{exact: make(map[string]bool, len(names))}
	for _, name := range names {
		idx.exact[name] = true
	}
	return idx
}

// has reports whether name is byte-equal to a real directory entry.
func (idx *zoneNameIndex) has(name string) bool {
	return idx.exact[name]
}

// alias returns the real entry that name aliases under NFC plus Unicode
// simple case folding (Ä/ä), the relation a case-insensitive volume applies.
func (idx *zoneNameIndex) alias(name string) (string, bool) {
	want := norm.NFC.String(name)
	// an entry whose NFC form equals the operand's is the one the OS names
	// (a normalization-only alias); case folding is the fallback, so a
	// case-sensitive volume holding both Ä and ä never answers by map order
	for real := range idx.exact {
		if norm.NFC.String(real) == want {
			return real, true
		}
	}
	for real := range idx.exact {
		if strings.EqualFold(norm.NFC.String(real), want) {
			return real, true
		}
	}
	return "", false
}

// aliasByIdentity returns the listing entry that is the same file as the one
// the OS names through dir/name (card t1804 F2). Go's simple folding is
// narrower than a volume's own table (ß folds to SS there), so a spelling the
// string tiers of alias miss is still decided by os.SameFile, never by a fold.
// Entries are scanned in sorted order so a hard-linked file answers the same
// way on every call.
func (idx *zoneNameIndex) aliasByIdentity(dir, name string) (string, bool) {
	named, err := os.Lstat(filepath.Join(dir, name))
	if err != nil {
		return "", false
	}
	entries := make([]string, 0, len(idx.exact))
	for real := range idx.exact {
		entries = append(entries, real)
	}
	sort.Strings(entries)
	for _, real := range entries {
		if fi, err := os.Lstat(filepath.Join(dir, real)); err == nil && os.SameFile(named, fi) {
			return real, true
		}
	}
	return "", false
}

// zoneCanonicalRel rewrites the components of a relative path (card t1804) to
// the spelling the directory really lists, wherever the operand's spelling and
// the listed one differ in non-ASCII text (ASCII-only case differences are the
// manifest fold's own business). The input is either the output of the physical
// walk or the operand's own cleaned lexical rel; a ".." ends the walk and is
// never popped as text. The manifest fold is ASCII-only, so a parent spelled
// as a non-ASCII case variant (`ä` for the protected `Ä`) reaches the zone on
// a case-insensitive volume while every other form keeps the operand's
// spelling and misses the entry. The entry is chosen by the NFC-equal name,
// then by simple case folding, then by file identity (os.SameFile) where the
// volume's folding is wider than Go's (ß/SS). A component is substituted only
// when the OS demonstrably names an existing entry through it (Lstat succeeds
// while the listing lacks the exact spelling), so a case-sensitive volume or a
// nonexistent spelling stays untouched. The caller ADDS the result next to the
// original form; it never replaces a form already found.
func zoneCanonicalRel(realRoot, rel string, listings map[string]*zoneNameIndex) string {
	parts := strings.Split(rel, "/")
	dir := filepath.FromSlash(realRoot)
	changed := false
	for i, part := range parts {
		if part == ".." {
			break
		}
		if part == "" || part == "." {
			dir = filepath.Join(dir, part)
			continue
		}
		var idx *zoneNameIndex
		if cached, hit := listings[dir]; hit {
			idx = cached
		} else if entries, err := zoneParentLister(dir); err == nil {
			names := make([]string, 0, len(entries))
			for _, e := range entries {
				names = append(names, e.Name())
			}
			idx = newZoneNameIndex(names)
			if listings != nil {
				listings[dir] = idx
			}
		}
		if idx == nil {
			if isASCII(part) {
				// an unreadable ancestor: later components may still be named
				dir = filepath.Join(dir, part)
				continue
			}
			break
		}
		if !idx.has(part) {
			if _, err := os.Lstat(filepath.Join(dir, part)); err != nil {
				break
			}
			real, ok := idx.alias(part)
			if !ok {
				real, ok = idx.aliasByIdentity(dir, part)
			}
			if !ok {
				break
			}
			// an ASCII spelling of an ASCII entry is the manifest fold's own
			// business (FoldZoneText lowers ASCII); only a spelling that crosses
			// into or out of non-ASCII text needs the listing's name
			if !isASCII(part) || !isASCII(real) {
				parts[i] = real
				changed = true
				part = real
			}
		}
		dir = filepath.Join(dir, part)
	}
	if !changed {
		return rel
	}
	return strings.Join(parts, "/")
}

// zoneLexicalCanon canonicalises the operand's own lexical rel against the
// real project root (card t1804 F1). The lexical form is never followed through
// a final symlink, so when that link leaves the zone the spelling the operand
// typed is the only form the zone can still be matched on. The caller ADDS the
// result next to rel.
func zoneLexicalCanon(root, rel string, listings map[string]*zoneNameIndex) string {
	nativeRoot := filepath.FromSlash(zoneNativeSlash(root))
	if root == "" || !filepath.IsAbs(nativeRoot) {
		return rel
	}
	realRoot, ok := zoneResolve(nativeRoot)
	if !ok {
		realRoot = nativeRoot
	}
	return zoneCanonicalRel(filepath.ToSlash(realRoot), rel, listings)
}

// isASCII reports whether s holds only ASCII bytes.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
