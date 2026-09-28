// factory_m3_test.go — SPEC-FACTORY-SELF-DISPATCH-001 M3 AC tests (card
// t1240): the per-card worktree of `moai factory next` (REQ-SD-011 /
// AC-SD-011), the `stage` transition and lease renewal (REQ-SD-012 /
// AC-SD-012), and the six MCP tools with their `project_root` resolution
// (REQ-SD-014/-024 / AC-SD-014, plus the M3 arms of AC-SD-010, -015, -016,
// -024).
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (§B of acceptance.md); ./internal/cli runs only
// through the anchored -run selectors naming one of these tests. The MCP
// handlers are called directly (the same functions the mcp-server registers),
// so no server process is started.
//
// Worktree-path note (t1292 absorption): the shared worktree materializer's
// L1 landing directory on this tree is <root>/.moai/worktrees/<name> — the
// AC-SD-011 tree assertions evaluate the materializer's actual landing
// directory. The AC's discriminative content — leaf name is the card id, the
// branch carries WT- without the card id, the record path equals the created
// directory, reuse never creates twice, a foreign directory refuses — is
// asserted in full.
package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// sdMoaiFixture is a factory fixture whose root is a MoAI project root (the
// .moai directory exists), so a caller-supplied project_root argument
// resolves through the same validation the existing project_root tools use.
func sdMoaiFixture(t *testing.T) (string, *kanban.BacklogStore) {
	t.Helper()
	root, store := fcFixture(t)
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root, store
}

// sdCallTool invokes an MCP handler the way the mcp-server dispatch does and
// flattens the result to its text payload; an IsError result surfaces as a
// non-nil error carrying the same text.
func sdCallTool(t *testing.T, fn func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) (string, error) {
	t.Helper()
	res, err := fn(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", nil
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	if res.IsError {
		return b.String(), errors.New(b.String())
	}
	return b.String(), nil
}

// sdCallToolErr is sdCallTool for arms that only need the error surface.
func sdCallToolErr(t *testing.T, fn func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) error {
	t.Helper()
	_, err := sdCallTool(t, fn, args)
	return err
}

// sdWorktreesDir is the shared materializer's landing directory under root.
func sdWorktreesDir(root string) string {
	return filepath.Join(root, ".moai", "worktrees")
}

// sdBranchOf reports the branch checked out in dir.
func sdBranchOf(t *testing.T, dir string) string {
	t.Helper()
	return fcGit(t, dir, "rev-parse", "--abbrev-ref", "HEAD")
}

// AC-SD-011 — worktree per card: create through the shared materializer,
// reuse the recorded tree, refuse a foreign directory.
func TestSD_AC011_CardWorktreeCreateReuseRefuse(t *testing.T) {
	t.Run("leased card with no recorded worktree gains one through the materializer", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStatePicked)
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)

		out, _, err := runFactory(t, "next", "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v\nstderr: %s", err, out)
		}
		wt := filepath.Join(sdWorktreesDir(root), "t1")
		info, statErr := os.Stat(wt)
		if statErr != nil || !info.IsDir() {
			t.Fatalf("card worktree %s does not exist: %v", wt, statErr)
		}
		branch := sdBranchOf(t, wt)
		if !strings.HasPrefix(branch, "WT-") {
			t.Errorf("card worktree branch = %q, want the WT- prefix", branch)
		}
		if strings.Contains(branch, "t1") {
			t.Errorf("card worktree branch %q contains the card id", branch)
		}
		if c := fcCard(t, root, "t1"); c.WorktreePath != wt {
			t.Errorf("recorded worktree path = %q, want %q", c.WorktreePath, wt)
		}
		if !strings.Contains(out, "t1") {
			t.Errorf("next output does not name the card: %q", out)
		}
	})

	t.Run("recorded worktree is reused without creating another", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStatePicked)
		sdRegisterLane(t, root, "lane-1")
		own := filepath.Join(root, "own-wt")
		fcGit(t, root, "worktree", "add", "-q", own, "-b", "WT-own-tree")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", WorktreePath: own})
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)

		if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
			t.Fatalf("next: %v", err)
		}
		if c := fcCard(t, root, "t1"); c.WorktreePath != own {
			t.Errorf("recorded worktree path = %q, want the card's own tree %q", c.WorktreePath, own)
		}
		if entries, err := os.ReadDir(sdWorktreesDir(root)); err == nil && len(entries) != 0 {
			t.Errorf("reuse created %d directories under the landing root, want 0", len(entries))
		}
	})

	t.Run("existing directory no card record names it: refuse, row unchanged", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStatePicked)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		foreign := filepath.Join(sdWorktreesDir(root), "t1")
		if err := os.MkdirAll(foreign, 0o755); err != nil {
			t.Fatal(err)
		}
		sdLaneEnv(t, "lane-1", "")
		before := fcCard(t, root, "t1")
		t.Chdir(root)

		_, _, err := runFactory(t, "next", "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), foreign) {
			t.Fatalf("next over a foreign directory: err = %v, want a refusal naming %s", err, foreign)
		}
		if after := fcCard(t, root, "t1"); after.Version != before.Version || after.State != before.State {
			t.Errorf("card row changed on the refusal: v%d %s → v%d %s", before.Version, before.State, after.Version, after.State)
		}
	})
}

// AC-SD-012 — `stage` applies the F1 edge with the lane label as actor and
// renews the lease; a non-holder is refused by F1 verbatim.
func TestSD_AC012_StageAppliesEdgeAndRenews(t *testing.T) {
	build := func(t *testing.T) (string, string) {
		t.Helper()
		root, _ := fcFixture(t)
		if err := os.WriteFile(filepath.Join(root, "plan.md"), []byte("plan\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, root, "add", "-A")
		fcGit(t, root, "commit", "-q", "-m", "plan artifact")
		sha := fcGit(t, root, "rev-parse", "HEAD")
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		fcPlace(t, root, homestate.Card{
			CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
			OwnerLabel: "lane-1", LeaseHolder: "lane-1",
			LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
			WorktreePath:   root,
		})
		return root, sha
	}

	t.Run("holder applies plan → plan-audit, actor recorded, lease renewed", func(t *testing.T) {
		root, sha := build(t)
		sdLaneEnv(t, "lane-1", "")
		if _, _, err := runFactory(t, "stage", "t1", "plan-audit", sha+":plan.md", "--run", fcRun); err != nil {
			t.Fatalf("stage: %v", err)
		}
		c := fcCard(t, root, "t1")
		if c.State != homestate.CardPlanAudit || c.Stage != homestate.CardPlanAudit {
			t.Fatalf("t1 = %s stage=%s, want plan-audit/plan-audit", c.State, c.Stage)
		}
		db := fcOpen(t, root)
		var payload string
		if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE kind='card.transition' ORDER BY seq DESC LIMIT 1`).Scan(&payload); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		if !strings.Contains(payload, `"actor":"lane-1"`) {
			t.Errorf("transition event actor missing: %s", payload)
		}
		expiry, err := time.Parse(time.RFC3339Nano, c.LeaseExpiresAt)
		if err != nil {
			t.Fatalf("lease expiry %q: %v", c.LeaseExpiresAt, err)
		}
		if want := fcNow.Add(10 * time.Minute); !expiry.After(want) {
			t.Errorf("lease expiry = %s, want moved forward past %s", expiry, want)
		}
	})

	t.Run("another lane is refused by F1 verbatim", func(t *testing.T) {
		root, sha := build(t)
		before := fcCard(t, root, "t1")
		sdLaneEnv(t, "lane-2", "")
		_, _, err := runFactory(t, "stage", "t1", "plan-audit", sha+":plan.md", "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), "does not hold the lease") {
			t.Fatalf("lane-2 stage: err = %v, want the F1 holder refusal", err)
		}
		if after := fcCard(t, root, "t1"); after.Version != before.Version || after.State != before.State {
			t.Errorf("card row changed on the refusal: v%d %s → v%d %s", before.Version, before.State, after.Version, after.State)
		}
	})
}

// AC-SD-014 — MCP ↔ CLI equivalence with project_root: twin fixtures, a
// success and a refusal per tool, equal record changes and refusal text.
func TestSD_AC014_MCPMatchesCLIWithProjectRoot(t *testing.T) {
	t.Run("todo_add appends identically; lane refusal text identical", func(t *testing.T) {
		// CLI twin.
		_, storeA := fcFixture(t)
		sdClearLaneEnv(t)
		cliOut, _, err := runTodo(t, "add", "mcp parity card")
		if err != nil {
			t.Fatalf("cli add: %v", err)
		}
		// MCP twin.
		rootB, storeB := sdMoaiFixture(t)
		mcpOut, err := sdCallTool(t, handleTodoAdd, map[string]any{"text": "mcp parity card", "project_root": rootB})
		if err != nil {
			t.Fatalf("mcp todo_add: %v", err)
		}
		if got := strings.TrimSpace(mcpOut); got != strings.TrimSpace(cliOut) {
			t.Errorf("mcp todo_add output = %q, cli = %q", got, strings.TrimSpace(cliOut))
		}
		readItem := func(store *kanban.BacklogStore) kanban.BacklogItem {
			t.Helper()
			rec, err := store.LoadPure()
			if err != nil {
				t.Fatal(err)
			}
			return rec.Items[0]
		}
		a, b := readItem(storeA), readItem(storeB)
		if a.ID != b.ID || a.State != b.State || a.Text != b.Text {
			t.Errorf("queue items diverge: cli %+v vs mcp %+v", a, b)
		}
		// Refusal twin: the lane boundary refusal, same text on both surfaces.
		sdLaneEnv(t, "lane-1", "")
		_, _, cliErr := runTodo(t, "add", "refused card")
		if cliErr == nil {
			t.Fatal("cli lane add was accepted")
		}
		_, mcpErr := sdCallTool(t, handleTodoAdd, map[string]any{"text": "refused card", "project_root": rootB})
		if mcpErr == nil {
			t.Fatal("mcp lane todo_add was accepted")
		}
		if !strings.Contains(mcpErr.Error(), cliErr.Error()) {
			t.Errorf("mcp refusal %q does not carry the cli refusal %q", mcpErr.Error(), cliErr.Error())
		}
	})

	t.Run("todo_list renders identically; an unusable root is refused identically", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, kanban.BacklogStateQueued)
		sdClearLaneEnv(t)
		cliOut, _, err := runTodo(t, "list")
		if err != nil {
			t.Fatalf("cli list: %v", err)
		}
		mcpOut, err := sdCallTool(t, handleTodoList, map[string]any{"project_root": root})
		if err != nil {
			t.Fatalf("mcp todo_list: %v", err)
		}
		if got := strings.TrimSpace(mcpOut); got != strings.TrimSpace(cliOut) {
			t.Errorf("mcp todo_list output = %q, cli = %q", got, strings.TrimSpace(cliOut))
		}
		bogus := filepath.Join(root, "no-such-tree")
		_, mcpErr := sdCallTool(t, handleTodoList, map[string]any{"project_root": bogus})
		if mcpErr == nil || !strings.Contains(mcpErr.Error(), "no-such-tree") {
			t.Fatalf("mcp todo_list bogus root: err = %v, want a rejection naming the path", mcpErr)
		}
	})

	t.Run("factory_next leases identically; not-a-lane refusal identical", func(t *testing.T) {
		lease := func(t *testing.T, viaMCP bool) (string, homestate.Card) {
			t.Helper()
			// A twin arm may run while the previous arm's lane environment is
			// still set; the fixture's own `todo add` needs a clean env.
			sdClearLaneEnv(t)
			root, store := sdMoaiFixture(t)
			fcQueue(t, store, kanban.BacklogStatePicked)
			sdRegisterLane(t, root, "lane-1")
			sdLaneEnv(t, "lane-1", "")
			t.Chdir(root)
			if viaMCP {
				if _, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun}); err != nil {
					t.Fatalf("mcp factory_next: %v", err)
				}
			} else {
				if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
					t.Fatalf("cli next: %v", err)
				}
			}
			return root, fcCard(t, root, "t1")
		}
		_, a := lease(t, false)
		rootB, b := lease(t, true)
		if a.State != b.State || a.Version != b.Version || a.LeaseHolder != b.LeaseHolder || a.OwnerLabel != b.OwnerLabel {
			t.Errorf("leased rows diverge: cli %+v vs mcp %+v", a, b)
		}
		if filepath.Base(b.WorktreePath) != "t1" || filepath.Dir(b.WorktreePath) != sdWorktreesDir(rootB) {
			t.Errorf("mcp worktree path = %q, want the landing directory leaf t1", b.WorktreePath)
		}

		// Refusal twin: label-only is not a lane, on either surface.
		sdClearLaneEnv(t)
		t.Setenv(config.EnvMoaiKanbanLabel, "lane-1")
		_, _, cliErr := runFactory(t, "next", "--run", fcRun)
		if cliErr == nil || !strings.Contains(cliErr.Error(), "not a lane session") {
			t.Fatalf("cli label-only next: err = %v", cliErr)
		}
		_, mcpErr := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": rootB, "run": fcRun})
		if mcpErr == nil {
			t.Fatal("mcp label-only factory_next was accepted")
		}
		if !strings.Contains(mcpErr.Error(), cliErr.Error()) {
			t.Errorf("mcp refusal %q does not carry the cli refusal %q", mcpErr.Error(), cliErr.Error())
		}
	})

	t.Run("factory_stage applies identically; holder refusal identical", func(t *testing.T) {
		build := func(t *testing.T) (string, string) {
			t.Helper()
			root, _ := sdMoaiFixture(t)
			if err := os.WriteFile(filepath.Join(root, "plan.md"), []byte("plan\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			fcGit(t, root, "add", "-A")
			fcGit(t, root, "commit", "-q", "-m", "plan artifact")
			sha := fcGit(t, root, "rev-parse", "HEAD")
			sdRegisterLane(t, root, "lane-1")
			sdRegisterLane(t, root, "lane-2")
			fcPlace(t, root, homestate.Card{
				CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
				OwnerLabel: "lane-1", LeaseHolder: "lane-1",
				LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
				WorktreePath:   root,
			})
			return root, sha
		}
		// CLI twin.
		rootA, shaA := build(t)
		sdLaneEnv(t, "lane-1", "")
		if _, _, err := runFactory(t, "stage", "t1", "plan-audit", shaA+":plan.md", "--run", fcRun); err != nil {
			t.Fatalf("cli stage: %v", err)
		}
		// MCP twin.
		rootB, shaB := build(t)
		sdLaneEnv(t, "lane-1", "")
		if _, err := sdCallTool(t, handleFactoryStage, map[string]any{"card": "t1", "state": "plan-audit", "evidence": shaB + ":plan.md", "run": fcRun, "project_root": rootB}); err != nil {
			t.Fatalf("mcp factory_stage: %v", err)
		}
		a, b := fcCard(t, rootA, "t1"), fcCard(t, rootB, "t1")
		if a.State != b.State || a.Version != b.Version || a.Stage != b.Stage || a.LeaseHolder != b.LeaseHolder {
			t.Errorf("staged rows diverge: cli %+v vs mcp %+v", a, b)
		}
		// Refusal twin: lane-2 does not hold the lease — the F1 refusal, the
		// same text on both surfaces.
		sdLaneEnv(t, "lane-2", "")
		_, _, cliErr := runFactory(t, "stage", "t1", "plan-audit", shaA+":plan.md", "--run", fcRun)
		if cliErr == nil {
			t.Fatal("cli lane-2 stage was accepted")
		}
		_, mcpErr := sdCallTool(t, handleFactoryStage, map[string]any{"card": "t1", "state": "plan-audit", "evidence": shaB + ":plan.md", "run": fcRun, "project_root": rootB})
		if mcpErr == nil {
			t.Fatal("mcp lane-2 factory_stage was accepted")
		}
		if !strings.Contains(mcpErr.Error(), cliErr.Error()) {
			t.Errorf("mcp refusal %q does not carry the cli refusal %q", mcpErr.Error(), cliErr.Error())
		}
	})

	t.Run("factory_complete merges identically; not-provisioned refusal identical", func(t *testing.T) {
		cliComplete := func(t *testing.T) homestate.Card {
			t.Helper()
			sdClearLaneEnv(t)
			root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
			sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
			sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", kanban.BranchSourceConfig, integWT, "t1")
			sdLaneEnv(t, "lane-1", "")
			t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
			t.Chdir(root)
			if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
				t.Fatalf("cli complete: %v", err)
			}
			return fcCard(t, root, "t1")
		}
		mcpComplete := func(t *testing.T) homestate.Card {
			t.Helper()
			sdClearLaneEnv(t)
			root, integWT, cards := sdMergeFixture(t, true, true, false, 1)
			sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
			sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", kanban.BranchSourceConfig, integWT, "t1")
			sdLaneEnv(t, "lane-1", "")
			t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
			t.Chdir(root)
			if err := sdCallToolErr(t, handleFactoryComplete, map[string]any{"card": "t1", "run": fcRun, "project_root": root}); err != nil {
				t.Fatalf("mcp factory_complete: %v", err)
			}
			return fcCard(t, root, "t1")
		}
		cliCard := cliComplete(t)
		mcpCard := mcpComplete(t)
		if cliCard.State != mcpCard.State || cliCard.Version != mcpCard.Version || cliCard.Stage != mcpCard.Stage {
			t.Errorf("completed rows diverge: cli %+v vs mcp %+v", cliCard, mcpCard)
		}
		if cliCard.State != homestate.CardMergedLocal {
			t.Errorf("cli card = %s, want merged-local", cliCard.State)
		}

		// Refusal twin: no provisioned integration worktree.
		cliRefuse := func(t *testing.T) (string, error) {
			t.Helper()
			sdClearLaneEnv(t)
			root, _, cards := sdMergeFixture(t, true, false, true, 1)
			sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
			sdLaneEnv(t, "lane-1", "")
			t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
			t.Chdir(root)
			_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
			return root, err
		}
		mcpRefuse := func(t *testing.T) (string, error) {
			t.Helper()
			sdClearLaneEnv(t)
			root, _, cards := sdMergeFixture(t, true, false, true, 1)
			sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
			sdLaneEnv(t, "lane-1", "")
			t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
			t.Chdir(root)
			return root, sdCallToolErr(t, handleFactoryComplete, map[string]any{"card": "t1", "run": fcRun, "project_root": root})
		}
		cliRoot, cliRefusal := cliRefuse(t)
		if cliRefusal == nil || !strings.Contains(cliRefusal.Error(), "not provisioned") {
			t.Fatalf("cli not-provisioned arm: err = %v", cliRefusal)
		}
		mcpRootName, mcpRefusal := mcpRefuse(t)
		if mcpRefusal == nil {
			t.Fatal("mcp factory_complete was accepted without a provisioned worktree")
		}
		// The refusal names the fixture's own parent checkout, so the twin
		// instances differ only in the temp path — normalize it and compare
		// the wording.
		cliText := strings.ReplaceAll(cliRefusal.Error(), cliRoot, "<root>")
		mcpText := strings.ReplaceAll(mcpRefusal.Error(), mcpRootName, "<root>")
		if !strings.Contains(mcpText, cliText) {
			t.Errorf("mcp refusal %q does not carry the cli refusal %q", mcpText, cliText)
		}
	})

	t.Run("factory_decide decides identically; lane refusal identical", func(t *testing.T) {
		build := func(t *testing.T) string {
			t.Helper()
			sdClearLaneEnv(t)
			root, store := sdMoaiFixture(t)
			fcQueue(t, store, kanban.BacklogStatePicked)
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardKickoff, DecisionGate: homestate.DecisionGateKickoff})
			return root
		}
		decide := map[string]any{"card": "t1", "gate": "kickoff", "choice": "approve", "run": fcRun}
		withRoot := func(root string) map[string]any {
			d := map[string]any{}
			for k, v := range decide {
				d[k] = v
			}
			d["project_root"] = root
			return d
		}
		// Refusal twin under the lane environment.
		rootRef := build(t)
		sdLaneEnv(t, "lane-1", "")
		_, _, cliErr := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun)
		if cliErr == nil {
			t.Fatal("cli lane decide was accepted")
		}
		mcpRefRoot := build(t)
		sdLaneEnv(t, "lane-1", "")
		_, mcpErr := sdCallTool(t, handleFactoryDecide, withRoot(mcpRefRoot))
		if mcpErr == nil {
			t.Fatal("mcp lane factory_decide was accepted")
		}
		if !strings.Contains(mcpErr.Error(), cliErr.Error()) {
			t.Errorf("mcp refusal %q does not carry the cli refusal %q", mcpErr.Error(), cliErr.Error())
		}
		_ = rootRef
		// Success twin with no lane variables.
		sdClearLaneEnv(t)
		rootA := build(t)
		if _, _, err := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun); err != nil {
			t.Fatalf("cli decide: %v", err)
		}
		rootB := build(t)
		if _, err := sdCallTool(t, handleFactoryDecide, withRoot(rootB)); err != nil {
			t.Fatalf("mcp factory_decide: %v", err)
		}
		a, b := fcCard(t, rootA, "t1"), fcCard(t, rootB, "t1")
		if a.State != b.State || a.Version != b.Version || a.Stage != b.Stage {
			t.Errorf("decided rows diverge: cli %+v vs mcp %+v", a, b)
		}
	})
}

// AC-SD-014 — `factory_next`, `factory_stage`, and `factory_complete`
// REQUIRE project_root: missing is rejected naming the argument, a path that
// is not a MoAI project root is rejected naming the path.
func TestSD_AC014_ProjectRootRequired(t *testing.T) {
	sdClearLaneEnv(t)
	root, _ := sdMoaiFixture(t)
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	for _, tc := range []struct {
		name string
		fn   func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		args map[string]any
	}{
		{"factory_next", handleFactoryNext, map[string]any{"run": fcRun}},
		{"factory_stage", handleFactoryStage, map[string]any{"card": "t1", "state": "run"}},
		{"factory_complete", handleFactoryComplete, map[string]any{"card": "t1"}},
	} {
		_, err := sdCallTool(t, tc.fn, tc.args)
		if err == nil || !strings.Contains(err.Error(), "project_root") {
			t.Errorf("%s without project_root: err = %v, want a rejection naming the argument", tc.name, err)
		}
		args := map[string]any{}
		for k, v := range tc.args {
			args[k] = v
		}
		args["project_root"] = filepath.Join(root, "not-a-root")
		_, err = sdCallTool(t, tc.fn, args)
		if err == nil || !strings.Contains(err.Error(), "not-a-root") {
			t.Errorf("%s with a non-root path: err = %v, want a rejection naming the path", tc.name, err)
		}
	}
}

// AC-SD-010 (MCP half) — factory_next evaluates its parent-checkout check
// against the project_root argument: a linked worktree is refused naming the
// parent, the parent checkout leases normally even though the server
// process's own working directory is the linked worktree.
func TestSD_AC010_MCPNextParentCheck(t *testing.T) {
	root, store := sdMoaiFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
	sdLaneEnv(t, "lane-1", "")

	wt := filepath.Join(filepath.Dir(root), filepath.Base(root)+"-wt")
	fcGit(t, root, "worktree", "add", "-q", wt, "-b", "WT-fixture")
	primary, _, err := identifyPrimaryCheckout(wt)
	if err != nil {
		t.Fatalf("identify primary: %v", err)
	}
	queueBefore := sdQueueBytes(t, store)
	cardBefore := fcCard(t, root, "t1")

	// The server process's own working directory is the linked worktree.
	t.Chdir(wt)
	_, mcpErr := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": wt, "run": fcRun})
	if mcpErr == nil || !strings.Contains(mcpErr.Error(), primary) {
		t.Fatalf("factory_next on the linked worktree: err = %v, want a refusal naming %s", mcpErr, primary)
	}
	if got := sdQueueBytes(t, store); got != queueBefore {
		t.Errorf("queue file changed on the refused factory_next")
	}
	if after := fcCard(t, root, "t1"); after.State != cardBefore.State || after.Version != cardBefore.Version {
		t.Errorf("card row changed on the refused factory_next")
	}

	// The parent checkout resolves through the argument: it leases normally.
	if _, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun}); err != nil {
		t.Fatalf("factory_next on the parent checkout: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased {
		t.Errorf("t1 = %s, want leased", c.State)
	}
}

// AC-SD-015 (todo_add MCP arm) — a lane environment and a Codex MCP
// environment both refuse todo_add with the queue bytes unchanged, and the
// Codex MCP environment also refuses factory_next as not a lane.
func TestSD_AC015_MCPTodoAddRefused(t *testing.T) {
	root, store := sdMoaiFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued)
	sdRegisterLane(t, root, "lane-1")

	sdLaneEnv(t, "lane-1", "")
	before := sdQueueBytes(t, store)
	if _, err := sdCallTool(t, handleTodoAdd, map[string]any{"text": "lane card", "project_root": root}); err == nil || !strings.Contains(err.Error(), "lane boundary") {
		t.Fatalf("lane todo_add: err = %v, want the lane-boundary refusal", err)
	}
	if got := sdQueueBytes(t, store); got != before {
		t.Errorf("lane todo_add refusal changed the queue bytes")
	}

	// Codex MCP environment: lane label + backend gpt, no role marker.
	sdClearLaneEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "lane-1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGPT)
	before = sdQueueBytes(t, store)
	if _, err := sdCallTool(t, handleTodoAdd, map[string]any{"text": "codex card", "project_root": root}); err == nil || !strings.Contains(err.Error(), "lane boundary") {
		t.Fatalf("codex-mcp todo_add: err = %v, want the lane-boundary refusal", err)
	}
	if got := sdQueueBytes(t, store); got != before {
		t.Errorf("codex-mcp todo_add refusal changed the queue bytes")
	}
	if _, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun}); err == nil || !strings.Contains(err.Error(), "not a lane session") {
		t.Fatalf("codex-mcp factory_next: err = %v, want the not-a-lane refusal", err)
	}
}

// AC-SD-016 (factory_decide MCP arm) — a lane and a Codex MCP environment
// are refused with the card row unchanged; an environment carrying none of
// the three lane variables succeeds.
func TestSD_AC016_MCPDecideRefused(t *testing.T) {
	root, store := sdMoaiFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardKickoff, DecisionGate: homestate.DecisionGateKickoff})
	before := fcCard(t, root, "t1")
	args := map[string]any{"card": "t1", "gate": "kickoff", "choice": "approve", "run": fcRun, "project_root": root}

	sdLaneEnv(t, "lane-1", "")
	if _, err := sdCallTool(t, handleFactoryDecide, args); err == nil || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("lane factory_decide: err = %v, want a refusal", err)
	}
	if after := fcCard(t, root, "t1"); after.State != before.State || after.Version != before.Version {
		t.Errorf("card row changed on the refused decide")
	}

	sdClearLaneEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "lane-1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGPT)
	if _, err := sdCallTool(t, handleFactoryDecide, args); err == nil || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("codex-mcp factory_decide: err = %v, want a refusal", err)
	}

	sdClearLaneEnv(t)
	if _, err := sdCallTool(t, handleFactoryDecide, args); err != nil {
		t.Fatalf("operator factory_decide: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardAssigned || c.Stage != homestate.CardRun {
		t.Errorf("t1 = %s stage=%s, want assigned/run", c.State, c.Stage)
	}
}

// AC-SD-024 (MCP) — the merge-ready → merging edge is refused through the
// MCP factory_stage and factory_complete tools while the backend identifies
// the Codex harness (marker set, so the harness check is the only refusal
// cause).
func TestSD_AC024_CodexMergeRefusedMCP(t *testing.T) {
	root, _, cards := sdMergeFixture(t, true, true, false, 1)
	sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
	before := fcCard(t, root, "t1")

	sdLaneEnv(t, "lane-1", kanban.BackendGPT)
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	t.Chdir(root)
	_, err := sdCallTool(t, handleFactoryStage, map[string]any{"card": "t1", "state": homestate.CardMerging, "run": fcRun, "project_root": root})
	if err == nil || !strings.Contains(err.Error(), factoryCodexMergeSentinel) {
		t.Fatalf("mcp factory_stage merging: err = %v, want the Codex merge-edge refusal", err)
	}
	_, err = sdCallTool(t, handleFactoryComplete, map[string]any{"card": "t1", "run": fcRun, "project_root": root})
	if err == nil || !strings.Contains(err.Error(), factoryCodexMergeSentinel) {
		t.Fatalf("mcp factory_complete: err = %v, want the Codex merge-edge refusal", err)
	}
	sdCardUnchanged(t, "codex mcp arms", root, "t1", before)
}

// The slug the card worktree branch derives from never contains the card id
// and stays within the naming rule (≤3 tokens, ≤24 chars, [a-z0-9-]).
func TestSD_WorktreeSlugShape(t *testing.T) {
	for _, tc := range []struct{ cardID, title, want string }{
		{"t1", "factory card 1", "factory-card-1"},
		{"t1240", "t1240 fix the flaky gate", "fix-the-flaky"}, // the id token is dropped
		{"t9", "Fix the ÜBER gate", "fix-the-ber"},             // non-ASCII falls to [a-z0-9]
		{"t2", "  spaced   out   words here ", "spaced-out-words"},
	} {
		got := factoryWorktreeSlug(tc.cardID, tc.title)
		if got != tc.want {
			t.Errorf("factoryWorktreeSlug(%q, %q) = %q, want %q", tc.cardID, tc.title, got, tc.want)
		}
		if len(got) > 24 {
			t.Errorf("slug %q exceeds 24 characters", got)
		}
		if strings.Contains(got, tc.cardID) {
			t.Errorf("slug %q contains the card id %q", got, tc.cardID)
		}
	}
	if got := factoryWorktreeSlug("t1", ""); got != "card" {
		t.Errorf("empty-title slug = %q, want the card fallback", got)
	}
}
