// Package userassets implements the per-user common-asset install
// (SPEC-USER-ASSET-INSTALL-001): the four user-folder roots, the per-user
// manifest at ~/.moai/user-assets.json (REQ-006), the pending-install
// recovery journal (final-class item 5 + R-e), the user-level
// read-modify-write lock (round-5 F4), and the REQ-023 backup home.
//
// The package is path-seamed: every entry point takes the user's HOME (the
// tests pass t.TempDir()), never os.UserHomeDir directly — no test may write
// the developer's real HOME (plan §D).
package userassets

import (
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// RootSlug names a user-folder root: one of the four fixed slugs, or a
// profile-derived slug `<base>@<profile>` (card t1702). The slugs are unique
// per root so the backup home `~/.moai/backups/<root-slug>/<relpath>`
// (iter4 D33) cannot collide two roots on the same relpath.
type RootSlug string

const (
	RootClaudeSkills RootSlug = "claude-skills" // ~/.claude/skills
	RootClaudeAgents RootSlug = "claude-agents" // ~/.claude/agents
	RootAgentsSkills RootSlug = "agents-skills" // $HOME/.agents/skills
	RootCodexAgents  RootSlug = "codex-agents"  // ~/.codex/agents
)

// Root is one resolved user-folder install root.
type Root struct {
	Slug RootSlug
	Dir  string // absolute path of the user folder
}

// RootKind returns a root slug's base kind. The four fixed slugs are their
// own kind; a profile-derived slug `claude-skills@<profile>` has the kind of
// its base slug, so per-kind logic (which roots a skill lands in) treats a
// profile root exactly like its home-level counterpart.
func RootKind(slug RootSlug) RootSlug {
	if base, _, found := strings.Cut(string(slug), "@"); found {
		return RootSlug(base)
	}
	return slug
}

// ProfileRootsBase returns the profile store base dir (~/.moai/claude-profiles).
// The store path's single definition is config.UserProfileStoreRel — this
// join is the typed seam over it (card t1702 audit iter4 F2).
//
// @MX:ANCHOR: [AUTO] profile-store base resolver — the single profile-base seam
// @MX:REASON: three consumers (ResolveRoots, RealProfileDirNames,
// ClaudeConfigRoots) derive every profile-root path from this one join;
// letting any of them re-derive the base independently would let a renamed
// profile store silently fork the enumeration.
func ProfileRootsBase(home string) string {
	return filepath.Join(home, filepath.FromSlash(config.UserProfileStoreRel))
}

// ValidProfileDirName reports whether name can name a real profile
// directory (card t1702). The rule's single definition lives in config
// (the shared derivation — card t1702 audit iter4 F2) so the hook's zone
// guard applies the SAME junk filter; this wrapper keeps the userassets
// API.
func ValidProfileDirName(name string) bool {
	return config.ValidProfileDirName(name)
}

// RealProfileDirNames returns the sorted names of the real profile
// directories under the profile base: (nil, nil) when the base is missing
// (the legitimate empty store), a wrapped error naming the store path for
// any other read failure (card t1702 audit round 4). The enumeration and
// its junk filter live in config — the single derivation shared with the
// hook's zone guard (card t1702 audit iter4 F2).
func RealProfileDirNames(home string) ([]string, error) {
	return config.UserProfileDirNames(home)
}

// ClaudeConfigRoots returns every Claude config directory a session can use
// (card t1702 gaps 1-2): ~/.claude, plus each real profile directory — a
// profile-launched session runs with CLAUDE_CONFIG_DIR pointed at its
// profile dir, so the dir carries the same settings.json/hooks shape as
// ~/.claude. The result is deterministic: ~/.claude first, then profile dirs
// sorted by name. Junk directories are never returned and nothing is
// created. A profile store that exists but cannot be read yields an error —
// never the bare ~/.claude that "no profiles exist" would produce (card
// t1702 audit round 4).
func ClaudeConfigRoots(home string) ([]string, error) {
	roots := []string{filepath.Join(home, ".claude")}
	names, err := RealProfileDirNames(home)
	if err != nil {
		return nil, err
	}
	base := ProfileRootsBase(home)
	for _, name := range names {
		roots = append(roots, filepath.Join(base, name))
	}
	return roots, nil
}

// fixedRoots returns the four home-level install roots — the roots that
// resolve regardless of the profile store's state. Derived from the shared
// fixed-root table in config (card t1702 audit iter4 F2).
func fixedRoots(home string) []Root {
	var roots []Root
	for _, r := range config.UserRootFixedDirs() {
		roots = append(roots, Root{Slug: RootSlug(r.Slug), Dir: filepath.Join(home, filepath.FromSlash(r.Rel))})
	}
	return roots
}

// ResolveRoots returns the install roots under the given home: the four
// fixed roots, plus — for each real profile directory under
// ~/.moai/claude-profiles — the two Claude-side roots that a
// profile-launched session can see (card t1702 gap 1). Profile-derived
// slugs are `claude-skills@<profile>` / `claude-agents@<profile>` so the
// manifest, journal, and backup home stay collision-free; the home-level
// roots (.agents, .codex) stay shared across profiles because only
// CLAUDE_CONFIG_DIR is per-profile. A profile store that exists but cannot
// be read yields an error — never the four fixed roots that "no profiles
// exist" would produce (card t1702 audit round 4).
func ResolveRoots(home string) ([]Root, error) {
	roots := fixedRoots(home)
	names, err := RealProfileDirNames(home)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		// The slug→directory derivation is config.UserRootSlugRel — the one
		// rule the hook's zone guard applies too (card t1702 audit iter4 F2).
		for _, b := range config.UserRootProfileBases() {
			slug := RootSlug(b.Slug + "@" + name)
			rel, ok := config.UserRootSlugRel(string(slug))
			if !ok {
				continue // unreachable: name passed the same filter upstream
			}
			roots = append(roots, Root{Slug: slug, Dir: filepath.Join(home, filepath.FromSlash(rel))})
		}
	}
	return roots, nil
}

// RootBySlugDir returns the absolute directory for a root slug (the
// string-only form the cli migration gate uses).
func RootBySlugDir(home string, slug RootSlug) string {
	r, ok := RootBySlug(home, slug)
	if !ok {
		return ""
	}
	return r.Dir
}

// RootBySlug returns the root with the given slug. The slug→directory rule
// is config.UserRootSlugRel — the one shared derivation; a fixed slug
// resolves without touching the profile store, a profile-derived slug
// resolves only through a successful enumeration. An unreadable store
// reports the slug as not-found, which every caller treats as the
// conservative refusal (the cli migration gate keeps the project file),
// never as a license to re-derive the directory independently (card t1702
// audit rounds 4 and iter4 F2).
func RootBySlug(home string, slug RootSlug) (Root, bool) {
	rel, ok := config.UserRootSlugRel(string(slug))
	if !ok {
		return Root{}, false
	}
	dir := filepath.Join(home, filepath.FromSlash(rel))
	if _, _, found := strings.Cut(string(slug), "@"); !found {
		return Root{Slug: slug, Dir: dir}, true
	}
	names, err := RealProfileDirNames(home)
	if err != nil {
		return Root{}, false
	}
	_, profile, _ := strings.Cut(string(slug), "@")
	for _, n := range names {
		if n == profile {
			return Root{Slug: slug, Dir: dir}, true
		}
	}
	return Root{}, false
}

// MoaiHome returns the moai user-level state home (~/.moai).
func MoaiHome(home string) string {
	return filepath.Join(home, ".moai")
}

// ManifestPath returns ~/.moai/user-assets.json (decision-index D-Q2).
func ManifestPath(home string) string {
	return filepath.Join(MoaiHome(home), "user-assets.json")
}

// JournalPath returns the pending-install journal's path — a manifest-path
// sibling under ~/.moai/ (design §2.2).
func JournalPath(home string) string {
	return filepath.Join(MoaiHome(home), "user-assets-journal.json")
}

// LockPath returns the user-level manifest lock's path (round-5 F4).
func LockPath(home string) string {
	return filepath.Join(MoaiHome(home), "user-assets.lock")
}

// GuardMarkerPath returns the acquire-guard marker's path — the marker both
// platform guards create under the same derivation (strings.TrimSuffix of
// the lock path + ".acquire-guard", then the guard suffix). M2
// (REQ-LOCK-001) makes it the ownership record the doctor's visible-
// recovery row reports on.
func GuardMarkerPath(home string) string {
	return strings.TrimSuffix(LockPath(home), ".lock") + ".acquire-guard.guard"
}

// BackupHome returns the REQ-023 backup home root (C2's sole out-of-root
// write carve-out for user-folder asset writes, iter4 D33).
func BackupHome(home string) string {
	return filepath.Join(MoaiHome(home), "backups")
}

// BackupPath returns the backup destination for a root file: the
// root-slug-prefixed layout ~/.moai/backups/<root-slug>/<relpath>.
func BackupPath(home string, slug RootSlug, rel string) (string, error) {
	clean, err := ValidateRelPath(rel)
	if err != nil {
		return "", err
	}
	return filepath.Join(BackupHome(home), string(slug), filepath.FromSlash(clean)), nil
}

// ValidateRelPath cleans a root-relative path and rejects escapes: no
// absolute paths, no `..` segments anywhere in the cleaned form.
func ValidateRelPath(p string) (string, error) {
	if p == "" {
		return "", ErrPathInvalid
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`) || filepath.IsAbs(p) {
		return "", ErrPathInvalid
	}
	clean := filepath.Clean(p)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == "." {
		return "", ErrPathInvalid
	}
	return filepath.ToSlash(clean), nil
}
