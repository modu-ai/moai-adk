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

// userRootForms returns the user-root forms of a raw target: when the
// target (after ~ / $HOME expansion, or as an absolute path under the
// home) lands under one of the user-install roots, the form is the
// namespaced "user-root:<slug>/<rest>" spelling the ZoneUserRoot entries
// match. Relative targets are project-relative and produce nothing here.
func userRootForms(raw string) []zoneForm {
	home, err := zoneHomeFn()
	if err != nil || home == "" {
		return nil
	}
	homeSlash := filepath.ToSlash(home)
	expanded := raw
	switch {
	case strings.HasPrefix(raw, "~/"):
		expanded = homeSlash + raw[1:]
	case strings.HasPrefix(raw, "$HOME/"):
		expanded = homeSlash + raw[len("$HOME"):]
	case strings.HasPrefix(raw, "${HOME}/"):
		expanded = homeSlash + raw[len("${HOME}"):]
	}
	if expanded == "" || !filepath.IsAbs(expanded) {
		return nil
	}
	slash := filepath.ToSlash(expanded)
	var out []zoneForm
	for slug, dir := range userRootSlugDirs {
		rest, ok := cutUserRootRest(homeSlash, slash, dir)
		if !ok {
			continue
		}
		display := userRootFormPrefix + slug + "/" + rest
		f := zoneForm{Display: display, Folded: config.FoldZoneText(display)}
		out = append(out, f)
	}
	return out
}

// cutUserRootRest reports rest when the slash path lies under
// <home>/<homeRelDir>/ — the home is joined HERE: a raw absolute path
// already lives under the home, and an expanded ~ path was joined above.
func cutUserRootRest(homeSlash, slashPath, homeRelDir string) (string, bool) {
	under := homeSlash + "/" + homeRelDir + "/"
	rest, ok := strings.CutPrefix(slashPath, under)
	if !ok {
		return "", false
	}
	if rest == "" {
		return "", false
	}
	return rest, true
}

// userRootFormTracked reports whether the file a user-root form names is
// TRACKED by the user-assets manifest — the REQ-GRD-002 limitation that
// keeps the protection on moai-managed assets only. An unreadable manifest
// reads as untracked (fail toward allowing a user's own file, never
// toward protecting it): the recovery surface for a corrupt manifest is
// the doctor row, not a guard denial.
func userRootFormTracked(home string, display string) bool {
	rest, ok := strings.CutPrefix(display, userRootFormPrefix)
	if !ok {
		return false
	}
	key := config.FoldZoneText(rest)
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
	// The manifest keys are written in the canonical slug/rest spelling;
	// compare on both the folded form and the raw spelling so a canonical
	// manifest matches case-exactly while a case-variant form still finds
	// its record.
	if _, ok := m.Files[key]; ok {
		return true
	}
	_, ok = m.Files[rest]
	return ok
}
