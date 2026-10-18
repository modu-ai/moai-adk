// factory_mcp_verb_parity_test.go — the verb-vs-MCP equivalence table for the lane verbs
// (card t1680, debts N1 and the decide/complete root items of SPEC-FACTORY-LEASE-RENEW-001).
// One implementation per verb means one rule per verb: for the same inputs the cobra verb and
// its MCP handler must resolve the same lane-label outcome and the same record and window
// roots. Every row builds its own tree under t.TempDir() with an isolated git config; the MCP
// handlers are called directly, no server process is started, and nothing reads the clock
// beyond the fixed factoryCardNow the fixture installs.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestMCPVerbParity_LaneLabel: a lane label the verb refuses (a legacy spelling, or an empty
// one) is refused by the MCP handler with the same wording, before any work. decide has no
// label check; its row pins the shared lane-boundary refusal instead.
func TestMCPVerbParity_LaneLabel(t *testing.T) {
	rows := []struct {
		name string
		verb []string
		tool func(t *testing.T, args map[string]any) error
		args map[string]any
	}{
		{"next", []string{"next", "--run", fcRun}, callFactoryTool(handleFactoryNext), map[string]any{"run": fcRun}},
		{"stage", []string{"stage", "t1", "plan-audit"}, callFactoryTool(handleFactoryStage), map[string]any{"card": "t1", "state": "plan-audit", "run": fcRun}},
		{"complete", []string{"complete", "t1", "--run", fcRun}, callFactoryTool(handleFactoryComplete), map[string]any{"card": "t1", "run": fcRun}},
		{"renew", []string{"renew", "t1", "--run", fcRun}, callFactoryTool(handleFactoryRenew), map[string]any{"card": "t1", "run": fcRun}},
	}
	labels := []struct{ name, label, want string }{
		{"legacy spelling", "worker-1", "legacy lane label"},
		{"empty", "", "is empty"},
	}
	for _, r := range rows {
		for _, l := range labels {
			t.Run(r.name+"/"+l.name, func(t *testing.T) {
				root, _ := sdMoaiFixture(t)
				sdLaneEnv(t, l.label, "")
				t.Chdir(root)

				_, _, verbErr := runFactory(t, r.verb...)
				if verbErr == nil || !strings.Contains(verbErr.Error(), l.want) {
					t.Fatalf("verb %s with label %q: err = %v, want a refusal containing %q", r.name, l.label, verbErr, l.want)
				}
				args := map[string]any{"project_root": root}
				for k, v := range r.args {
					args[k] = v
				}
				toolErr := r.tool(t, args)
				if toolErr == nil || !strings.Contains(toolErr.Error(), verbErr.Error()) {
					t.Errorf("tool %s with label %q: err = %v, want the verb's refusal %q", r.name, l.label, toolErr, verbErr)
				}
			})
		}
	}
	t.Run("decide/lane boundary", func(t *testing.T) {
		root, _ := sdMoaiFixture(t)
		sdLaneEnv(t, "worker-1", "")
		t.Chdir(root)
		_, _, verbErr := runFactory(t, "decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun)
		if verbErr == nil || !strings.Contains(verbErr.Error(), factoryLaneBoundarySentinel) {
			t.Fatalf("verb decide in a lane session: err = %v, want the lane-boundary refusal", verbErr)
		}
		toolErr := callFactoryTool(handleFactoryDecide)(t, map[string]any{"card": "t1", "gate": "kickoff", "choice": "approve", "run": fcRun, "project_root": root})
		if toolErr == nil || !strings.Contains(toolErr.Error(), verbErr.Error()) {
			t.Errorf("tool decide in a lane session: err = %v, want the verb's refusal %q", toolErr, verbErr)
		}
	})
}

// callFactoryTool adapts an MCP handler to the error-only surface the table compares.
func callFactoryTool(fn func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(t *testing.T, args map[string]any) error {
	return func(t *testing.T, args map[string]any) error {
		t.Helper()
		return sdCallToolErr(t, fn, args)
	}
}

// parityTree is a MoAI project root (the parent checkout) plus one linked worktree of it —
// the shape a lane's MCP call arrives from: project_root names the linked tree while the
// factory record, the queue and the integration window live at the parent.
func parityTree(t *testing.T) (root, wt string) {
	t.Helper()
	root, _ = sdMoaiFixture(t)
	wt = filepath.Join(filepath.Dir(root), filepath.Base(root)+"-wt")
	fcGit(t, root, "worktree", "add", "-q", wt, "-b", "WT-parity")
	return root, wt
}

// TestMCPVerbParity_RecordRoot: the verb (run from the parent, as the launcher does) and the
// MCP handler (told a linked worktree as project_root) act on the same factory record, so the
// same card ends in the same state on both surfaces.
func TestMCPVerbParity_RecordRoot(t *testing.T) {
	live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
	rows := []struct {
		name    string
		env     func(t *testing.T, root string)
		place   func(t *testing.T, root string)
		verb    []string
		tool    func(t *testing.T, args map[string]any) error
		args    map[string]any
		observe func(c homestate.Card) string
	}{
		{
			name: "decide",
			env:  func(t *testing.T, _ string) { sdClearLaneEnv(t) },
			place: func(t *testing.T, root string) {
				fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardKickoff, DecisionGate: homestate.DecisionGateKickoff})
			},
			verb: []string{"decide", "t1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun},
			tool: callFactoryTool(handleFactoryDecide),
			args: map[string]any{"card": "t1", "gate": "kickoff", "choice": "approve", "run": fcRun},
			observe: func(c homestate.Card) string {
				return strings.Join([]string{c.State, c.Stage}, "/")
			},
		},
		{
			name: "renew",
			env: func(t *testing.T, root string) {
				sdRegisterLane(t, root, "lane-1")
				sdLaneEnv(t, "lane-1", "")
			},
			place: func(t *testing.T, root string) {
				fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
					OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root})
			},
			verb:    []string{"renew", "t1", "--run", fcRun},
			tool:    callFactoryTool(handleFactoryRenew),
			args:    map[string]any{"card": "t1", "run": fcRun},
			observe: func(c homestate.Card) string { return c.LeaseExpiresAt },
		},
		{
			name: "stage",
			env: func(t *testing.T, root string) {
				sdRegisterLane(t, root, "lane-1")
				sdLaneEnv(t, "lane-1", "")
			},
			place: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "plan.md"), []byte("plan\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				fcGit(t, root, "add", "plan.md")
				fcGit(t, root, "commit", "-q", "-m", "plan artifact")
				fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
					OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root})
			},
			verb:    nil, // built per arm: the evidence names the arm's own commit
			tool:    callFactoryTool(handleFactoryStage),
			args:    map[string]any{"card": "t1", "state": homestate.CardPlanAudit, "run": fcRun},
			observe: func(c homestate.Card) string { return strings.Join([]string{c.State, c.Stage, c.LeaseExpiresAt}, "/") },
		},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			arm := func(viaMCP bool) (string, string) {
				root, wt := parityTree(t)
				r.env(t, root)
				r.place(t, root)
				t.Chdir(root)
				var err error
				if viaMCP {
					args := map[string]any{"project_root": wt}
					for k, v := range r.args {
						args[k] = v
					}
					if r.name == "stage" {
						args["evidence"] = fcGit(t, root, "rev-parse", "HEAD") + ":plan.md"
					}
					err = r.tool(t, args)
				} else {
					verb := r.verb
					if r.name == "stage" {
						verb = []string{"stage", "t1", homestate.CardPlanAudit, fcGit(t, root, "rev-parse", "HEAD") + ":plan.md", "--run", fcRun}
					}
					_, _, err = runFactory(t, verb...)
				}
				got := "ok"
				if err != nil {
					got = "refused"
				}
				return got, r.observe(fcCard(t, root, "t1"))
			}
			verbRes, verbObs := arm(false)
			toolRes, toolObs := arm(true)
			if verbRes != "ok" {
				t.Fatalf("verb arm = %s, want ok (the fixture must let the verb succeed)", verbRes)
			}
			if toolRes != verbRes || toolObs != verbObs {
				t.Errorf("tool arm = %s / %q, verb arm = %s / %q: the handler acted on a different record than the verb", toolRes, toolObs, verbRes, verbObs)
			}
		})
	}
}

// TestMCPVerbParity_CompleteWindowRoot: complete reads and writes the integration window and
// the re-measure record at the parent checkout on both surfaces (the verb's integrationLockRoot),
// not at the caller's linked tree. Row one: a window another lane holds at the parent refuses
// the second lane naming the holder. Row two: a re-measure record written at the parent by the
// lane's usual measurement is found, and the card reaches merged-local.
func TestMCPVerbParity_CompleteWindowRoot(t *testing.T) {
	arm := func(t *testing.T, viaMCP, holderOther bool) (string, string, string) {
		t.Helper()
		root, integWT, cards := sdMergeFixture(t, true, true, false, 2)
		sdPlaceMergeReady(t, root, "t1", "lane-1", cards[0])
		sdPlaceMergeReady(t, root, "t2", "lane-2", cards[1])
		lane, card, tree := "lane-1", "t1", cards[0]
		if holderOther {
			lane, card, tree = "lane-2", "t2", cards[1]
		}
		sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", factory.BranchSourceConfig, integWT, "t1")
		if _, err := factory.RunRemeasure(root, tree.wt, "develop", "true"); err != nil {
			t.Fatalf("remeasure at the parent: %v", err)
		}
		sdLaneEnv(t, lane, "")
		t.Setenv(config.EnvClaudeCodeSessionID, "sess-"+lane)
		t.Chdir(root)
		var err error
		if viaMCP {
			err = callFactoryTool(handleFactoryComplete)(t, map[string]any{"card": card, "run": fcRun, "project_root": tree.wt})
		} else {
			_, _, err = runFactory(t, "complete", card, "--run", fcRun)
		}
		errText := ""
		if err != nil {
			errText = err.Error()
		}
		return errText, fcCard(t, root, card).State, sdWindow(t, root).SessionID
	}
	t.Run("a window another lane holds refuses the second lane naming the holder", func(t *testing.T) {
		verbErr, verbState, _ := arm(t, false, true)
		if !strings.Contains(verbErr, "lane-1") {
			t.Fatalf("verb arm: err = %q, want a refusal naming lane-1", verbErr)
		}
		toolErr, toolState, _ := arm(t, true, true)
		if !strings.Contains(toolErr, verbErr) {
			t.Errorf("tool arm: err = %q, want the verb's refusal %q", toolErr, verbErr)
		}
		if toolState != verbState {
			t.Errorf("tool arm left the card %q, verb arm %q", toolState, verbState)
		}
	})
	t.Run("a re-measure record written at the parent is found and the card lands", func(t *testing.T) {
		verbErr, verbState, _ := arm(t, false, false)
		if verbErr != "" || verbState != homestate.CardMergedLocal {
			t.Fatalf("verb arm: err = %q state = %q, want a clean landing", verbErr, verbState)
		}
		toolErr, toolState, _ := arm(t, true, false)
		if toolErr != "" || toolState != homestate.CardMergedLocal {
			t.Errorf("tool arm: err = %q state = %q, want the verb's clean landing", toolErr, toolState)
		}
	})
}

// TestMCPVerbParity_DecideConfigRoot: decide reads the git-flow integration target and the
// queue record at the parent checkout on both surfaces. A push decision on a merged-local
// card needs both (the target, then the card's backlog identity for the leader receipt), so
// a handler reading the linked worktree's empty config refuses where the verb lands the card.
func TestMCPVerbParity_DecideConfigRoot(t *testing.T) {
	arm := func(viaMCP bool) homestate.Card {
		root, wt := parityTree(t)
		sdClearLaneEnv(t)
		fcGitFlowConfig(t, root)
		dir, merge := fcRepo(t, true)
		fcGit(t, dir, "push", "-q", "origin", "integration")
		store := factory.NewBacklogStore(todoBacklogPath(root))
		cardID := addShapeCard(t, "decide config root card")
		fcLinkRuntime(t, root, cardID)
		fcBindDispatch(t, root, cardID, fcRun)
		fcPlace(t, root, homestate.Card{CardID: cardID, RunID: fcRun, State: homestate.CardMergedLocal, OwnerLabel: "lane-1",
			Version: 1, MergeSHA: merge, WorktreePath: dir, EvidenceSHA: merge})
		fcPlaceApprovalRaw(t, root, homestate.LeaderApproval{
			CardUUID: recheckUUID(t, root, store, cardID), RunID: fcRun, CardID: cardID, FactoryVersion: 1,
			EvidenceHash: merge, Issuer: "lead", IssuerRole: homestate.ApprovalIssuerLeader,
		})
		t.Chdir(root)
		var err error
		if viaMCP {
			err = callFactoryTool(handleFactoryDecide)(t, map[string]any{"card": cardID, "gate": "push", "run": fcRun, "project_root": wt})
		} else {
			_, _, err = runFactory(t, "decide", cardID, "--gate", "push", "--run", fcRun)
		}
		if err != nil {
			t.Logf("decide (viaMCP=%v): %v", viaMCP, err)
		}
		return fcCard(t, root, cardID)
	}
	verb := arm(false)
	if verb.State != homestate.CardPushed {
		t.Fatalf("verb arm: card = %s, want pushed (the fixture must let the verb land it)", verb.State)
	}
	if tool := arm(true); tool.State != verb.State || tool.Version != verb.Version {
		t.Errorf("tool arm: card = %s v%d, verb arm = %s v%d: the handler read another tree's config or queue", tool.State, tool.Version, verb.State, verb.Version)
	}
}
