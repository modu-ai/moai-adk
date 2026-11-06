// userroots.go — the shared user-root derivation (card t1702 audit iter4
// F2): the ONE source of truth for the user-install slug→directory rule
// and the profile-store junk filter. userassets (the install roots) and
// hook (the user-root zone guard) both derive from these — neither keeps
// its own table, so the REQ-GRD-002 parity cannot drift again. hook's
// import-direction constraint (production hook ← config only) is why this
// lives here and not in userassets.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// UserRootDir is one fixed user-install root: its slug and its
// home-relative directory (slash form).
type UserRootDir struct {
	Slug string
	Rel  string
}

// userRootFixedDirs is the ordered fixed-root table — the single
// definition userassets' ResolveRoots and the hook's zone lookup both
// derive from.
//
// @MX:ANCHOR: [AUTO] user-root fixed table SSOT — userassets and the hook zone guard both derive their root set from this one list
// @MX:REASON: the hook may not import userassets (import direction hook ←
// config only), so before this table existed both sides kept their own
// copy and the profile roots drifted out of the hook's protection zone
// (card t1702 audit iter4 F2); a second hand-maintained table anywhere
// reopens exactly that drift.
var userRootFixedDirs = []UserRootDir{
	{Slug: "claude-skills", Rel: ".claude/skills"},
	{Slug: "claude-agents", Rel: ".claude/agents"},
	{Slug: "agents-skills", Rel: ".agents/skills"},
	{Slug: "codex-agents", Rel: ".codex/agents"},
}

// UserRootFixedDirs returns the four fixed user-install roots in a stable
// order.
func UserRootFixedDirs() []UserRootDir {
	out := make([]UserRootDir, len(userRootFixedDirs))
	copy(out, userRootFixedDirs)
	return out
}

// UserProfileStoreRel is the home-relative profile store base — the dir
// whose real subdirectories are Claude profile config roots (card t1702).
const UserProfileStoreRel = ".moai/claude-profiles"

// reservedProfileDirName is the sentinel directory name that is never a
// real profile (card t1702): tooling creates it as a placeholder for "a
// profile that does not exist", so a directory bearing it under the
// profile base is junk by definition.
const reservedProfileDirName = "__no_such_profile__"

// ValidProfileDirName reports whether name can name a real profile
// directory (card t1702): traversal-safe (no separators, no leading dot,
// not absolute), whitespace-free, and not the reserved sentinel. Profile
// names carrying spaces or sentinel shapes are junk a directory sweep can
// leave behind and are never enumerated as install targets.
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

// LauncherProfileDirName reports whether the profile launcher accepts name as
// a profile directory name. It mirrors internal/profile.isValidProfileName
// (no separator, not absolute, no leading dot) and is deliberately WIDER than
// ValidProfileDirName: a name with a space, or the reserved sentinel, is
// junk for install targets yet still startable. A read-only judgment that
// must not miss a startable profile (card t1702 audit iter6 B1) uses this to
// notice entries the install enumeration drops. internal/profile cannot be
// imported from here, so a test in that package pins the two rules together.
func LauncherProfileDirName(name string) bool {
	if name == "" {
		return false
	}
	return !(strings.Contains(name, "/") || strings.Contains(name, "\\") ||
		strings.HasPrefix(name, ".") || filepath.IsAbs(name))
}

// UserProfileDirNames returns the sorted names of the real profile
// directories under home's profile store. A missing base is the legitimate
// empty store and yields (nil, nil); any OTHER ReadDir error (permissions,
// I/O) yields a wrapped error naming the store path (card t1702 audit
// round 4) — swallowing it made Install and prune finish "successfully"
// while nothing landed in the actual profiles.
//
// @MX:ANCHOR: [AUTO] real-profile enumeration SSOT — the junk-filtered store listing both userassets and the hook derive from
// @MX:REASON: install roots, settings cleanup, the doctor's active-root
// summary, and the hook's zone derivation all enumerate through this one
// listing and its one junk filter; a looser copy anywhere would start
// installing into (or protecting) junk directories.
func UserProfileDirNames(home string) ([]string, error) {
	base := filepath.Join(home, filepath.FromSlash(UserProfileStoreRel))
	dirEntries, err := os.ReadDir(base)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read profile store %s: %w", base, err)
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

// UserRootProfileBase names one base slug whose root exists PER PROFILE:
// a profile-launched session runs with CLAUDE_CONFIG_DIR pointed at its
// profile dir, so the two Claude-side roots live inside each profile. The
// .agents/.codex roots stay home-level (only CLAUDE_CONFIG_DIR is
// per-profile).
type UserRootProfileBase struct {
	Slug string
}

// userRootProfileBases is the ordered per-profile base list — the single
// definition of which fixed roots are profile-derivable.
var userRootProfileBases = []UserRootProfileBase{
	{Slug: "claude-skills"},
	{Slug: "claude-agents"},
}

// UserRootProfileBases returns the per-profile base slugs in a stable
// order.
func UserRootProfileBases() []UserRootProfileBase {
	out := make([]UserRootProfileBase, len(userRootProfileBases))
	copy(out, userRootProfileBases)
	return out
}

// fixedUserRootRel looks a fixed slug up in the fixed-root table.
func fixedUserRootRel(slug string) (string, bool) {
	for _, r := range userRootFixedDirs {
		if slug == r.Slug {
			return r.Rel, true
		}
	}
	return "", false
}

// isUserRootProfileBase reports whether slug is one of the per-profile
// base slugs.
func isUserRootProfileBase(slug string) bool {
	for _, b := range userRootProfileBases {
		if slug == b.Slug {
			return true
		}
	}
	return false
}

// UserRootSlugRel derives the home-relative directory (slash form) for a
// user-root slug: a fixed slug maps to its fixed directory; a
// profile-derived slug `<base>@<profile>` (base claude-skills or
// claude-agents) maps to `<store>/<profile>/<subdir>` where subdir is the
// last segment of the base's fixed directory. The profile name must pass
// ValidProfileDirName; LIVE EXISTENCE is not judged here — callers that
// need it enumerate UserProfileDirNames (card t1702 audit iter4 F2).
func UserRootSlugRel(slug string) (string, bool) {
	if rel, ok := fixedUserRootRel(slug); ok {
		return rel, true
	}
	base, profile, found := strings.Cut(slug, "@")
	if !found || !isUserRootProfileBase(base) || !ValidProfileDirName(profile) {
		return "", false
	}
	rel, ok := fixedUserRootRel(base)
	if !ok {
		return "", false
	}
	return path.Join(UserProfileStoreRel, profile, path.Base(rel)), true
}
