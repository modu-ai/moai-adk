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
	// Gate round 29-2: the RAW candidate spelling is preserved for
	// resolution — zoneSlash would convert a literal backslash to a
	// separator on POSIX, where a backslash is an ordinary filename
	// character (an `alias\dir` symlink would check a nonexistent path).
	// zoneResolve walks the raw path exactly like the interpreting process.
	if !zoneIsAbs(zoneSlash(abs)) {
		if cwd, cwdErr := zoneGetwd(); cwdErr == nil && cwd != "" {
			abs = cwd + "/" + abs
		}
	}
	if !zoneIsAbs(zoneSlash(abs)) {
		return nil
	}
	// Gate round 27-4: EACH INSTALL ROOT is resolved through the same walk
	// as the target — a dotfile-manager symlink on ~/.claude relocates the
	// root, and only root-resolved prefixes can match a target-resolved
	// path.
	var out []zoneForm
	for slug, dir := range userRootSlugDirs {
		resolvedRoot, ok := zoneResolve(filepath.Join(home, filepath.FromSlash(dir)))
		if !ok {
			continue // root absent — nothing of it is on disk to protect
		}
		resolvedTarget, ok := zoneResolve(abs)
		if !ok {
			return nil // unresolvable target: no evidence, fail to allow
		}
		under := filepath.ToSlash(resolvedRoot) + "/"
		rest, ok := strings.CutPrefix(filepath.ToSlash(resolvedTarget), under)
		if !ok {
			continue
		}
		// Gate round 27-2: the ROOT ITSELF (rest == "") is a protected form
		// — removing it removes every tracked descendant. The bare form
		// matches via the entry's dir semantics and the tracked containment.
		display := userRootFormPrefix + slug
		if rest != "" {
			display += "/" + rest
		}
		out = append(out, zoneForm{Display: display, Folded: config.FoldZoneText(display)})
	}
	return out
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
	return false
}

// zoneUserRootCovered judges the user-root arm for one shell candidate:
// the resolved target's namespaced forms are matched against ZoneUserRoot
// entries only, filtered by the manifest-tracked containment. Returns the
// entry's category, the namespaced display form, and whether covered.
func zoneUserRootCovered(cand string, load config.ProtectedZoneLoad) (string, string, bool) {
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
