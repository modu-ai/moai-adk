package hook

// user_root_zone.go — the user-root half of the zone guard (SPEC-
// USERASSET-DEPLOY-GUARD-001 M4, REQ-GRD-001/002, design §5).
//
// The config list may declare entries of the kind "user-root:<slug>/…"
// naming files inside one of the four USER-INSTALL roots; the zone
// resolver emits the matching namespaced forms for targets under the user
// home, and the deny path further limits protection to files the user
// manifest TRACKS — a user-created file in a managed directory is not a
// managed asset and stays editable (the over-protection guard).
//
// Import direction: hook does not import userassets. The slug→directory
// table below is pinned to userassets.ResolveRoots by a parity test
// (test-only import — the production dependency direction stays
// hook ← config only), so a drift turns the parity test red.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// userRootSlugDirs maps the user-install root slugs to their home-relative
// directories. MUST agree with userassets.ResolveRoots — the parity test
// pins it.
var userRootSlugDirs = map[string]string{
	"claude-skills": ".claude/skills",
	"claude-agents": ".claude/agents",
	"agents-skills": ".agents/skills",
	"codex-agents":  ".codex/agents",
}

// userRootFormPrefix namespaces a user-root target form.
const userRootFormPrefix = "user-root:"

// zoneHomeFn resolves the user's home directory for the user-root match.
// Tests replace it.
var zoneHomeFn = os.UserHomeDir

// userRootZoneForms is the deny sites' user-root entry point (gate round
// 26 #1/#3): it judges the RESOLVED target — the candidate is absolutized
// against the hook cwd, ~/$HOME spellings are expanded, and the whole path
// goes through zoneResolve (symlinks and physical ".." follow the same
// walk the shell would) — so /./, /../ and symlink spellings of a tracked
// file all land on the same resolved inode path and produce the same
// namespaced form. The forms match ONLY ZoneUserRoot entries; the
// project/baseline streams never see them.
func userRootZoneForms(cand string) []zoneForm {
	home, err := zoneHomeFn()
	if err != nil || home == "" {
		return nil
	}
	abs := cand
	switch {
	case strings.HasPrefix(cand, "~/"):
		abs = home + cand[1:]
	case strings.HasPrefix(cand, "$HOME/"):
		abs = home + cand[len("$HOME"):]
	case strings.HasPrefix(cand, "${HOME}/"):
		abs = home + cand[len("${HOME}"):]
	}
	// Gate round 40-1: the ABSOLUTENESS judgment runs on the RAW form —
	// zoneSlash would make a leading `\alias` read as absolute on POSIX,
	// where a backslash is an ordinary filename character (the cwd-join
	// would be skipped and the resolution would go wrong).
	if !filepath.IsAbs(abs) && !strings.HasPrefix(abs, "/") {
		if cwd, cwdErr := zoneGetwd(); cwdErr == nil && cwd != "" {
			abs = cwd + "/" + abs
		}
	}
	if !filepath.IsAbs(abs) && !strings.HasPrefix(abs, "/") {
		return nil
	}
	// Gate round 27-4: EACH INSTALL ROOT is resolved through the same walk
	// as the target — a dotfile-manager symlink on ~/.claude relocates the
	// root, and only root-resolved prefixes can match a target-resolved
	// path. Gate round 34-2: the prefix comparison itself is CASE-FOLDED —
	// on a case-insensitive filesystem .CLAUDE names the same directory as
	// .claude, and a case-alias target must not bypass the match.
	// Gate round 39-NEW-1: the target's UNRESOLVED spelling is emitted as
	// an ALIAS form beside the resolved one — a symlinked installed file
	// (moai-demo -> real-demo) is manifest-tracked under its LINK path,
	// while the resolved path reads unregistered; both forms are judged.
	var out []zoneForm
	addForm := func(display string) {
		f := zoneForm{Display: display, Folded: config.FoldZoneText(display)}
		for _, have := range out {
			if have.Folded == f.Folded {
				return
			}
		}
		out = append(out, f)
	}
	unresolvedSlash := filepath.ToSlash(abs)
	for slug, dir := range userRootSlugDirs {
		resolvedRoot, ok := zoneResolve(filepath.Join(home, filepath.FromSlash(dir)))
		if !ok {
			continue // root absent — nothing of it is on disk to protect
		}
		rootSlash := config.FoldZoneText(filepath.ToSlash(resolvedRoot))
		// Gate round 44-1: the dual forms are PAIRED BY BASIS — the
		// unresolved candidate is compared against the UNRESOLVED root
		// (home + dir, the link-path form) and the resolved candidate
		// against the resolved root. Mixing (unresolved candidate vs
		// resolved root) loses the tracked key when the root ITSELF is a
		// symlink: the link-spelled candidate never prefix-matches the
		// resolved root, the link-path alias form is never emitted, and a
		// manifest key recorded under the link path (demo/SKILL.md) reads
		// as unregistered real-demo/SKILL.md.
		unresolvedRoot := filepath.ToSlash(filepath.Join(home, filepath.FromSlash(dir)))
		// the UNRESOLVED candidate spelling (the link-path form). A `.`
		// segment has no real-path effect and is dropped TEXTUALLY so the
		// alias still interprets the link-path key (gate round 47-2). A
		// `..` segment is NEVER collapsed textually (gate round 48-3): its
		// meaning depends on which components are symlinks — collapsing it
		// would deny healthy edits through a link+`..` spelling that
		// actually lands on an untracked file — so the RESOLVED form (the
		// filesystem walk) carries that judgment.
		if rest, ok := strings.CutPrefix(config.FoldZoneText(unresolvedSlash), config.FoldZoneText(unresolvedRoot)+"/"); ok && !hasDotDotSegment(rest) {
			if alias := dropDotSegments(rest); alias != "" {
				addForm(userRootFormPrefix + slug + "/" + alias)
			}
		}
		// the RESOLVED target spelling (the real-path form)
		resolvedTarget, ok := zoneResolve(abs)
		if !ok {
			continue
		}
		if rest, ok := strings.CutPrefix(config.FoldZoneText(filepath.ToSlash(resolvedTarget)), rootSlash+"/"); ok {
			display := userRootFormPrefix + slug
			if rest != "" {
				display += "/" + rest
			}
			addForm(display)
		}
		// Gate round 27-2 + 34-1: the ROOT ITSELF and every ANCESTOR of a
		// root (rm -rf ~/.claude deletes every tracked file inside it) are
		// protected forms; the tracked containment decides.
		if targetSlash := config.FoldZoneText(filepath.ToSlash(resolvedTarget)); targetSlash == rootSlash || strings.HasPrefix(rootSlash, targetSlash+"/") {
			addForm(userRootFormPrefix + slug)
		}
	}
	return out
}

// hasDotDotSegment reports whether a slash-separated rest carries a ".."
// segment — such a spelling keeps real-path semantics and is judged by the
// resolved form only (gate round 48-3).
func hasDotDotSegment(rest string) bool {
	for _, seg := range strings.Split(rest, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

// dropDotSegments removes "." (and empty) components textually — a dot has
// no real-path effect, so the alias keeps the link-path identity (gate
// round 47-2).
func dropDotSegments(rest string) string {
	segs := strings.Split(rest, "/")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		if seg == "." || seg == "" {
			continue
		}
		out = append(out, seg)
	}
	return strings.Join(out, "/")
}

// userRootKey parses a namespaced form into the manifest key
// "<slug>/<rest>".
func userRootKey(display string) (string, bool) {
	rest, ok := strings.CutPrefix(display, userRootFormPrefix)
	if !ok || rest == "" {
		return "", false
	}
	return rest, true
}

// userRootTracksAny reports whether the user manifest tracks the keyed
// file OR anything UNDER it (gate round 26 #2: the manifest records
// FILES, but a mutation can target a DIRECTORY — the prefix containment
// closes that escape). Gate round 27-3: BOTH sides fold before comparing —
// on a case-insensitive filesystem `skill.md` names the same inode as
// `SKILL.md`, and a case-alias write must not bypass the tracked judgment.
// An unreadable or non-regular manifest reads as untracked (the
// over-protection guard's fail direction; gate round 30: a FIFO at the
// manifest path must not hang the read).
func userRootTracksAny(home, display string) bool {
	key, ok := userRootKey(display)
	if !ok {
		return false
	}
	foldedKey := config.FoldZoneText(strings.TrimSuffix(key, "/"))
	if home == "" {
		if h, err := zoneHomeFn(); err == nil {
			home = h
		}
	}
	if home == "" {
		return false
	}
	data, ok := readRegularFile(filepath.Join(home, ".moai", "user-assets.json"))
	if !ok {
		return false
	}
	var m struct {
		Files map[string]json.RawMessage `json:"files"`
	}
	if json.Unmarshal(data, &m) != nil {
		return false
	}
	for k := range m.Files {
		folded := config.FoldZoneText(strings.TrimSuffix(k, "/"))
		if folded == foldedKey || strings.HasPrefix(folded, foldedKey+"/") {
			return true
		}
	}
	// Gate round 50-1: the dual-form mapping runs BOTH directions. The
	// manifest may track the LINK path while the candidate arrives spelled
	// as the REAL path (or the reverse) — the folded-key comparison alone
	// then reads a tracked file unregistered and lets the write through.
	// Resolve the candidate's real location and compare it against every
	// tracked key's real location (the filesystem's own mapping).
	slug, rest, cutOK := strings.Cut(strings.TrimSuffix(key, "/"), "/")
	dir, dirOK := userRootSlugDirs[slug]
	if !cutOK || !dirOK {
		return false
	}
	candReal, resOK := zoneResolve(filepath.Join(home, filepath.FromSlash(dir), filepath.FromSlash(rest)))
	if !resOK {
		return false
	}
	candRealFolded := config.FoldZoneText(filepath.ToSlash(candReal))
	for k := range m.Files {
		kslug, krest, ok := strings.Cut(strings.TrimSuffix(k, "/"), "/")
		kdir, kdirOK := userRootSlugDirs[kslug]
		if !ok || !kdirOK {
			continue
		}
		kReal, kresOK := zoneResolve(filepath.Join(home, filepath.FromSlash(kdir), filepath.FromSlash(krest)))
		if !kresOK {
			continue
		}
		if candRealFolded == config.FoldZoneText(filepath.ToSlash(kReal)) {
			return true
		}
		// Gate round 51-1: the containment check applies on the real path
		// too — a candidate DIRECTORY whose real location contains a
		// tracked key's real file (rm -rf real-demo, where the tracked
		// key's link lives under real-demo) is as destructive as a direct
		// file delete, and equality alone let it through.
		kRealFolded := config.FoldZoneText(filepath.ToSlash(kReal))
		if strings.HasPrefix(kRealFolded+"/", candRealFolded+"/") {
			return true
		}
	}
	return false
}

// userManifestRel is the home-relative path of the user-assets manifest —
// the ROOT of the whole protection model (gate rounds 37-1/38-2): deleting
// or emptying it turns every previously-managed file unregistered, so the
// frozen identity's destructive changes to the file are denied outright
// (no tracked-ness filter — the manifest needs no tracking evidence).
var userManifestRel = filepath.ToSlash(filepath.Join(".moai", "user-assets.json"))

// userMoaiDirRel is the manifest's containing directory (home-relative).
var userMoaiDirRel = ".moai"

// userManifestProtectForms returns the protected forms for a candidate
// that targets the user manifest itself or its containing directory —
// judged independently of the four install roots (which never contain
// ~/.moai, so the user-root forms alone cannot see it).
func userManifestProtectForms(cand string) []zoneForm {
	home, err := zoneHomeFn()
	if err != nil || home == "" {
		return nil
	}
	abs := cand
	switch {
	case strings.HasPrefix(cand, "~/"):
		abs = home + cand[1:]
	case strings.HasPrefix(cand, "$HOME/"):
		abs = home + cand[len("$HOME"):]
	case strings.HasPrefix(cand, "${HOME}/"):
		abs = home + cand[len("${HOME}"):]
	}
	if !filepath.IsAbs(abs) && !strings.HasPrefix(abs, "/") {
		if cwd, cwdErr := zoneGetwd(); cwdErr == nil && cwd != "" {
			abs = cwd + "/" + abs
		}
	}
	if !filepath.IsAbs(abs) && !strings.HasPrefix(abs, "/") {
		return nil
	}
	resolved, ok := zoneResolve(abs)
	if !ok {
		return nil
	}
	// the home is resolved through the same walk (gate 27-4 — a symlinked
	// home would otherwise never prefix-match a resolved target)
	resolvedHome, ok := zoneResolve(home)
	if !ok {
		resolvedHome = home
	}
	// Gate round 43-2: the protection is the manifest ITSELF and its
	// containing directory (whose deletion takes the manifest with it) —
	// nothing wider. The former ~/.moai/** prefix arm over-protected (an
	// untracked user note under ~/.moai was denied a deletion that the
	// normal tracked-ness judgment should decide).
	//
	// Gate rounds 46-1/46-2/47-1: the protection basis matches the
	// candidate standard BOTH ways. The rests are CASE-FOLDED before
	// comparison — a .MOAI/USER-ASSETS.JSON alias names the same inode on
	// a case-insensitive filesystem, and the resolver returns the caller's
	// spelling, so the fold is the only normalization that holds. And the
	// candidate is judged in BOTH scopes: the resolved real path AND the
	// unresolved (link-path) spelling — a symlinked ~/.moai relocates the
	// resolved manifest outside the home prefix, and only the link-path
	// scope still sees it.
	formsFor := func(rest string) []zoneForm {
		rest = strings.TrimSuffix(rest, "/")
		// Gate round 48-1: the symlink+dot combination (~/.moai/./user-
		// assets.json) drops the dot TEXTUALLY so the link-path rest still
		// matches. A `..` keeps real-path semantics (gate round 48-3) and
		// is left to the resolved branch / the direct real-path equality
		// below.
		if hasDotDotSegment(rest) {
			return nil
		}
		rest = dropDotSegments(rest)
		folded := config.FoldZoneText(rest)
		if folded == config.FoldZoneText(userManifestRel) || folded == config.FoldZoneText(userMoaiDirRel) {
			return []zoneForm{{Display: folded, Folded: folded}}
		}
		return nil
	}
	if rest, ok := strings.CutPrefix(filepath.ToSlash(abs), filepath.ToSlash(home)+"/"); ok {
		if f := formsFor(rest); len(f) > 0 {
			return f
		}
	}
	if rest, ok := strings.CutPrefix(filepath.ToSlash(resolved), filepath.ToSlash(resolvedHome)+"/"); ok {
		if f := formsFor(rest); len(f) > 0 {
			return f
		}
	}
	// Gate round 48-1 + 50-2: direct real-path judgment — the filesystem's
	// own answer for the symlink+`..` combinations the home-relative rests
	// cannot name. The containment walk covers the resolved manifest's
	// FULL ancestor chain: the manifest is the protection model's root, so
	// anything that deletes it at ANY ancestor level (rm -rf outside,
	// where ~/.moai links into outside/) is protected — not just the file
	// and its direct parent.
	resolvedFolded := config.FoldZoneText(filepath.ToSlash(resolved))
	if resolvedManifest, ok := zoneResolve(filepath.Join(home, ".moai", "user-assets.json")); ok {
		manFolded := config.FoldZoneText(filepath.ToSlash(resolvedManifest))
		if resolvedFolded == manFolded || strings.HasPrefix(manFolded, resolvedFolded+"/") {
			f := config.FoldZoneText(userManifestRel)
			return []zoneForm{{Display: f, Folded: f}}
		}
	}
	return nil
}

// zoneUserRootCovered judges the user-root arm for one shell candidate:
// the manifest-protection forms FIRST (gate rounds 37-1/38-2 — no
// tracked-ness filter: the manifest IS the evidence), then the resolved
// target's namespaced forms matched against ZoneUserRoot entries filtered
// by the manifest-tracked containment. Returns the entry's category, the
// namespaced display form, and whether covered.
func zoneUserRootCovered(cand string, load config.ProtectedZoneLoad) (string, string, bool) {
	if load.State == config.ZoneStateOK {
		if forms := userManifestProtectForms(cand); len(forms) > 0 {
			for i := range load.Zone.Entries {
				e := &load.Zone.Entries[i]
				if e.Kind == config.ZoneUserRoot {
					return e.Category, forms[0].Display, true
				}
			}
			// No ZoneUserRoot entry declared: the manifest protection is
			// compiled-in for the frozen identity regardless (gate 37-1 —
			// the bypass chain does not depend on the overlay declaring
			// the manifest).
			return "user_manifest", forms[0].Display, true
		}
	}
	for _, uf := range userRootZoneForms(cand) {
		for i := range load.Zone.Entries {
			e := &load.Zone.Entries[i]
			if e.Kind != config.ZoneUserRoot || !e.Match(uf.Folded) {
				continue
			}
			if !userRootTracksAny("", uf.Display) {
				continue
			}
			return e.Category, uf.Display, true
		}
	}
	return "", "", false
}
