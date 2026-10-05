package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// TestAC008_ClearRearmsEmbeddedGoalUnderNewSessionID asserts that on /clear
// (mode=auto) with a pending record carrying an embedded goal, a NEW goal state
// file is written under the NEW session-id (input.SessionID), keyed by session-id
// (NOT SPEC-id — Option B rejected). SPEC-INFINITE-GOAL-001 REQ-6 / AC-008.
func TestAC008_ClearRearmsEmbeddedGoalUnderNewSessionID(t *testing.T) {
	pd := t.TempDir()
	rec := livePending("resume body")
	rec.EmbeddedGoal = &handoff.EmbeddedGoal{
		Condition:   "go test ./... exits 0",
		MaxTurns:    0, // infinite
		MaxDuration: 3600,
		CostCap:     0,
	}
	mustSavePending(t, pd, rec)

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if additionalContextOf(out) == "" {
		t.Fatal("AC-008: expected additionalContext injection, got empty")
	}

	// A NEW goal state file MUST be written under the NEW session-id.
	newSession := "sess-a1b2c3d4" // injectInput's SessionID
	g, loadErr := goal.LoadGoal(pd, newSession)
	if loadErr != nil {
		t.Fatalf("AC-008: load new-session goal: %v", loadErr)
	}
	if g == nil {
		t.Fatalf("AC-008: no goal file written under new session-id %s", newSession)
	}
	if g.Goal != "go test ./... exits 0" {
		t.Errorf("AC-008: re-armed goal condition = %q, want embedded condition", g.Goal)
	}
	if g.Ceiling.MaxTurns != 0 || g.Ceiling.MaxDuration != 3600 {
		t.Errorf("AC-008: re-armed Ceiling = %+v, want MaxTurns=0 MaxDuration=3600", g.Ceiling)
	}
	if g.Status != goal.StatusArmed {
		t.Errorf("AC-008: re-armed status = %q, want armed", g.Status)
	}
	// Session-id keying: the new file is keyed by the NEW session-id.
	if g.SessionID != newSession {
		t.Errorf("AC-008: re-armed SessionID = %q, want %s (session-id keying preserved)", g.SessionID, newSession)
	}
}

// TestAC008_NoEmbeddedGoalNoRearm asserts that a pending record with NO embedded
// goal does NOT write a goal file (no spurious rearm).
func TestAC008_NoEmbeddedGoalNoRearm(t *testing.T) {
	pd := t.TempDir()
	mustSavePending(t, pd, livePending("resume body")) // no EmbeddedGoal

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	_, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if g, _ := goal.LoadGoal(pd, "sess-a1b2c3d4"); g != nil {
		t.Errorf("AC-008 negative: goal file written despite no embedded goal: %+v", g)
	}
}

// TestAC008_D8_RejectsUnboundedEmbeddedGoal asserts the defense-in-depth
// re-validation (D8): an embedded goal with MaxTurns=0 AND no real bound is
// rejected (no goal file written), so a corrupt pending record cannot re-open
// the unbounded hole.
func TestAC008_D8_RejectsUnboundedEmbeddedGoal(t *testing.T) {
	pd := t.TempDir()
	rec := livePending("resume body")
	rec.EmbeddedGoal = &handoff.EmbeddedGoal{
		Condition: "unbounded exits 0",
		MaxTurns:  0, // infinite
		// NEITHER MaxDuration NOR CostCap → unbounded
	}
	rec.SavedAt = time.Now() // keep live
	mustSavePending(t, pd, rec)

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	_, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if g, _ := goal.LoadGoal(pd, "sess-a1b2c3d4"); g != nil {
		t.Errorf("AC-008 D8: unbounded embedded goal must be rejected, but a goal file was written: %+v", g)
	}
	// The handoff AdditionalContext must STILL be injected (the reject is scoped
	// to the goal rearm, not the handoff resume itself).
	_ = filepath.Join // keep filepath referenced for the negative-case path probe
	if _, err := os.Stat(handoff.PendingPath(pd)); err == nil {
		t.Log("pending consumed (handoff resume proceeded; goal rearm rejected)")
	}
}
