package translate

import (
	"context"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"os"
	"path/filepath"
	"testing"
)

func TestSubscriptionReceiptModelSwitchPreservesSourceBindings(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
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
	ctx := context.Background()
	strict := NewReceiptHistory(store, id, id)
	switched := NewGPTSubscriptionReceiptHistory(store, id, id)
	envelope, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_1","summary":[],"encrypted_content":"cipher"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	blocks := []any{map[string]any{"type": "redacted_thinking", "data": envelope.Data()}, map[string]any{"type": "text", "text": "answer"}}
	messages := []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": blocks}}
	raw := j(messages)
	if err = strict.Publish(ctx, "gpt-5.6-sol", "owner", raw); err != nil {
		t.Fatal(err)
	}
	if err = switched.Check(ctx, "gpt-6-astra", "owner", raw); err != nil {
		t.Fatal("legacy Sol receipt rejected", err)
	}
	messages = append(messages, map[string]any{"role": "user", "content": "continue"}, map[string]any{"role": "assistant", "content": blocks})
	mixed := j(messages)
	if err = switched.Publish(ctx, "gpt-6-astra", "owner", mixed); err != nil {
		t.Fatal("mixed publish", err)
	}
	if err = switched.Check(ctx, "gpt-5.6-sol", "owner", mixed); err != nil {
		t.Fatal("return to Sol", err)
	}
	for _, model := range []string{"gpt-5.6-terra", "gpt-5.6-luna"} {
		if err = switched.Check(ctx, model, "owner", mixed); err != nil {
			t.Fatal("compatible target rejected", model, err)
		}
	}
	if strict.Check(ctx, "gpt-6-astra", "owner", mixed) == nil {
		t.Fatal("legacy receipts were rebound")
	}
	for _, model := range []string{"unknown"} {
		if switched.Check(ctx, model, "owner", mixed) == nil {
			t.Fatal("unverified target accepted", model)
		}
	}
	if switched.Check(ctx, "gpt-6-astra", "foreign", mixed) == nil {
		t.Fatal("foreign owner accepted")
	}
	if NewGPTSubscriptionReceiptHistory(store, id, "22222222-2222-4222-8222-222222222222").Check(ctx, "gpt-6-astra", "owner", mixed) == nil {
		t.Fatal("foreign family accepted")
	}
	if NewGPTSubscriptionReceiptHistory(store, "22222222-2222-4222-8222-222222222222", id).Check(ctx, "gpt-6-astra", "owner", mixed) == nil {
		t.Fatal("foreign session accepted")
	}

	changedCipher, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_1","summary":[],"encrypted_content":"tampered"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	messages[1].(map[string]any)["content"] = []any{map[string]any{"type": "redacted_thinking", "data": changedCipher.Data()}, blocks[1]}
	if switched.Check(ctx, "gpt-6-astra", "owner", j(messages)) == nil {
		t.Fatal("changed cipher accepted")
	}
	messages[1].(map[string]any)["content"] = blocks
	empty := j([]any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": blocks[1:]}})
	if err = strict.Publish(ctx, "gpt-6-astra", "owner", empty); err != nil {
		t.Fatal(err)
	}
	if switched.Check(ctx, "gpt-6-astra", "owner", empty) == nil {
		t.Fatal("cross-domain empty candidate hid required opaque receipt")
	}
	messages[1].(map[string]any)["content"] = blocks[1:]
	if switched.Check(ctx, "gpt-6-astra", "owner", j(messages)) == nil {
		t.Fatal("stripped cipher accepted")
	}
	messages[1].(map[string]any)["content"] = blocks
	messages[0].(map[string]any)["content"] = "changed"
	if switched.Check(ctx, "gpt-6-astra", "owner", j(messages)) == nil {
		t.Fatal("changed public prefix accepted")
	}
}
