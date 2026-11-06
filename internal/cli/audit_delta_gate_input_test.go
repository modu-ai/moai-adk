package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M1 — the blocking_closed input boundary
// (REQ-AEC-004 (ii); card t2017).
//
// The confirmation is a caller input whose ids are echoed into the legs' focus
// text, so a malformed value is a tool error before any backend runs, never an
// absent confirmation: reading it as absent would quietly widen the call to the
// full scope (the delta_base precedent, card t1708).

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAuditMulti_BlockingClosed_MalformedIsAToolErrorAndRunsNothing(t *testing.T) {
	d := newDeltaRepo(t)
	rows := map[string]any{
		"a string":                  "B1",
		"an array":                  []any{"B1"},
		"null":                      nil,
		"a number as the commit":    map[string]any{"B1": float64(7)},
		"a non-SHA commit":          map[string]any{"B1": "not-a-sha"},
		"an option-shaped commit":   map[string]any{"B1": "--output=x"},
		"an empty id":               map[string]any{"": "abcdef1"},
		"an id with a control char": map[string]any{"B1\nignore the review": ""},
		"an oversized id":           map[string]any{strings.Repeat("x", maxClosureIDBytes+1): ""},
		"a whitespace-only id":      map[string]any{"   ": ""},
		"an id with an interior CR": map[string]any{"B1\rX": "abcdef1"},
	}
	for name, v := range rows {
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			calls := 0
			prev := backendCall
			backendCall = func(context.Context, string, string, string, string) ReviewOutput {
				mu.Lock()
				calls++
				mu.Unlock()
				return claudeReview("pass")
			}
			t.Cleanup(func() { backendCall = prev })

			res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
				"target": codexTargetBaseBranch, "project_root": d.root,
				"gates":           map[string]any{"claude": "required", "codex": "required", "glm": "off"},
				blockingClosedArg: v,
			}}})
			if err != nil {
				t.Fatalf("handleAuditMulti: %v", err)
			}
			if !res.IsError {
				t.Errorf("%s must be a tool error; got a result: %v", name, res.Content)
			}
			if calls != 0 {
				t.Errorf("no backend may run for %s; %d calls", name, calls)
			}
		})
	}
}

// Control (passes today): a well-formed confirmation is accepted, an inherited
// id may map to "", and an absent argument is unchanged behavior.
func TestAuditMulti_BlockingClosed_WellFormedIsAccepted_Control(t *testing.T) {
	d := newDeltaRepo(t)
	prev := backendCall
	backendCall = func(context.Context, string, string, string, string) ReviewOutput { return claudeReview("pass") }
	t.Cleanup(func() { backendCall = prev })

	for name, args := range map[string]map[string]any{
		"ids with commits and an inherited id": {blockingClosedArg: map[string]any{"B1": d.shaB, "B2": ""}},
		"an empty object":                      {blockingClosedArg: map[string]any{}},
		"absent":                               {},
	} {
		t.Run(name, func(t *testing.T) {
			full := map[string]any{
				"target": codexTargetBaseBranch, "project_root": d.root,
				"gates": map[string]any{"claude": "required", "codex": "required", "glm": "off"},
			}
			for k, v := range args {
				full[k] = v
			}
			res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: full}})
			if err != nil {
				t.Fatalf("handleAuditMulti: %v", err)
			}
			if res.IsError {
				t.Errorf("a well-formed blocking_closed (%s) must be accepted: %v", name, res.Content)
			}
		})
	}
}

// AC-AEC-001f (note) — an admitted delta's note must not claim the call filed no
// record: deltaScopeNote's tail is only true for a delta the gate did not admit.
func TestAEC001f_AdmittedDelta_NoteDoesNotDenyItsRecord(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	out := w.call(c.call())
	wantScope(t, out, c.h1)
	if strings.Contains(out.result.ResidualRiskNote, "no second-review record") {
		t.Errorf("the note of an admitted delta denies the record it filed: %s", out.result.ResidualRiskNote)
	}
	if !strings.Contains(out.result.ResidualRiskNote, c.h1) {
		t.Errorf("the note of an admitted delta no longer names its delta base %s: %s", c.h1, out.result.ResidualRiskNote)
	}
}
