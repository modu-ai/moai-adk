package gateway

import (
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"testing"
)

func TestAppServerPreparedModelOnlyDefersPureContinuation(t *testing.T) {
	valid := codexbridge.Request{Model: "gpt-5.6-terra", ExpectedPrefix: "prefix", Results: []codexbridge.ToolResult{{ID: "pending"}}}
	if !appServerPreparedModelAllowed("gpt-5.6-sol", valid) {
		t.Fatal("trusted pure continuation cannot finish on current model")
	}
	for _, kind := range []string{"unknown-model", "foreign-model", "no-results", "new-input", "no-prefix", "resume", "ephemeral", "fork"} {
		q := valid
		switch kind {
		case "unknown-model":
			q.Model = "gpt-6"
		case "foreign-model":
			q.Model = "claude-opus-5"
		case "no-results":
			q.Results = nil
		case "new-input":
			q.Input = []any{"new input"}
		case "no-prefix":
			q.ExpectedPrefix = ""
		case "resume":
			q.Resume = true
		case "ephemeral":
			q.Ephemeral = true
		case "fork":
			q.Fork = true
		}
		if appServerPreparedModelAllowed("gpt-5.6-sol", q) {
			t.Errorf("untrusted model override accepted: %s", kind)
		}
	}
	if !appServerPreparedModelAllowed(valid.Model, codexbridge.Request{Model: valid.Model, Input: []any{"ordinary"}}) {
		t.Fatal("same selected model rejected")
	}
}
