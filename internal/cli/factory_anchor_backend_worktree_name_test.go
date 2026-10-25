package cli

// factory_anchor_backend_worktree_name_test.go — card t1958 (issue #1847): two
// worktree-name spellings on which the leader anchor guard and the backend
// disagreed about the tree the launch enters.
//
//   - `-w -`          the backend (claude 2.1.303) takes a lone `-` as the optional
//                     VALUE of -w/--worktree; the launch view used to read it as
//                     "no value", so the guard saw no selection at all.
//   - `-w feature/foo` the backend maps a slash in the name to `+` for the native
//                     directory (.claude/worktrees/feature+foo); the guard looked
//                     at .claude/worktrees/feature/foo, which does not exist.
//
// The cells drive the real runCC / runGLM leader entry through the harness of
// factory_anchor_guard_test.go; only the engine launch, the claude binary and
// the --branch materializer are substituted.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// addBackendNamedTree links a tree at .claude/worktrees/<dir> on branch.
func addBackendNamedTree(t *testing.T, fx *leaderFixture, dir, branch string) string {
	t.Helper()
	path := filepath.Join(fx.root, ".claude", "worktrees", dir)
	anchorGit(t, fx.root, "worktree", "add", "-q", "-b", branch, path, fx.base)
	return path
}

func TestLauncherAnchorBackendWorktreeNames(t *testing.T) {
	pinSnapshotModel(t)
	cases := []struct {
		name   string
		args   []string
		dir    string // directory under .claude/worktrees the backend enters ("" = the fixture's WT-card)
		branch string
		refuse bool
	}{
		{"lone dash, short, separator", []string{"-f", "--", "-w", "-"}, "-", "WT-dash", true},
		{"lone dash, short, no separator", []string{"-f", "-w", "-"}, "-", "WT-dash", true},
		{"lone dash, long", []string{"-f", "--worktree", "-"}, "-", "WT-dash", true},
		{"slash, short, separator", []string{"-f", "--", "-w", "feature/foo"}, "feature+foo", "WT-slash", true},
		{"slash, long equals", []string{"-f", "--worktree=feature/foo"}, "feature+foo", "WT-slash", true},
		{"slash, nested", []string{"-f", "-w", "a/b/c"}, "a+b+c", "WT-nested", true},
		{"lone dash, plain branch", []string{"-f", "--", "-w", "-"}, "-", "plain-dash", false},
		{"slash, plain branch", []string{"-f", "--", "-w", "feature/foo"}, "feature+foo", "plain-slash", false},
		{"short name, card tree (existing control)", []string{"-f", "-w", "WT-card"}, "", "", true},
	}
	for _, arm := range leaderArms {
		for _, tc := range cases {
			t.Run(arm+"/"+tc.name, func(t *testing.T) {
				fx := newLeaderFixture(t)
				tree, branch := fx.card, "WT-card"
				if tc.dir != "" {
					tree, branch = addBackendNamedTree(t, fx, tc.dir, tc.branch), tc.branch
				}
				obs := driveLeaderLaunch(t, fx, arm, tc.args)
				if tc.refuse {
					assertRefusedBeforeMutation(t, tc.name, obs)
					refusalNames(t, tc.name, obs, tree, branch)
					return
				}
				if obs.err != nil || !obs.launched {
					t.Errorf("%s: a tree on a plain branch must launch: err=%v launched=%v", tc.name, obs.err, obs.launched)
				}
			})
		}
	}
}

// TestLauncherAnchorBackendWorktreeNameSelection pins the unified meaning of a
// lone `-` in the launch view: the backend model reads it as the worktree
// option's value, and only there (a lone `-` standing alone is still a
// positional, and the consumer-scoped stage reads are untouched).
func TestLauncherAnchorBackendWorktreeNameSelection(t *testing.T) {
	pinSnapshotModel(t)
	cases := []struct {
		name string
		argv []string
		want worktreeSelection
	}{
		{"short", []string{"-w", "-"}, worktreeSelection{Present: true, Spelling: "-w", Value: "-", HasValue: true}},
		{"long", []string{"--worktree", "-"}, worktreeSelection{Present: true, Spelling: "--worktree", Value: "-", HasValue: true}},
		{"cluster", []string{"-pw", "-"}, worktreeSelection{Present: true, Spelling: "-<cluster>w<value>", Value: "-", HasValue: true}},
		{"attached dash", []string{"-w-"}, worktreeSelection{Present: true, Spelling: "-w<value>", Value: "-", HasValue: true}},
		{"a flag-shaped token is still not a value", []string{"-w", "-x"}, worktreeSelection{Present: true, Spelling: "-w"}},
		{"the separator is still not a value", []string{"-w", "--", "-"}, worktreeSelection{Present: true, Spelling: "-w"}},
		{"a lone dash before the option stays positional", []string{"-", "-w", "x"}, worktreeSelection{Present: true, Spelling: "-w", Value: "x", HasValue: true, Index: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := selectBackendWorktree(tc.argv); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("selectBackendWorktree(%q) = %+v, want %+v", tc.argv, got, tc.want)
			}
		})
	}

	// The consumer-scoped reads keep reading `-w -` as a bare option (SPEC-LAUNCHER-ARGV-001 §2).
	view := resolveLaunchView([]string{"-w", "-"})
	if view.StageA.Has || view.StageB.Has || view.Consumed.ChainTarget != "" {
		t.Errorf("stage reads of `-w -` = A %+v B %+v chain %q, want no value on all three", view.StageA, view.StageB, view.Consumed.ChainTarget)
	}
	if want := []string{"--worktree", "-"}; !reflect.DeepEqual(view.FinalArgv, want) {
		t.Errorf("FinalArgv = %q, want %q", view.FinalArgv, want)
	}
}

// TestLauncherAnchorBackendWorktreeNameTarget pins the name-to-directory
// mapping the guard inspects: the slash of a short name becomes `+`, an
// absolute value names its tree directly.
func TestLauncherAnchorBackendWorktreeNameTarget(t *testing.T) {
	pinSnapshotModel(t)
	fx := newLeaderFixture(t)
	for value, wantDir := range map[string]string{
		"plain":       "plain",
		"WT-card":     "WT-card",
		"-":           "-",
		"feature/foo": "feature+foo",
		"a/b/c":       "a+b+c",
		"feature+foo": "feature+foo",
	} {
		got := filepath.ToSlash(resolveWorktreeLaunchTarget(fx.root, value))
		if want := "/.claude/worktrees/" + wantDir; !strings.HasSuffix(got, want) {
			t.Errorf("resolveWorktreeLaunchTarget(%q) = %q, want a path ending in %q", value, got, want)
		}
	}
	abs := filepath.Join(t.TempDir(), "a", "b")
	if got := resolveWorktreeLaunchTarget(fx.root, abs); got != abs {
		t.Errorf("resolveWorktreeLaunchTarget(%q) = %q, want the absolute path unchanged", abs, got)
	}
}

// TestLauncherAnchorBackendWorktreeNameNonLeader proves the non-leader launch
// of `-w -` is untouched: the guard is a leader-only reader of the selection,
// so the launch proceeds even where the tree the backend would enter is a card
// tree.
func TestLauncherAnchorBackendWorktreeNameNonLeader(t *testing.T) {
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		t.Run(arm, func(t *testing.T) {
			fx := newLeaderFixture(t)
			addBackendNamedTree(t, fx, "-", "WT-dash")
			obs := driveLeaderLaunch(t, fx, arm, []string{"--", "-w", "-"})
			if obs.err != nil || !obs.launched {
				t.Errorf("non-leader `-w -` launch: err=%v launched=%v, want allowed", obs.err, obs.launched)
			}
		})
	}
}

// TestLauncherAnchorBackendWorktreeNameSlashIgnoresOtherRoots proves a value
// containing a slash is judged at the native .claude/worktrees directory only.
// The real launch never enters a home-root or .moai tree for such a name
// (existingWorktreeEntryDir rejects it), so a card-branch tree of the mapped
// name under another root must not condemn a launch the backend sends to
// .claude/worktrees/feature+foo.
func TestLauncherAnchorBackendWorktreeNameSlashIgnoresOtherRoots(t *testing.T) {
	pinSnapshotModel(t)
	assertOtherRootsIgnored(t, "feature/foo", "feature+foo")
}

// TestLauncherAnchorBackendWorktreeNameDashIgnoresOtherRoots is the same for a
// lone `-`: existingWorktreeEntryDir reads `-w -` as no value, so the launch
// enters .claude/worktrees/- and a card tree named `-` under another root must
// not condemn it.
func TestLauncherAnchorBackendWorktreeNameDashIgnoresOtherRoots(t *testing.T) {
	pinSnapshotModel(t)
	assertOtherRootsIgnored(t, "-", "-")
}

// TestLauncherAnchorBackendWorktreeNameFromLinkedWorktree launches the leader
// from a linked worktree on a plain branch. Claude 2.1.309 creates and enters
// `-w <name>` under the MAIN repository's .claude/worktrees (the startup path
// resolves the canonical git root, not the launch tree's top level), so a WT-
// tree there is the anchor and a plain one is not.
func TestLauncherAnchorBackendWorktreeNameFromLinkedWorktree(t *testing.T) {
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		for _, tc := range []struct {
			value, dir string
		}{{"feature/foo", "feature+foo"}, {"-", "-"}} {
			for _, branch := range []string{"WT-main-side", "plain-main-side"} {
				t.Run(arm+"/"+tc.dir+"/"+branch, func(t *testing.T) {
					fx := newLeaderFixture(t)
					tree := addBackendNamedTree(t, fx, tc.dir, branch)
					t.Chdir(filepath.Join(fx.root, ".claude", "worktrees", "plain"))
					obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", tc.value})
					if strings.HasPrefix(branch, "WT-") {
						assertRefusedBeforeMutation(t, tc.dir, obs)
						refusalNames(t, tc.dir, obs, tree, branch)
						return
					}
					if obs.err != nil || !obs.launched {
						t.Errorf("a plain tree under the main repository must launch: err=%v launched=%v", obs.err, obs.launched)
					}
				})
			}
		}
	}
}

// assertOtherRootsIgnored launches `-f -- -w <value>` on both arms with a
// WT- tree named dir under the home root or the .moai root and the native
// .claude/worktrees/<dir> on a plain branch or absent: every launch proceeds.
func assertOtherRootsIgnored(t *testing.T, value, dir string) {
	t.Helper()
	for _, arm := range leaderArms {
		for _, other := range []string{"home", "legacy-moai"} {
			for _, native := range []string{"plain branch", "absent"} {
				t.Run(arm+"/"+other+"/native "+native, func(t *testing.T) {
					fx := newLeaderFixture(t)
					otherPath := filepath.Join(fx.root, ".moai", "worktrees", dir)
					if other == "home" {
						var err error
						if otherPath, err = worktreeroot.PathFor(fx.root, dir); err != nil {
							t.Fatalf("resolve the home-root destination: %v", err)
						}
					}
					anchorGit(t, fx.root, "worktree", "add", "-q", "-b", "WT-other-root", otherPath, fx.base)
					if native == "plain branch" {
						addBackendNamedTree(t, fx, dir, "plain-native")
					}
					obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", value})
					if obs.err != nil || !obs.launched {
						t.Errorf("a card tree under the %s root must not condemn a launch into .claude/worktrees/%s: err=%v launched=%v", other, dir, obs.err, obs.launched)
					}
				})
			}
		}
	}
}

// backendRootValues are the two values the guard resolves to the native
// directory without the root lookup.
var backendRootValues = []struct{ value, dir string }{{"feature/foo", "feature+foo"}, {"-", "-"}}

// assertMainRootRefused launches the leader from a linked worktree on a plain
// branch, with a WT- tree at the MAIN repository's .claude/worktrees/<dir>
// (where claude 2.1.309 enters), after setup has disturbed the environment: the
// launch must be refused.
func assertMainRootRefused(t *testing.T, setup func(t *testing.T, fx *leaderFixture)) {
	t.Helper()
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		for _, v := range backendRootValues {
			t.Run(arm+"/"+v.dir, func(t *testing.T) {
				fx := newLeaderFixture(t)
				tree := addBackendNamedTree(t, fx, v.dir, "WT-native")
				t.Chdir(filepath.Join(fx.root, ".claude", "worktrees", "plain"))
				setup(t, fx)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", v.value})
				assertRefusedBeforeMutation(t, v.dir, obs)
				refusalNames(t, v.dir, obs, tree, "WT-native")
			})
		}
	}
}

// TestLauncherAnchorBackendWorktreeNameResolveFailure: the backend enters the
// main root whatever MOAI_HOME holds, so a worktreeroot.Resolve that fails
// (MOAI_HOME/worktrees is a regular file) must not change the guard's answer.
func TestLauncherAnchorBackendWorktreeNameResolveFailure(t *testing.T) {
	assertMainRootRefused(t, func(t *testing.T, _ *leaderFixture) {
		if err := os.WriteFile(filepath.Join(os.Getenv("MOAI_HOME"), "worktrees"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := worktreeroot.Resolve("."); err == nil {
			t.Fatal("the fixture no longer makes worktreeroot.Resolve fail")
		}
	})
}

// TestLauncherAnchorBackendWorktreeNameInheritedGitDir: an inherited GIT_DIR
// outranks `git -C`; the root lookup must scrub it like the rest of the guard
// (anchorBranchOfTree, shortWorktreeRoot), and the backend, which walks the
// file system, is unaffected by it.
func TestLauncherAnchorBackendWorktreeNameInheritedGitDir(t *testing.T) {
	assertMainRootRefused(t, func(t *testing.T, _ *leaderFixture) {
		other, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		anchorGit(t, other, "init", "-q")
		t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	})
}

// TestLauncherAnchorBackendWorktreeNameSeparateGitDir: in a `git init
// --separate-git-dir` checkout the common dir is not named `.git`, so it has no
// "main checkout above it". Claude 2.1.309 reads the checkout's own `.git` file,
// finds no commondir and uses the checkout root itself; the guard falls back to
// the launch tree's top level, not the parent of the metadata directory.
func TestLauncherAnchorBackendWorktreeNameSeparateGitDir(t *testing.T) {
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		for _, v := range backendRootValues {
			t.Run(arm+"/"+v.dir, func(t *testing.T) {
				fx := newSeparateGitDirFixture(t, filepath.Join("meta", "store"))
				tree := addBackendNamedTree(t, fx, v.dir, "WT-native")
				t.Chdir(fx.root)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", v.value})
				assertRefusedBeforeMutation(t, v.dir, obs)
				refusalNames(t, v.dir, obs, tree, "WT-native")
			})
		}
	}
}

// newSeparateGitDirFixture creates tmp/proj with its git directory at tmp/<meta>
// (`git init --separate-git-dir`), one seed commit, and the project environment
// the leader harness reads. The returned fixture's root is tmp/proj.
func newSeparateGitDirFixture(t *testing.T, meta string) *leaderFixture {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(tmp, meta)), 0o755); err != nil {
		t.Fatal(err)
	}
	anchorGit(t, tmp, "init", "-q", "--separate-git-dir", filepath.Join(tmp, meta), repo)
	anchorGit(t, repo, "config", "user.email", "anchor-test@example.com")
	anchorGit(t, repo, "config", "user.name", "Anchor Test")
	anchorGit(t, repo, "config", "commit.gpgsign", "false")
	if err := os.MkdirAll(filepath.Join(repo, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	anchorGit(t, repo, "add", "README.md")
	anchorGit(t, repo, "commit", "-q", "-m", "seed")
	t.Setenv(config.EnvClaudeProjectDir, repo)
	withCodexProjectRoot(t, repo)
	t.Setenv("MOAI_HOME", t.TempDir())
	netScrubLaneEnv(t)
	return &leaderFixture{root: repo, base: anchorGit(t, repo, "branch", "--show-current")}
}

// TestLauncherAnchorBackendWorktreeNameSeparateGitDirNamedDotGit: a NON-linked
// checkout whose separate git directory happens to be named `.git`. Claude
// 2.1.309 decides "linked" from the checkout's gitdir/commondir pair, not from
// the directory name: with no commondir file it uses the checkout root, so the
// metadata directory's parent is not the root.
func TestLauncherAnchorBackendWorktreeNameSeparateGitDirNamedDotGit(t *testing.T) {
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		for _, v := range backendRootValues {
			t.Run(arm+"/"+v.dir, func(t *testing.T) {
				fx := newSeparateGitDirFixture(t, filepath.Join("meta", ".git"))
				tree := addBackendNamedTree(t, fx, v.dir, "WT-native")
				t.Chdir(fx.root)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", v.value})
				assertRefusedBeforeMutation(t, v.dir, obs)
				refusalNames(t, v.dir, obs, tree, "WT-native")
			})
		}
	}
}

// assertCommonDirRoot launches the leader from the linked worktree lt (on a
// plain branch) with a WT- tree at <common>/.claude/worktrees/<dir>: for a
// linked worktree whose common directory is not named `.git` and holds no
// `.git` of its own, Claude 2.1.309 returns the common directory itself as the
// root, so the launch must be refused.
func assertCommonDirRoot(t *testing.T, host, common, lt, base string) {
	t.Helper()
	pinSnapshotModel(t)
	if err := os.MkdirAll(filepath.Join(lt, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	trees := map[string]string{}
	for _, v := range backendRootValues {
		trees[v.dir] = filepath.Join(common, ".claude", "worktrees", v.dir)
		anchorGit(t, host, "worktree", "add", "-q", "-b", "WT-native-"+strings.ReplaceAll(v.dir, "+", "-"), trees[v.dir], base)
	}
	for _, arm := range leaderArms {
		for _, v := range backendRootValues {
			t.Run(arm+"/"+v.dir, func(t *testing.T) {
				t.Setenv(config.EnvClaudeProjectDir, lt)
				withCodexProjectRoot(t, lt)
				t.Chdir(lt)
				fx := &leaderFixture{root: lt, base: base}
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", v.value})
				assertRefusedBeforeMutation(t, v.dir, obs)
				refusalNames(t, v.dir, obs, trees[v.dir], "WT-native-")
			})
		}
	}
}

// TestLauncherAnchorBackendWorktreeNameBareLinkedWorktree: a linked worktree of
// a bare repository (common dir `<tmp>/bare.git`).
func TestLauncherAnchorBackendWorktreeNameBareLinkedWorktree(t *testing.T) {
	seed := newAnchorRepo(t)
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bare := filepath.Join(tmp, "bare.git")
	anchorGit(t, tmp, "clone", "-q", "--bare", seed.root, bare)
	lt := filepath.Join(tmp, "lt")
	anchorGit(t, bare, "worktree", "add", "-q", lt, seed.base)
	t.Setenv("MOAI_HOME", t.TempDir())
	netScrubLaneEnv(t)
	assertCommonDirRoot(t, bare, bare, lt, seed.base)
}

// TestLauncherAnchorBackendWorktreeNameSeparateGitDirLinkedWorktree: a linked
// worktree of a separate-git-dir repository (common dir `<tmp>/meta/store`).
func TestLauncherAnchorBackendWorktreeNameSeparateGitDirLinkedWorktree(t *testing.T) {
	fx := newSeparateGitDirFixture(t, filepath.Join("meta", "store"))
	lt := filepath.Join(filepath.Dir(fx.root), "lt")
	anchorGit(t, fx.root, "worktree", "add", "-q", "-b", "plain-lt", lt, fx.base)
	assertCommonDirRoot(t, fx.root, filepath.Join(filepath.Dir(fx.root), "meta", "store"), lt, fx.base)
}
