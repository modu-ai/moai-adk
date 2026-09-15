package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

type structuredIsolationRPC struct {
	*acceptedDeltaRPC
	threads atomic.Int32
}

func (r *structuredIsolationRPC) Call(_ context.Context, method string, params any, out any) error {
	p := params.(map[string]any)
	var result any = map[string]any{}
	switch method {
	case "thread/start":
		result = map[string]any{"thread": map[string]string{"id": fmt.Sprintf("thread-%d", r.threads.Add(1))}}
	case "turn/start":
		seq := r.turns.Add(1)
		thread := p["threadId"].(string)
		turn := fmt.Sprintf("turn-%d", seq)
		result = map[string]any{"turn": map[string]string{"id": turn}}
		if seq == 1 {
			r.emit("item/agentMessage/delta", nil, map[string]string{"threadId": thread, "turnId": turn, "itemId": "text", "delta": "started"})
		} else if p["model"] == "gpt-5.6-luna" {
			r.emit("error", nil, map[string]any{"threadId": thread, "turnId": turn, "error": map[string]string{"message": "synthetic schema failure"}})
		} else {
			r.emit("turn/completed", nil, map[string]any{"threadId": thread, "turn": map[string]string{"id": turn, "status": "completed"}})
		}
	case "turn/interrupt":
		r.emit("turn/completed", nil, map[string]any{"threadId": p["threadId"], "turn": map[string]any{"id": p["turnId"], "status": "interrupted"}})
	default:
		return fmt.Errorf("unexpected %s", method)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func TestStructuredUtilityFailureAfterCancelPreservesMainModel(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	store := managedGPTTestStore(t)
	rpc := &structuredIsolationRPC{acceptedDeltaRPC: &acceptedDeltaRPC{events: make(chan codexapp.Message, 32)}}
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("utility-family", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)
	request := func(model string, structured bool) (codexbridge.Request, error) {
		entry := gateway.ModelEntry{RouteID: model, UpstreamID: model, Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
		grant, err := authority.Authorize(ctx, entry)
		if err != nil {
			return codexbridge.Request{}, err
		}
		body := map[string]any{"max_tokens": 100, "messages": []any{map[string]string{"role": "user", "content": "synthetic task"}}}
		if structured {
			body["output_config"] = map[string]any{"format": map[string]any{"type": "json_schema", "schema": map[string]any{"type": "object", "properties": map[string]any{"ok": map[string]string{"type": "boolean"}, "reason": map[string]string{"type": "string"}}, "required": []string{"ok", "reason"}, "additionalProperties": false}}}
		}
		raw, err := json.Marshal(body)
		if err != nil {
			return codexbridge.Request{}, err
		}
		return prepare(ctx, gateway.RoutedRequest{Entry: entry, Managed: grant, Body: raw, Headers: http.Header{"X-Claude-Code-Session-Id": {"utility-family"}}})
	}
	main, err := request("gpt-6-astra", false)
	if err != nil {
		t.Fatal(err)
	}
	turnCtx, cancel := context.WithCancel(ctx)
	_, err = engine.StepStream(turnCtx, main, func(string) error { cancel(); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel=%v", err)
	}
	var before codexbridge.Barrier
	for {
		barrier, _, err := store.Barrier(main.Owner)
		if err != nil {
			t.Fatal(err)
		}
		if barrier.Phase == "idle" {
			before = barrier
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("cancel not acknowledged")
		case <-time.After(time.Millisecond):
		}
	}
	utility, err := request("gpt-5.6-luna", true)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := request("gpt-5.6-luna", true)
	if err != nil || retry.Owner != utility.Owner {
		t.Fatal("snapshot retry identity changed", err)
	}
	if !utility.Ephemeral || utility.Owner == main.Owner || len(utility.Tools) != 0 || len(utility.Results) != 0 {
		t.Error("utility is not isolated read-only snapshot")
	}
	if _, err := engine.Step(ctx, utility); !errors.Is(err, codexbridge.ErrProtocol) {
		t.Fatalf("utility failure=%v", err)
	}
	after, _, err := store.Barrier(main.Owner)
	if err != nil || !reflect.DeepEqual(after, before) {
		t.Errorf("main barrier poisoned: before=%+v after=%+v err=%v", before, after, err)
	}
	followup, err := request("gpt-6-astra", false)
	if err != nil {
		t.Fatalf("main prepare after utility failure: %v", err)
	}
	if seg, err := engine.Step(ctx, followup); err != nil || !seg.Done {
		t.Fatalf("main did not complete: %v", err)
	}
}
