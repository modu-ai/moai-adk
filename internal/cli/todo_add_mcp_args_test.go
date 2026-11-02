package cli

import (
	"maps"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card t2011, audit sync-audit-1 defect D4 — an optional MCP argument that is
// PRESENT but not a string (attach: 1, independent_scope: null, ...) was read as
// "not given", so a caller who meant to attach text to a card got a NEW card
// instead. A present key with a wrong type is refused before any write; an
// absent key still means "not given".
func TestTodoAddMCP_PresentButWrongTypedOptionalArgIsRefusedBeforeAnyWrite(t *testing.T) {
	root, store := sdMoaiFixture(t)
	t.Setenv(config.EnvFactoryRole, "operator")
	seedItems(t, store, "first card without any issue reference")
	before := queueDigest(t, store)

	for name, extra := range map[string]map[string]any{
		"attach number":             {"attach": 1},
		"attach bool":               {"attach": true},
		"attach null":               {"attach": nil},
		"attach list":               {"attach": []any{"t1"}},
		"independent_scope null":    {"independent_scope": nil},
		"independent_scope number":  {"independent_scope": 1},
		"independent_scope bool":    {"independent_scope": true},
		"attach and wrong scope":    {"attach": "t1", "independent_scope": 7},
		"wrong attach, valid scope": {"attach": 7, "independent_scope": "separate docs"},
	} {
		t.Run(name, func(t *testing.T) {
			args := map[string]any{"text": "a plain note that cites no issue", "project_root": root}
			maps.Copy(args, extra)
			out, err := sdCallTool(t, handleTodoAdd, args)
			if err == nil {
				t.Fatalf("todo_add admitted %v (out %q), want a refusal", extra, out)
			}
			if got := queueDigest(t, store); got != before {
				t.Fatalf("the queue changed on a refused call: %v", extra)
			}
		})
	}

	// Control: the keys absent is the ordinary add and still creates the card,
	// so the refusals above are about the type, not about the call shape.
	if _, err := sdCallTool(t, handleTodoAdd, map[string]any{"text": "a plain note that cites no issue", "project_root": root}); err != nil {
		t.Fatalf("plain todo_add: %v", err)
	}
	if queueDigest(t, store) == before {
		t.Fatal("the plain todo_add wrote nothing")
	}
}
