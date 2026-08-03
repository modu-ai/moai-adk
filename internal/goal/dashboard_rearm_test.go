package goal

import (
	"strings"
	"testing"
)

// TestRenderDashboardReArm_EmbeddedGoalIndicator verifies AC-GHF-007 state (a):
// when pending.json carries a non-nil embedded_goal whose IsUnbounded() is
// false, the dashboard renders a "re-arm on /clear" indicator carrying the
// embedded condition text + embedded ceiling.
func TestRenderDashboardReArm_EmbeddedGoalIndicator(t *testing.T) {
	g := NewGoal("sess-a", "live goal text", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	reArm := &ReArmContext{
		EmbeddedCondition:   "embedded condition text",
		EmbeddedMaxTurns:    20,
		EmbeddedMaxDuration: 300,
		EmbeddedUnbounded:   false,
	}
	raw, err := RenderDashboardReArm(g, nil, reArm)
	if err != nil {
		t.Fatalf("RenderDashboardReArm: %v", err)
	}
	bt := bodyText(parseHTML(t, raw))
	if !strings.Contains(bt, "re-arm") || !strings.Contains(strings.ToLower(bt), "/clear") {
		t.Errorf("missing re-arm indicator; got:\n%s", bt)
	}
	if !strings.Contains(bt, "embedded condition text") {
		t.Errorf("re-arm indicator missing embedded condition text")
	}
	if !strings.Contains(bt, "20") || !strings.Contains(bt, "300") {
		t.Errorf("re-arm indicator missing embedded ceiling (max_turns/max_duration)")
	}
}

// TestRenderDashboardReArm_PostClearReArmedView verifies AC-GHF-007 state (b):
// when a post-/clear new-session goal file exists, the dashboard renders a
// "re-armed under <new-session-id>" view with a pointer to the new goal file.
func TestRenderDashboardReArm_PostClearReArmedView(t *testing.T) {
	g := NewGoal("old-sess", "prior goal", []Condition{
		{Type: ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	reArm := &ReArmContext{
		NewSessionID: "new-sess-abc",
	}
	raw, err := RenderDashboardReArm(g, nil, reArm)
	if err != nil {
		t.Fatalf("RenderDashboardReArm: %v", err)
	}
	bt := bodyText(parseHTML(t, raw))
	if !strings.Contains(strings.ToLower(bt), "re-armed") {
		t.Errorf("missing 're-armed' view; got:\n%s", bt)
	}
	if !strings.Contains(bt, "new-sess-abc") {
		t.Errorf("re-armed view missing new session id")
	}
}

// TestRenderDashboardReArm_UnboundedBanner verifies AC-GHF-007 state (c): when
// the embedded goal IsUnbounded() is true, the dashboard renders a D8-rejection
// banner naming the unbounded embedded goal as the cause.
func TestRenderDashboardReArm_UnboundedBanner(t *testing.T) {
	g := NewGoal("sess-c", "live goal", []Condition{
		{Type: ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	reArm := &ReArmContext{
		EmbeddedCondition: "unbounded embedded goal",
		EmbeddedMaxTurns:  0, // infinite
		EmbeddedUnbounded: true,
	}
	raw, err := RenderDashboardReArm(g, nil, reArm)
	if err != nil {
		t.Fatalf("RenderDashboardReArm: %v", err)
	}
	bt := bodyText(parseHTML(t, raw))
	if !strings.Contains(strings.ToLower(bt), "unbounded") || !strings.Contains(strings.ToLower(bt), "reject") {
		t.Errorf("missing D8 unbounded-rejection banner; got:\n%s", bt)
	}
}

// TestRenderDashboardReArm_NilContextMatchesBase verifies that a nil reArm
// context produces output identical to the base RenderDashboard call (the
// re-arm path is purely additive; zero re-arm state = base dashboard).
func TestRenderDashboardReArm_NilContextMatchesBase(t *testing.T) {
	g := NewGoal("s", "g", []Condition{
		{Type: ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	base, err := RenderDashboard(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	reArm, err := RenderDashboardReArm(g, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(base) != string(reArm) {
		t.Errorf("nil reArm context must produce byte-identical output to base RenderDashboard")
	}
}
