package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// writePendingRecord writes a minimal pending.json carrying the given embedded goal.
func writePendingRecord(t *testing.T, root string, eg *handoff.EmbeddedGoal) {
	t.Helper()
	rec := &handoff.PendingRecord{
		SchemaVersion: handoff.PendingSchemaVersion,
		SpecID:        "SPEC-TEST",
		Phase:         "run",
		Body:          "test resume body",
		EmbeddedGoal:  eg,
	}
	if err := handoff.SavePending(root, rec); err != nil {
		t.Fatalf("SavePending: %v", err)
	}
}

// TestRunGoalRender_ReArmIndicatorFromEmbeddedGoal verifies AC-WIRE-007/008(a):
// with pending.json carrying a bounded EmbeddedGoal, runGoalRender constructs a
// *goal.ReArmContext and the rendered DOM shows the "re-arm on /clear" indicator
// mentioning the embedded condition.
func TestRunGoalRender_ReArmIndicatorFromEmbeddedGoal(t *testing.T) {
	root := t.TempDir()
	sid := "sess-rearm-a"
	writeGoalFixture(t, root, sid)
	writePendingRecord(t, root, &handoff.EmbeddedGoal{
		Condition:   "go test ./... exits 0",
		MaxTurns:    20,
		MaxDuration: 300,
	})

	_, _, errBuf, execErr := runGoalRenderWithSession(t, root, sid, false)
	if execErr != nil {
		t.Fatalf("runGoalRender failed: %v; stderr=%s", execErr, errBuf.String())
	}
	raw, err := os.ReadFile(goal.HTMLPath(root, sid))
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	body := domBodyText(t, raw)

	if !strings.Contains(body, "re-arm on /clear") {
		t.Errorf("DOM missing the re-arm indicator; body:\n%s", body)
	}
	if !strings.Contains(body, "go test ./... exits 0") {
		t.Errorf("DOM missing the embedded condition text; body:\n%s", body)
	}
	// Bounded case → no D8 banner.
	if strings.Contains(body, "D8 rejection") {
		t.Errorf("DOM unexpectedly shows D8 banner for bounded goal; body:\n%s", body)
	}
}

// TestRunGoalRender_ReArmedUnderNewSession verifies AC-WIRE-008(b): with
// pending.json EmbeddedGoal AND a post-/clear new-session goal file, the DOM
// shows the "re-armed under <new-session-id>" view mentioning the new session.
func TestRunGoalRender_ReArmedUnderNewSession(t *testing.T) {
	root := t.TempDir()
	oldSid := "sess-rearm-old"
	newSid := "sess-rearm-new-9b7"
	condition := "go test ./... exits 0"
	writeGoalFixture(t, root, oldSid)
	// The "new session" goal file (the rearm write signature: Goal text matches
	// the embedded condition).
	newG := &goal.Goal{
		SessionID: newSid,
		Goal:      condition,
		Conditions: []goal.Condition{
			{Type: goal.ConditionMechanical, Cmd: condition, ExpectExit: 0},
		},
		Ceiling: goal.Ceiling{MaxTurns: 20},
		Status:  goal.StatusArmed,
	}
	if err := goal.SaveGoal(root, newG); err != nil {
		t.Fatalf("SaveGoal new session: %v", err)
	}
	writePendingRecord(t, root, &handoff.EmbeddedGoal{
		Condition: condition,
		MaxTurns:  20,
	})

	_, _, errBuf, execErr := runGoalRenderWithSession(t, root, oldSid, false)
	if execErr != nil {
		t.Fatalf("runGoalRender failed: %v; stderr=%s", execErr, errBuf.String())
	}
	raw, err := os.ReadFile(goal.HTMLPath(root, oldSid))
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	body := domBodyText(t, raw)

	if !strings.Contains(body, "Re-armed under session") {
		t.Errorf("DOM missing the re-armed-under view; body:\n%s", body)
	}
	if !strings.Contains(body, newSid) {
		t.Errorf("DOM missing the new session id %q; body:\n%s", newSid, body)
	}
}

// TestRunGoalRender_D8BannerForUnbounded verifies AC-WIRE-008(c): with
// pending.json carrying an unbounded EmbeddedGoal, the D8-rejection banner
// renders (and the indicator does NOT — D8 takes precedence per applyReArm).
func TestRunGoalRender_D8BannerForUnbounded(t *testing.T) {
	root := t.TempDir()
	sid := "sess-rearm-c"
	writeGoalFixture(t, root, sid)
	writePendingRecord(t, root, &handoff.EmbeddedGoal{
		Condition: "unbounded goal condition",
		MaxTurns:  0, // infinite, no real bound → IsUnbounded() == true
	})

	_, _, errBuf, execErr := runGoalRenderWithSession(t, root, sid, false)
	if execErr != nil {
		t.Fatalf("runGoalRender failed: %v; stderr=%s", execErr, errBuf.String())
	}
	raw, err := os.ReadFile(goal.HTMLPath(root, sid))
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	body := domBodyText(t, raw)

	if !strings.Contains(body, "D8 rejection") {
		t.Errorf("DOM missing D8 rejection banner; body:\n%s", body)
	}
	if strings.Contains(body, "re-arm on /clear") {
		t.Errorf("DOM unexpectedly shows the indicator for an unbounded goal (D8 takes precedence); body:\n%s", body)
	}
}

// TestRunGoalRender_NoReArmStateIsByteIdentical verifies AC-WIRE-009: no
// pending.json → runGoalRender produces byte-identical output to
// RenderDashboard(g, v) for the same (g, v). The re-arm path is purely additive.
func TestRunGoalRender_NoReArmStateIsByteIdentical(t *testing.T) {
	root := t.TempDir()
	sid := "sess-rearm-none"
	g := writeGoalFixture(t, root, sid)
	// No pending.json, no verdict sidecar.

	_, _, errBuf, execErr := runGoalRenderWithSession(t, root, sid, false)
	if execErr != nil {
		t.Fatalf("runGoalRender failed: %v; stderr=%s", execErr, errBuf.String())
	}
	raw, err := os.ReadFile(goal.HTMLPath(root, sid))
	if err != nil {
		t.Fatalf("read html: %v", err)
	}

	// Compare against a direct RenderDashboard(g, nil) for the same g.
	direct, err := goal.RenderDashboard(g, nil)
	if err != nil {
		t.Fatalf("RenderDashboard direct: %v", err)
	}
	if string(raw) != string(direct) {
		t.Errorf("AC-WIRE-009 byte-identity regression: runGoalRender output != RenderDashboard(g, nil)")
	}
}
