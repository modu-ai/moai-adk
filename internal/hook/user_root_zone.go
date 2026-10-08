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
	// Gate round 26 #1: BOTH sides are resolved — the target through
	// zoneResolve (symlinks, physical ".."), and the home through the same
	// walk. On macOS /var is a symlink to /private/var: an unresolved home
	// would never prefix-match a resolved target and every spelling would
	// bypass the match.
	homeResolved, ok := zoneResolve(home)
	if !ok {
		homeResolved = home
	}
	homeSlash := filepath.ToSlash(homeResolved)
	abs := cand
	switch {
	case strings.HasPrefix(cand, "~/"):
		abs = homeSlash + cand[1:]
	case strings.HasPrefix(cand, "$HOME/"):
		abs = homeSlash + cand[len("$HOME"):]
	case strings.HasPrefix(cand, "${HOME}/"):
		abs = homeSlash + cand[len("${HOME}"):]
	}
	if !zoneIsAbs(zoneSlash(abs)) {
		if cwd, cwdErr := zoneGetwd(); cwdErr == nil && cwd != "" {
			abs = zoneSlash(cwd) + "/" + abs
		}
	}
	if !zoneIsAbs(zoneSlash(abs)) {
		return nil
	}
	resolved, ok := zoneResolve(filepath.FromSlash(zoneSlash(abs)))
	if !ok {
		// Unresolvable — the tracking evidence cannot be judged on the
		// real path, so no user-root match is formed (the over-protection
		// guard's fail direction is allow).
		return nil
	}
	slash := filepath.ToSlash(resolved)
	var out []zoneForm
	for slug, dir := range userRootSlugDirs {
		under := homeSlash + "/" + dir + "/"
		rest, ok := strings.CutPrefix(slash, under)
		if !ok || rest == "" {
			continue
		}
		display := userRootFormPrefix + slug + "/" + rest
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
// closes that escape). An unreadable manifest reads as untracked (the
// over-protection guard's fail direction).
func userRootTracksAny(home, display string) bool {
	key, ok := userRootKey(display)
	if !ok {
		return false
	}
	if home == "" {
		if h, err := zoneHomeFn(); err == nil {
			home = h
		}
	}
	if home == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(home, ".moai", "user-assets.json"))
	if err != nil {
		return false
	}
	var m struct {
		Files map[string]json.RawMessage `json:"files"`
	}
	if json.Unmarshal(data, &m) != nil {
		return false
	}
	if _, ok := m.Files[key]; ok {
		return true
	}
	prefix := key + "/"
	for k := range m.Files {
		if strings.HasPrefix(k, prefix) {
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
