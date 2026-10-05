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
)

// RootSlug names one of the four user-folder roots. The slugs are fixed so
// the backup home `~/.moai/backups/<root-slug>/<relpath>` (iter4 D33) cannot
// collide two roots on the same relpath.
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

// ResolveRoots returns the four install roots under the given home.
func ResolveRoots(home string) []Root {
	return []Root{
		{Slug: RootClaudeSkills, Dir: filepath.Join(home, ".claude", "skills")},
		{Slug: RootClaudeAgents, Dir: filepath.Join(home, ".claude", "agents")},
		{Slug: RootAgentsSkills, Dir: filepath.Join(home, ".agents", "skills")},
		{Slug: RootCodexAgents, Dir: filepath.Join(home, ".codex", "agents")},
	}
}

// RootBySlug returns the root with the given slug.
func RootBySlug(home string, slug RootSlug) (Root, bool) {
	for _, r := range ResolveRoots(home) {
		if r.Slug == slug {
			return r, true
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
