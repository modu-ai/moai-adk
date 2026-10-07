package cli

// session_worktree_landing_squash_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-A
// (card t1453), AC-GFD-002 / REQ-GFD-002 second clause: the SESSION-EXIT landing
// predicate answers from layers 1-2 only (ancestry, cumulative patch-id) against
// the CONFIGURED integration target, and makes no network call — in particular no
// `gh` call. Real scratch repositories; a `gh` shim first on PATH records any
// invocation, so a mutant that runs the PR layer on this path is caught at the
// process boundary, not by a mocked seam.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const gfdSessGitHubFlowYAML = "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n"
const gfdSessGitFlowYAML = "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n"

type gfdSessFixture struct {
	repo, tree, origin string
	integration        string // the branch the fixture integrates into
}

func gfdSessGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// gfdSessRepo builds bare origin + a repo on `integration` (pushed) + a card
// worktree on a WT- branch. yaml == "" leaves the project without any
// git-strategy configuration (the empty-target case).
func gfdSessRepo(t *testing.T, integration, yaml string) *gfdSessFixture {
	t.Helper()
	tmp := canonicalTempDir(t)
	f := &gfdSessFixture{
		repo:        filepath.Join(tmp, "repo"),
		tree:        filepath.Join(tmp, "wt"),
		origin:      filepath.Join(tmp, "origin.git"),
		integration: integration,
	}
	if err := os.MkdirAll(f.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gfdSessGit(t, tmp, "init", "-q", "--bare", f.origin)
	gfdSessGit(t, f.repo, "init", "-q", "-b", integration)
	gfdSessGit(t, f.repo, "config", "user.email", "t@example.com")
	gfdSessGit(t, f.repo, "config", "user.name", "t")
	gfdSessGit(t, f.repo, "config", "commit.gpgsign", "false")
	var seed strings.Builder
	for i := 1; i <= 20; i++ {
		seed.WriteString("line " + string(rune('A'+i-1)) + "\n")
	}
	if err := os.WriteFile(filepath.Join(f.repo, "f.txt"), []byte(seed.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	gfdSessGit(t, f.repo, "add", "f.txt")
	gfdSessGit(t, f.repo, "commit", "-q", "-m", "seed")
	gfdSessGit(t, f.repo, "remote", "add", "origin", f.origin)
	gfdSessGit(t, f.repo, "push", "-q", "-u", "origin", integration)
	gfdSessGit(t, f.repo, "worktree", "add", "-q", "-b", "WT-sess-card", f.tree)
	if yaml != "" {
		dir := filepath.Join(f.repo, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

// editLine appends " <suffix>" to line n (1-based, letters A.. as content) of dir/f.txt.
func gfdSessEdit(t *testing.T, dir string, n int, suffix string) {
	t.Helper()
	p := filepath.Join(dir, "f.txt")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	lines[n-1] += " " + suffix
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *gfdSessFixture) cardCommit(t *testing.T, line int, suffix string) {
	t.Helper()
	gfdSessEdit(t, f.tree, line, suffix)
	gfdSessGit(t, f.tree, "commit", "-q", "-a", "-m", "card "+suffix)
}

func (f *gfdSessFixture) mainCommit(t *testing.T, line int, suffix string) {
	t.Helper()
	gfdSessEdit(t, f.repo, line, suffix)
	gfdSessGit(t, f.repo, "commit", "-q", "-a", "-m", "main "+suffix)
}

func (f *gfdSessFixture) squash(t *testing.T) {
	t.Helper()
	gfdSessGit(t, f.repo, "merge", "--squash", "WT-sess-card")
	gfdSessGit(t, f.repo, "commit", "-q", "-m", "squash card")
}

func (f *gfdSessFixture) push(t *testing.T) {
	t.Helper()
	gfdSessGit(t, f.repo, "push", "-q", "origin", f.integration)
}

// withGHShim puts a `gh` first on PATH that records every invocation and fails,
// and returns the marker path. An empty marker file == gh was never run.
func withGHShim(t *testing.T) string {
	t.Helper()
	bin := t.TempDir()
	marker := filepath.Join(bin, "gh-was-called")
	script := "#!/bin/sh\necho \"$@\" >> '" + marker + "'\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return marker
}

func ghWasCalled(t *testing.T, marker string) bool {
	t.Helper()
	_, err := os.Stat(marker)
	return err == nil
}

// TestSessionExitLandingNoNetwork is AC-GFD-002's session-exit half. F1-F3 are
// "landed" from layers 1-2; F5 — which only the PR layer can confirm — is
// "preserve" (the next sweep's layer 3 decides); no cell runs `gh`.
func TestSessionExitLandingNoNetwork(t *testing.T) {
	marker := withGHShim(t)

	cells := []struct {
		name  string
		build func(t *testing.T, f *gfdSessFixture)
		want  bool
	}{
		{"F1_one_commit_squash", func(t *testing.T, f *gfdSessFixture) {
			f.cardCommit(t, 5, "card")
			f.squash(t)
			f.push(t)
		}, true},
		{"F2_two_commit_squash", func(t *testing.T, f *gfdSessFixture) {
			f.cardCommit(t, 5, "card")
			f.cardCommit(t, 15, "card")
			f.squash(t)
			f.push(t)
		}, true},
		{"F3_merge_commit", func(t *testing.T, f *gfdSessFixture) {
			f.cardCommit(t, 5, "card")
			gfdSessGit(t, f.repo, "merge", "--no-ff", "-q", "-m", "merge card", "WT-sess-card")
			f.push(t)
		}, true},
		{"F4_unmerged", func(t *testing.T, f *gfdSessFixture) {
			f.cardCommit(t, 5, "card")
		}, false},
		{"F5_squash_context_drifted_only_layer_3_confirms", func(t *testing.T, f *gfdSessFixture) {
			f.cardCommit(t, 10, "card")
			f.mainCommit(t, 13, "other") // main touches a nearby line BEFORE the squash
			f.squash(t)
			f.mainCommit(t, 10, "later") // and the same region again AFTER it
			f.push(t)
		}, false},
	}
	for _, c := range cells {
		t.Run(c.name, func(t *testing.T) {
			f := gfdSessRepo(t, "main", gfdSessGitHubFlowYAML)
			c.build(t, f)
			got, err := gitBranchLandedReal(f.tree)
			if err != nil {
				t.Fatalf("landing check errored: %v", err)
			}
			if got != c.want {
				t.Errorf("landed = %v, want %v", got, c.want)
			}
			if ghWasCalled(t, marker) {
				t.Fatal("the session-exit landing predicate ran gh (REQ-GFD-002 second clause: no network call)")
			}
		})
	}
}

// TestSessionExitLandingFollowsIntegrationTarget: the integration ref is the
// configured target, not a literal develop (REQ-GFD-003 / M2(c)). A card landed
// on origin/develop under a github-flow project is NOT landed; and under git-flow
// the develop-based answer is unchanged.
func TestSessionExitLandingFollowsIntegrationTarget(t *testing.T) {
	t.Run("git_flow_develop_unchanged", func(t *testing.T) {
		f := gfdSessRepo(t, "develop", gfdSessGitFlowYAML)
		f.cardCommit(t, 5, "card")
		gfdSessGit(t, f.repo, "merge", "--no-ff", "-q", "-m", "merge card", "WT-sess-card")
		f.push(t)
		got, err := gitBranchLandedReal(f.tree)
		if err != nil || !got {
			t.Fatalf("git-flow: landed on origin/develop = (%v, %v), want (true, nil)", got, err)
		}
	})
	t.Run("github_flow_ignores_a_develop_only_landing", func(t *testing.T) {
		f := gfdSessRepo(t, "main", gfdSessGitHubFlowYAML)
		f.cardCommit(t, 5, "card")
		// land on a develop branch the configuration does not name
		gfdSessGit(t, f.repo, "branch", "develop")
		gfdSessGit(t, f.repo, "worktree", "add", "-q", filepath.Join(filepath.Dir(f.repo), "dev"), "develop")
		dev := filepath.Join(filepath.Dir(f.repo), "dev")
		gfdSessGit(t, dev, "merge", "--no-ff", "-q", "-m", "merge card", "WT-sess-card")
		gfdSessGit(t, dev, "push", "-q", "origin", "develop")
		got, err := gitBranchLandedReal(f.tree)
		if err == nil && got {
			t.Fatal("github-flow: a landing on origin/develop must not read as landed on the configured target (main)")
		}
	})
	t.Run("empty_target_preserves_and_names_the_key", func(t *testing.T) {
		f := gfdSessRepo(t, "develop", "")
		f.cardCommit(t, 5, "card")
		gfdSessGit(t, f.repo, "merge", "--no-ff", "-q", "-m", "merge card", "WT-sess-card")
		f.push(t)
		got, err := gitBranchLandedReal(f.tree)
		if got {
			t.Fatal("no integration target configured: the predicate must never answer landed")
		}
		if err == nil {
			t.Fatal("no integration target configured: want an error the caller preserves on, got nil")
		}
		if !strings.Contains(err.Error(), "git_strategy") {
			t.Errorf("the error must name the config key to set (M1b guidance), got: %v", err)
		}
	})
}
