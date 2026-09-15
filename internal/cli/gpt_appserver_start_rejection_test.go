package cli

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

type rejectedStartFixture struct {
	events        chan codexapp.Message
	starts, turns int
}

func (r *rejectedStartFixture) DiscardRequest(json.RawMessage) error { return nil }

func (r *rejectedStartFixture) Events() <-chan codexapp.Message { return r.events }
func (r *rejectedStartFixture) Err() error                      { return nil }
func (r *rejectedStartFixture) Respond(context.Context, json.RawMessage, any) error {
	return errors.New("unexpected tool response")
}
func (r *rejectedStartFixture) Call(_ context.Context, method string, _ any, out any) error {
	switch method {
	case "thread/start":
		r.starts++
		if r.starts == 1 {
			return &codexapp.RPCError{Code: -32603}
		}
		return json.Unmarshal([]byte(`{"thread":{"id":"accepted-thread"}}`), out)
	case "turn/start":
		r.turns++
		r.events <- codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"accepted-thread","turn":{"id":"accepted-turn","status":"completed"}}`)}
		return json.Unmarshal([]byte(`{"turn":{"id":"accepted-turn"}}`), out)
	default:
		return errors.New("unexpected RPC")
	}
}

func TestManagedPrepareAndEngineRetryRejectedThreadStart(t *testing.T) {
	for _, compacted := range []bool{false, true} {
		t.Run(map[bool]string{false: "main", true: "compacted"}[compacted], func(t *testing.T) { testManagedRejectedStart(t, compacted) })
	}
}

func testManagedRejectedStart(t *testing.T, compacted bool) {
	ctx := context.Background()
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gatewayGPTModels()[1]
	grant, err := authority.Authorize(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	store := managedGPTTestStore(t)
	rpc := &rejectedStartFixture{events: make(chan codexapp.Message, 4)}
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	prepare := newManagedGPTPrepare("prepare-rejected", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 1 << 20}, store)
	routed := gateway.RoutedRequest{Entry: entry, Managed: grant, Body: []byte(`{"model":"gpt-5.6-sol","max_tokens":64,"messages":[{"role":"user","content":"Synthetic retry"}],"tools":[]}`)}
	if compacted {
		summary := "This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.\n\nSynthetic summary. If you need specific details from before compaction (like exact code snippets, error messages, or content you generated), read the full transcript at: synthetic\nPick up the last task as if the break never happened.\n"
		routed.Body, err = json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 64, "messages": []any{map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": summary}}}}, "tools": []any{}})
		if err != nil {
			t.Fatal(err)
		}
	}
	first, err := prepare(ctx, routed)
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Step(ctx, first)
	var rpcErr *codexapp.RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected rejected start: %v", err)
	}
	retry, err := prepare(ctx, routed)
	if err != nil {
		t.Fatal(err)
	}
	if retry.ExpectedPrefix != first.ExpectedPrefix {
		t.Fatal("rejected input became accepted prefix")
	}
	if !reflect.DeepEqual(first.Input, retry.Input) {
		t.Fatal("retry lost inherited context")
	}
	seg, err := engine.Step(ctx, retry)
	if err != nil || !seg.Done {
		t.Fatalf("retry failed: %v", err)
	}
	if rpc.starts != 2 || rpc.turns != 1 {
		t.Fatalf("starts=%d turns=%d", rpc.starts, rpc.turns)
	}
}
