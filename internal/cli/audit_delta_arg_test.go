package cli

// Card t1708 (sync-audit F1) — delta_base must be a string. mcp-go's GetString
// returns the default for any non-string value, so an array, object, number,
// bool or null used to read as "absent" and the call became a silent full-scope
// review. A non-string delta_base is a tool error before any backend runs.

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// nonStringDeltaBases are the JSON values a client can send that are not strings.
func nonStringDeltaBases() map[string]any {
	return map[string]any{
		"array":  []any{"abc1234"},
		"object": map[string]any{"sha": "abc1234"},
		"number": float64(1234567),
		"bool":   true,
		"null":   nil,
	}
}

func TestCodexAudit_DeltaBase_NonStringIsAToolErrorNotAFallback(t *testing.T) {
	d := newDeltaRepo(t)
	for name, v := range nonStringDeltaBases() {
		t.Run(name, func(t *testing.T) {
			sess := withCodexSession(t, codexSessionScript("clean change, no findings"))
			res := callCodexAudit(t, deltaAuditArgs(d, map[string]any{"delta_base": v}))
			if !res.IsError {
				t.Fatalf("a %s delta_base must be a tool error; got a result: %v", name, res.Content)
			}
			if len(sess.sent) != 0 {
				t.Errorf("no codex request may be sent for a %s delta_base; sent=%v", name, sess.sent)
			}
		})
	}
}

func TestAuditMulti_DeltaBase_NonStringIsAToolErrorAndRunsNothing(t *testing.T) {
	d := newDeltaRepo(t)
	for name, v := range nonStringDeltaBases() {
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

			const session = "sess-nonstring"
			res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
				"target": codexTargetBaseBranch, "project_root": d.root, "session_id": session,
				"gates":      map[string]any{"claude": "required", "codex": "required", "glm": "off"},
				"delta_base": v,
			}}})
			if err != nil {
				t.Fatalf("handleAuditMulti: %v", err)
			}
			if !res.IsError {
				t.Fatalf("a %s delta_base must be a tool error; got a result: %v", name, res.Content)
			}
			if calls != 0 {
				t.Errorf("no backend may run for a %s delta_base; %d calls", name, calls)
			}
			state := filepath.Join(d.root, ".moai", "state", "audit-multi", session+".json")
			if _, statErr := os.Stat(state); statErr == nil {
				t.Errorf("a rejected call must write no session state; found %s", state)
			}
		})
	}
}

// Control: an explicitly empty string is still read as absent, as before.
func TestAuditMulti_DeltaBase_EmptyStringStillMeansAbsent(t *testing.T) {
	d := newDeltaRepo(t)
	prev := backendCall
	backendCall = func(context.Context, string, string, string, string) ReviewOutput { return claudeReview("pass") }
	t.Cleanup(func() { backendCall = prev })

	res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"target": codexTargetBaseBranch, "project_root": d.root,
		"gates":      map[string]any{"claude": "required", "codex": "required", "glm": "off"},
		"delta_base": "",
	}}})
	if err != nil {
		t.Fatalf("handleAuditMulti: %v", err)
	}
	if res.IsError {
		t.Fatalf("an empty delta_base must behave as absent, not a tool error: %v", res.Content)
	}
}
