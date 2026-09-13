package translate

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// newClassifyFixture publishes three completed sol turns whose assistant
// messages each carry a gateway-issued opaque reasoning envelope and a bound
// tool marker, exactly the client-visible shape a real sol conversation
// produces. It returns the authority and a builder for the cumulative
// history after N turns.
func newClassifyFixture(t *testing.T) (HistoryAuthority, func(turn int) []byte) {
	t.Helper()
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
	t.Cleanup(func() { _ = store.Close() })
	h := NewGPTSubscriptionReceiptHistory(store, id, id)
	envelope, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_1","summary":[],"encrypted_content":"cipher"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	marker, err := opaque.BindToolID("call_1", envelope)
	if err != nil {
		t.Fatal(err)
	}
	assistant := []any{
		map[string]any{"type": "redacted_thinking", "data": envelope.Data()},
		map[string]any{"type": "tool_use", "id": marker, "name": "read_file", "input": map[string]any{"path": "a.go"}},
		map[string]any{"type": "text", "text": "done"},
	}
	raw := func(turn int) []byte {
		messages := []any{}
		for range turn {
			messages = append(messages,
				map[string]any{"role": "user", "content": "prompt"},
				map[string]any{"role": "assistant", "content": assistant})
		}
		return j(messages)
	}
	for turn := 1; turn <= 3; turn++ {
		if err = h.Publish(context.Background(), "gpt-5.6-sol", "owner", raw(turn)); err != nil {
			t.Fatal(err)
		}
	}
	return h, raw
}

// stripThinking drops every redacted_thinking block from the replayed
// history while keeping the bound tool markers, the transformation a client
// applies when it re-encodes prior turns under a newly selected model.
func stripThinking(t *testing.T, raw []byte) []byte {
	t.Helper()
	var messages []map[string]any
	if json.Unmarshal(raw, &messages) != nil {
		t.Fatal("fixture history must be a message array")
	}
	for _, m := range messages {
		if m["role"] != "assistant" {
			continue
		}
		blocks, ok := m["content"].([]any)
		if !ok {
			continue
		}
		kept := []any{}
		for _, b := range blocks {
			block, ok := b.(map[string]any)
			if ok && block["type"] == "redacted_thinking" {
				continue
			}
			kept = append(kept, b)
		}
		m["content"] = kept
	}
	return j(messages)
}

// A mid-conversation model switch inside one verified family must accept the
// unmodified history: the compatible-domain check exists for exactly this.
func TestModelSwitchWithinFamilyAcceptsUnmodifiedHistory(t *testing.T) {
	h, raw := newClassifyFixture(t)
	ctx := context.Background()
	for _, model := range []string{"gpt-5.6-luna", "gpt-5.6-sol"} {
		if err := h.Check(ctx, model, "owner", raw(3)); err != nil {
			t.Fatalf("unmodified %s replay must pass: %v", model, err)
		}
	}
}

// When the client strips the gateway-issued reasoning on the switch turn the
// replay is untranslatable and must still reject — but with the classified
// recovery guidance, never the bare receipt.ErrInvalid body production
// showed (card t703).
func TestModelSwitchStrippedEnvelopeRejectionIdentifiesReason(t *testing.T) {
	h, raw := newClassifyFixture(t)
	ctx := context.Background()
	stripped := stripThinking(t, raw(3))
	for _, model := range []string{"gpt-5.6-luna", "gpt-5.6-sol"} {
		err := h.Check(ctx, model, "owner", stripped)
		var replay HistoryReplayError
		if !errors.As(err, &replay) {
			t.Fatalf("%s stripped-envelope rejection must carry recovery guidance, got %v", model, err)
		}
		if replay.Cause != CauseReasoning {
			t.Fatalf("%s cause = %d, want CauseReasoning", model, replay.Cause)
		}
		if !strings.HasPrefix(err.Error(), historyReplayGuidance) || err.Error() == historyReplayGuidance {
			t.Fatalf("%s message must extend the fixed guidance, got %q", model, err.Error())
		}
		if errors.Is(err, receipt.ErrInvalid) {
			t.Fatalf("%s bare receipt.ErrInvalid must not leak to clients", model)
		}
	}
}
