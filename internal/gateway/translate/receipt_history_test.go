package translate

import (
	"context"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"os"
	"path/filepath"
	"testing"
)

func TestReceiptHistoryRejectsStrippingForeignScopeAndFamily(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := receipt.OpenStore(context.Background(), dir, id, true)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	history := NewReceiptHistory(store, id, id)
	envelope, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_1","summary":[],"encrypted_content":"cipher"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	blocks := []any{map[string]any{"type": "redacted_thinking", "data": envelope.Data()}, map[string]any{"type": "text", "text": "answer"}}
	messages := []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": blocks}}
	raw := j(messages)
	if err = history.Publish(context.Background(), "gpt-6-astra", "account-a", raw); err != nil {
		t.Fatal(err)
	}
	if err = history.Check(context.Background(), "gpt-6-astra", "account-a", raw); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"gpt-5.6-sol", "account-a"}, {"gpt-6-astra", "account-b"}} {
		if history.Check(context.Background(), pair[0], pair[1], raw) == nil {
			t.Fatal("foreign scope authorized")
		}
	}
	messages[1].(map[string]any)["content"] = blocks[1:]
	if history.Check(context.Background(), "gpt-6-astra", "account-a", j(messages)) == nil {
		t.Fatal("stripped reasoning authorized")
	}
}

func TestReceiptHistoryAuthorizedForkRetainsScope(t *testing.T) {
	const parentID = "11111111-1111-4111-8111-111111111111"
	const childID = "22222222-2222-4222-8222-222222222222"
	open := func(id string) *receipt.Store {
		p, e := filepath.EvalSymlinks(t.TempDir())
		if e != nil {
			t.Fatal(e)
		}
		if e = os.Chmod(p, 0700); e != nil {
			t.Fatal(e)
		}
		s, e := receipt.OpenStore(context.Background(), p, id, true)
		if e != nil {
			t.Fatal(e)
		}
		t.Cleanup(func() { s.Close() })
		return s
	}
	parent := open(parentID)
	child := open(childID)
	history := NewReceiptHistory(parent, parentID, parentID)
	raw := j([]any{map[string]any{"role": "assistant", "content": "answer"}})
	if err := history.Publish(context.Background(), "gpt-6-astra", "account-a", raw); err != nil {
		t.Fatal(err)
	}
	snapshot, err := parent.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fork, err := snapshot.Fork(childID)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range fork.Candidates() {
		if err = child.Publish(context.Background(), candidate); err != nil {
			t.Fatal(err)
		}
	}
	if err = NewReceiptHistory(child, childID, parentID).Check(context.Background(), "gpt-6-astra", "account-a", raw); err != nil {
		t.Fatal("authorized fork lost history", err)
	}
	if NewReceiptHistory(child, childID, childID).Check(context.Background(), "gpt-6-astra", "account-a", raw) == nil {
		t.Fatal("foreign family accepted history")
	}
}

func TestReceiptHistoryRejectsMalformedAndClosedAuthority(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	os.Chmod(dir, 0700)
	store, err := receipt.OpenStore(context.Background(), dir, id, true)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	h := NewReceiptHistory(store, id, id)
	envelope, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_1","summary":[],"encrypted_content":"cipher"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	carrier := map[string]any{"type": "redacted_thinking", "data": envelope.Data()}
	toolID, err := opaque.BindToolID("call_1", envelope)
	if err != nil {
		t.Fatal(err)
	}
	good := j([]any{map[string]any{"role": "assistant", "content": []any{carrier, map[string]any{"type": "tool_use", "id": toolID, "name": "read", "input": map[string]any{}}}}})
	if err = h.Publish(context.Background(), "gpt-5.6-sol", "owner", good); err != nil {
		t.Fatal(err)
	}
	if err = h.Check(context.Background(), "gpt-5.6-terra", "owner", good); err != nil {
		t.Fatal("same family rejected", err)
	}
	cases := [][]byte{[]byte(`{`), []byte(`null`), []byte(`[{}]`), j([]any{map[string]any{"role": "assistant", "content": 1}}), j([]any{map[string]any{"role": "assistant", "content": []any{carrier, carrier}}}), j([]any{map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "redacted_thinking", "data": false}}}}), j([]any{map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "redacted_thinking", "data": "wrong"}}}}), j([]any{map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": false}}}}), j([]any{map[string]any{"role": "assistant", "content": []any{carrier, map[string]any{"type": "tool_use", "id": "wrong"}}}}), j([]any{map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": toolID}}}}), j([]any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": toolID, "content": "result"}}}})}
	for i, raw := range cases {
		if h.Check(context.Background(), "gpt-6-astra", "owner", raw) == nil {
			t.Errorf("malformed history %d accepted", i)
		}
		if h.Publish(context.Background(), "gpt-6-astra", "owner", raw) == nil {
			t.Errorf("malformed publication %d accepted", i)
		}
	}
	for _, pair := range [][2]string{{"foreign", "owner"}, {"gpt-6-astra", ""}} {
		if h.Check(context.Background(), pair[0], pair[1], good) == nil {
			t.Fatal("missing authority accepted")
		}
	}
	if h.Publish(context.Background(), "gpt-6-astra", "owner", []byte(`[]`)) == nil {
		t.Fatal("empty publication accepted")
	}
	chained := []any{map[string]any{"role": "assistant", "content": "unknown"}, map[string]any{"role": "assistant", "content": "new"}}
	if h.Publish(context.Background(), "gpt-6-astra", "owner", j(chained)) == nil {
		t.Fatal("unreceipted prefix published")
	}
	store.Close()
	if h.Check(context.Background(), "gpt-6-astra", "owner", good) == nil || h.Publish(context.Background(), "gpt-6-astra", "owner", good) == nil {
		t.Fatal("closed authority accepted")
	}
}

func TestReceiptHistoryRejectsPublicLayoutMutation(t *testing.T) {
	const id = "33333333-3333-4333-8333-333333333333"
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := receipt.OpenStore(context.Background(), dir, id, true)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	h := NewReceiptHistory(store, id, id)
	items := []opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_layout","summary":[],"encrypted_content":"synthetic"}`)}}
	layout := []opaque.PublicItem{{OutputIndex: 1, Type: "message", Blocks: 1, Phase: "final_answer"}}
	env, err := opaque.EncodeWithPublic(items, layout)
	if err != nil {
		t.Fatal(err)
	}
	carrier := map[string]any{"type": "redacted_thinking", "data": env.Data()}
	messages := []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": []any{carrier, map[string]any{"type": "text", "text": "answer"}}}}
	if err = h.Publish(context.Background(), "gpt-6-astra", "owner", j(messages)); err != nil {
		t.Fatal(err)
	}
	if err = h.Check(context.Background(), "gpt-6-astra", "owner", j(messages)); err != nil {
		t.Fatal(err)
	}
	layout[0].Phase = "commentary"
	forged, err := opaque.EncodeWithPublic(items, layout)
	if err != nil {
		t.Fatal(err)
	}
	carrier["data"] = forged.Data()
	if h.Check(context.Background(), "gpt-6-astra", "owner", j(messages)) == nil {
		t.Fatal("changed public metadata authorized by matching public text")
	}
}
