// mcp_refusal_test.go — SPEC-MCP-NEUTRALITY-001 M1 (card t1632): the
// structured policy-refusal wire contract (REQ-MN-001, design §D1), the actor
// argument with env-as-default and its owned / stale-run refusals
// (REQ-MN-003 / AC-MN-004), and the positive operator mark on the
// operator-only verbs (REQ-MN-003 / AC-MN-005).
//
// The refusals are asserted on the WIRE (is_error false, a structured payload
// whose refused field is true, the same JSON in the text content block)
// through a local decoder, deliberately decoupled from the production type:
// the contract under test is the payload shape, not the Go struct.
package cli

import (
	"context"
	"encoding/json"
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

// mcpRefusalShape is the test-side decoder for the refusal payload the wire
// carries (design §D1): refused, an optional closed-set code, a human reason,
// an optional hint.
type mcpRefusalShape struct {
	Refused bool   `json:"refused"`
	Code    string `json:"code"`
	Reason  string `json:"reason"`
	Hint    string `json:"hint"`
}

// mcpCallRaw invokes an MCP handler the way the mcp-server dispatch does and
// returns the raw CallToolResult, so a structured refusal is asserted on the
// wire fields instead of the flattened error surface sdCallTool builds
// (factory_m3_test.go).
func mcpCallRaw(t *testing.T, fn func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := fn(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	if err != nil {
		t.Fatalf("handler returned a Go error: %v", err)
	}
	if res == nil {
		t.Fatal("handler returned a nil result")
	}
	return res
}

// mcpTextOf flattens a result's text content blocks.
func mcpTextOf(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

// assertStructuredRefusal asserts the policy-refusal wire shape
// (SPEC-MCP-NEUTRALITY-001 REQ-MN-001, design §D1): is_error false, the text
// block parses to the refusal JSON with refused true, and the structured
// content block serializes to the SAME bytes. wantCode pins the closed-set
// code; an empty wantCode demands NO code (the identity class the closed set
// does not name). Returns the reason for further assertions.
func assertStructuredRefusal(t *testing.T, res *mcp.CallToolResult, wantCode string, wantReasonContains ...string) string {
	t.Helper()
	if res.IsError {
		t.Fatalf("refusal rode the error channel (is_error true); text=%s", mcpTextOf(res))
	}
	var payload mcpRefusalShape
	if err := json.Unmarshal([]byte(mcpTextOf(res)), &payload); err != nil {
		t.Fatalf("text block is not the refusal JSON: %v; text=%s", err, mcpTextOf(res))
	}
	if !payload.Refused {
		t.Fatalf("payload refused = false: %+v", payload)
	}
	if payload.Code != wantCode {
		t.Errorf("refusal code = %q, want %q", payload.Code, wantCode)
	}
	for _, frag := range wantReasonContains {
		if !strings.Contains(payload.Reason, frag) {
			t.Errorf("refusal reason %q does not carry %q", payload.Reason, frag)
		}
	}
	structured, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	if string(structured) != mcpTextOf(res) {
		t.Errorf("structured content %s differs from the text block %s", structured, mcpTextOf(res))
	}
	return payload.Reason
}

// sdSeedRun inserts (or re-states) a runs row so run-id validation sees it.
func sdSeedRun(t *testing.T, root, runID, status string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,created_at,updated_at) VALUES(?,?,?,?) ON CONFLICT(run_id) DO UPDATE SET status=excluded.status`, runID, status, "2026-09-26T00:00:00Z", "2026-09-26T00:00:00Z"); err != nil {
		t.Fatalf("seed run %s: %v", runID, err)
	}
}

// AC-MN-001 (debt N3 serial-slot instance) — the serial-slot policy refusal
// rides the structured result, never the error channel: a live serial-slot
// lease held by another card refuses a nominated lease with refused: true,
// code serial-slot, is_error false.
func TestMCPRefusal_SerialSlotStructured(t *testing.T) {
	root, store := sdMoaiFixture(t)
	sdSeedRun(t, root, fcRun, "active")
	sdRegisterLane(t, root, "lane-1")
	// t-slot holds the serial slot: in flight, driven, live lease.
	fcPlace(t, root, homestate.Card{
		CardID: "t-slot", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
		WorktreePath:   root,
	})
	fcQueue(t, store, factory.BacklogStatePicked) // t1: the nominee
	// SPEC-FACTORY-DEFAULT-MODE-001: the absent read now derives
	// parallelizable, so the serial-slot refusal needs the holder's serial
	// judgment recorded on the queue record — the flSeedDrivenQueueItem mold
	// (queue item + explicit serial classification on the holder).
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items = append(rec.Items, factory.BacklogItem{
			ID:      "t-slot",
			Text:    "serial slot holder",
			AddedAt: "2026-09-26T00:00:00Z",
			State:   factory.BacklogStatePicked,
		})
		return nil
	}); err != nil {
		t.Fatalf("seed t-slot queue item: %v", err)
	}
	fcClassify(t, store, "t-slot", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	// The nominee is serial too: the structured serial-slot refusal is a
	// serial nominee against a serial holder, and after the flip an unrecorded
	// nominee reads parallelizable — the refusal would never arm.
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	res := mcpCallRaw(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun, "card": "t1"})
	assertStructuredRefusal(t, res, "serial-slot", "serial slot")
}

// AC-MN-004 — the actor argument: the explicit actor decides whose label acts,
// env is the default only. actor lane-2 on a card the record holds for lane-1
// refuses owned; the same actor on a completed run refuses stale-run; actor
// lane-1 on the active run proceeds.
func TestMCPFactoryActor_OwnedAndStaleRun(t *testing.T) {
	build := func(t *testing.T) (string, string) {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		sdSeedRun(t, root, fcRun, "active")
		sdSeedRun(t, root, "run-retired", "retired")
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		if err := os.WriteFile(filepath.Join(root, "plan.md"), []byte("plan\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, root, "add", "-A")
		fcGit(t, root, "commit", "-q", "-m", "plan artifact")
		sha := fcGit(t, root, "rev-parse", "HEAD")
		fcPlace(t, root, homestate.Card{
			CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
			OwnerLabel: "lane-1", LeaseHolder: "lane-1",
			LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
			WorktreePath:   root,
		})
		return root, sha
	}

	t.Run("actor lane-2 on a lane-1 card refuses owned, is_error false", func(t *testing.T) {
		root, sha := build(t)
		sdLaneEnv(t, "lane-1", "") // the spawn is recorded as lane-1
		t.Chdir(root)
		res := mcpCallRaw(t, handleFactoryStage, map[string]any{
			"card": "t1", "state": "plan-audit", "evidence": sha + ":plan.md",
			"run": fcRun, "project_root": root, "actor": "lane-2",
		})
		assertStructuredRefusal(t, res, "owned", "lane-2")
	})

	t.Run("actor lane-1 on a retired run refuses stale-run, is_error false", func(t *testing.T) {
		root, sha := build(t)
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		res := mcpCallRaw(t, handleFactoryStage, map[string]any{
			"card": "t1", "state": "plan-audit", "evidence": sha + ":plan.md",
			"run": "run-retired", "project_root": root, "actor": "lane-1",
		})
		assertStructuredRefusal(t, res, "stale-run", "run-retired")
	})

	t.Run("actor lane-1 on the active run proceeds", func(t *testing.T) {
		root, sha := build(t)
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		res := mcpCallRaw(t, handleFactoryStage, map[string]any{
			"card": "t1", "state": "plan-audit", "evidence": sha + ":plan.md",
			"run": fcRun, "project_root": root, "actor": "lane-1",
		})
		if res.IsError {
			t.Fatalf("lane-1 stage refused: %s", mcpTextOf(res))
		}
		if c := fcCard(t, root, "t1"); c.State != homestate.CardPlanAudit {
			t.Errorf("t1 = %s, want plan-audit", c.State)
		}
	})
}

// The factory_renew MCP guard rails: the actor argument refuses legacy
// spellings as an argument error (AC-SD-021's rule, actor form), a malformed
// run id is an argument error — never a policy refusal — and an explicit
// actor renews the lane's own lease.
func TestMCPRenew_ActorAndRunGuard(t *testing.T) {
	build := func(t *testing.T) string {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		sdRegisterLane(t, root, "lane-1")
		fcPlace(t, root, homestate.Card{
			CardID: "t1", State: homestate.CardPlan, Stage: homestate.CardPlan,
			OwnerLabel: "lane-1", LeaseHolder: "lane-1",
			LeaseExpiresAt: fcNow.Add(10 * time.Minute).Format(time.RFC3339Nano),
			WorktreePath:   root,
		})
		return root
	}

	t.Run("a legacy actor label is an argument error, not a refusal", func(t *testing.T) {
		root := build(t)
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		_, err := sdCallTool(t, handleFactoryRenew, map[string]any{"card": "t1", "run": fcRun, "project_root": root, "actor": "worker-1"})
		if err == nil || !strings.Contains(err.Error(), "legacy lane label") {
			t.Fatalf("renew with a legacy actor: err = %v, want the legacy-label rejection", err)
		}
	})

	t.Run("a malformed run id is an argument error, not a policy refusal", func(t *testing.T) {
		root := build(t)
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		res := mcpCallRaw(t, handleFactoryRenew, map[string]any{"card": "t1", "run": "../bad", "project_root": root, "actor": "lane-1"})
		if !res.IsError {
			t.Fatalf("malformed run accepted: %s", mcpTextOf(res))
		}
		if strings.Contains(mcpTextOf(res), "\"refused\"") {
			t.Errorf("malformed run id classified as a policy refusal: %s", mcpTextOf(res))
		}
	})

	t.Run("an explicit actor renews the lane's own lease", func(t *testing.T) {
		root := build(t)
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		out, err := sdCallTool(t, handleFactoryRenew, map[string]any{"card": "t1", "run": fcRun, "project_root": root, "actor": "lane-1"})
		if err != nil || !strings.Contains(out, "renewed") {
			t.Fatalf("renew with actor lane-1: out=%q err=%v", out, err)
		}
	})
}

// AC-MN-005 — the operator-only verbs require a POSITIVE operator mark
// (MOAI_FACTORY_ROLE=operator): an unmarked session and a lane session refuse
// with refused: true (unknown attribution carries no closed-set code; the lane
// arm keeps the CLI-identical boundary text), and the marked session proceeds.
func TestMCPOperatorMark_Required(t *testing.T) {
	type verbFixture struct {
		name string
		args func(root string) map[string]any
		fn   func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}
	build := func(t *testing.T) string {
		t.Helper()
		root, store := sdMoaiFixture(t)
		sdSeedRun(t, root, fcRun, "active")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardKickoff, DecisionGate: homestate.DecisionGateKickoff})
		// A queued card so the todo_claim arms fail (or proceed) on the
		// operator gate, never on an empty queue.
		if _, _, err := storeAdd(t, store, "claimable card"); err != nil {
			t.Fatalf("seed queue: %v", err)
		}
		return root
	}
	verbs := []verbFixture{
		{
			name: "factory_decide",
			args: func(root string) map[string]any {
				return map[string]any{"card": "t1", "gate": "kickoff", "choice": "approve", "run": fcRun, "project_root": root}
			},
			fn: handleFactoryDecide,
		},
		{
			name: "todo_claim",
			args: func(root string) map[string]any { return map[string]any{"project_root": root} },
			fn:   handleTodoClaim,
		},
		{
			name: "todo_add",
			args: func(root string) map[string]any { return map[string]any{"text": "operator card", "project_root": root} },
			fn:   handleTodoAdd,
		},
	}

	for _, v := range verbs {
		t.Run(v.name+" refuses an unmarked session (unknown attribution)", func(t *testing.T) {
			sdClearLaneEnv(t)
			root := build(t)
			res := mcpCallRaw(t, v.fn, v.args(root))
			assertStructuredRefusal(t, res, "", "operator")
		})

		t.Run(v.name+" refuses a lane session with the boundary text", func(t *testing.T) {
			sdClearLaneEnv(t)
			root := build(t)
			sdLaneEnv(t, "lane-1", "")
			// Card t1632 main-absorb r9: factory_decide's boundary refusal
			// rides the ERROR channel with the CLI-identical text (the t1914
			// parity contract); the todo verbs keep the structured refusal.
			if v.name == "factory_decide" {
				err := sdCallToolErr(t, v.fn, v.args(root))
				if err == nil || !strings.Contains(err.Error(), "lane boundary") {
					t.Errorf("%s lane session: err = %v, want the boundary refusal", v.name, err)
				}
				return
			}
			res := mcpCallRaw(t, v.fn, v.args(root))
			assertStructuredRefusal(t, res, "")
		})

		t.Run(v.name+" proceeds under the operator mark", func(t *testing.T) {
			sdClearLaneEnv(t)
			root := build(t)
			// The operator role VALUE is the launcher-stamped contract
			// (decision-index Q2 default); pinned as the literal here the way
			// envkeys_factory_role_test.go pins the variable name.
			t.Setenv(config.EnvFactoryRole, "operator")
			res := mcpCallRaw(t, v.fn, v.args(root))
			if res.IsError {
				t.Fatalf("%s refused under the operator mark: %s", v.name, mcpTextOf(res))
			}
		})
	}
}
