package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

// This uses the already-running subscription owner and synthetic echo only.
// It never launches a second owner or executes a shell on the model's behalf.
func TestSharedGPTLiveSummaryPreservesPendingChild(t *testing.T) {
	if os.Getenv("MOAI_GPT_LIVE") != "1" {
		t.Skip("authenticated summary isolation requires MOAI_GPT_LIVE=1")
	}
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, _ = filepath.Abs(binary)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	client, err := codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		t.Fatalf("start product shared owner first: %v", err)
	}
	defer func() { _ = client.Close() }() // Best-effort teardown; primary assertions own failures.
	if _, err = client.Initialize(ctx, "moai-summary-live", "1"); err != nil {
		t.Fatal(err)
	}
	authority, err := gateway.NewAppServerAuthority(client, "chatgpt", func(context.Context) (string, error) { return "summary-live", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	store := managedGPTTestStore(t)
	engine, err := codexbridge.New(ctx, client, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	const family = "a7d15361-2e26-47e7-ae2e-476e210f77b3"
	prepare := newManagedGPTPrepare(family, t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxOutputBytes: 1 << 20, MaxEventBytes: 1 << 20}, store)
	tool := map[string]any{"name": "echo", "description": "Return the supplied text unchanged", "input_schema": map[string]any{"type": "object", "properties": map[string]any{"text": map[string]string{"type": "string"}}, "required": []string{"text"}, "additionalProperties": false}}
	request := func(messages []any) codexbridge.Request {
		t.Helper()
		body, err := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 256, "thinking": map[string]any{"type": "adaptive"}, "output_config": map[string]any{"effort": "low"}, "tools": []any{tool}, "messages": messages})
		if err != nil {
			t.Fatal(err)
		}
		h := http.Header{}
		h.Set("X-Claude-Code-Session-Id", family)
		h.Set("X-Claude-Code-Agent-Id", "a123456789abcdef0")
		q, err := prepare(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: body, Headers: h})
		if err != nil {
			t.Fatal(err)
		}
		return q
	}
	initial := map[string]any{"role": "user", "content": "Call echo exactly once with text SUMMARY_CHILD_OK, then report its result verbatim."}
	child := request([]any{initial})
	segment, err := engine.Step(ctx, child)
	if err != nil || segment.Tool == nil || segment.Tool.Name != "echo" {
		t.Fatalf("child did not reach echo boundary: %+v %v", segment, err)
	}
	before, found, err := store.Barrier(child.Owner)
	if err != nil || !found || before.Phase != "waiting" {
		t.Fatalf("pending child barrier: %+v %v", before, err)
	}
	var arguments map[string]any
	if err := json.Unmarshal(segment.Tool.Arguments, &arguments); err != nil {
		t.Fatal(err)
	}
	call := map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": segment.Tool.ID, "name": "echo", "input": arguments}}}
	result := map[string]any{"type": "tool_result", "tool_use_id": segment.Tool.ID, "content": "SUMMARY_CHILD_OK"}
	// Include public tool history, but leave the actual child's RPC unanswered.
	// The snapshot must not replay that result into its independent summary.
	previousPrompt := strings.Replace(observedAgentSummaryPrompt, "\n\nGood:", "\n\nPrevious: \"Calling echo function\" — say something NEW.\n\nGood:", 1)
	for _, prompt := range []string{observedAgentSummaryPrompt, previousPrompt} {
		summary := request([]any{initial, call, map[string]any{"role": "user", "content": []any{result, map[string]any{"type": "text", "text": prompt}}}})
		if summary.Owner.ConversationID == child.Owner.ConversationID || !summary.Ephemeral || len(summary.Tools) != 0 || len(summary.Results) != 0 || len(summary.References) != 0 {
			t.Fatalf("summary not isolated: owner_same=%v ephemeral=%v tools=%d results=%d references=%d", summary.Owner.ConversationID == child.Owner.ConversationID, summary.Ephemeral, len(summary.Tools), len(summary.Results), len(summary.References))
		}
		started := time.Now()
		out, err := engine.Step(ctx, summary)
		if err != nil || !out.Done || out.Tool != nil || strings.TrimSpace(out.Text) == "" {
			t.Fatalf("summary completion: %+v %v", out, err)
		}
		after, found, err := store.Barrier(child.Owner)
		if err != nil || !found || !reflect.DeepEqual(before, after) {
			t.Fatalf("summary changed pending child barrier: before=%+v after=%+v err=%v", before, after, err)
		}
		t.Logf("summary completed while child pending: elapsed=%s tools=0 result=%q", time.Since(started).Round(time.Millisecond), out.Text)
	}
	continued := request([]any{initial, call, map[string]any{"role": "user", "content": []any{result}}})
	out, err := engine.Step(ctx, continued)
	if err != nil || !out.Done || !strings.Contains(out.Text, "SUMMARY_CHILD_OK") {
		t.Fatalf("child continuation: %+v %v", out, err)
	}
	t.Logf("child resumed successfully after two independent summaries: result=%q", out.Text)
}
