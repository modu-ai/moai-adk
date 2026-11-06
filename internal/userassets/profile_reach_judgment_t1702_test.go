package userassets

// Card t1702, round 7: the READ-ONLY profile-reach judgment.
//
// Operator ruling (relayed 11-06): before a project copy is removed, read each
// profile root and check it holds the same copy; absent or uncertain keeps the
// project copy; NOTHING is written to a profile root. These tests pin the
// judgment (CheckProfileReach / SafeToRemoveProjectCopy) and its read-only
// property. Every test uses a temporary HOME.

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const reachRel = "moai-alpha/SKILL.md"

var reachBody = []byte("alpha skill body\n")

func reachHash(b []byte) [32]byte { return sha256.Sum256(b) }

// reachProfile makes profile `name` under home and, when body != nil, writes
// skills/<reachRel> with that body.
func reachProfile(t *testing.T, home, name string, body []byte) string {
	t.Helper()
	dir := filepath.Join(ProfileRootsBase(home), name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if body != nil {
		p := filepath.Join(dir, "skills", filepath.FromSlash(reachRel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func reachStates(rs []ProfileReach) map[string]ReachState {
	m := map[string]ReachState{}
	for _, r := range rs {
		m[r.Profile] = r.State
	}
	return m
}

// treeSnapshot lists every entry under dir with its kind and size, so a
// read-only call can be shown to leave the tree byte-for-byte as it was.
func treeSnapshot(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		b.WriteString(rel + "|" + info.Mode().String() + "|" + info.ModTime().String() + "\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestCheckProfileReach_SameCopyInEveryProfileIsSafeToRemove(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", reachBody)
	reachProfile(t, home, "beta", reachBody)
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	got := reachStates(rs)
	if got["alpha"] != ReachSame || got["beta"] != ReachSame || len(rs) != 2 {
		t.Fatalf("want alpha and beta Same, got %+v", rs)
	}
	for _, r := range rs {
		if r.Root == "" || r.Reason == "" {
			t.Errorf("a row must name the root it read and a reason: %+v", r)
		}
	}
	if !SafeToRemoveProjectCopy(rs) {
		t.Fatal("every profile holds the same copy: removal is safe")
	}
}

func TestCheckProfileReach_AbsentKeepsTheProjectCopy(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", reachBody)
	reachProfile(t, home, "beta", nil) // profile exists, no skills root at all
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	if got := reachStates(rs); got["alpha"] != ReachSame || got["beta"] != ReachAbsent {
		t.Fatalf("want alpha Same, beta Absent; got %+v", rs)
	}
	if SafeToRemoveProjectCopy(rs) {
		t.Fatal("one profile lacks the copy: the project copy must stay")
	}
}

func TestCheckProfileReach_AbsentFileInExistingRootIsAbsent(t *testing.T) {
	home := t.TempDir()
	dir := reachProfile(t, home, "alpha", nil)
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	if got := reachStates(rs); got["alpha"] != ReachAbsent {
		t.Fatalf("the root exists but not the copy: want Absent, got %+v", rs)
	}
}

func TestCheckProfileReach_DifferingContentKeepsTheProjectCopy(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", []byte("the profile's own, different skill\n"))
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	if got := reachStates(rs); got["alpha"] != ReachDiffers {
		t.Fatalf("want Differs, got %+v", rs)
	}
	if SafeToRemoveProjectCopy(rs) {
		t.Fatal("a differing copy must keep the project copy")
	}
}

func TestCheckProfileReach_UncertainShapesAreUnknownNotSame(t *testing.T) {
	home := t.TempDir()
	// dir-as-file: the copy path is a directory.
	d := reachProfile(t, home, "isdir", nil)
	if err := os.MkdirAll(filepath.Join(d, "skills", filepath.FromSlash(reachRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	// link leaving the root: the skill dir is a link to a tree outside.
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outside, "moai-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "moai-alpha", "SKILL.md"), reachBody, 0o644); err != nil {
		t.Fatal(err)
	}
	l := reachProfile(t, home, "outlink", nil)
	if err := os.MkdirAll(filepath.Join(l, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "moai-alpha"), filepath.Join(l, "skills", "moai-alpha")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// the file itself is a link (even to identical content): not a regular file.
	f := reachProfile(t, home, "filelink", nil)
	if err := os.MkdirAll(filepath.Join(f, "skills", "moai-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "moai-alpha", "SKILL.md"), filepath.Join(f, "skills", "moai-alpha", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	got := reachStates(rs)
	for _, name := range []string{"isdir", "outlink", "filelink"} {
		if got[name] != ReachUnknown {
			t.Errorf("profile %s: want Unknown, got %v (%+v)", name, got[name], rs)
		}
	}
	if SafeToRemoveProjectCopy(rs) {
		t.Fatal("an uncertain profile must keep the project copy")
	}
}

func TestSafeToRemoveProjectCopy_NoProfilesAndEmptyKeepTheCopy(t *testing.T) {
	home := t.TempDir() // no profile store at all
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatalf("a missing store is the legitimate empty store, not an error: %v", err)
	}
	if len(rs) != 0 {
		t.Fatalf("no profiles: want no rows, got %+v", rs)
	}
	if SafeToRemoveProjectCopy(rs) || SafeToRemoveProjectCopy(nil) {
		t.Fatal("zero profiles must keep the project copy")
	}
}

func TestSafeToRemoveProjectCopy_AnyNonSameStateKeepsTheCopy(t *testing.T) {
	same := ProfileReach{Profile: "a", State: ReachSame}
	for _, other := range []ReachState{ReachAbsent, ReachDiffers, ReachUnknown, ReachState(99)} {
		if SafeToRemoveProjectCopy([]ProfileReach{same, {Profile: "b", State: other}}) {
			t.Errorf("state %v beside a Same row must keep the copy", other)
		}
	}
	if !SafeToRemoveProjectCopy([]ProfileReach{same}) {
		t.Error("a single Same row is safe")
	}
}

// Audit iter6 B1 (reproduced): a store entry the launcher can start but the
// shared enumeration drops (a linked directory, a name with a space, the
// reserved sentinel) is a profile whose roots were never read. It must show up
// as an Unknown row, so that "every other profile is Same" cannot license a
// removal. Entries the launcher cannot start (plain files, hidden names,
// dangling links) are not profiles and give no row.
func TestCheckProfileReach_LaunchableProfilesTheSharedFilterDropsAreUnknown(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", reachBody)
	reachProfile(t, home, "has space", reachBody)
	reachProfile(t, home, "__no_such_profile__", reachBody)
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outside, "skills", "moai-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "skills", filepath.FromSlash(reachRel)), reachBody, 0o644); err != nil {
		t.Fatal(err)
	}
	base := ProfileRootsBase(home)
	if err := os.Symlink(outside, filepath.Join(base, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(outside, "no-such-target"), filepath.Join(base, "dangling")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(base, ".hidden"), 0o755); err != nil {
		t.Fatal(err)
	}
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	got := reachStates(rs)
	want := map[string]ReachState{
		"alpha": ReachSame, "has space": ReachUnknown,
		"__no_such_profile__": ReachUnknown, "linked": ReachUnknown,
	}
	if len(got) != len(want) || len(rs) != len(want) {
		t.Fatalf("rows = %+v, want exactly %v", rs, want)
	}
	for name, st := range want {
		if got[name] != st {
			t.Errorf("profile %q: state %v, want %v", name, got[name], st)
		}
	}
	if SafeToRemoveProjectCopy(rs) {
		t.Fatal("a launchable profile that was not read must keep the project copy")
	}
}

// Audit iter7 N1: a store entry that cannot even be inspected (a link that
// loops, so os.Stat fails with something other than "does not exist") is not a
// plain file or a dangling link; the launcher's own os.Stat would fail the same
// way, but "cannot tell" must keep the project copy, so it is an Unknown row.
func TestCheckProfileReach_AStoreEntryThatCannotBeInspectedIsUnknown(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", reachBody)
	if err := os.Symlink("loop", filepath.Join(ProfileRootsBase(home), "loop")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	if got := reachStates(rs); got["alpha"] != ReachSame || got["loop"] != ReachUnknown || len(rs) != 2 {
		t.Fatalf("want alpha Same and loop Unknown, got %+v", rs)
	}
	if SafeToRemoveProjectCopy(rs) {
		t.Fatal("an entry that cannot be inspected must keep the project copy")
	}
}

// Audit iter7 N4: between the Lstat of the copy and its open, the name is
// swapped for a link to ANOTHER file inside the same root. The open follows it,
// so the handle must be shown to be the file that was inspected.
func TestCheckProfileReach_ACopySwappedForAnInRootLinkAfterTheLstatIsUnknown(t *testing.T) {
	home := t.TempDir()
	d := reachProfile(t, home, "alpha", reachBody)
	other := filepath.Join(d, "skills", "moai-alpha", "other.md")
	if err := os.WriteFile(other, reachBody, 0o644); err != nil { // same content: only identity differs
		t.Fatal(err)
	}
	orig := afterReachLstat
	skipSwap := false
	afterReachLstat = func(dir, rel string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.Remove(p); err != nil {
			t.Errorf("swap, step 1: %v", err)
			return
		}
		if err := os.Symlink("other.md", p); err != nil {
			// no links here (for example an unprivileged Windows account): the
			// swap cannot be staged, which is not a defect of the judgment
			skipSwap = true
		}
	}
	t.Cleanup(func() { afterReachLstat = orig })
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	if skipSwap {
		t.Skip("symlinks unavailable: the swap could not be staged")
	}
	if got := reachStates(rs); got["alpha"] != ReachUnknown {
		t.Fatalf("the opened file is not the one that was inspected: want Unknown, got %+v", rs)
	}
}

// Audit iter6 S1: shapes that must stay Unknown and were not pinned.
func TestCheckProfileReach_RootAndFileShapesThatMustStayUnknown(t *testing.T) {
	home := t.TempDir()
	// skills root is itself a link (even to a directory holding the same copy).
	real := reachProfile(t, home, "realtree", reachBody)
	rl := reachProfile(t, home, "rootlink", nil)
	if err := os.Symlink(filepath.Join(real, "skills"), filepath.Join(rl, "skills")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// the file is a link that stays inside the root.
	fl := reachProfile(t, home, "inrootlink", reachBody)
	if err := os.Rename(filepath.Join(fl, "skills", "moai-alpha", "SKILL.md"), filepath.Join(fl, "skills", "moai-alpha", "real.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.md", filepath.Join(fl, "skills", "moai-alpha", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	// the file is larger than the read bound.
	big := reachProfile(t, home, "oversize", nil)
	if err := os.MkdirAll(filepath.Join(big, "skills", "moai-alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(big, "skills", filepath.FromSlash(reachRel)), make([]byte, maxReachFileBytes+1), 0o644); err != nil {
		t.Fatal(err)
	}
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	got := reachStates(rs)
	for _, name := range []string{"rootlink", "inrootlink", "oversize"} {
		if got[name] != ReachUnknown {
			t.Errorf("profile %s: want Unknown, got %v", name, got[name])
		}
	}
	if got["realtree"] != ReachSame {
		t.Errorf("the real tree holds the copy: want Same, got %v", got["realtree"])
	}
}

func TestCheckProfileReach_AgentsRootAndFixedSlugs(t *testing.T) {
	home := t.TempDir()
	d := reachProfile(t, home, "alpha", nil)
	if err := os.MkdirAll(filepath.Join(d, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte("agent body\n")
	if err := os.WriteFile(filepath.Join(d, "agents", "x.md"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	rs, err := CheckProfileReach(home, RootClaudeAgents, "x.md", reachHash(body))
	if err != nil {
		t.Fatal(err)
	}
	if got := reachStates(rs); got["alpha"] != ReachSame {
		t.Fatalf("the agents root is read too: %+v", rs)
	}
	// agents-skills / codex-agents are shared across profiles: they have no
	// profile root, so there is nothing to read and removal is not licensed.
	for _, slug := range []RootSlug{RootAgentsSkills, RootCodexAgents} {
		rs, err := CheckProfileReach(home, slug, "x.md", reachHash(body))
		if err != nil {
			t.Fatal(err)
		}
		if len(rs) != 0 || SafeToRemoveProjectCopy(rs) {
			t.Errorf("slug %s has no profile root: want no rows and keep, got %+v", slug, rs)
		}
	}
}

func TestCheckProfileReach_RejectsEscapingRelPath(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", reachBody)
	for _, rel := range []string{"../x", "/abs", "", "a/../../x"} {
		if _, err := CheckProfileReach(home, RootClaudeSkills, rel, reachHash(reachBody)); err == nil {
			t.Errorf("rel %q must be rejected", rel)
		}
	}
}

func TestCheckProfileReach_WritesNothingAndCreatesNothing(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", reachBody)
	reachProfile(t, home, "beta", nil)       // no skills dir: must stay without one
	reachProfile(t, home, "gamma", []byte{}) // empty differing copy
	before := treeSnapshot(t, home)
	if _, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody)); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckProfileReach(home, RootClaudeAgents, "x.md", reachHash(reachBody)); err != nil {
		t.Fatal(err)
	}
	if after := treeSnapshot(t, home); after != before {
		t.Fatalf("the judgment changed the tree\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestCheckProfileReach_UnreadableStoreIsAnErrorNotAnEmptyAnswer(t *testing.T) {
	home := t.TempDir()
	// The profile store path is a regular file: it exists but cannot be listed.
	if err := os.MkdirAll(filepath.Dir(ProfileRootsBase(home)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ProfileRootsBase(home), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err == nil {
		t.Fatalf("an unreadable store must be an error, got rows %+v", rs)
	}
	if SafeToRemoveProjectCopy(rs) {
		t.Fatal("on an error the project copy stays")
	}
}
