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

func TestCheckProfileReach_JunkProfileDirectoriesAreNotRead(t *testing.T) {
	home := t.TempDir()
	reachProfile(t, home, "alpha", reachBody)
	reachProfile(t, home, "has space", nil)
	reachProfile(t, home, "__no_such_profile__", nil)
	rs, err := CheckProfileReach(home, RootClaudeSkills, reachRel, reachHash(reachBody))
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 1 || rs[0].Profile != "alpha" {
		t.Fatalf("junk directories must not count as profiles: %+v", rs)
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
