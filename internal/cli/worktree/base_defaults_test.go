package worktree

// base_defaults_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1 (card t1453),
// AC-GFD-001 / AC-GFD-003. The remote integration base the landing machinery
// compares against — `moai worktree sweep`'s --base default (ledger row E-02)
// and `moai worktree done`'s origin-landing check (row E-01) — follows the
// configured integration target, the D2 interpretation table behind
// config.LoadGitFlowIntegrationConfig, instead of a literal develop.
//
//   - TestBaseDefaultsGitFlowUnchanged pins the git-flow behaviour (origin/develop)
//     and is green before and after the swap.
//   - TestBaseDefaultsFollowIntegrationTarget runs every interpretation-table row
//     through both surfaces. A row that resolves nothing is a refusal that
//     preserves the tree (sweep errors before fetching; done refuses fail-closed),
//     never a silently applied default. The done cells discriminate the chosen
//     base by WHERE the card landed: only on the row's target (disposed) versus
//     only on a different branch (refused naming the target).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// baseRow is one row of the interpretation table: a git-strategy.yaml fixture
// and the target the interpreter answers for it ("" = resolves nothing).
type baseRow struct {
	name   string
	body   string
	absent bool
	want   string
}

func baseRows() []baseRow {
	manual := func(profile string) string {
		return "git_strategy:\n    mode: manual\n    manual:\n" + profile
	}
	return []baseRow{
		{name: "github-flow", body: manual("        workflow: github-flow\n"), want: "main"},
		{name: "git-flow develop", body: manual("        workflow: git-flow\n        develop_branch: develop\n"), want: "develop"},
		{name: "git-flow custom develop branch", body: manual("        workflow: git-flow\n        develop_branch: staging\n"), want: "staging"},
		{name: "git-flow empty develop_branch", body: manual("        workflow: git-flow\n        develop_branch: \"\"\n"), want: ""},
		{name: "git-flow outside the manual gate", body: "git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: develop\n", want: ""},
		{name: "gitlab-flow environment", body: manual("        workflow: gitlab-flow\n        environment: production\n"), want: "production"},
		{name: "release-flow prefix", body: manual("        workflow: release-flow\n        release_branch_prefix: release/\n"), want: "release/"},
		{name: "unknown workflow", body: manual("        workflow: svn-flow\n"), want: ""},
		{name: "no git-strategy.yaml", absent: true, want: ""},
		{name: "unparseable git-strategy.yaml", body: "git_strategy: [unterminated\n", want: ""},
	}
}

func gitFlowBaseRow() baseRow { return baseRows()[1] }

// installBaseRow writes the row's git-strategy.yaml under root.
func installBaseRow(t *testing.T, root string, row baseRow) {
	t.Helper()
	if row.absent {
		return
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(row.body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// sweepDefaultBaseFor runs the sweep WITHOUT --base against a mock environment
// whose project root carries the row's configuration, and returns the bases the
// fetch seam observed plus the run error.
func sweepDefaultBaseFor(t *testing.T, row baseRow) (fetched []string, err error) {
	t.Helper()
	m := sweepMockEnv(t, nil)
	root := t.TempDir()
	installBaseRow(t, root, row)
	WorktreeProvider = &mockWorktreeManager{
		rootPath: root,
		listFunc: func() ([]git.Worktree, error) { return nil, nil },
	}
	_, err = runSweepCmd(t, map[string]string{})
	return m.fetchedBases, err
}

// landDoneFixtureOn pushes the card branch to origin/<branch>, the way a
// merged card reaches that remote branch.
func landDoneFixtureOn(t *testing.T, f *landingFixture, branch string) {
	t.Helper()
	landingGit(t, f.repo, "push", "-q", "origin", f.branch+":refs/heads/"+branch)
}

// TestBaseDefaultsGitFlowUnchanged pins the git-flow outputs: the sweep's
// default base is origin/develop and `done` confirms a card's landing against
// origin/develop. Green before and after the swap.
func TestBaseDefaultsGitFlowUnchanged(t *testing.T) {
	t.Run("sweep default base is origin/develop", func(t *testing.T) {
		fetched, err := sweepDefaultBaseFor(t, gitFlowBaseRow())
		if err != nil {
			t.Fatalf("sweep under git-flow: %v", err)
		}
		if len(fetched) != 1 || fetched[0] != "origin/develop" {
			t.Fatalf("sweep must fetch origin/develop exactly once, observed %v", fetched)
		}
	})

	t.Run("done refuses a card absent from origin/develop", func(t *testing.T) {
		f := newLandingFixture(t)
		installBaseRow(t, f.repo, gitFlowBaseRow())
		err := executeDoneForTierGuard(t, "--auto", f.branch)
		if err == nil || !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") || !strings.Contains(err.Error(), "origin/develop") {
			t.Fatalf("done must refuse with MERGE_NOT_ON_ORIGIN naming origin/develop, got: %v", err)
		}
		assertTreeSurvives(t, f.tree)
	})

	t.Run("done disposes a card landed on origin/develop", func(t *testing.T) {
		f := newLandingFixture(t)
		installBaseRow(t, f.repo, gitFlowBaseRow())
		f.landCard(t)
		if err := executeDoneForTierGuard(t, "--auto", f.branch); err != nil {
			t.Fatalf("a card landed on origin/develop must dispose under git-flow: %v", err)
		}
	})
}

// TestBaseDefaultsFollowIntegrationTarget: both landing surfaces resolve their
// base from the interpretation table, row by row.
func TestBaseDefaultsFollowIntegrationTarget(t *testing.T) {
	for _, row := range baseRows() {
		t.Run("sweep/"+row.name, func(t *testing.T) {
			fetched, err := sweepDefaultBaseFor(t, row)
			if row.want == "" {
				if err == nil {
					t.Fatalf("a row with no target must fail the sweep before any fetch; fetched %v", fetched)
				}
				if len(fetched) != 0 {
					t.Fatalf("no fetch may run when the base is unresolved, observed %v", fetched)
				}
				return
			}
			if err != nil {
				t.Fatalf("sweep: %v", err)
			}
			if want := "origin/" + row.want; len(fetched) != 1 || fetched[0] != want {
				t.Fatalf("sweep must fetch %s exactly once, observed %v", want, fetched)
			}
		})
	}

	t.Run("sweep/explicit --base wins over an unresolved target", func(t *testing.T) {
		m := sweepMockEnv(t, nil)
		root := t.TempDir() // no configuration at all
		WorktreeProvider = &mockWorktreeManager{
			rootPath: root,
			listFunc: func() ([]git.Worktree, error) { return nil, nil },
		}
		if _, err := runSweepCmd(t, map[string]string{"base": "origin/custom"}); err != nil {
			t.Fatalf("sweep with an explicit base: %v", err)
		}
		if len(m.fetchedBases) != 1 || m.fetchedBases[0] != "origin/custom" {
			t.Fatalf("explicit base must be used verbatim, observed %v", m.fetchedBases)
		}
	})

	for _, row := range baseRows() {
		t.Run("done/"+row.name, func(t *testing.T) {
			if row.want == "" {
				f := newLandingFixture(t)
				installBaseRow(t, f.repo, row)
				// Landed on both candidate branches: still refused, because no
				// target resolved and an unconfirmable landing is not a confirmed one.
				landDoneFixtureOn(t, f, "develop")
				landDoneFixtureOn(t, f, "main")
				err := executeDoneForTierGuard(t, "--auto", f.branch)
				if err == nil || !strings.Contains(err.Error(), "ORIGIN_LANDING_UNCONFIRMED") {
					t.Fatalf("done must refuse fail-closed when no target resolves, got: %v", err)
				}
				assertTreeSurvives(t, f.tree)
				return
			}
			if strings.HasSuffix(row.want, "/") {
				// A release prefix names no branch: nothing can be fetched.
				f := newLandingFixture(t)
				installBaseRow(t, f.repo, row)
				landDoneFixtureOn(t, f, "develop")
				err := executeDoneForTierGuard(t, "--auto", f.branch)
				if err == nil || !strings.Contains(err.Error(), "ORIGIN_LANDING_UNCONFIRMED") || !strings.Contains(err.Error(), row.want) {
					t.Fatalf("done must refuse naming the unfetchable target %q, got: %v", row.want, err)
				}
				assertTreeSurvives(t, f.tree)
				return
			}

			// Landed only on the row's target: disposed.
			f := newLandingFixture(t)
			installBaseRow(t, f.repo, row)
			landDoneFixtureOn(t, f, row.want)
			if err := executeDoneForTierGuard(t, "--auto", f.branch); err != nil {
				t.Fatalf("a card landed on origin/%s must dispose: %v", row.want, err)
			}

			// Landed only on a different branch: refused, naming the target.
			other := "develop"
			if row.want == "develop" {
				other = "main"
			}
			g := newLandingFixture(t)
			installBaseRow(t, g.repo, row)
			// The target branch exists on the remote (the seed commit) but does
			// not carry the card.
			landingGit(t, g.repo, "push", "-q", "origin", "develop:refs/heads/"+row.want)
			landDoneFixtureOn(t, g, other)
			err := executeDoneForTierGuard(t, "--auto", g.branch)
			if err == nil || !strings.Contains(err.Error(), "MERGE_NOT_ON_ORIGIN") || !strings.Contains(err.Error(), "origin/"+row.want) {
				t.Fatalf("a card not on origin/%s must be refused naming it, got: %v", row.want, err)
			}
			assertTreeSurvives(t, g.tree)
		})
	}
}
