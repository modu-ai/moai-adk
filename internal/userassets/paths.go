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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// reservedProfileDirName is the sentinel directory name that is never a real
// profile (card t1702): tooling creates it as a placeholder for "a profile
// that does not exist", so a directory bearing it under the profile base is
// junk by definition.
const reservedProfileDirName = "__no_such_profile__"

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
//
// @MX:ANCHOR: [AUTO] profile-store base resolver — the single profile-base seam
// @MX:REASON: three consumers (ResolveRoots, RealProfileDirNames,
// ClaudeConfigRoots) derive every profile-root path from this one join;
// letting any of them re-derive the base independently would let a renamed
// profile store silently fork the enumeration.
func ProfileRootsBase(home string) string {
	return filepath.Join(MoaiHome(home), "claude-profiles")
}

// ValidProfileDirName reports whether name can name a real profile directory
// (card t1702): traversal-safe (no separators, no leading dot, not absolute),
// whitespace-free, and not the reserved sentinel. Profile names carrying
// spaces or sentinel shapes are junk a directory sweep can leave behind and
// are never enumerated as install targets.
func ValidProfileDirName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.HasPrefix(name, ".") {
		return false
	}
	if filepath.IsAbs(name) {
		return false
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return false
	}
	return name != reservedProfileDirName
}

// RealProfileDirNames returns the sorted names of the real profile
// directories under the profile base. A missing base is the legitimate
// empty store and yields (nil, nil); any OTHER ReadDir error (permissions,
// I/O) yields a wrapped error naming the store path — swallowing it made
// Install and prune finish "successfully" while nothing landed in the
// actual profiles (card t1702 audit round 4). Entries that are not
// directories or whose name fails ValidProfileDirName are ignored.
//
// @MX:ANCHOR: [AUTO] real-profile enumeration — the junk-filter SSOT
// @MX:REASON: the install roots, the settings cleanup, and the doctor's
// active-root summary all enumerate through this one predicate; a looser
// copy elsewhere would start installing into (or cleaning) junk directories.
func RealProfileDirNames(home string) ([]string, error) {
	base := ProfileRootsBase(home)
	dirEntries, err := os.ReadDir(base)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("userassets: read profile store %s: %w", base, err)
	}
	var names []string
	for _, e := range dirEntries {
		if !e.IsDir() || !ValidProfileDirName(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
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
// resolve regardless of the profile store's state.
func fixedRoots(home string) []Root {
	return []Root{
		{Slug: RootClaudeSkills, Dir: filepath.Join(home, ".claude", "skills")},
		{Slug: RootClaudeAgents, Dir: filepath.Join(home, ".claude", "agents")},
		{Slug: RootAgentsSkills, Dir: filepath.Join(home, ".agents", "skills")},
		{Slug: RootCodexAgents, Dir: filepath.Join(home, ".codex", "agents")},
	}
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
	base := ProfileRootsBase(home)
	names, err := RealProfileDirNames(home)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		pdir := filepath.Join(base, name)
		roots = append(roots,
			Root{Slug: RootSlug(string(RootClaudeSkills) + "@" + name), Dir: filepath.Join(pdir, "skills")},
			Root{Slug: RootSlug(string(RootClaudeAgents) + "@" + name), Dir: filepath.Join(pdir, "agents")},
		)
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

// RootBySlug returns the root with the given slug. The four fixed roots
// resolve without touching the profile store; a profile-derived slug
// resolves only through a successful enumeration — an unreadable store
// reports the slug as not-found, which every caller treats as the
// conservative refusal (the cli migration gate keeps the project file),
// never as a license to re-derive the directory independently (card t1702
// audit round 4).
func RootBySlug(home string, slug RootSlug) (Root, bool) {
	for _, r := range fixedRoots(home) {
		if r.Slug == slug {
			return r, true
		}
	}
	if _, _, found := strings.Cut(string(slug), "@"); !found {
		return Root{}, false
	}
	names, err := RealProfileDirNames(home)
	if err != nil {
		return Root{}, false
	}
	base := ProfileRootsBase(home)
	for _, name := range names {
		switch slug {
		case RootSlug(string(RootClaudeSkills) + "@" + name):
			return Root{Slug: slug, Dir: filepath.Join(base, name, "skills")}, true
		case RootSlug(string(RootClaudeAgents) + "@" + name):
			return Root{Slug: slug, Dir: filepath.Join(base, name, "agents")}, true
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
