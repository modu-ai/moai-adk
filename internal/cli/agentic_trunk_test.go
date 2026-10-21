package cli

// agentic_trunk_test.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M2 (card
// t1831).
//
// The consumer tests for the fifth workflow value: the factory delivery
// path (REQ-AGT-006), the session-exit auto-merge gate and notice
// (REQ-AGT-007), the integration acquire window (REQ-AGT-008), the
// ignored keys (REQ-AGT-010), and the contract push check over the
// projected develop-branch seam (REQ-AGT-025 Judge 4). The harness
// helpers (fcFixture, sdMergeFixture shapes, mergeReadyFixture shapes,
// scratchRepo, the closuretest fixture) are reused, never re-derived.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/testguard"
)

// agenticTrunkGitStrategy is the manual agentic-trunk profile body the
// fixtures write: the value and the git-flow-only keys together, the exact
// state REQ-AGT-018 leaves behind at the flip (the keys ignored, no
// diagnostic).
const agenticTrunkGitStrategy = "git_strategy:\n    mode: manual\n    worktree_base_branch: main\n    manual:\n        workflow: agentic-trunk\n" +
	"        develop_branch: develop\n        release_branch_prefix: release/\n" +
	"        rc_version_format: vX.Y.Z-rc.N\n        lead_push_threshold: 20\n        merge_method: merge\n"

// writeAgenticTrunkConfig stamps the agentic-trunk git-strategy body under
// root (the hard-coded shape writeGitStrategyFixture cannot express).
func writeAgenticTrunkConfig(t *testing.T, root string) {
	t.Helper()
	writeGitStrategyBody(t, root, agenticTrunkGitStrategy)
}

// ─── REQ-AGT-006 — the factory delivery path (AC-AGT-005) ───

// agenticTrunkMergeFixture is the sdMergeFixture shape with main as the
// integration branch: the parent moves off main, a provisioned integration
// worktree holds main, one card worktree carries one commit, the config
// names manual agentic-trunk, and the card sits merge-ready under a live
// lease. The re-measure record is keyed to the candidate tree as REQ-MWQ-019
// step 3 requires.
func agenticTrunkMergeFixture(t *testing.T) (root, integWT string, card sdCardTree) {
	t.Helper()
	sdClearLaneEnv(t)
	r, store := fcFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", r)
	fcQueue(t, store, factory.BacklogStatePicked)
	// todoFixture names no default branch; name it main first (ghfNew's
	// precedent), then vacate it so a dedicated integration worktree can
	// hold it — the shape the parent-refusal rule (REQ-SD-023 step 3) admits.
	fcGit(t, r, "branch", "-M", "main")
	fcGit(t, r, "checkout", "-q", "-b", "wtroot")
	writeAgenticTrunkConfig(t, r)
	integWT = filepath.Join(r, ".claude", "worktrees", "main")
	fcGit(t, r, "worktree", "add", "-q", integWT, "main")
	wt := filepath.Join(r, ".claude", "worktrees", "t1")
	fcGit(t, r, "worktree", "add", "-q", "-b", "WT-alpha", wt)
	if err := os.WriteFile(filepath.Join(wt, "alpha.txt"), []byte("alpha\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, wt, "add", "-A")
	fcGit(t, wt, "commit", "-q", "-m", "card change alpha")
	sdPlaceMergeReady(t, r, "t1", "lane-1", sdCardTree{wt: wt, branch: "WT-alpha"})
	if _, err := factory.RunRemeasure(r, wt, "main", "true"); err != nil {
		t.Fatalf("place re-measure record: %v", err)
	}
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	return r, integWT, sdCardTree{wt: wt, branch: "WT-alpha"}
}

// TestAgenticTrunkFactoryCompleteTakesWindow is AC-AGT-005: under manual
// agentic-trunk, `factory complete` takes the integration window, merges the
// card into main --no-ff in the worktree holding main, and calls gh never.
func TestAgenticTrunkFactoryCompleteTakesWindow(t *testing.T) {
	root, integWT, card := agenticTrunkMergeFixture(t)
	before := factoryGHUnexpectedCalls.Load()
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v", err)
	}
	c := fcCard(t, root, "t1")
	if c.State != homestate.CardMergedLocal || c.MergeSHA == "" {
		t.Fatalf("card = %s merge=%q, want merged-local", c.State, c.MergeSHA)
	}
	if got := fcGit(t, integWT, "rev-list", "--parents", "-n", "1", "HEAD"); len(strings.Fields(got)) != 3 {
		t.Errorf("main HEAD is not a two-parent merge: %s", got)
	}
	if lock := sdWindow(t, root); lock.Held() {
		t.Errorf("window after complete = held=%v by %q, want released after the transitions", lock.Held(), lock.SessionID)
	}
	if got := factoryGHUnexpectedCalls.Load(); got != before {
		t.Errorf("agentic-trunk complete called gh %d time(s) — it opens no pull request", got-before)
	}
	_ = card
}

// agenticTrunkReadyFixture is the mergeReadyFixture shape with main as the
// integration branch: a repo on main, a WT-card branch that absorbed main,
// a SPEC directory with the given sync status, and the re-measure record the
// fourth condition reads.
func agenticTrunkReadyFixture(t *testing.T, syncStatus string) (repo, lockRoot, specID string) {
	t.Helper()
	repo = t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(rel, content string) {
		if err := os.WriteFile(filepath.Join(repo, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-b", "main")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	write("base.txt", "base\n")
	git("add", ".")
	git("commit", "-m", "base")
	git("checkout", "-b", "WT-card")
	write("card.txt", "card\n")
	git("add", ".")
	git("commit", "-m", "card work")
	git("checkout", "main")
	write("dev.txt", "dev\n")
	git("add", ".")
	git("commit", "-m", "main work")
	git("checkout", "WT-card")
	git("merge", "--no-ff", "-m", "absorb main", "main")

	specID = "SPEC-MERGE-READY-FIXTURE-001"
	section := ""
	if syncStatus != "" {
		section = "## §E.4 Sync-phase Audit-Ready Signal\n\nsync_status: " + syncStatus + "\n"
	}
	specDir := filepath.Join(repo, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"),
		[]byte("## §E.2 Run-phase Evidence\n\n...\n\n"+section), 0o644); err != nil {
		t.Fatal(err)
	}

	lockRoot = t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", lockRoot)
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-9")
	tree := strings.TrimSpace(func() string {
		cmd := exec.Command("git", "rev-parse", "WT-card^{tree}")
		cmd.Dir = repo
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("read candidate tree: %v", err)
		}
		return string(out)
	}())
	base := strings.TrimSpace(func() string {
		cmd := exec.Command("git", "merge-base", "main", "WT-card")
		cmd.Dir = repo
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("read absorbed base: %v", err)
		}
		return string(out)
	}())
	if err := factory.WriteRemeasureRecord(lockRoot, tree, factory.RemeasureRecord{
		Base:          base,
		Command:       "true",
		ExitCode:      0,
		BuildIdentity: "moai merge-ready fixture",
		RecordedAt:    time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}
	writeAgenticTrunkConfig(t, repo)
	t.Chdir(repo)
	return repo, lockRoot, specID
}

// TestAgenticTrunkMergeReadyTakesWindow is AC-AGT-005's second half: under
// manual agentic-trunk, `factory merge ready` evaluates the triple and takes
// the integration window (the local-merge path), opening no pull request.
func TestAgenticTrunkMergeReadyTakesWindow(t *testing.T) {
	_, lockRoot, specID := agenticTrunkReadyFixture(t, "complete")

	out, err := runFactoryMerge(t, "merge", "ready", "--card", "t9001", "--spec", specID, "--branch", "WT-card", "--develop", "main", "--session", "sess-lane-9", "--json")
	if err != nil {
		t.Fatalf("merge ready: %v", err)
	}
	if !strings.Contains(out, `"verdict":"cleared"`) {
		t.Fatalf("ready did not clear: %s", out)
	}
	lock := readLockRoot(t, lockRoot)
	if !lock.Held() || lock.SessionName != "lane-9" || lock.Card != "t9001" {
		t.Fatalf("window not recorded for lane-9/t9001: %+v", lock)
	}
	if lock.Branch != "main" {
		t.Errorf("window branch = %q, want main (the fixed integration target)", lock.Branch)
	}
}

// ─── REQ-AGT-007 — session-exit auto-merge (AC-AGT-006) ───

// autoMergeAgenticTrunkFixture is autoMergeGitFixture's shape with the
// agentic-trunk config: the repo's branch is main, the config names manual
// agentic-trunk with the git-flow keys present and ignored, and the session
// worktree carries one commit ahead of main.
func autoMergeAgenticTrunkFixture(t *testing.T) (root, swt string) {
	t.Helper()
	root = t.TempDir()
	git := func(args ...string) string {
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
		return string(out)
	}
	git("-C", root, "init")
	git("-C", root, "config", "user.name", "t1831-test")
	git("-C", root, "config", "user.email", "t1831@test.local")
	git("-C", root, "symbolic-ref", "HEAD", "refs/heads/main")
	if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".moai/state/\n.claude/worktrees/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml"), []byte(agenticTrunkGitStrategy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "card.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("-C", root, "add", ".")
	git("-C", root, "commit", "-m", "init")

	swt = filepath.Join(root, ".claude", "worktrees", "WT-am-test")
	git("-C", root, "worktree", "add", "-b", "WT-am-test", swt)
	if err := os.WriteFile(filepath.Join(swt, "card.txt"), []byte("card work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("-C", swt, "add", ".")
	git("-C", swt, "commit", "-m", "card work")
	return root, swt
}

// TestAgenticTrunkAutoMerge is AC-AGT-006: session-exit auto-merge enabled
// and manual agentic-trunk — a clean exit with commits ahead of main merges
// the session branch into main under the guards manual git-flow has today; a
// non-manual agentic-trunk project emits the skip notice naming BOTH
// workflows.
func TestAgenticTrunkAutoMerge(t *testing.T) {
	t.Run("manual merges into main", func(t *testing.T) {
		testguard.RedirectHome(t)
		root, swt := autoMergeAgenticTrunkFixture(t)
		gitOut := func(args ...string) string {
			out, err := exec.Command("git", args...).Output()
			if err != nil {
				t.Fatalf("git %v failed: %v\n%s", args, err, out)
			}
			return strings.TrimSpace(string(out))
		}
		swapAutoMergeSeams(t, amSeams{
			lockRoot:  func() string { return root },
			sessionID: func() string { return "sess-fixture" },
			wtForBranch: func(branch string) string {
				out, err := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
				if err != nil {
					return ""
				}
				return worktreeForBranchFromList(string(out), branch)
			},
		})
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), swt, true, out)

		parents := strings.Fields(gitOut("-C", root, "rev-list", "--parents", "-n", "1", "HEAD"))
		if len(parents) != 3 {
			t.Fatalf("main HEAD is not a two-parent merge commit: %v (notice: %q)", parents, out.String())
		}
		swtTip := gitOut("-C", swt, "rev-parse", "HEAD")
		if parents[2] != swtTip {
			t.Errorf("merge second parent %s != session branch tip %s", parents[2], swtTip)
		}
		lock, err := factory.ReadIntegrationLock(root)
		if err != nil {
			t.Fatalf("read lock: %v", err)
		}
		if lock.Held() {
			t.Errorf("window must be released after the merge, still held by %s", lock.SessionID)
		}
		notice := out.String()
		if !strings.Contains(notice, "main") {
			t.Errorf("success notice must name main as the target: %q", notice)
		}
		requireNoticePrefixed(t, notice)
	})

	t.Run("non-manual skip notice names both workflows", func(t *testing.T) {
		testguard.RedirectHome(t)
		rec := &amRecorder{}
		seams := baseAMSeams(rec)
		seams.gitFlow = func(string) config.GitFlowIntegrationConfig {
			return config.GitFlowIntegrationConfig{Manual: false, AgenticTrunkWorkflow: true, Workflow: "agentic-trunk", Mode: "team", IntegrationTarget: "main"}
		}
		swapAutoMergeSeams(t, seams)
		out := &bytes.Buffer{}

		sessionExitAutoMerge(amCfg(true, false), amSessionWt, true, out)

		notice := out.String()
		if len(rec.gitLog) != 0 || len(rec.lockLog) != 0 {
			t.Errorf("a non-manual agentic-trunk project must not merge: gitLog=%v lockLog=%v", rec.gitLog, rec.lockLog)
		}
		if !strings.Contains(notice, "git-flow") || !strings.Contains(notice, "agentic-trunk") {
			t.Errorf("skip notice must name both workflows (git-flow and agentic-trunk): %q", notice)
		}
		if !strings.Contains(notice, "develop_branch") || !strings.Contains(notice, "git_strategy") {
			t.Errorf("skip notice must keep naming the config key: %q", notice)
		}
		requireNoticePrefixed(t, notice)
	})
}

// ─── REQ-AGT-008 — integration acquire (AC-AGT-007) ───

// TestAgenticTrunkIntegrationAcquire is AC-AGT-007: in a manual
// agentic-trunk project, acquire without --branch records the window against
// the worktree holding main and writes no caller-fallback warning; in a
// non-manual agentic-trunk project the t886 unwired-target warning fires
// exactly as it does for github-flow.
func TestAgenticTrunkIntegrationAcquire(t *testing.T) {
	t.Run("manual records the worktree holding main", func(t *testing.T) {
		repo, _ := scratchRepo(t)
		writeAgenticTrunkConfig(t, repo)
		chdirRepo(t, repo)

		_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12", "--name", "lane", "--card", "c1")
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		if got := warningLines(stderr); len(got) != 0 {
			t.Errorf("a projected main must fall back never: %q (stderr %q)", got, stderr)
		}
		lock := readLock(t, repo)
		if lock.BranchSource != factory.BranchSourceConfig {
			t.Errorf("branch_source = %q, want %q", lock.BranchSource, factory.BranchSourceConfig)
		}
		if lock.Branch != "main" {
			t.Errorf("branch = %q, want main (the fixed target)", lock.Branch)
		}
		if lock.Worktree != realPath(t, repo) {
			t.Errorf("worktree = %q, want the tree holding main %q", lock.Worktree, realPath(t, repo))
		}
	})

	t.Run("team mode falls back with the unwired-target warning", func(t *testing.T) {
		repo, otherWT := scratchRepo(t)
		writeGitStrategyBody(t, repo, "git_strategy:\n    mode: team\n    team:\n        workflow: agentic-trunk\n")
		// The caller stands in a NON-main tree (the second worktree): a
		// caller already standing in the named target is the carve-out that
		// silences the t886 warning, and this scenario is the disagreement.
		chdirRepo(t, otherWT)

		_, stderr, err := runIntegrationStreams(t, repo, "acquire", "--session", "sess-lane12")
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		lines := warningLines(stderr)
		if len(lines) != 1 {
			t.Fatalf("want exactly one warning line, got %d: %q", len(lines), stderr)
		}
		for _, token := range []string{"agentic-trunk", "main"} {
			if !strings.Contains(lines[0], token) {
				t.Errorf("unwired-target warning does not name %q: %q", token, lines[0])
			}
		}
		lock := readLock(t, repo)
		if lock.BranchSource != factory.BranchSourceCaller {
			t.Errorf("branch_source = %q, want %q", lock.BranchSource, factory.BranchSourceCaller)
		}
	})
}

// ─── REQ-AGT-010 — ignored keys (AC-AGT-009) ───

// TestAgenticTrunkIgnoredKeys is AC-AGT-009: the reader and the doctor stay
// silent about develop_branch, release_branch_prefix and rc_version_format
// under agentic-trunk, while lead_push_threshold, merge_method and
// worktree_base_branch keep their present meaning.
func TestAgenticTrunkIgnoredKeys(t *testing.T) {
	root := t.TempDir()
	writeAgenticTrunkConfig(t, root)
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.LoadGitFlowIntegrationConfig(root)
	if cfg.IntegrationTarget != "main" {
		t.Errorf("IntegrationTarget = %q, want main", cfg.IntegrationTarget)
	}
	if cfg.MergeMethod != "merge" {
		t.Errorf("MergeMethod = %q, want merge (the key keeps its meaning)", cfg.MergeMethod)
	}

	loaded, err := config.NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil || loaded == nil {
		t.Fatalf("load full config: %v", err)
	}
	if got := loaded.GitStrategy.Manual.LeadPushThreshold; got != 20 {
		t.Errorf("LeadPushThreshold = %d, want 20 (the key keeps its meaning)", got)
	}
	if got := loaded.GitStrategy.WorktreeBaseBranch; got != "main" {
		t.Errorf("WorktreeBaseBranch = %q, want main (the key keeps its meaning)", got)
	}

	check := checkGitStrategyWorkflow(root, true)
	diagnostic := check.Message + " " + check.Detail
	for _, key := range []string{"develop_branch", "release_branch_prefix", "rc_version_format"} {
		if strings.Contains(diagnostic, key) {
			t.Errorf("diagnostic mentions the ignored key %q: %q", key, diagnostic)
		}
	}
	if !strings.Contains(check.Message, "main") {
		t.Errorf("doctor message does not name the target main: %q", check.Message)
	}
}

// ─── REQ-AGT-025 Judge 4 — contract push check (AC-AGT-025) ───

// runAgenticTrunkPushCheck is runPushCheck with the agentic-trunk config
// (no develop_branch key) instead of the git-flow fixture body.
func runAgenticTrunkPushCheck(t *testing.T, f *closuretest.Fixture, args ...string) (string, int, error) {
	t.Helper()
	f.Write(filepath.Join(f.Root, ".moai", "config", "sections", "workflow.yaml"),
		"workflow:\n  autonomy:\n    mode: contract\n    contract:\n      second_review: required\n")
	f.Write(filepath.Join(f.Root, ".moai", "config", "sections", "git-strategy.yaml"),
		"git_strategy:\n  mode: manual\n  manual:\n    workflow: agentic-trunk\n")
	fixtureSeams(t, f)
	c := &cobra.Command{}
	out := &bytes.Buffer{}
	c.SetOut(out)
	c.SetErr(out)
	err := runContractPushCheck(c, args)
	return out.String(), exitCodeOf(t, err), err
}

// TestAgenticTrunkPushCheckIntegration is AC-AGT-025 Judge 4: under contract
// autonomy with the develop_branch key absent, the push check does not print
// its inactive line and evaluates the push against the projected main — the
// ready case answers ready, and the unevidenced case answers NOT ready with
// the SPEC named (the evaluation ran), never inactive.
func TestAgenticTrunkPushCheckIntegration(t *testing.T) {
	// agenticTrunkPushFixture builds the closure fixture with a local main
	// branch (the projection's answer) pushed to origin and CHECKED OUT in
	// the primary, so the push the command evaluates by default (source =
	// the projected main) carries the SPEC once the card branch is merged.
	agenticTrunkPushFixture := func(t *testing.T) *closuretest.Fixture {
		t.Helper()
		f := closuretest.New(t)
		f.Git(f.Root, "branch", "main", "develop")
		f.Git(f.Root, "push", "-q", "origin", "main")
		f.Git(f.Root, "checkout", "-q", "main")
		queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
		// The queue/report seams the evidence builder reads are the command
		// seams — set before the evidence is built.
		fixtureSeams(t, f)
		return f
	}

	t.Run("the check stays active and answers ready over the projection", func(t *testing.T) {
		f := agenticTrunkPushFixture(t)
		// The report and the performed second review make the contract
		// genuinely ready; the merge puts the SPEC in the pushed range so the
		// evaluation has a candidate.
		buildReadyEvidence(t, f)
		f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
		out, code, err := runAgenticTrunkPushCheck(t, f)
		if code != 0 {
			t.Fatalf("exit %d err %v out %s, want ready", code, err, out)
		}
		if strings.Contains(out, "inactive (no git-flow integration branch configured)") {
			t.Fatalf("the projected main left the check inactive: %q", out)
		}
		if !strings.Contains(out, "ready") {
			t.Fatalf("out = %q, want ready", out)
		}
	})

	t.Run("an unevidenced push of main is evaluated, not inactive", func(t *testing.T) {
		f := agenticTrunkPushFixture(t)
		// The closure report builds the evidence the push evaluation reads;
		// landing the SPEC commit on main (the same shape the git-flow test
		// builds on develop) puts the contract in the pushed range — then the
		// missing second review answers not ready.
		if _, err := runReport(t, closuretest.Card); err != nil {
			t.Fatalf("report: %v", err)
		}
		f.Git(f.Root, "merge", "--no-ff", closuretest.Branch)
		out, code, err := runAgenticTrunkPushCheck(t, f)
		if code != 1 {
			t.Fatalf("exit %d err %v out %s, want 1 (the evaluation ran and answered not ready)", code, err, out)
		}
		if strings.Contains(out, "inactive (no git-flow integration branch configured)") {
			t.Fatalf("the projected main left the check inactive: %q", out)
		}
		if !strings.Contains(out, closuretest.SpecID) {
			t.Fatalf("out = %q, want the evaluated SPEC named", out)
		}
	})
}
