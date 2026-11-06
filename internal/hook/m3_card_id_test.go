package hook

// m3_card_id_test.go — SPEC-WORKTREE-HOME-ROOT-001 M3 RED (card t1640):
// AC-WHR-008 (no misidentification at the new depth) and the hook half of
// AC-WHR-007 (the SessionStart card-id derivations).
//
// The tests are written before the readers learn the home-root form
// <MoaiHome>/worktrees/projects/<project-key>/<name>: today
// cardIDFromPath walks to the first ancestor whose parent is named
// "worktrees" and returns "projects" for the new depth, and
// cardIDFromRegisteredWorktree walks straight past the home tier because
// its parent-of-parent is not .moai/.claude. That failure is the RED
// (behavioral, per research §R13's M3 row — the tests name M1 and existing
// symbols only). The legacy shapes are pinned as characterization: they
// must keep passing through the M3 change.

import (
	"path/filepath"
	"testing"
)

// m3HomeRootDirs returns the two cwd spellings AC-WHR-008 names, under a
// synthetic home root (the readers are pure path walkers; MOAI_HOME is
// irrelevant to the RED and is carried by the GREEN resolver wiring).
func m3HomeRootDirs(home string) (deep, treeRoot string) {
	treeRoot = filepath.Join(home, "worktrees", "projects", "proj", "t7")
	deep = filepath.Join(treeRoot, "internal", "cli")
	return deep, treeRoot
}

// TestCardIDFromPath_HomeRootDepth pins AC-WHR-008: a session whose cwd is
// deep inside a home-root tree derives the card id t7 — never "projects",
// never the project key "proj".
func TestCardIDFromPath_HomeRootDepth(t *testing.T) {
	deep, treeRoot := m3HomeRootDirs(filepath.Join(string(filepath.Separator), "home-m3", "home"))
	for name, dir := range map[string]string{"tree root": treeRoot, "deep cwd": deep} {
		if got := cardIDFromPath(dir); got != "t7" {
			t.Errorf("%s: cardIDFromPath(%s) = %q, want t7 (REQ-WHR-008: the parent segment is the project key, not the card)", name, dir, got)
		}
	}
}

// TestCardIDFromRegisteredWorktree_HomeRoot pins the registered-tier
// derivation on the home root: the walk must recognize
// <MoaiHome>/worktrees/projects/<key>/<name> and return the tree name.
func TestCardIDFromRegisteredWorktree_HomeRoot(t *testing.T) {
	deep, treeRoot := m3HomeRootDirs(filepath.Join(string(filepath.Separator), "home-m3", "home"))
	for name, dir := range map[string]string{"tree root": treeRoot, "deep cwd": deep} {
		if got := cardIDFromRegisteredWorktree(dir); got != "t7" {
			t.Errorf("%s: cardIDFromRegisteredWorktree(%s) = %q, want t7", name, dir, got)
		}
	}
}

// TestCardIDDerivations_LegacyShapesPinned pins the legacy behavior the M3
// change must preserve: both derivations resolve the two in-repo tiers at
// any cwd depth.
func TestCardIDDerivations_LegacyShapesPinned(t *testing.T) {
	cases := []struct {
		dir  string
		want string
	}{
		{filepath.Join(string(filepath.Separator), "repo", ".moai", "worktrees", "t7b", "internal", "hook"), "t7b"},
		{filepath.Join(string(filepath.Separator), "repo", ".moai", "worktrees", "t7b"), "t7b"},
		{filepath.Join(string(filepath.Separator), "repo", ".claude", "worktrees", "t7c"), "t7c"},
	}
	for _, c := range cases {
		if got := cardIDFromPath(c.dir); got != c.want {
			t.Errorf("cardIDFromPath(%s) = %q, want %q", c.dir, got, c.want)
		}
		if got := cardIDFromRegisteredWorktree(c.dir); got != c.want {
			t.Errorf("cardIDFromRegisteredWorktree(%s) = %q, want %q", c.dir, got, c.want)
		}
	}
}
