package cli

// delivery_base_target_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M1 (card t1453),
// AC-GFD-001 / AC-GFD-003. Base-branch resolution for the card-delivery path
// (the two delivery-base readers of ledger row E-28) and for the two cli-side
// defaults (the card diff base `cardBaseBranch` and the mission contract
// `MergeTarget`, ledger row E-29) follows the configured integration target —
// the D2 interpretation table behind config.LoadGitFlowIntegrationConfig —
// instead of a hardcoded develop.
//
// Two groups per surface:
//
//   - the git-flow characterization tests (…GitFlowUnchanged) pin today's output
//     under a git-flow + develop configuration and are green on the pre-change
//     tree, so the swap provably leaves the git-flow path alone;
//   - the table-driven tests (…FollowIntegrationTarget / …ResolvesFromIntegrationTarget)
//     run every row of the interpretation table, including the rows that resolve
//     nothing (empty target key, non-git-flow profile under the manual gate,
//     absent file, unparseable file). A resolved nothing is a refusal or a
//     caller fallback at each site, never a silently applied default.
//
// This file deliberately imports nothing from the kanban package — the
// delivery-base source is compared by its wire value — so the planned package
// rename touches no file added here.

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/mission"
)

// Wire values of the delivery-base source recorded on the integration window.
const (
	baseSourceConfigWire = "config"
	baseSourceCallerWire = "caller"
)

// targetRow is one row of the interpretation table as a git-strategy.yaml
// fixture plus the target the interpreter answers for it ("" = resolves
// nothing).
type targetRow struct {
	name   string
	body   string // git-strategy.yaml content; ignored when absent
	absent bool   // no git-strategy.yaml at all
	want   string // the interpreter's answer; "" = no target
}

// interpretationRows covers every row of the D2 table plus the failure rows.
func interpretationRows() []targetRow {
	manual := func(profile string) string {
		return "git_strategy:\n    mode: manual\n    manual:\n" + profile
	}
	return []targetRow{
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

// installTargetRow writes the row's fixture under root.
func installTargetRow(t *testing.T, root string, row targetRow) {
	t.Helper()
	if row.absent {
		return
	}
	writeGitStrategyBody(t, root, row.body)
}

// seedGitFlowPrecondition models a git-flow project (card t1453): it writes the
// git-flow git-strategy.yaml (manual, git-flow, develop) under root and keeps
// it out of git status through .git/info/exclude, so fixtures whose assertions
// read the working tree stay byte-identical. It is a precondition seed — the
// integration target is now read from configuration — not an assertion change.
func seedGitFlowPrecondition(t *testing.T, root string) {
	t.Helper()
	writeGitStrategyFixture(t, root, "git-flow", "develop")
	info := filepath.Join(root, ".git", "info")
	if err := os.MkdirAll(info, 0o755); err != nil {
		return // not a git checkout: nothing to exclude from
	}
	f, err := os.OpenFile(filepath.Join(info, "exclude"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(".moai/config/\n"); err != nil {
		t.Fatal(err)
	}
}

// gitFlowRow is the characterization configuration: git-flow with develop.
func gitFlowRow() targetRow {
	for _, r := range interpretationRows() {
		if r.name == "git-flow develop" {
			return r
		}
	}
	panic("git-flow develop row missing")
}

// --- delivery base (ledger row E-28) ---

// TestDeliveryBaseGitFlowUnchanged pins the git-flow behaviour of the two
// delivery-base readers: `factory complete` resolves the configured develop
// branch with source "config", and `factory merge ready` takes it as its merge
// target. It passes before and after the swap; the shapes with no configured
// branch keep today's refusal and caller fallback.
func TestDeliveryBaseGitFlowUnchanged(t *testing.T) {
	t.Run("complete resolves the configured develop branch", func(t *testing.T) {
		root := t.TempDir()
		installTargetRow(t, root, gitFlowRow())
		branch, source := factoryResolveIntegrationBranch(root, homestate.Card{})
		if branch != "develop" || source != baseSourceConfigWire {
			t.Fatalf("factoryResolveIntegrationBranch = (%q, %q), want (develop, config)", branch, source)
		}
	})

	t.Run("complete falls back to the caller when develop_branch is empty", func(t *testing.T) {
		root := t.TempDir()
		for _, r := range interpretationRows() {
			if r.name == "git-flow empty develop_branch" {
				installTargetRow(t, root, r)
			}
		}
		branch, source := factoryResolveIntegrationBranch(root, homestate.Card{})
		if branch != "" || source != baseSourceCallerWire {
			t.Fatalf("factoryResolveIntegrationBranch = (%q, %q), want the caller fallback (\"\", caller)", branch, source)
		}
	})

	t.Run("merge ready takes the configured develop branch as its target", func(t *testing.T) {
		run, err := runMergeReadyWithRow(t, gitFlowRow())
		if err != nil {
			t.Fatalf("merge ready under git-flow: %v", err)
		}
		if run == nil || run.Develop != "develop" {
			t.Fatalf("recorded merge target = %+v, want develop", run)
		}
	})

	t.Run("merge ready refuses to guess a target when develop_branch is empty", func(t *testing.T) {
		var row targetRow
		for _, r := range interpretationRows() {
			if r.name == "git-flow empty develop_branch" {
				row = r
			}
		}
		run, err := runMergeReadyWithRow(t, row)
		if err == nil || !strings.Contains(err.Error(), "no integration branch configured") {
			t.Fatalf("err = %v, want the no-integration-branch refusal", err)
		}
		if run != nil {
			t.Fatalf("a refused resolution must record no run, got %+v", run)
		}
	})
}

// TestDeliveryBaseResolvesFromIntegrationTarget runs every interpretation-table
// row through both delivery-base readers. The reader result must equal the
// interpreter's answer for the row: main for github-flow, develop for git-flow,
// the environment and release-prefix answers for the other two flows, and no
// target (caller fallback / refusal) for every row the interpreter leaves empty.
func TestDeliveryBaseResolvesFromIntegrationTarget(t *testing.T) {
	for _, row := range interpretationRows() {
		t.Run("complete/"+row.name, func(t *testing.T) {
			root := t.TempDir()
			installTargetRow(t, root, row)
			branch, source := factoryResolveIntegrationBranch(root, homestate.Card{})
			if row.want == "" {
				if branch != "" || source != baseSourceCallerWire {
					t.Fatalf("(%q, %q): a row with no target must fall back to the caller", branch, source)
				}
				return
			}
			if branch != row.want || source != baseSourceConfigWire {
				t.Fatalf("(%q, %q), want (%q, config)", branch, source, row.want)
			}
		})
	}
	for _, row := range interpretationRows() {
		t.Run("merge-ready/"+row.name, func(t *testing.T) {
			run, err := runMergeReadyWithRow(t, row)
			if row.want == "" {
				if err == nil || !strings.Contains(err.Error(), "no integration branch configured") {
					t.Fatalf("err = %v, want the no-integration-branch refusal", err)
				}
				if run != nil {
					t.Fatalf("a refused resolution must record no run, got %+v", run)
				}
				return
			}
			if run == nil {
				t.Fatalf("no merge-check run recorded for target %q (err=%v)", row.want, err)
			}
			if run.Develop != row.want {
				t.Fatalf("merge target = %q, want %q", run.Develop, row.want)
			}
		})
	}
}

// runMergeReadyWithRow runs `factory merge ready` WITHOUT --develop against a
// fixture repository whose lock root carries the row's git-strategy.yaml, and
// returns the merge-check run it recorded (nil when it recorded none). The
// merge target is then the configuration's, so the recorded Develop field is
// the observable the reader decided.
func runMergeReadyWithRow(t *testing.T, row targetRow) (*factorylane.MergeCheckRun, error) {
	t.Helper()
	_, lockRoot, specID := mergeReadyFixture(t, "complete", "lane-9", "sess-lane-9")
	installTargetRow(t, lockRoot, row)

	_, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", specID,
		"--branch", "WT-card", "--session", "sess-lane-9", "--json")
	run, loadErr := factorylane.NewStore(".", nil).LatestMergeCheckRun("lane-9", "t9001")
	if loadErr != nil {
		t.Fatalf("read the recorded run: %v", loadErr)
	}
	return run, err
}

// --- card diff base and mission contract target (ledger row E-29) ---

// cardBaseRepo builds a linear history main -> staging -> develop -> card on
// branch WT-card and returns the repository directory. Every candidate target
// branch tip is a distinct commit, so the merge base of each with the card
// head names which branch the resolver chose.
func cardBaseRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	cardScopeGit(t, dir, "init", "-q", "-b", "main")
	commit := func(name string) {
		writeCardFile(t, dir, name, name+"\n")
		cardScopeGit(t, dir, "add", name)
		cardScopeGit(t, dir, "commit", "-q", "-m", name)
	}
	commit("a.txt")
	cardScopeGit(t, dir, "checkout", "-q", "-b", "staging")
	commit("m.txt")
	cardScopeGit(t, dir, "checkout", "-q", "-b", "develop")
	commit("b.txt")
	cardScopeGit(t, dir, "checkout", "-q", "-b", "WT-card")
	commit("c.txt")
	return dir
}

// TestBaseDefaultsGitFlowUnchanged pins the git-flow outputs of the cli-side
// defaults: the card diff base is the merge base with develop and the mission
// contract names develop as its merge target. Green before and after the swap.
func TestBaseDefaultsGitFlowUnchanged(t *testing.T) {
	t.Run("card merge base is measured against develop", func(t *testing.T) {
		dir := cardBaseRepo(t)
		installTargetRow(t, dir, gitFlowRow())
		got, err := cardMergeBase(dir)
		if err != nil {
			t.Fatalf("cardMergeBase: %v", err)
		}
		want := cardScopeGit(t, dir, "rev-parse", "develop")
		if got != want {
			t.Fatalf("card merge base = %s, want the develop tip %s", got, want)
		}
	})

	t.Run("mission contract merge target is develop", func(t *testing.T) {
		if got := approveAndReadMergeTarget(t, gitFlowRow()); got != "develop" {
			t.Fatalf("MergeTarget = %q, want develop", got)
		}
	})
}

// TestBaseDefaultsFollowIntegrationTarget runs every interpretation-table row
// through the card diff base and the mission contract target. A resolved
// target is used verbatim; a row that resolves nothing is an error at both
// sites — the card diff base cannot be measured, the contract cannot be
// sealed — rather than a default.
func TestBaseDefaultsFollowIntegrationTarget(t *testing.T) {
	for _, row := range interpretationRows() {
		t.Run("card-base/"+row.name, func(t *testing.T) {
			dir := cardBaseRepo(t)
			installTargetRow(t, dir, row)
			got, err := cardMergeBase(dir)
			switch {
			case row.want == "":
				if err == nil {
					t.Fatalf("a row with no target must fail the base measurement, got %q", got)
				}
			case cardScopeBranchExists(t, dir, row.want):
				if err != nil {
					t.Fatalf("cardMergeBase: %v", err)
				}
				if want := cardScopeGit(t, dir, "rev-parse", row.want); got != want {
					t.Fatalf("card merge base = %s, want the %s tip %s", got, row.want, want)
				}
			default:
				// A target that names no branch of this repository (the release
				// prefix) cannot be measured either; the point is that the chosen
				// ref is the interpreter's, not develop.
				if err == nil {
					t.Fatalf("target %q names no branch here, the measurement must fail, got %q", row.want, got)
				}
			}
		})
	}
	for _, row := range interpretationRows() {
		t.Run("mission-target/"+row.name, func(t *testing.T) {
			if row.want == "" {
				if _, _, err := approveMissionWithRow(t, row); err == nil {
					t.Fatal("a row with no target must refuse the contract approval")
				}
				return
			}
			if got := approveAndReadMergeTarget(t, row); got != row.want {
				t.Fatalf("MergeTarget = %q, want %q", got, row.want)
			}
		})
	}
}

// cardScopeBranchExists reports whether branch resolves in dir.
func cardScopeBranchExists(t *testing.T, dir, branch string) bool {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	return cmd.Run() == nil
}

// approveMissionWithRow arms an auto mission under a fresh project root that
// carries the row's configuration, runs `goal approve`, and returns the root,
// the session id, and the approve error (nil on success).
func approveMissionWithRow(t *testing.T, row targetRow) (root, session string, err error) {
	t.Helper()
	root = t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	installTargetRow(t, root, row)
	session = "018f4f4a-7b7c-7a11-8f4d-bbbbbbbbbbbb"
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	if armErr := run("--auto", "--session", session, "approve the base target"); armErr != nil {
		t.Fatalf("arm auto mission: %v", armErr)
	}
	err = run("approve", "--session", session, "--scope", "repo:"+root,
		"--action", "commit", "--completion-evidence", "commit")
	return root, session, err
}

// approveAndReadMergeTarget approves a mission under the row's configuration
// and returns the sealed contract's MergeTarget.
func approveAndReadMergeTarget(t *testing.T, row targetRow) string {
	t.Helper()
	root, session, err := approveMissionWithRow(t, row)
	if err != nil {
		t.Fatalf("goal approve: %v", err)
	}
	state, loadErr := mission.LoadAutoMission(root, session)
	if loadErr != nil || state == nil || state.Contract == nil {
		t.Fatalf("approved mission carries no contract (state=%+v err=%v)", state, loadErr)
	}
	return state.Contract.MergeTarget
}
