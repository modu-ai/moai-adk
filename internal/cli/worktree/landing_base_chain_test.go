package worktree

// landing_base_chain_test.go — SPEC-GITHUB-FLOW-CI-RESIDUE-001 M1,
// REQ-GFC-001..004 / AC-GFC-001..004. The worktree landing surfaces (`sweep`
// default --base and `done`'s origin-landing check) resolve their base through
// the SAME chain the todo surface answers from
// (factory.LandedRefForWithLevel, SPEC-TODO-LANDING-ATTRIBUTION-001):
//
//	1. git_strategy.worktree_base_branch (configured — this repository: main)
//	2. refs/remotes/origin/HEAD (the branch the repository itself records)
//	3. the compiled-in default (origin/main)
//
// Every fixture is a REAL scratch repository under t.TempDir() with a LOCAL
// bare remote standing in for origin; gh never runs (the package-level
// failing double stays installed). The symref is controlled explicitly with
// `git symbolic-ref` — push -u and worktree add do NOT create origin/HEAD
// (measured 2026-10-07, git 2.54), so a fixture that wants level 2 sets it
// and a fixture that wants level 3 deletes it.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// chainFixture is one sandbox: bare origin carrying BOTH main and develop,
// a repository cloned-shaped on main, and a card worktree on a WT- branch
// with one unlanded commit.
type chainFixture struct {
	base   string // the t.TempDir() root
	origin string // the bare remote
	repo   string // the working repository (main)
	tree   string // the card worktree
	branch string // the card branch, WT- prefixed
}

func newChainFixture(t *testing.T) *chainFixture {
	t.Helper()
	base := t.TempDir()
	f := &chainFixture{
		base:   base,
		origin: filepath.Join(base, "origin.git"),
		repo:   filepath.Join(base, "repo"),
		tree:   filepath.Join(base, "card-wt"),
		branch: "WT-chain-card",
	}
	chainGit(t, base, "init", "-q", "--bare", f.origin)
	if err := os.MkdirAll(f.repo, 0o755); err != nil {
		t.Fatal(err)
	}
	chainGit(t, f.repo, "init", "-q", "-b", "main")
	chainGit(t, f.repo, "config", "user.email", "chain-test@example.com")
	chainGit(t, f.repo, "config", "user.name", "Chain Test")
	chainGit(t, f.repo, "config", "commit.gpgsign", "false")
	chainGit(t, f.repo, "config", "advice.detachedHead", "false")
	chainGit(t, f.repo, "remote", "add", "origin", f.origin)
	if err := os.WriteFile(filepath.Join(f.repo, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	chainGit(t, f.repo, "add", ".")
	chainGit(t, f.repo, "commit", "-q", "-m", "seed")
	chainGit(t, f.repo, "push", "-q", "origin", "main")
	// develop exists on the remote too, so a test that lands on develop can —
	// and a test that wants the develop-absent world simply never merges there.
	chainGit(t, f.repo, "branch", "develop")
	chainGit(t, f.repo, "push", "-q", "origin", "develop")
	chainGit(t, f.repo, "worktree", "add", "-q", "-b", f.branch, f.tree)
	if err := os.WriteFile(filepath.Join(f.tree, "work.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	chainGit(t, f.tree, "add", ".")
	chainGit(t, f.tree, "commit", "-q", "-m", "card work")
	withTierTestEnv(t, f.repo)
	return f
}

// chainGit runs a git subcommand in dir, failing the test on error.
func chainGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// chainBody is the git-strategy.yaml body used for the configured level: the
// shape THIS repository ships (git-flow flow config plus worktree_base_branch:
// main) — the exact case REQ-GFC-001 must flip to origin/main.
const chainBody = "git_strategy:\n    mode: manual\n    worktree_base_branch: main\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n"

// chainGitFlowBody is a git-flow project WITHOUT worktree_base_branch — the
// compatibility shape: level 2 (origin/HEAD) must keep answering origin/develop.
const chainGitFlowBody = "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n"

// installChainBody writes body as the repo's git-strategy.yaml. An empty body
// removes the file — the unconfigured level.
func (f *chainFixture) installChainBody(t *testing.T, body string) {
	t.Helper()
	installBaseRow(t, f.repo, baseRow{body: body, absent: body == ""})
}

// setOriginHEAD points refs/remotes/origin/HEAD at the given remote branch
// (level 2). clearOriginHEAD deletes it (level 3).
func (f *chainFixture) setOriginHEAD(t *testing.T, branch string) {
	t.Helper()
	chainGit(t, f.repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/"+branch)
}

func (f *chainFixture) clearOriginHEAD(t *testing.T) {
	t.Helper()
	cmd := exec.Command("git", "symbolic-ref", "-d", "refs/remotes/origin/HEAD")
	cmd.Dir = f.repo
	if out, err := cmd.CombinedOutput(); err != nil && !strings.Contains(string(out), "not a symbolic ref") {
		t.Fatalf("delete origin/HEAD: %v\n%s", err, out)
	}
}

// landCardOn merges the card branch into local branch (no-ff, the lane
// discipline) and pushes that branch to origin. The repo sits on main, so
// the merge must CHECK OUT the target branch first — `git merge` always
// merges into the current branch.
func (f *chainFixture) landCardOn(t *testing.T, branch string) {
	t.Helper()
	chainGit(t, f.repo, "checkout", "-q", branch)
	chainGit(t, f.repo, "merge", "--no-ff", "-q", "-m", "merge card into "+branch, f.branch)
	chainGit(t, f.repo, "push", "-q", "origin", branch)
	chainGit(t, f.repo, "checkout", "-q", "main")
}

// sweepDefaultBaseFor runs the sweep WITHOUT --base against the fixture repo
// and returns the bases the fetch seam observed plus the run error.
func (f *chainFixture) sweepDefaultBaseFor(t *testing.T) (fetched []string, err error) {
	t.Helper()
	m := sweepMockEnv(t, nil)
	sweepConfigRoot = func() string { return f.repo }
	_, err = runSweepCmd(t, map[string]string{})
	return m.fetchedBases, err
}

// TestSweepDefaultBaseFollowsLandedRefChain (AC-GFC-001/002): the sweep's
// default --base resolves through the landed-ref chain, level by level.
func TestSweepDefaultBaseFollowsLandedRefChain(t *testing.T) {
	t.Run("level1_configured_worktree_base_branch_wins", func(t *testing.T) {
		f := newChainFixture(t)
		f.installChainBody(t, chainBody)
		fetched, err := f.sweepDefaultBaseFor(t)
		if err != nil {
			t.Fatalf("sweep with a configured worktree_base_branch: %v", err)
		}
		if len(fetched) != 1 || fetched[0] != "origin/main" {
			t.Fatalf("configured worktree_base_branch: main must answer origin/main, observed %v", fetched)
		}
	})

	t.Run("level2_origin_HEAD_answers_when_unconfigured", func(t *testing.T) {
		f := newChainFixture(t)
		f.installChainBody(t, "")
		f.setOriginHEAD(t, "develop")
		fetched, err := f.sweepDefaultBaseFor(t)
		if err != nil {
			t.Fatalf("sweep with origin/HEAD: %v", err)
		}
		if len(fetched) != 1 || fetched[0] != "origin/develop" {
			t.Fatalf("origin/HEAD -> develop must answer origin/develop, observed %v", fetched)
		}
	})

	t.Run("level3_compiled_in_default_when_both_absent", func(t *testing.T) {
		f := newChainFixture(t)
		f.installChainBody(t, "")
		f.clearOriginHEAD(t)
		fetched, err := f.sweepDefaultBaseFor(t)
		if err != nil {
			t.Fatalf("sweep with neither level: %v", err)
		}
		if len(fetched) != 1 || fetched[0] != "origin/main" {
			t.Fatalf("the compiled-in default must answer origin/main, observed %v", fetched)
		}
	})
}

// TestDoneLandingBaseFollowsLandedRefChain (AC-GFC-001/002/004): done's
// origin-landing check asks the same chain — and the answered base decides
// the disposal, with the provenance disclosed on the refusal.
func TestDoneLandingBaseFollowsLandedRefChain(t *testing.T) {
	t.Run("level1_configured_main_decides_the_verdict", func(t *testing.T) {
		f := newChainFixture(t)
		f.installChainBody(t, chainBody)
		// Landed ONLY on origin/develop: refused, naming the base the chain
		// answered (origin/main) and the level it answered from.
		f.landCardOn(t, "develop")
		err := executeDoneForTierGuard(t, "--auto", f.branch)
		if err == nil || !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") {
			t.Fatalf("a card absent from the configured base must be refused, got: %v", err)
		}
		if !strings.Contains(err.Error(), "origin/main") || !strings.Contains(err.Error(), "git_strategy.worktree_base_branch") {
			t.Errorf("refusal must name origin/main from git_strategy.worktree_base_branch, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)

		// Landed on origin/main: disposed — the configured key is what decides.
		g := newChainFixture(t)
		g.installChainBody(t, chainBody)
		g.landCardOn(t, "main")
		if err := executeDoneForTierGuard(t, "--auto", g.branch); err != nil {
			t.Fatalf("a card landed on the configured base must dispose, got: %v", err)
		}
	})

	t.Run("level2_origin_HEAD_compatibility_keeps_git_flow_projects_working", func(t *testing.T) {
		// git-flow flow config, NO worktree_base_branch, origin/HEAD -> develop:
		// still origin/develop (design D-1 compatibility clause).
		f := newChainFixture(t)
		f.installChainBody(t, chainGitFlowBody)
		f.setOriginHEAD(t, "develop")
		f.landCardOn(t, "develop")
		if err := executeDoneForTierGuard(t, "--auto", f.branch); err != nil {
			t.Fatalf("git-flow compatibility: a card landed on origin/develop must dispose, got: %v", err)
		}
	})

	t.Run("level3_default_answers_when_origin_never_had_develop", func(t *testing.T) {
		// AC-GFC-004: the develop remote is GONE and the chain answers
		// origin/main — a landed card disposes; there is no PRESERVE-everything
		// and no blanket done refusal.
		f := newChainFixture(t)
		f.installChainBody(t, "")
		f.clearOriginHEAD(t)
		chainGit(t, f.repo, "push", "-q", "origin", "--delete", "develop")
		f.landCardOn(t, "main")
		if err := executeDoneForTierGuard(t, "--auto", f.branch); err != nil {
			t.Fatalf("AC-GFC-004: a landed card must dispose with develop gone (base origin/main), got: %v", err)
		}
	})

	t.Run("fetch_failure_of_the_resolved_base_stays_fail_closed", func(t *testing.T) {
		// AC-GFC-003: the three-way contract does not weaken — the chain may
		// always name a base, but a base whose fetch fails still refuses.
		f := newChainFixture(t)
		f.installChainBody(t, "")
		f.clearOriginHEAD(t)
		f.landCardOn(t, "main")
		chainGit(t, f.repo, "remote", "set-url", "origin", filepath.Join(f.base, "missing.git"))
		err := executeDoneForTierGuard(t, "--auto", f.branch)
		if err == nil || !strings.Contains(err.Error(), "ORIGIN_LANDING_UNCONFIRMED") {
			t.Fatalf("a failing fetch of the resolved base must refuse fail-closed, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)
	})
}

// TestLandingBaseDisclosesItsChainLevel (AC-GFC-002): landingBase returns the
// bare branch beside the chain level that answered it, in the disclosure
// vocabulary the refusal text prints.
func TestLandingBaseDisclosesItsChainLevel(t *testing.T) {
	cells := []struct {
		name        string
		body        string
		headBranch  string // "" = delete the symref
		wantBase    string
		wantSources []string
	}{
		{
			name:        "level1_configured",
			body:        chainBody,
			wantBase:    "main",
			wantSources: []string{"git_strategy.worktree_base_branch"},
		},
		{
			name:        "level2_origin_HEAD",
			body:        "",
			headBranch:  "develop",
			wantBase:    "develop",
			wantSources: []string{"refs/remotes/origin/HEAD"},
		},
		{
			name:        "level3_default",
			body:        "",
			wantBase:    "main",
			wantSources: []string{"the compiled-in default"},
		},
	}
	for _, tc := range cells {
		t.Run(tc.name, func(t *testing.T) {
			f := newChainFixture(t)
			f.installChainBody(t, tc.body)
			switch {
			case tc.headBranch != "":
				f.setOriginHEAD(t, tc.headBranch)
			default:
				f.clearOriginHEAD(t)
			}
			base, provenance, err := landingBase(f.tree)
			if err != nil {
				t.Fatalf("landingBase: %v", err)
			}
			if base != tc.wantBase {
				t.Errorf("base = %q, want %q", base, tc.wantBase)
			}
			for _, want := range tc.wantSources {
				if !strings.Contains(provenance, want) {
					t.Errorf("provenance %q does not name %q", provenance, want)
				}
			}
		})
	}
}

func assertContainsAll(t *testing.T, what, got string, wants []string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Errorf("%s must contain %q, got:\n%s", what, w, got)
		}
	}
}

// TestEmptyTargetRowsFollowTheLandedRefChain: every git-strategy.yaml shape
// that used to leave the integration target EMPTY — and made both landing
// surfaces REFUSE — now resolves through the landed-ref chain (design D-1.1:
// the chain always answers, the no-target error branch is gone). The sweep
// fetches the chain's answer (origin/main in these bare fixture roots)
// instead of refusing before any fetch.
func TestEmptyTargetRowsFollowTheLandedRefChain(t *testing.T) {
	executed := 0
	for _, row := range baseRows() {
		if row.want != "" {
			continue
		}
		t.Run("sweep/"+row.name, func(t *testing.T) {
			executed++
			fetched, err := sweepDefaultBaseFor(t, row)
			if err != nil {
				t.Fatalf("the chain must answer where the interpretation table refused: %v", err)
			}
			if len(fetched) != 1 || fetched[0] != "origin/main" {
				t.Fatalf("the unconfigured chain must fall through to origin/main, observed %v", fetched)
			}
		})
	}
	if executed == 0 {
		t.Fatal("empty sweep: no empty-target row was executed")
	}
}

// TestMergeNotOnOriginNamesItsChainLevel: the MERGE_NOT_ON_ORIGIN refusal
// says which chain level made the landing base what it is, so a mismatched
// base reads as a resolution, not as an unmerged card (REQ-GFC-002).
func TestMergeNotOnOriginNamesItsChainLevel(t *testing.T) {
	cells := []struct {
		name       string
		body       string
		headBranch string // "" = delete the symref
		landOn     string // the branch the card is merged to (NOT the resolved base)
		want       []string
	}{
		{
			name:       "level2_symref_provenance",
			body:       chainGitFlowBody,
			headBranch: "develop",
			landOn:     "main",
			want:       []string{"landing base origin/develop from refs/remotes/origin/HEAD"},
		},
		{
			name:   "level1_configured_provenance",
			body:   chainBody,
			landOn: "develop",
			want:   []string{"landing base origin/main from git_strategy.worktree_base_branch"},
		},
	}
	for _, tc := range cells {
		t.Run(tc.name, func(t *testing.T) {
			f := newChainFixture(t)
			f.installChainBody(t, tc.body)
			if tc.headBranch != "" {
				f.setOriginHEAD(t, tc.headBranch)
			} else {
				f.clearOriginHEAD(t)
			}
			f.landCardOn(t, tc.landOn)
			err := executeDoneForTierGuard(t, "--auto", f.branch)
			if err == nil || !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") {
				t.Fatalf("done must refuse with MERGE_NOT_ON_ORIGIN, got: %v", err)
			}
			assertContainsAll(t, "the MERGE_NOT_ON_ORIGIN refusal", err.Error(), tc.want)
			assertTreeSurvives(t, f.tree)
		})
	}
}
