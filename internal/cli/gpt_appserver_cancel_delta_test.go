package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

type acceptedDeltaRPC struct {
	events    chan codexapp.Message
	turns     atomic.Int32
	responses atomic.Int32
}

func (r *acceptedDeltaRPC) emit(method string, id json.RawMessage, value any) {
	raw, _ := json.Marshal(value)
	r.events <- codexapp.Message{Method: method, ID: id, Params: raw}
}
func (r *acceptedDeltaRPC) Call(_ context.Context, method string, params any, out any) error {
	var value any = map[string]any{}
	switch method {
	case "thread/start", "thread/resume":
		value = map[string]any{"thread": map[string]string{"id": "cancel-thread"}}
	case "turn/start":
		seq := r.turns.Add(1)
		turn := fmt.Sprintf("turn-%d", seq)
		value = map[string]any{"turn": map[string]string{"id": turn}}
		if seq == 1 {
			r.emit("item/tool/call", json.RawMessage(`"rpc-tool"`), map[string]any{"threadId": "cancel-thread", "turnId": turn, "callId": "call-one", "tool": "Skill", "arguments": map[string]any{}})
		} else {
			r.emit("turn/completed", nil, map[string]any{"threadId": "cancel-thread", "turn": map[string]string{"id": turn, "status": "completed"}})
		}
	case "turn/interrupt":
		r.emit("turn/completed", nil, map[string]any{"threadId": "cancel-thread", "turn": map[string]string{"id": "turn-1", "status": "interrupted"}})
	default:
		return fmt.Errorf("unexpected call %s", method)
	}
	raw, _ := json.Marshal(value)
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}
func (r *acceptedDeltaRPC) Respond(_ context.Context, _ json.RawMessage, _ any) error {
	r.responses.Add(1)
	r.emit("item/agentMessage/delta", nil, map[string]string{"threadId": "cancel-thread", "turnId": "turn-1", "itemId": "text", "delta": "started"})
	return nil
}
func (r *acceptedDeltaRPC) DiscardRequest(json.RawMessage) error { return nil }
func (r *acceptedDeltaRPC) Events() <-chan codexapp.Message      { return r.events }
func (r *acceptedDeltaRPC) Err() error                           { return nil }

func TestManagedGPTCancelAfterToolReplyAcceptsFreshInputExactlyOnce(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	store := managedGPTTestStore(t)
	rpc := &acceptedDeltaRPC{events: make(chan codexapp.Message, 8)}
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	authority, _ := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, _ := authority.Authorize(ctx, entry)
	prepare := newManagedGPTPrepare("cancel-family", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)
	root := map[string]any{"max_tokens": 1024, "tools": []any{map[string]any{"name": "Skill", "input_schema": map[string]any{"type": "object"}}}}
	messages := []any{map[string]any{"role": "user", "content": "work"}}
	request := func() codexbridge.Request {
		t.Helper()
		root["messages"] = messages
		body, _ := json.Marshal(root)
		q, err := prepare(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: body, Headers: http.Header{"X-Claude-Code-Session-Id": {"cancel-family"}}})
		if err != nil {
			t.Fatal(err)
		}
		return q
	}
	first, err := engine.Step(ctx, request())
	if err != nil || first.Tool == nil {
		t.Fatalf("first tool: %+v %v", first, err)
	}
	messages = append(messages, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": first.Tool.ID, "name": "Skill", "input": map[string]any{}}}}, map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": first.Tool.ID, "content": "accepted tool output"}, map[string]string{"type": "text", "text": "skill context"}}})
	accepted := request()
	turnCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err = engine.StepStream(turnCtx, accepted, func(string) error { cancel(); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		barrier, found, err := store.Barrier(accepted.Owner)
		if err != nil {
			t.Fatal(err)
		}
		if found && barrier.Phase == "idle" {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("cancel acknowledgement not idle")
		case <-tick.C:
		}
	}
	messages = append(messages, map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": "new question"}, map[string]any{"type": "image", "source": map[string]string{"type": "base64", "media_type": "image/png", "data": "aW1hZ2U="}}}})
	next := request()
	// The old last-assistant projection resubmitted this consumed result.
	// Show the real engine rejects that replay without mutating the idle owner.
	legacy := next
	legacy.Input, legacy.Results = nil, accepted.Results
	if _, err := engine.Step(ctx, legacy); !errors.Is(err, codexbridge.ErrScope) {
		t.Fatalf("unproven replay accepted: %v", err)
	}
	if len(next.Input) != 2 || len(next.Results) != 0 {
		t.Fatalf("bad fresh delta: %+v", next)
	}
	seg, err := engine.Step(ctx, next)
	if err != nil || !seg.Done || rpc.responses.Load() != 1 || rpc.turns.Load() != 2 {
		t.Fatalf("followup: done=%v err=%v responses=%d turns=%d", seg.Done, err, rpc.responses.Load(), rpc.turns.Load())
	}
}
