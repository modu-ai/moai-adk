package translate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// Card t672: rejection-cause classification and the legitimate replay shapes.
// The truncated-fork cells lock the tolerance that already exists (a tail
// truncation of published, pairing-valid boundaries replays); every rejection
// cell locks the class named in HistoryReplayError.Cause and its fixed
// guidance, with no loosening of the authorization property.

func causeStore(t *testing.T, id string) *receipt.Store {
	t.Helper()
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
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func causeEnvelope(t *testing.T, id, cipher string) *opaque.Envelope {
	t.Helper()
	envelope, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"` + id + `","summary":[],"encrypted_content":"` + cipher + `"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	return envelope
}

// causeConversation builds a three-turn conversation whose second turn carries
// an envelope-bound tool_use/tool_result pair, then publishes it turn by turn
// exactly the way the response path does.
func causeConversation(t *testing.T, h HistoryAuthority, cipher string) (turn1, turn2, turn3 []any) {
	t.Helper()
	ctx := context.Background()
	env1 := causeEnvelope(t, "rs_1", cipher+"-1")
	env2 := causeEnvelope(t, "rs_2", cipher+"-2")
	env3 := causeEnvelope(t, "rs_3", cipher+"-3")
	marker, err := opaque.BindToolID("call_1", env2)
	if err != nil {
		t.Fatal(err)
	}
	carrier := func(e *opaque.Envelope) map[string]any {
		return map[string]any{"type": "redacted_thinking", "data": e.Data()}
	}
	turn1 = []any{
		map[string]any{"role": "user", "content": "hello"},
		map[string]any{"role": "assistant", "content": []any{carrier(env1), map[string]any{"type": "text", "text": "one"}}},
	}
	turn2 = []any{
		map[string]any{"role": "user", "content": "run"},
		map[string]any{"role": "assistant", "content": []any{carrier(env2), map[string]any{"type": "tool_use", "id": marker, "name": "read", "input": map[string]any{}}}},
		map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": marker, "content": "data"}}},
	}
	turn3 = []any{
		map[string]any{"role": "user", "content": "more"},
		map[string]any{"role": "assistant", "content": []any{carrier(env3), map[string]any{"type": "text", "text": "three"}}},
	}
	published := append([]any{}, turn1...)
	if err = h.Publish(ctx, "gpt-5.6-sol", "owner", j(published)); err != nil {
		t.Fatal(err)
	}
	published = append(published, turn2...)
	if err = h.Publish(ctx, "gpt-5.6-sol", "owner", j(published)); err != nil {
		t.Fatal(err)
	}
	published = append(published, turn3...)
	if err = h.Publish(ctx, "gpt-5.6-sol", "owner", j(published)); err != nil {
		t.Fatal(err)
	}
	return turn1, turn2, turn3
}

func replayCauseOf(t *testing.T, err error) HistoryReplayError {
	t.Helper()
	if err == nil {
		t.Fatal("expected rejection")
	}
	var replay HistoryReplayError
	if !errors.As(err, &replay) {
		t.Fatalf("not a HistoryReplayError: %v", err)
	}
	if !strings.Contains(err.Error(), historyReplayGuidance) {
		t.Fatalf("guidance sentence lost: %v", err)
	}
	return replay
}

func TestReceiptHistoryAcceptsTruncatedForkReplay(t *testing.T) {
	const id = "44444444-4444-4444-8444-444444444444"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, turn2, _ := causeConversation(t, h, "cipher")
	ctx := context.Background()
	cases := []struct {
		name  string
		model string
		tail  []any
	}{
		{"prefix at turn boundary", "gpt-5.6-sol", []any{map[string]any{"role": "user", "content": "summarize"}}},
		{"prefix keeps tool pairing", "gpt-5.6-sol", append(append([]any{}, turn2[2:]...), map[string]any{"role": "user", "content": "summarize"})},
		{"compatible domain replay", "gpt-6-astra", []any{map[string]any{"role": "user", "content": "summarize"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			messages := append(append([]any{}, turn1...), turn2...)
			messages = append(messages, c.tail...)
			if err := h.Check(ctx, c.model, "owner", j(messages)); err != nil {
				t.Fatalf("legitimate truncated fork rejected: %v", err)
			}
		})
	}
}

func TestReceiptHistoryRejectsMidChainTruncationWithGuidance(t *testing.T) {
	const id = "45454545-4545-4545-8545-454545454545"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, _, turn3 := causeConversation(t, h, "cipher")
	// Dropping the middle assistant boundary (with its tool_result) breaks the
	// recorded Previous chain at turn 3.
	messages := append(append([]any{}, turn1...), turn3...)
	messages = append(messages, map[string]any{"role": "user", "content": "summarize"})
	replay := replayCauseOf(t, h.Check(context.Background(), "gpt-5.6-sol", "owner", j(messages)))
	if replay.Cause != CauseChain {
		t.Fatalf("cause = %d, want CauseChain", replay.Cause)
	}
	if !strings.Contains(replay.Error(), "(reason: replayed history does not match the recorded receipt chain") {
		t.Fatal(replay.Error())
	}
}

func TestReceiptHistoryRejectsMidPairTruncationBeforeHistory(t *testing.T) {
	const id = "46464646-4646-4646-8646-464646464646"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	_, turn2, _ := causeConversation(t, h, "cipher")
	// A slice that ends between the assistant tool_use and its tool_result is
	// refused before translation: the receipt-layer Check runs first and the
	// sliced boundary no longer matches the recorded chain.
	messages := append([]any{map[string]any{"role": "user", "content": "hello"}}, turn2[:2]...)
	_, _, err := Request("gpt-5.6-sol", j(map[string]any{"max_tokens": 100, "messages": messages}), Limits{History: h, CredentialScope: "owner"})
	if err == nil {
		t.Fatal("mid-pair slice accepted")
	}
	replay := replayCauseOf(t, err)
	if replay.Cause != CauseChain {
		t.Fatalf("cause = %d, want CauseChain", replay.Cause)
	}
}

func TestReceiptHistoryRejectsForeignItemReplay(t *testing.T) {
	const id = "47474747-4747-4747-8747-474747474747"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, turn2, _ := causeConversation(t, h, "cipher")
	// Same shape, same public content, but a forged envelope whose items were
	// never issued by this gateway: the digest cannot match any receipt.
	forged := causeEnvelope(t, "rs_1", "forged")
	blocks := []any{map[string]any{"type": "redacted_thinking", "data": forged.Data()}, map[string]any{"type": "text", "text": "one"}}
	foreign := []any{
		map[string]any{"role": "user", "content": "hello"},
		map[string]any{"role": "assistant", "content": blocks},
	}
	replay := replayCauseOf(t, h.Check(context.Background(), "gpt-5.6-sol", "owner", j(foreign)))
	if replay.Cause != CauseChain {
		t.Fatalf("cause = %d, want CauseChain", replay.Cause)
	}
	// A published conversation replayed into an unrelated empty root is a
	// lineage miss: no metadata-selectable seeding exists.
	other := NewGPTSubscriptionReceiptHistory(causeStore(t, "48484848-4848-4848-8848-484848484848"), "48484848-4848-4848-8848-484848484848", id)
	messages := append(append([]any{}, turn1...), turn2...)
	replay = replayCauseOf(t, other.Check(context.Background(), "gpt-5.6-sol", "owner", j(messages)))
	if replay.Cause != CauseLineage {
		t.Fatalf("cause = %d, want CauseLineage", replay.Cause)
	}
	if !strings.Contains(replay.Error(), "(reason: no recorded history exists for this session") {
		t.Fatal(replay.Error())
	}
}

func TestReceiptHistorySpawnShapedFreshHistoryNeedsNoLineageSeeding(t *testing.T) {
	const id = "49494949-4949-4949-8949-494949494949"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	// A model-arg spawn's first turn is fresh: system prompt plus task, no
	// assistant boundary, so the empty receipt root authorizes it as a branch
	// without any lineage seeding.
	spawn := []any{
		map[string]any{"role": "system", "content": "You are a coordinator with an Agent tool."},
		map[string]any{"role": "user", "content": "Investigate and report."},
	}
	for _, model := range []string{"gpt-5.6-sol", "gpt-6-astra"} {
		limits := Limits{History: h, CredentialScope: "owner"}
		if _, _, err := Request(model, j(map[string]any{"max_tokens": 100, "messages": spawn}), limits); err != nil {
			t.Fatalf("spawn-shaped request rejected on %s: %v", model, err)
		}
		if err := h.Check(context.Background(), model, "owner", j(spawn)); err != nil {
			t.Fatalf("spawn-shaped history rejected on %s: %v", model, err)
		}
	}
}

func TestReceiptHistoryKeepsStrippedReasoningRejection(t *testing.T) {
	const id = "50505050-5050-5050-8550-505050505050"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, _, _ := causeConversation(t, h, "cipher")
	stripped := []any{
		turn1[0],
		map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "one"}}},
	}
	replay := replayCauseOf(t, h.Check(context.Background(), "gpt-5.6-sol", "owner", j(stripped)))
	if replay.Cause != CauseReasoning {
		t.Fatalf("cause = %d, want CauseReasoning", replay.Cause)
	}
	if !strings.Contains(replay.Error(), "(reason: replayed history omits gateway-issued reasoning") {
		t.Fatal(replay.Error())
	}
}

func TestReceiptHistoryClassifiesDesyncWedgeAfterUnpublishedTurn(t *testing.T) {
	const id = "51515151-5151-5151-8551-515151515151"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, _, _ := causeConversation(t, h, "cipher")
	// The desync wedge shape (family 0efb66f7): the client recorded an
	// assistant turn the gateway never published, so every later replay of it
	// is rejected as a chain mismatch while a bare retry still passes.
	phantom := causeEnvelope(t, "rs_phantom", "never-published")
	messages := append(append([]any{}, turn1...),
		map[string]any{"role": "assistant", "content": []any{
			map[string]any{"type": "redacted_thinking", "data": phantom.Data()},
			map[string]any{"type": "text", "text": "half answer"},
		}},
		map[string]any{"role": "user", "content": "continue"},
	)
	replay := replayCauseOf(t, h.Check(context.Background(), "gpt-5.6-sol", "owner", j(messages)))
	if replay.Cause != CauseChain {
		t.Fatalf("cause = %d, want CauseChain", replay.Cause)
	}
	if err := h.Check(context.Background(), "gpt-5.6-sol", "owner", j(turn1)); err != nil {
		t.Fatalf("exact retry after the failed turn rejected: %v", err)
	}
}
