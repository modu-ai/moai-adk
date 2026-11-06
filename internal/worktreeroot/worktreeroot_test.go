package worktreeroot

// worktreeroot_test.go - SPEC-WORKTREE-HOME-ROOT-001 M1 RED commit (card t1640).
//
// These tests are written before the resolver exists: the package under test
// has no production file, so `go test` fails to compile it. That compile
// failure is the RED (plan.md section C, test compilation order).
//
// Acceptance mapping: AC-WHR-001 (home-root form, MOAI_HOME), AC-WHR-002a
// (key derivation and owner marker), AC-WHR-002b clauses 1-3 (stability),
// AC-WHR-003a (case-insensitive collision), AC-WHR-003b (refusals), AC-WHR-004
// (operator pin), AC-WHR-022 (Candidates, no preferred root), AC-WHR-023
// (Same, spelling-safe equality).
//
// Surface fixed by this file (the GREEN stage implements exactly this):
//
//	type RootKind string                       // "home" | "legacy-moai" | "legacy-claude"
//	type Candidate struct{ Path string; Kind RootKind }
//	type Resolution struct{ PrimaryRoot, Key, Root, Notice string }
//	func Resolve(repoDir string) (Resolution, error)
//	func Key(repoDir string) (string, error)
//	func Root(repoDir string) (string, error)
//	func PathFor(repoDir, name string) (string, error)
//	func Candidates(repoDir, name string) ([]Candidate, error)
//	func Same(a, b string) bool
//
// repoDir is any directory of the repository: the primary checkout or one of
// its linked worktrees (AC-WHR-002a: the key comes from the primary root,
// never from the invoking worktree).
//
// Every fixture is a scratch repository under t.TempDir() with its own
// MOAI_HOME; no test touches the real MoaiHome or this repository's worktrees.
// The linked trees made with `git worktree add` are scratch fixtures, the
// sanctioned exception to AGENTS.md section 3.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// ---------------------------------------------------------------- fixtures

// wrReal resolves symlinks (the macOS /var -> /private/var alias) so paths the
// resolver derives from git agree with the paths asserted here.
func wrReal(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// wrGit runs one git command in dir and fails the test on error.
func wrGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	// Async maintenance/auto-gc must never run inside the fixture: its
	// `.git/objects/maintenance.lock` churn was the natural CI failure of
	// TestPlan_IsReadOnly (card t1954, run 40178638653 — "REMOVED
	// plan-repo/.git/objects/maintenance.lock" between the two hashes;
	// written and removed by git itself, never by Plan). Blocking it here
	// keeps the read-only proof about Plan, not about git's own housekeeping.
	args = append([]string{"-c", "maintenance.auto=false", "-c", "gc.auto=0"}, args...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// wrRepo creates a scratch repository named name (under its own temp parent)
// with one commit, and returns its symlink-resolved root. --template= keeps
// the sample hooks out of .git so a text search of .git is not polluted.
func wrRepo(t *testing.T, name string) string {
	t.Helper()
	root := filepath.Join(wrReal(t, t.TempDir()), name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	wrGit(t, root, "init", "-q", "--template=", "-b", "main")
	wrGit(t, root, "config", "user.email", "wr-test@example.com")
	wrGit(t, root, "config", "user.name", "WR Test")
	wrGit(t, root, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	wrGit(t, root, "add", "README.md")
	wrGit(t, root, "commit", "-q", "-m", "seed")
	return root
}

// wrLinked adds a linked worktree of repo (outside repo) and returns its path.
func wrLinked(t *testing.T, repo, name string) string {
	t.Helper()
	path := filepath.Join(wrReal(t, t.TempDir()), name)
	wrGit(t, repo, "worktree", "add", "-q", "-b", "WT-"+name, "--", path)
	return wrReal(t, path)
}

// wrHome points MOAI_HOME at a fresh temp directory and returns it.
func wrHome(t *testing.T) string {
	t.Helper()
	home := wrReal(t, t.TempDir())
	t.Setenv("MOAI_HOME", home)
	return home
}

func wrProjects(home string) string { return filepath.Join(home, "worktrees", "projects") }

func wrMarker(home, key string) string {
	return filepath.Join(wrProjects(home), key, ".moai-project.json")
}

// wrPin writes git_strategy.worktree_project_key into the repository's
// git-strategy.yaml (the file the config reader of REQ-WHR-004 consults).
func wrPin(t *testing.T, repo, value string) {
	t.Helper()
	dir := filepath.Join(repo, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "git_strategy:\n    worktree_project_key: '" + value + "'\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
}

// wrSuffixKey is the disambiguated key REQ-WHR-003 prescribes for the
// later-arriving repository: <project-name>-<first 8 hex of SHA-256 of the
// case-folded canonical primary root>. "Canonical" is read as symlink-resolved
// and "case-folded" as strings.ToLower; the fixture roots are already
// symlink-resolved, so the formula is the one the GREEN stage must reproduce.
func wrSuffixKey(t *testing.T, repo string) string {
	t.Helper()
	sum := sha256.Sum256([]byte(strings.ToLower(wrReal(t, repo))))
	return filepath.Base(repo) + "-" + hex.EncodeToString(sum[:])[:8]
}

func wrKey(t *testing.T, dir string) string {
	t.Helper()
	k, err := Key(dir)
	if err != nil {
		t.Fatalf("Key(%s): %v", dir, err)
	}
	return k
}

func wrRoot(t *testing.T, dir string) string {
	t.Helper()
	r, err := Root(dir)
	if err != nil {
		t.Fatalf("Root(%s): %v", dir, err)
	}
	return r
}

// wrEntries lists the names in dir, sorted; a missing dir lists as empty.
func wrEntries(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("readdir %s: %v", dir, err)
	}
	var names []string
	for _, e := range ents {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

// wrCaseInsensitive reports whether the volume under the temp dir folds case.
func wrCaseInsensitive(t *testing.T) bool {
	t.Helper()
	dir := wrReal(t, t.TempDir())
	if err := os.WriteFile(filepath.Join(dir, "CaseProbe"), nil, 0o644); err != nil {
		t.Fatalf("probe file: %v", err)
	}
	_, err := os.Stat(filepath.Join(dir, "caseprobe"))
	return err == nil
}

// ------------------------------------------------- AC-WHR-001 home-root form

// TestWorktreeRoot_PathFor_HomeRootForm pins the path form
// <MoaiHome>/worktrees/projects/<project-key>/<name> and that MOAI_HOME
// relocates the whole tree (REQ-WHR-001).
func TestWorktreeRoot_PathFor_HomeRootForm(t *testing.T) {
	repo := wrRepo(t, "proj")

	t.Run("absolute MOAI_HOME", func(t *testing.T) {
		home := wrHome(t)
		want := filepath.Join(home, "worktrees", "projects", "proj", "t9")
		got, err := PathFor(repo, "t9")
		if err != nil {
			t.Fatalf("PathFor: %v", err)
		}
		if got != want {
			t.Errorf("PathFor = %q, want %q", got, want)
		}
		if root := wrRoot(t, repo); root != filepath.Dir(want) {
			t.Errorf("Root = %q, want %q", root, filepath.Dir(want))
		}
		// The path computation creates no tree.
		if _, err := os.Lstat(want); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("PathFor created %s (lstat err=%v); computing a path must create no tree", want, err)
		}
		// The first call records the assignment of AC-WHR-002a.
		if _, err := os.Stat(wrMarker(home, "proj")); err != nil {
			t.Errorf("owner marker missing after the first resolution: %v", err)
		}
	})

	t.Run("MOAI_HOME unset falls back to HOME/.moai", func(t *testing.T) {
		t.Setenv("MOAI_HOME", "")
		fakeHome := wrReal(t, t.TempDir())
		t.Setenv("HOME", fakeHome)
		want := filepath.Join(fakeHome, ".moai", "worktrees", "projects", "proj", "t9")
		got, err := PathFor(repo, "t9")
		if err != nil {
			t.Fatalf("PathFor: %v", err)
		}
		if got != want {
			t.Errorf("PathFor = %q, want %q", got, want)
		}
	})

	t.Run("relative MOAI_HOME is disregarded", func(t *testing.T) {
		t.Setenv("MOAI_HOME", "relative/moai-home")
		fakeHome := wrReal(t, t.TempDir())
		t.Setenv("HOME", fakeHome)
		want := filepath.Join(fakeHome, ".moai", "worktrees", "projects", "proj", "t9")
		got, err := PathFor(repo, "t9")
		if err != nil {
			t.Fatalf("PathFor: %v", err)
		}
		if got != want {
			t.Errorf("PathFor = %q, want %q", got, want)
		}
	})
}

// ------------------------------------- AC-WHR-002a key derivation and marker

// TestWorktreeRoot_Key_SanitizedPrimaryBasenameAndMarker pins the key as the
// sanitized basename of the PRIMARY root, derived even when the resolver is
// first invoked from a linked worktree, and the owner marker content
// (REQ-WHR-002).
func TestWorktreeRoot_Key_SanitizedPrimaryBasenameAndMarker(t *testing.T) {
	cases := []struct{ dir, want string }{
		{"proj", "proj"},
		{"my proj (v2)", "my-proj--v2-"},
		{"a_b.c-d", "a_b.c-d"},
		{"MixedCase", "MixedCase"},
		{"x@y#z", "x-y-z"},
	}
	for _, c := range cases {
		t.Run(c.dir, func(t *testing.T) {
			home := wrHome(t)
			repo := wrRepo(t, c.dir)
			linked := wrLinked(t, repo, "t1")

			// First resolution is from the linked worktree.
			if got := wrKey(t, linked); got != c.want {
				t.Fatalf("Key from linked worktree = %q, want %q", got, c.want)
			}
			if got := wrKey(t, repo); got != c.want {
				t.Errorf("Key from primary = %q, want %q", got, c.want)
			}

			raw, err := os.ReadFile(wrMarker(home, c.want))
			if err != nil {
				t.Fatalf("owner marker: %v", err)
			}
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("owner marker is not a JSON object: %v\n%s", err, raw)
			}
			// schema version, owner identity, primary root last seen, creation time.
			if len(m) < 4 {
				t.Errorf("owner marker has %d fields, want at least 4: %s", len(m), raw)
			}
			sawRoot := false
			for _, v := range m {
				if s, ok := v.(string); ok && s == repo {
					sawRoot = true
				}
			}
			if !sawRoot {
				t.Errorf("owner marker does not record the primary root %q: %s", repo, raw)
			}
		})
	}
}

// TestWorktreeRoot_Key_AssignmentIsRecordedInTheRepositoryAndTheMarker pins
// "recorded twice": the marker under MoaiHome and a record that travels with
// the repository (its git common directory). A pin different from the
// directory name makes the search non-vacuous: the key text exists nowhere in
// .git unless the resolver wrote it there.
func TestWorktreeRoot_Key_AssignmentIsRecordedInTheRepositoryAndTheMarker(t *testing.T) {
	home := wrHome(t)
	repo := wrRepo(t, "rec-dir")
	wrPin(t, repo, "pinned-uniq-q7")

	if got := wrKey(t, repo); got != "pinned-uniq-q7" {
		t.Fatalf("Key = %q, want the pin on first assignment", got)
	}
	if _, err := os.Stat(wrMarker(home, "pinned-uniq-q7")); err != nil {
		t.Errorf("owner marker: %v", err)
	}
	found := false
	_ = filepath.WalkDir(filepath.Join(repo, ".git"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if b, rerr := os.ReadFile(p); rerr == nil && strings.Contains(string(b), "pinned-uniq-q7") {
			found = true
		}
		return nil
	})
	if !found {
		t.Errorf("no file under %s/.git records the assigned key; the assignment must travel with the repository", repo)
	}

	// The record, not the pin, keeps the key once assigned.
	if err := os.Remove(filepath.Join(repo, ".moai", "config", "sections", "git-strategy.yaml")); err != nil {
		t.Fatalf("remove pin: %v", err)
	}
	if got := wrKey(t, repo); got != "pinned-uniq-q7" {
		t.Errorf("Key after the pin was removed = %q, want the recorded key", got)
	}
}

// ------------------------------------------------ AC-WHR-002b clauses 1-3

// TestWorktreeRoot_Key_StableAcrossResolutionMoveAndRename pins I-8: once
// assigned, a key never changes - not on re-resolution, not when the checkout
// directory is moved or renamed, not when a same-name repository arrives.
func TestWorktreeRoot_Key_StableAcrossResolutionMoveAndRename(t *testing.T) {
	home := wrHome(t)
	first := wrRepo(t, "proj")

	// Clause 1: re-resolution with no change.
	if got := wrKey(t, first); got != "proj" {
		t.Fatalf("first assignment = %q, want proj", got)
	}
	if got := wrKey(t, first); got != "proj" {
		t.Errorf("re-resolution = %q, want proj", got)
	}

	// Clause 2: the checkout directory is renamed and moved; the
	// repository-side record moves with it.
	moved := filepath.Join(wrReal(t, t.TempDir()), "other-name")
	if err := os.Rename(first, moved); err != nil {
		t.Fatalf("rename checkout: %v", err)
	}
	if got := wrKey(t, moved); got != "proj" {
		t.Errorf("key from the renamed checkout = %q, want proj (never other-name)", got)
	}
	if got, want := wrRoot(t, moved), filepath.Join(wrProjects(home), "proj"); got != want {
		t.Errorf("Root from the renamed checkout = %q, want %q", got, want)
	}

	// Clause 3: a different repository named proj arrives later.
	later := wrRepo(t, "proj")
	kl := wrKey(t, later)
	if !strings.HasPrefix(kl, "proj-") || len(kl) != len("proj-")+8 {
		t.Errorf("later arrival key = %q, want proj-<8 hex>", kl)
	}
	if got := wrKey(t, moved); got != "proj" {
		t.Errorf("first repository after the arrival = %q, want proj (the first keeps the plain name)", got)
	}
	if got := wrKey(t, later); got != kl {
		t.Errorf("later arrival re-resolved = %q, want the stable %q", got, kl)
	}
}

// TestWorktreeRoot_Key_RecordedKeySurvivesMoaiHomeRelocation: the recorded
// key is returned on every later resolution even when the plain name would
// now be free (a relocated MOAI_HOME), so a key is a property of the
// repository, not of the folder state.
func TestWorktreeRoot_Key_RecordedKeySurvivesMoaiHomeRelocation(t *testing.T) {
	wrHome(t)
	a := wrRepo(t, "proj")
	b := wrRepo(t, "proj")
	if got := wrKey(t, a); got != "proj" {
		t.Fatalf("first = %q, want proj", got)
	}
	kb := wrKey(t, b)
	if kb == "proj" {
		t.Fatalf("second repository got the plain name %q; a case-insensitive collision must suffix it", kb)
	}

	homeB := wrReal(t, t.TempDir())
	t.Setenv("MOAI_HOME", homeB)
	if got := wrKey(t, b); got != kb {
		t.Errorf("key after relocating MOAI_HOME = %q, want the recorded %q", got, kb)
	}
	if got, want := wrRoot(t, b), filepath.Join(wrProjects(homeB), kb); got != want {
		t.Errorf("Root after relocation = %q, want %q", got, want)
	}
	if _, err := os.Stat(wrMarker(homeB, kb)); err != nil {
		t.Errorf("owner marker is not written under the relocated MOAI_HOME: %v", err)
	}
	if got := wrKey(t, a); got != "proj" {
		t.Errorf("first repository after relocation = %q, want proj", got)
	}
}

// ------------------------------------------------------ AC-WHR-003a collision

// TestWorktreeRoot_Key_CaseInsensitiveCollisionSuffix pins REQ-WHR-003: the
// folder-name test is case-insensitive; the later arrival takes
// <name>-<8 hex of SHA-256 of the case-folded canonical root>; the first keeps
// its plain name and its folder stays untouched.
func TestWorktreeRoot_Key_CaseInsensitiveCollisionSuffix(t *testing.T) {
	cases := []struct{ first, later string }{
		{"Proj", "proj"},
		{"proj", "PROJ"},
		{"proj", "proj"},
	}
	for _, c := range cases {
		t.Run(c.first+"_then_"+c.later, func(t *testing.T) {
			home := wrHome(t)
			a := wrRepo(t, c.first)
			b := wrRepo(t, c.later)

			if got := wrKey(t, a); got != c.first {
				t.Fatalf("first key = %q, want the plain %q", got, c.first)
			}
			markerBefore, err := os.ReadFile(wrMarker(home, c.first))
			if err != nil {
				t.Fatalf("first owner marker: %v", err)
			}

			kb := wrKey(t, b)
			if want := wrSuffixKey(t, b); kb != want {
				t.Errorf("later arrival key = %q, want %q", kb, want)
			}

			// The first repository is untouched and keeps its plain name.
			markerAfter, err := os.ReadFile(wrMarker(home, c.first))
			if err != nil {
				t.Fatalf("first owner marker after the arrival: %v", err)
			}
			if string(markerAfter) != string(markerBefore) {
				t.Errorf("the arrival rewrote the first repository's owner marker")
			}
			if got := wrKey(t, a); got != c.first {
				t.Errorf("first key after the arrival = %q, want %q", got, c.first)
			}

			// No tree of the later repository is placed under the first folder.
			rb := wrRoot(t, b)
			if strings.HasPrefix(strings.ToLower(rb), strings.ToLower(filepath.Join(wrProjects(home), c.first))+string(filepath.Separator)) {
				t.Errorf("later arrival Root %q lies under the first repository's folder %q", rb, c.first)
			}
			if want := filepath.Join(wrProjects(home), kb); rb != want {
				t.Errorf("later arrival Root = %q, want %q", rb, want)
			}
		})
	}
}

// ------------------------------------------------------ AC-WHR-003b refusals

// TestWorktreeRoot_Key_RefusesWhenTheSuffixedFolderHasAThirdOwner: the
// disambiguated folder is itself owned by a third repository. A pin is the
// way a fixture makes another repository own an arbitrary folder name without
// hand-writing a marker.
func TestWorktreeRoot_Key_RefusesWhenTheSuffixedFolderHasAThirdOwner(t *testing.T) {
	home := wrHome(t)
	first := wrRepo(t, "Proj")
	third := wrRepo(t, "proj")
	squatter := wrRepo(t, "squatter")

	collide := wrSuffixKey(t, third)
	wrPin(t, squatter, collide)

	if got := wrKey(t, first); got != "Proj" {
		t.Fatalf("first = %q, want Proj", got)
	}
	if got := wrKey(t, squatter); got != collide {
		t.Fatalf("squatter pin = %q, want %q", got, collide)
	}
	before := wrEntries(t, wrProjects(home))

	_, err := Resolve(third)
	if err == nil {
		t.Fatalf("Resolve succeeded; the suffixed folder %q is owned by a third repository", collide)
	}
	if want := wrMarker(home, collide); !strings.Contains(err.Error(), want) {
		t.Errorf("refusal %q does not name the marker path %q", err, want)
	}
	if _, perr := PathFor(third, "t1"); perr == nil {
		t.Errorf("PathFor succeeded after the refusal; no tree may be created")
	}
	if after := wrEntries(t, wrProjects(home)); strings.Join(after, ",") != strings.Join(before, ",") {
		t.Errorf("the refusal changed the projects folder: before %v, after %v", before, after)
	}
}

// TestWorktreeRoot_Key_RefusesAnUnreadableMarker: a marker that exists but
// cannot be read or parsed makes creation refuse, naming the marker path,
// also when the folder differs from the candidate only by case.
func TestWorktreeRoot_Key_RefusesAnUnreadableMarker(t *testing.T) {
	bodies := map[string]string{
		"garbage":   "{ not json",
		"empty":     "",
		"truncated": `{"schema":`,
	}
	folders := []string{"proj", "Proj"}
	for name, body := range bodies {
		for _, folder := range folders {
			t.Run(name+"/"+folder, func(t *testing.T) {
				home := wrHome(t)
				repo := wrRepo(t, "proj")
				dir := filepath.Join(wrProjects(home), folder)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				marker := wrMarker(home, folder)
				if err := os.WriteFile(marker, []byte(body), 0o644); err != nil {
					t.Fatalf("write marker: %v", err)
				}

				_, err := Resolve(repo)
				if err == nil {
					t.Fatalf("Resolve succeeded over an unreadable marker")
				}
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(marker)) {
					t.Errorf("refusal %q does not name the marker path %q", err, marker)
				}
				if _, perr := PathFor(repo, "t1"); perr == nil {
					t.Errorf("PathFor succeeded after the refusal; no tree may be created")
				}
				if _, serr := os.Lstat(filepath.Join(dir, "t1")); !errors.Is(serr, fs.ErrNotExist) {
					t.Errorf("a tree directory exists under the refused folder (lstat err=%v)", serr)
				}
			})
		}
	}
}

// --------------------------------------------------------- AC-WHR-004 the pin

// TestWorktreeRoot_Key_OperatorPin pins REQ-WHR-004: the pin names the
// first-assigned key only; it never re-keys an assigned repository; a value
// that is not one safe path segment is rejected; a folder owned by another
// repository makes the pin refuse (never suffix).
func TestWorktreeRoot_Key_OperatorPin(t *testing.T) {
	t.Run("first assignment takes the pin, also from a linked worktree", func(t *testing.T) {
		home := wrHome(t)
		repo := wrRepo(t, "proj")
		linked := wrLinked(t, repo, "t1")
		wrPin(t, repo, "fixedkey")
		if got := wrKey(t, linked); got != "fixedkey" {
			t.Errorf("Key = %q, want fixedkey", got)
		}
		if _, err := os.Stat(wrMarker(home, "fixedkey")); err != nil {
			t.Errorf("owner marker under the pinned key: %v", err)
		}
	})

	t.Run("an empty or blank pin equals the derivation", func(t *testing.T) {
		for _, value := range []string{"", "   "} {
			wrHome(t)
			repo := wrRepo(t, "proj")
			wrPin(t, repo, value)
			if got := wrKey(t, repo); got != "proj" {
				t.Errorf("pin %q: Key = %q, want proj", value, got)
			}
		}
	})

	t.Run("a later differing pin keeps the recorded key and notices", func(t *testing.T) {
		home := wrHome(t)
		repo := wrRepo(t, "proj")
		res, err := Resolve(repo)
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if res.Key != "proj" || res.Notice != "" {
			t.Fatalf("first resolution = key %q notice %q, want proj and no notice", res.Key, res.Notice)
		}
		wrPin(t, repo, "other")
		res, err = Resolve(repo)
		if err != nil {
			t.Fatalf("Resolve with a differing pin: %v", err)
		}
		if res.Key != "proj" {
			t.Errorf("Key with a differing pin = %q, want the recorded proj", res.Key)
		}
		if !strings.Contains(res.Notice, "other") || !strings.Contains(res.Notice, "proj") {
			t.Errorf("Notice %q must name both the pin and the recorded key", res.Notice)
		}
		if got := wrEntries(t, wrProjects(home)); strings.Join(got, ",") != "proj" {
			t.Errorf("projects folder = %v, want only proj (a pin never re-keys)", got)
		}
		wrPin(t, repo, "proj")
		if res, err = Resolve(repo); err != nil || res.Notice != "" {
			t.Errorf("pin equal to the recorded key: notice %q err %v, want none", res.Notice, err)
		}
	})

	t.Run("a pin that is not one safe segment is rejected", func(t *testing.T) {
		for _, bad := range []string{"a/b", `a\b`, "..", "../x", "x/..", "."} {
			home := wrHome(t)
			repo := wrRepo(t, "proj")
			wrPin(t, repo, bad)
			if _, err := Resolve(repo); err == nil {
				t.Errorf("pin %q was accepted, want rejection", bad)
			}
			if got := wrEntries(t, wrProjects(home)); len(got) != 0 {
				t.Errorf("pin %q created %v under projects, want nothing", bad, got)
			}
		}
	})

	t.Run("a pin naming another repository's folder refuses and never suffixes", func(t *testing.T) {
		for _, pin := range []string{"taken", "TAKEN"} {
			home := wrHome(t)
			owner := wrRepo(t, "taken")
			other := wrRepo(t, "other")
			if got := wrKey(t, owner); got != "taken" {
				t.Fatalf("owner key = %q, want taken", got)
			}
			wrPin(t, other, pin)
			if _, err := Resolve(other); err == nil {
				t.Errorf("pin %q over another repository's folder was accepted", pin)
			}
			if got := wrEntries(t, wrProjects(home)); strings.Join(got, ",") != "taken" {
				t.Errorf("pin %q changed the projects folder to %v, want only taken", pin, got)
			}
		}
	})
}

// ------------------------------------------------------- AC-WHR-022 Candidates

// TestWorktreeRoot_Candidates_AcrossThreeRoots pins REQ-WHR-022: Candidates
// returns every distinct tree of the name found at the home root and at both
// legacy roots, with no root preferred and no ambiguity judgment (the caller
// reports ambiguity when it receives two or more).
func TestWorktreeRoot_Candidates_AcrossThreeRoots(t *testing.T) {
	wrHome(t)
	repo := wrRepo(t, "proj")
	linked := wrLinked(t, repo, "lk")
	roots := map[string]string{
		"home":   wrRoot(t, repo),
		"moai":   filepath.Join(repo, ".moai", "worktrees"),
		"claude": filepath.Join(repo, ".claude", "worktrees"),
	}
	kinds := map[string]string{"home": "home", "moai": "legacy-moai", "claude": "legacy-claude"}

	cases := []struct {
		name string
		at   []string
	}{
		{"t8", []string{"home"}},
		{"t9", []string{"claude"}},
		{"t7", []string{"home", "moai"}},
		{"t6", []string{"moai", "claude"}},
		{"t4", []string{"home", "claude"}},
		{"t3", []string{"home", "moai", "claude"}},
		{"t0", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var want []string
			for _, at := range c.at {
				p := filepath.Join(roots[at], c.name)
				if err := os.MkdirAll(p, 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", p, err)
				}
				want = append(want, kinds[at]+"="+p)
			}
			sort.Strings(want)

			for from, dir := range map[string]string{"primary": repo, "linked worktree": linked} {
				got, err := Candidates(dir, c.name)
				if err != nil {
					t.Fatalf("Candidates from %s: %v", from, err)
				}
				var have []string
				for _, cand := range got {
					have = append(have, string(cand.Kind)+"="+cand.Path)
				}
				sort.Strings(have)
				if strings.Join(have, "\n") != strings.Join(want, "\n") {
					t.Errorf("Candidates(%s, %q) from %s:\n got  %v\n want %v", dir, c.name, from, have, want)
				}
			}
		})
	}
}

// TestWorktreeRoot_Candidates_TwoSpellingsOfOneDirectoryCollapse: the same
// directory reached under two spellings is one tree, not an ambiguity.
func TestWorktreeRoot_Candidates_TwoSpellingsOfOneDirectoryCollapse(t *testing.T) {
	wrHome(t)
	repo := wrRepo(t, "proj")
	homeTree := filepath.Join(wrRoot(t, repo), "t5")
	if err := os.MkdirAll(homeTree, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacy := filepath.Join(repo, ".moai", "worktrees")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(homeTree, filepath.Join(legacy, "t5")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	got, err := Candidates(repo, "t5")
	if err != nil {
		t.Fatalf("Candidates: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("Candidates returned %d trees for one directory under two spellings: %v", len(got), got)
	}
}

// -------------------------------------------------------- AC-WHR-023 Same

// TestWorktreeRoot_Same_SpellingSafeEquality pins REQ-WHR-023 for the
// resolver helper: two spellings of one directory are equal, distinct
// directories are not, and a path that no longer exists compares by its
// cleaned spelling.
func TestWorktreeRoot_Same_SpellingSafeEquality(t *testing.T) {
	base := wrReal(t, t.TempDir())
	a := filepath.Join(base, "a")
	b := filepath.Join(base, "b")
	for _, d := range []string{a, b, filepath.Join(a, "child")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}

	t.Run("one directory, several spellings", func(t *testing.T) {
		for name, other := range map[string]string{
			"identical":      a,
			"trailing slash": a + string(filepath.Separator),
			"dot segments":   filepath.Join(a, "child", ".."),
		} {
			if !Same(a, other) {
				t.Errorf("Same(%q, %q) = false (%s), want true", a, other, name)
			}
		}
	})

	t.Run("symlink alias", func(t *testing.T) {
		link := filepath.Join(base, "alias")
		if err := os.Symlink(a, link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		if !Same(a, link) {
			t.Errorf("Same(%q, %q) = false, want true after symlink resolution", a, link)
		}
	})

	t.Run("distinct directories", func(t *testing.T) {
		if Same(a, b) {
			t.Errorf("Same(%q, %q) = true, want false", a, b)
		}
		if Same(a, filepath.Join(a, "child")) {
			t.Errorf("Same(parent, child) = true, want false")
		}
	})

	t.Run("paths that do not exist compare by cleaned spelling", func(t *testing.T) {
		gone := filepath.Join(base, "gone", "t7")
		if !Same(gone, gone+string(filepath.Separator)) {
			t.Errorf("Same of one missing path with a trailing slash = false, want true")
		}
		if Same(gone, filepath.Join(base, "gone", "t8")) {
			t.Errorf("Same of two different missing paths = true, want false")
		}
	})

	t.Run("case-aliased directory on a case-insensitive volume", func(t *testing.T) {
		if !wrCaseInsensitive(t) {
			t.Skip("volume is case-sensitive; the case-sensitive test below covers it")
		}
		dir := filepath.Join(base, "CaseDir")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		alias := filepath.Join(base, "casedir")
		if !Same(dir, alias) {
			t.Errorf("Same(%q, %q) = false on a case-insensitive volume, want true", dir, alias)
		}
	})

	t.Run("directories differing only by case on a case-sensitive volume stay distinct", func(t *testing.T) {
		if wrCaseInsensitive(t) {
			t.Skip("volume is case-insensitive; the case-aliased test above covers it")
		}
		upper := filepath.Join(base, "Dup")
		lower := filepath.Join(base, "dup")
		for _, d := range []string{upper, lower} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
		}
		if Same(upper, lower) {
			t.Errorf("Same(%q, %q) = true on a case-sensitive volume, want false", upper, lower)
		}
	})
}
