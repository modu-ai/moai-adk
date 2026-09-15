package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

const hookVerifier = "You are verifying a stop condition in Claude Code. Your task is to verify that the agent completed the given plan. The conversation transcript is available at: /synthetic/transcript.jsonl\nUse as few steps as possible - be efficient and direct.\n\nWhen done, return your result using the StructuredOutput tool with:\n- ok: true if the condition is met\n- ok: false with reason if the condition is not met"
const hookResultSchema = `{"type":"object","properties":{"ok":{"type":"boolean","description":"Whether the condition was met"},"reason":{"type":"string","description":"Reason, if the condition was not met"}},"required":["ok"],"additionalProperties":false}`

func hookAgentBody(t *testing.T) map[string]any {
	t.Helper()
	var schema any
	if err := json.Unmarshal([]byte(hookResultSchema), &schema); err != nil {
		t.Fatal(err)
	}
	return map[string]any{"max_tokens": 32000, "system": []any{map[string]any{"type": "text", "text": hookVerifier}}, "messages": []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": "synthetic condition\n\nARGUMENTS: {}", "cache_control": map[string]string{"type": "ephemeral"}}}}}, "tools": []any{map[string]any{"name": "Read", "input_schema": map[string]any{"type": "object", "properties": map[string]any{"file_path": map[string]string{"type": "string"}}}}, map[string]any{"name": "StructuredOutput", "input_schema": schema}}}
}
func hookJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestHookAgentShapeStableOwnerAndNearMatches(t *testing.T) {
	body := hookAgentBody(t)
	first := managedGPTHookAgentDigest(hookJSON(t, body), http.Header{})
	if first == "" {
		t.Fatal("observed hook not classified")
	}
	msg := body["messages"].([]any)[0].(map[string]any)
	delete(msg["content"].([]any)[0].(map[string]any), "cache_control")
	body["messages"] = append(body["messages"].([]any), map[string]any{"role": "assistant", "content": []any{map[string]string{"type": "tool_use", "name": "Read", "id": "read-1"}}}, map[string]any{"role": "user", "content": []any{map[string]string{"type": "tool_result", "tool_use_id": "read-1", "content": "synthetic"}}})
	if got := managedGPTHookAgentDigest(hookJSON(t, body), http.Header{}); got != first {
		t.Fatal("tool continuation or cache metadata changed owner")
	}
	for _, mutate := range []func(map[string]any){
		func(b map[string]any) { b["system"] = "ordinary main" },
		func(b map[string]any) { b["system"] = hookVerifier + " extra" },
		func(b map[string]any) { b["tools"] = b["tools"].([]any)[:1] },
		func(b map[string]any) {
			b["tools"].([]any)[1].(map[string]any)["input_schema"].(map[string]any)["required"] = []any{"ok", "reason"}
		},
		func(b map[string]any) {
			b["tools"].([]any)[1].(map[string]any)["input_schema"].(map[string]any)["additionalProperties"] = true
		},
	} {
		b := hookAgentBody(t)
		mutate(b)
		if managedGPTHookAgentDigest(hookJSON(t, b), http.Header{}) != "" {
			t.Error("near match classified as hook")
		}
	}
	if managedGPTHookAgentDigest(hookJSON(t, hookAgentBody(t)), http.Header{"X-Claude-Code-Agent-Id": {"ordinary-child"}}) != "" {
		t.Error("normal child header classified as hook")
	}
	stringBody := hookAgentBody(t)
	stringBody["system"] = hookVerifier
	if managedGPTHookAgentDigest(hookJSON(t, stringBody), http.Header{}) != first {
		t.Error("equivalent string system changed digest")
	}
	extraSystem := hookAgentBody(t)
	extraSystem["system"] = append(extraSystem["system"].([]any), map[string]string{"type": "text", "text": "additional synthetic policy"})
	if managedGPTHookAgentDigest(hookJSON(t, extraSystem), http.Header{}) == first {
		t.Error("changed system text collided with prior owner")
	}
	formatBody := hookAgentBody(t)
	formatBody["output_config"] = map[string]any{"format": map[string]string{"type": "json_schema"}}
	if managedGPTHookAgentDigest(hookJSON(t, formatBody), http.Header{}) != "" {
		t.Error("contradictory hook output format accepted")
	}
}

func TestHookAgentProtocolFailureDoesNotRepinMain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	store := managedGPTTestStore(t)
	rpc := &structuredIsolationRPC{acceptedDeltaRPC: &acceptedDeltaRPC{events: make(chan codexapp.Message, 32)}}
	rpc.turns.Store(1)
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("hook-family", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)
	request := func(model string, body map[string]any) codexbridge.Request {
		t.Helper()
		entry := gateway.ModelEntry{RouteID: model, UpstreamID: model, Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
		grant, err := authority.Authorize(ctx, entry)
		if err != nil {
			t.Fatal(err)
		}
		q, err := prepare(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: hookJSON(t, body), Headers: http.Header{"X-Claude-Code-Session-Id": {"hook-family"}}})
		if err != nil {
			t.Fatal(err)
		}
		return q
	}
	mainBody := map[string]any{"max_tokens": 100, "messages": []any{map[string]string{"role": "user", "content": "synthetic main"}}}
	main := request("gpt-6-astra", mainBody)
	if _, err := engine.Step(ctx, main); err != nil {
		t.Fatal(err)
	}
	before, _, err := store.Barrier(main.Owner)
	if err != nil {
		t.Fatal(err)
	}
	near := hookAgentBody(t)
	near["system"] = hookVerifier + " changed"
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-luna", UpstreamID: "gpt-5.6-luna", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := prepare(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: hookJSON(t, near)}); !errors.Is(err, codexbridge.ErrProtocol) {
		t.Fatalf("hook-like near match reached main: %v", err)
	}
	unchanged, _, err := store.Barrier(main.Owner)
	if err != nil || !reflect.DeepEqual(before, unchanged) {
		t.Fatal("near match mutated main")
	}
	hook := request("gpt-5.6-luna", hookAgentBody(t))
	if !hook.Ephemeral || !strings.Contains(hook.Owner.ConversationID, ":hook-agent:") || len(hook.Tools) != 2 {
		t.Fatal("hook not isolated with tools retained")
	}
	if _, err := engine.Step(ctx, hook); !errors.Is(err, codexbridge.ErrProtocol) {
		t.Fatalf("failure=%v", err)
	}
	after, _, err := store.Barrier(main.Owner)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("hook failed main barrier")
	}
	followup := request("gpt-6-astra", mainBody)
	if seg, err := engine.Step(ctx, followup); err != nil || !seg.Done {
		t.Fatalf("main followup %v", err)
	}
}

type hookReadRPC struct {
	*acceptedDeltaRPC
	terminal        bool
	terminalAtStart bool
	interrupted     bool
	interruptError  bool
}

func (r *hookReadRPC) Call(_ context.Context, method string, _ any, out any) error {
	var value any
	switch method {
	case "turn/interrupt":
		if r.interruptError {
			return errors.New("synthetic turn already completed")
		}
		if r.interrupted {
			r.emit("turn/completed", nil, map[string]any{"threadId": "hook-read", "turn": map[string]string{"id": "read-turn", "status": "interrupted"}})
		}
		return nil
	case "thread/start":
		value = map[string]any{"thread": map[string]string{"id": "hook-read"}}
	case "turn/start":
		name := "Read"
		var args any = map[string]string{"file_path": "/synthetic"}
		if r.terminalAtStart {
			name = "StructuredOutput"
			args = map[string]bool{"ok": true}
		}
		r.emit("item/tool/call", json.RawMessage(`"read-rpc"`), map[string]any{"threadId": "hook-read", "turnId": "read-turn", "callId": "read-call", "tool": name, "arguments": args})
		value = map[string]any{"turn": map[string]string{"id": "read-turn"}}
	default:
		return errors.New("unexpected hook RPC")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}
func (r *hookReadRPC) Respond(context.Context, json.RawMessage, any) error {
	r.responses.Add(1)
	if r.interrupted {
		return nil
	}
	if r.terminal && r.responses.Load() == 1 {
		r.emit("item/tool/call", json.RawMessage(`"terminal-rpc"`), map[string]any{"threadId": "hook-read", "turnId": "read-turn", "callId": "terminal-call", "tool": "StructuredOutput", "arguments": map[string]any{"ok": true}})
		return nil
	}
	if r.terminal {
		r.emit("item/agentMessage/delta", nil, map[string]string{"threadId": "hook-read", "turnId": "read-turn", "delta": "discarded post-decision native output"})
	}
	r.emit("turn/completed", nil, map[string]any{"threadId": "hook-read", "turn": map[string]string{"id": "read-turn", "status": "completed"}})
	return nil
}

func TestHookTerminalMarkerDoesNotAffectOrdinaryTools(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		marker, ephemeral, hookOwner bool
		reject                       bool
	}{
		{"ordinary", false, false, false, false}, {"ordinaryEphemeral", false, true, false, false}, {"markerRequiresEphemeral", true, false, true, true}, {"markerRequiresHookOwner", true, true, false, true}, {"interruptedTerminal", true, true, true, false}, {"completedWinsInterruptError", true, true, true, false}, {"missingAckInterruptError", true, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			store := managedGPTTestStore(t)
			rpc := &hookReadRPC{acceptedDeltaRPC: &acceptedDeltaRPC{events: make(chan codexapp.Message, 16)}, terminalAtStart: true, interrupted: true}
			if tc.name == "completedWinsInterruptError" {
				rpc.interrupted = false
				rpc.interruptError = true
			}
			if tc.name == "missingAckInterruptError" {
				rpc.interruptError = true
			}
			e, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
			if err != nil {
				t.Fatal(err)
			}
			entry := gateway.ModelEntry{RouteID: "gpt-5.6-luna", UpstreamID: "gpt-5.6-luna", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
			grant, err := authority.Authorize(ctx, entry)
			if err != nil {
				t.Fatal(err)
			}
			q, err := newManagedGPTPrepare("terminal", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: hookJSON(t, hookAgentBody(t))})
			if err != nil {
				t.Fatal(err)
			}
			q.HookAgentTerminalTool, q.Ephemeral = tc.marker, tc.ephemeral
			if !tc.hookOwner {
				q.Owner.ConversationID = "ordinary"
			}
			seg, err := e.Step(ctx, q)
			if tc.name == "missingAckInterruptError" {
				if !errors.Is(err, codexbridge.ErrProtocol) || seg.Done {
					t.Fatalf("missing terminal ack synthesized success: %v", err)
				}
				return
			}
			if tc.reject {
				if !errors.Is(err, codexbridge.ErrScope) {
					t.Fatal(err)
				}
				return
			}
			if err != nil || seg.Tool == nil {
				t.Fatalf("tool result=%+v error=%v", seg, err)
			}
			if tc.marker {
				if !seg.Done || rpc.responses.Load() != 1 {
					t.Fatal("interrupted terminal not acknowledged")
				}
				if _, found, err := store.Barrier(q.Owner); err != nil || found {
					t.Fatal("terminal barrier retained")
				}
			} else if seg.Done || rpc.responses.Load() != 0 {
				t.Fatal("ordinary StructuredOutput auto-acknowledged")
			}
		})
	}
}

func TestHookAgentReadContinuationCompletesAndReleases(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	store := managedGPTTestStore(t)
	rpc := &hookReadRPC{acceptedDeltaRPC: &acceptedDeltaRPC{events: make(chan codexapp.Message, 16)}, terminal: true}
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store, MaxConversations: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-luna", UpstreamID: "gpt-5.6-luna", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("hook-read-family", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)
	body := hookAgentBody(t)
	request := func() codexbridge.Request {
		t.Helper()
		q, err := prepare(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: hookJSON(t, body)})
		if err != nil {
			t.Fatal(err)
		}
		return q
	}
	first := request()
	if !first.HookAgentTerminalTool {
		t.Fatal("hook terminal contract missing")
	}
	seg, err := engine.Step(ctx, first)
	if err != nil || seg.Tool == nil {
		t.Fatalf("Read not surfaced: %v", err)
	}
	firstUser := body["messages"].([]any)[0].(map[string]any)
	delete(firstUser["content"].([]any)[0].(map[string]any), "cache_control")
	body["messages"] = append(body["messages"].([]any), map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": seg.Tool.ID, "name": "Read", "input": map[string]string{"file_path": "/synthetic"}}}}, map[string]any{"role": "user", "content": []any{map[string]string{"type": "tool_result", "tool_use_id": seg.Tool.ID, "content": "synthetic read"}}})
	next := request()
	if next.Owner != first.Owner || len(next.Results) != 1 || len(next.Tools) != 2 {
		t.Fatal("Read result owner/tools lost")
	}
	if seg, err := engine.Step(ctx, next); err != nil || !seg.Done || seg.Tool == nil || seg.Tool.Name != "StructuredOutput" || seg.Text != "" || rpc.responses.Load() != 2 {
		t.Fatalf("Read continuation failed: %v", err)
	}
	if _, found, err := store.Barrier(first.Owner); err != nil || found {
		t.Fatalf("completed hook durable barrier retained: found=%v err=%v", found, err)
	}
	body = hookAgentBody(t)
	if seg, err := engine.Step(ctx, request()); err != nil || seg.Tool == nil {
		t.Fatalf("same completed hook owner cannot be reused: %v", err)
	}
}

func TestTitleRequiresExactSchemaNotJustRequiredTitle(t *testing.T) {
	schema := map[string]any{"type": "object", "properties": map[string]any{"title": map[string]string{"type": "string"}}, "required": []string{"title"}, "additionalProperties": false}
	body := map[string]any{"tools": []any{}, "output_config": map[string]any{"format": map[string]any{"type": "json_schema", "schema": schema}}}
	if !managedGPTTitleRequest(hookJSON(t, body)) {
		t.Fatal("observed title not recognized")
	}
	schema["properties"].(map[string]any)["ok"] = map[string]string{"type": "boolean"}
	if managedGPTTitleRequest(hookJSON(t, body)) {
		t.Fatal("structured hook aliases title owner")
	}
}

// This opt-in acceptance probe sends only a synthetic satisfied condition and
// the observed StructuredOutput schema to the already-running subscription
// owner. It verifies the real App Server acknowledgement/interrupt ordering;
// no project content or executable tool is exposed to the model.
func TestSharedGPTLiveHookAgentTerminal(t *testing.T) {
	if os.Getenv("MOAI_GPT_HOOK_LIVE") != "1" {
		t.Skip("authenticated hook terminal probe requires MOAI_GPT_HOOK_LIVE=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	client, err := codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	if _, err = client.Initialize(ctx, "moai-hook-terminal-live", "1"); err != nil {
		t.Fatal(err)
	}
	store := managedGPTTestStore(t)
	engine, err := codexbridge.New(ctx, client, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	authority, err := gateway.NewAppServerAuthority(client, "chatgpt", func(context.Context) (string, error) { return "hook-terminal-live", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-luna", UpstreamID: "gpt-5.6-luna", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	family := fmt.Sprintf("hook-terminal-live-%d", time.Now().UnixNano())
	body := hookAgentBody(t)
	body["tools"] = body["tools"].([]any)[1:]
	first := body["messages"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)
	first["text"] = "The synthetic condition is satisfied. Return ok=true now using StructuredOutput."
	headers := http.Header{"X-Claude-Code-Session-Id": {family}}
	prepare := newManagedGPTPrepare(family, t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxOutputBytes: 1 << 20, MaxEventBytes: 1 << 20}, store)
	q, err := prepare(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: hookJSON(t, body), Headers: headers})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	seg, err := engine.Step(ctx, q)
	if err != nil || !seg.Done || seg.Tool == nil || seg.Tool.Name != "StructuredOutput" || seg.Text != "" {
		t.Fatalf("live terminal result=%+v error=%v", seg, err)
	}
	var result struct {
		OK bool `json:"ok"`
	}
	if json.Unmarshal(seg.Tool.Arguments, &result) != nil || !result.OK {
		t.Fatalf("unexpected hook decision: %s", seg.Tool.Arguments)
	}
	if _, found, err := store.Barrier(q.Owner); err != nil || found {
		t.Fatalf("live terminal barrier retained: found=%v err=%v", found, err)
	}
	t.Logf("live hook terminal acknowledged and released in %s", time.Since(started).Round(time.Millisecond))
}
