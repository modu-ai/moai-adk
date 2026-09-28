// factory_complete_test.go — SPEC-FACTORY-SELF-DISPATCH-001 M2 AC tests
// (card t1240): `moai factory complete` through the integration worktree,
// the REQ-SD-023 integration-window refusals, and the REQ-SD-025 Codex
// merge-edge refusal on the CLI paths. AC-SD-013, AC-SD-024 (complete and
// stage halves), AC-SD-025. The MCP half of AC-SD-024 lands with the tools
// in M3 and consumes the same factoryRefuseCodexMergeEdge check wired here.
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (§B of acceptance.md); ./internal/cli runs only
// through the anchored -run selectors naming one of these tests.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// sdCardTree names one card's worktree and its WT- branch.
type sdCardTree struct{ wt, branch string }

// sdGitFlowDevelop names `develop` as the project's git-flow integration
// branch (fcGitFlowConfig hardcodes a different branch name).
func sdGitFlowDevelop(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: develop\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// sdMergeFixture builds the git shape `factory complete` operates on: a
// parent checkout (root) whose repository carries a `develop` integration
// branch — checked out in a provisioned integration worktree, checked out in
// the parent itself, or held by no tree — plus one worktree per card on its
// own WT- branch carrying one commit. withConfig stamps the git-flow
// integration branch; without it the project reads as github-flow.
func sdMergeFixture(t *testing.T, withConfig, provisioned, onParent bool, n int) (string, string, []sdCardTree) {
	t.Helper()
	root, store := fcFixture(t)
	states := make([]kanban.BacklogState, n)
	for i := range states {
		states[i] = kanban.BacklogStatePicked
	}
	fcQueue(t, store, states...)
	fcGit(t, root, "branch", "develop")
	integWT := ""
	switch {
	case onParent:
		fcGit(t, root, "checkout", "-q", "develop")
	case provisioned:
		integWT = filepath.Join(root, ".claude", "worktrees", "develop")
		fcGit(t, root, "worktree", "add", "-q", integWT, "develop")
	}
	if withConfig {
		sdGitFlowDevelop(t, root)
	}
	names := []string{"alpha", "beta"}
	cards := make([]sdCardTree, 0, n)
	for i := 0; i < n; i++ {
		branch := "WT-" + names[i]
		wt := filepath.Join(root, ".claude", "worktrees", fmt.Sprintf("t%d", i+1))
		fcGit(t, root, "worktree", "add", "-q", "-b", branch, wt)
		if err := os.WriteFile(filepath.Join(wt, names[i]+".txt"), []byte(names[i]+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, wt, "add", "-A")
		fcGit(t, wt, "commit", "-q", "-m", "card change "+names[i])
		cards = append(cards, sdCardTree{wt: wt, branch: branch})
	}
	return root, integWT, cards
}

// sdPlaceMergeReady records a merge-ready card leased by lane with tree as
// its worktree.
func sdPlaceMergeReady(t *testing.T, root, cardID, lane string, tree sdCardTree) {
	t.Helper()
	fcPlace(t, root, homestate.Card{
		CardID: cardID, State: homestate.CardMergeReady, Stage: homestate.CardMergeReady,
		OwnerLabel: lane, LeaseHolder: lane, LeaseExpiresAt: "2026-10-01T00:00:00Z",
		WorktreePath: tree.wt,
	})
}

// sdHoldWindow records a held integration window (pid 0 reads live, the
// shape an owning session leaves behind).
func sdHoldWindow(t *testing.T, root, sessionID, name, branch, source, wt, card string) {
	t.Helper()
	if _, err := kanban.AcquireIntegrationLock(root, kanban.IntegrationLock{
		SessionID: sessionID, SessionName: name, Branch: branch, BranchSource: source, Worktree: wt, Card: card,
	}, false); err != nil {
		t.Fatalf("hold window: %v", err)
	}
}

// sdWindow reads the recorded window for assertions.
func sdWindow(t *testing.T, root string) *kanban.IntegrationLock {
	t.Helper()
	lock, err := kanban.ReadIntegrationLock(root)
	if err != nil {
		t.Fatalf("read window: %v", err)
	}
	return lock
}

// sdCardUnchanged asserts the card row did not move on a refusal.
func sdCardUnchanged(t *testing.T, where string, root, cardID string, before homestate.Card) {
	t.Helper()
	after := fcCard(t, root, cardID)
	if after.State != before.State || after.Version != before.Version || after.MergeSHA != before.MergeSHA {
		t.Errorf("%s: card row changed: %s v%d merge=%s → %s v%d merge=%s",
			where, before.State, before.Version, before.MergeSHA, after.State, after.Version, after.MergeSHA)
	}
}

// AC-SD-013 — Claude `complete` through the integration worktree.
func TestSD_AC013_ClaudeCompleteViaIntegrationWorktree(t *testing.T) {
	t.Run("pre-merged card reaches merged-local; the window stays held", func(t *testing.T) {
		root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", kanban.BranchSourceConfig, integWT, "t1")
		// The lane already merged --no-ff and wrote the re-measure record.
		fcGit(t, integWT, "merge", "-q", "--no-ff", "-m", "M t1", cards[0].branch)
		merge := fcGit(t, integWT, "rev-parse", "HEAD")
		remeasure := filepath.Join(root, ".moai", "reports", "t1", "remeasure.txt")
		if err := os.MkdirAll(filepath.Dir(remeasure), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(remeasure, []byte("re-measure on merge "+merge+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		if _, _, err := runFactory(t, "complete", "t1", remeasure, "--run", fcRun); err != nil {
			t.Fatalf("complete: %v", err)
		}
		c := fcCard(t, root, "t1")
		if c.State != homestate.CardMergedLocal {
			t.Fatalf("t1 = %s, want merged-local", c.State)
		}
		if c.MergeSHA != merge {
			t.Errorf("merge SHA = %q, want the pre-made merge %q", c.MergeSHA, merge)
		}
		if c.RemeasurePath != remeasure {
			t.Errorf("remeasure path = %q, want %q", c.RemeasurePath, remeasure)
		}
		// Release is the lane's NEXT step: complete itself does not release.
		lock := sdWindow(t, root)
		if !lock.Held() || lock.SessionID != "sess-lane-1" {
			t.Errorf("window after complete = held=%v by %q, want still held by sess-lane-1", lock.Held(), lock.SessionID)
		}
	})

	t.Run("complete performs the merge itself and records the re-measure evidence", func(t *testing.T) {
		root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", kanban.BranchSourceConfig, integWT, "t1")

		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
			t.Fatalf("complete: %v", err)
		}
		c := fcCard(t, root, "t1")
		if c.State != homestate.CardMergedLocal {
			t.Fatalf("t1 = %s, want merged-local", c.State)
		}
		if c.MergeSHA == "" {
			t.Fatal("merged-local without a merge SHA")
		}
		// The merge commit is a real --no-ff merge of the card branch into
		// develop, inside the integration worktree.
		head := fcGit(t, integWT, "rev-parse", "HEAD")
		if head != c.MergeSHA {
			t.Errorf("integration worktree HEAD = %s, want the recorded merge %s", head, c.MergeSHA)
		}
		if got := fcGit(t, integWT, "rev-list", "--parents", "-n", "1", head); len(strings.Fields(got)) != 3 {
			t.Errorf("HEAD is not a two-parent merge: %s", got)
		}
		if c.RemeasurePath == "" {
			t.Fatal("no re-measure path recorded")
		}
		raw, err := os.ReadFile(c.RemeasurePath)
		if err != nil {
			t.Fatalf("read remeasure: %v", err)
		}
		if !strings.Contains(string(raw), c.MergeSHA[:12]) {
			t.Errorf("remeasure record %q does not name merge %s", raw, c.MergeSHA)
		}
	})

	t.Run("integration branch held only by the parent checkout: not provisioned", func(t *testing.T) {
		root, _, cards := sdMergeFixture(t, true, false, true, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		before := fcCard(t, root, "t1")

		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), "not provisioned") {
			t.Fatalf("parent-checkout arm: err = %v, want a not-provisioned refusal", err)
		}
		sdCardUnchanged(t, "parent-checkout arm", root, "t1", before)
	})

	t.Run("integration branch held by no tree: not provisioned", func(t *testing.T) {
		root, _, cards := sdMergeFixture(t, true, false, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		before := fcCard(t, root, "t1")

		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), "not provisioned") {
			t.Fatalf("no-tree arm: err = %v, want a not-provisioned refusal", err)
		}
		sdCardUnchanged(t, "no-tree arm", root, "t1", before)
	})

	t.Run("caller-source window: refused naming --branch", func(t *testing.T) {
		// A github-flow fixture: no configured integration branch, so the
		// lane's own acquire fell back to its card worktree.
		root, _, cards := sdMergeFixture(t, false, false, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		before := fcCard(t, root, "t1")
		sdHoldWindow(t, root, "sess-lane-1", "lane-1", cards[0].branch, kanban.BranchSourceCaller, cards[0].wt, "t1")

		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), "--branch") {
			t.Fatalf("caller-source arm: err = %v, want a refusal naming --branch", err)
		}
		if c := fcCard(t, root, "t1"); c.State == homestate.CardMergedLocal {
			t.Error("card reached merged-local on the caller-source refusal")
		}
		sdCardUnchanged(t, "caller-source arm", root, "t1", before)
	})

	t.Run("window naming the card's own branch: refused", func(t *testing.T) {
		// A window acquired with --branch <the card's own WT- branch>: source
		// flag, tree = the card worktree.
		root, _, cards := sdMergeFixture(t, true, false, false, 1)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		before := fcCard(t, root, "t1")
		sdHoldWindow(t, root, "sess-lane-1", "lane-1", cards[0].branch, kanban.BranchSourceFlag, cards[0].wt, "t1")

		sdLaneEnv(t, "lane-1", "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), "own branch") {
			t.Fatalf("card-branch arm: err = %v, want the card's-own-branch refusal", err)
		}
		sdCardUnchanged(t, "card-branch arm", root, "t1", before)
	})
}

// AC-SD-024 — Codex merge edge refused on every path (the complete half of
// the CLI; the MCP half lands with the tools in M3 on the same check).
func TestSD_AC024_CodexMergeRefusedComplete(t *testing.T) {
	root, _, cards := sdMergeFixture(t, true, true, false, 1)
	sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
	before := fcCard(t, root, "t1")

	// Marker set, so admission holds and the harness check is the only
	// possible refusal cause.
	sdLaneEnv(t, "lane-1", kanban.BackendGPT)
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
	if err == nil || !strings.Contains(err.Error(), factoryCodexMergeSentinel) {
		t.Fatalf("codex complete: err = %v, want the Codex merge-edge refusal", err)
	}
	sdCardUnchanged(t, "codex complete", root, "t1", before)
}

// AC-SD-024 — the stage half: `stage <card> merging` refused under the Codex
// backend (the transition behavior itself is M3; the refusal is wired now).
func TestSD_AC024_CodexMergeRefusedStage(t *testing.T) {
	root, _, cards := sdMergeFixture(t, true, true, false, 1)
	sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
	before := fcCard(t, root, "t1")

	sdLaneEnv(t, "lane-1", kanban.BackendGPT)
	_, _, err := runFactory(t, "stage", "t1", "merging", "--run", fcRun)
	if err == nil || !strings.Contains(err.Error(), factoryCodexMergeSentinel) {
		t.Fatalf("codex stage merging: err = %v, want the Codex merge-edge refusal", err)
	}
	sdCardUnchanged(t, "codex stage", root, "t1", before)

	// Negative control: a Claude backend does not take this refusal (the
	// stage behavior itself is the M3 stub).
	sdLaneEnv(t, "lane-1", "")
	_, _, err = runFactory(t, "stage", "t1", "merging", "--run", fcRun)
	if err != nil && strings.Contains(err.Error(), factoryCodexMergeSentinel) {
		t.Errorf("claude stage merging took the Codex refusal: %v", err)
	}
	sdCardUnchanged(t, "claude stage", root, "t1", before)
}

// AC-SD-025 — the integration window serializes lanes.
func TestSD_AC025_IntegrationWindowSerializes(t *testing.T) {
	root, integWT, cards := sdMergeFixture(t, true, true, false, 2)
	sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
	sdPlaceMergeReady(t, root, "t2", "lane-2", cards[1])
	before2 := fcCard(t, root, "t2")

	// lane-1 holds the window; lane-2's complete is refused naming lane-1.
	sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", kanban.BranchSourceConfig, integWT, "t1")
	sdLaneEnv(t, "lane-2", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-2")
	_, _, err := runFactory(t, "complete", "t2", "--run", fcRun)
	if err == nil || !strings.Contains(err.Error(), "lane-1") {
		t.Fatalf("lane-2 complete against lane-1's window: err = %v, want a refusal naming lane-1", err)
	}
	sdCardUnchanged(t, "serialized lane-2", root, "t2", before2)

	// lane-1 completes on its own window, then releases.
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("lane-1 complete: %v", err)
	}
	merge1 := fcCard(t, root, "t1").MergeSHA
	if _, err := kanban.ReleaseIntegrationLock(root, "sess-lane-1", 0, false); err != nil {
		t.Fatalf("release: %v", err)
	}

	// lane-2 absorbs develop into its card branch (the lane duty), completes,
	// and both merges are on develop in the integration worktree.
	fcGit(t, cards[1].wt, "merge", "-q", "--no-ff", "-m", "absorb develop", "develop")
	sdLaneEnv(t, "lane-2", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-2")
	if _, _, err := runFactory(t, "complete", "t2", "--run", fcRun); err != nil {
		t.Fatalf("lane-2 complete after release: %v", err)
	}
	merge2 := fcCard(t, root, "t2").MergeSHA
	for _, sha := range []string{merge1, merge2} {
		if sha == "" {
			t.Fatal("a completed card carries no merge SHA")
		}
		out := fcGit(t, integWT, "merge-base", "--is-ancestor", sha, "develop")
		if strings.TrimSpace(out) != "" {
			t.Errorf("merge %s not an ancestor of develop: %s", sha, out)
		}
	}
	// The window is now held by lane-2's session (release stays its next step).
	if lock := sdWindow(t, root); !lock.Held() || lock.SessionID != "sess-lane-2" {
		t.Errorf("window after lane-2 complete = held=%v by %q, want held by sess-lane-2", lock.Held(), lock.SessionID)
	}
}
