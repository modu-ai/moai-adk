package translate

// Card t708 (SPEC-GATEWAY-ENVELOPE-REPAIR-001 M2): the wedge-policy matrix as
// a GREEN-at-arrival characterization lock over the unchanged receipt-history
// validator. t672 named the rejection classes; this matrix adds the
// byte-identical golden error bodies (AC-EVR-002) and the accept-side wedge
// policies (AC-EVR-001): a trailing never-published boundary is a chain
// mismatch, while removing that trailing turn re-roots the tail and replays
// cleanly. Every cell locks the behavior documented in the SPEC; a failing
// cell at arrival would be a SPEC-defect finding, never a code change here.

import (
	"context"
	"testing"
)

// Golden 400 bodies: the fixed guidance sentence plus t672's three reason
// clauses, asserted byte-for-byte (AC-EVR-002). A drift in either the
// sentence or a clause is a wire-visible contract change and must trip these
// cells; the strings are deliberately duplicated from receipt_history.go as
// full literals so an edit to either half of the composed body cannot slip
// through a prefix or substring check.
const (
	goldenChainBody = "conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation" +
		" (reason: replayed history does not match the recorded receipt chain; start a new conversation)"
	goldenLineageBody = "conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation" +
		" (reason: no recorded history exists for this session; resume or fork through the moai launcher, or start a new conversation)"
	goldenReasoningBody = "conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation" +
		" (reason: replayed history omits gateway-issued reasoning recorded at this position; replay the history unmodified or start a new conversation)"
)

// goldenReplayOf asserts a rejection carries the expected class and the exact
// golden body — the whole string, never a substring.
func goldenReplayOf(t *testing.T, err error, cause ReplayCause, body string) HistoryReplayError {
	t.Helper()
	replay := replayCauseOf(t, err)
	if replay.Cause != cause {
		t.Fatalf("cause = %d, want %d", replay.Cause, cause)
	}
	if got := replay.Error(); got != body {
		t.Fatalf("error body drifted from the golden contract:\n got: %q\nwant: %q", got, body)
	}
	return replay
}

// Group 1 — the wedge shape (a trailing never-published assistant boundary,
// family 0efb66f7) is rejected, chain-classified, and carries the
// byte-identical golden chain body.
func TestWedgePolicyTrailingUnpublishedBoundaryIsChainRejected(t *testing.T) {
	const id = "70800001-0000-4708-8708-000000000001"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, _, _ := causeConversation(t, h, "cipher")
	phantom := causeEnvelope(t, "rs_phantom", "never-published")
	messages := append(append([]any{}, turn1...),
		map[string]any{"role": "assistant", "content": []any{
			map[string]any{"type": "redacted_thinking", "data": phantom.Data()},
			map[string]any{"type": "text", "text": "half answer"},
		}},
		map[string]any{"role": "user", "content": "continue"},
	)
	err := h.Check(context.Background(), "gpt-5.6-sol", "owner", j(messages))
	_ = goldenReplayOf(t, err, CauseChain, goldenChainBody)
}

// Group 2 — removing the trailing unpublished turn re-roots the tail: the
// client drops its never-published turn and continues from the recorded
// history, which must replay cleanly.
func TestWedgePolicyTailRerootedHistoryIsAccepted(t *testing.T) {
	const id = "70800002-0000-4708-8708-000000000002"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, _, _ := causeConversation(t, h, "cipher")
	rerooted := append(append([]any{}, turn1...), map[string]any{"role": "user", "content": "continue"})
	if err := h.Check(context.Background(), "gpt-5.6-sol", "owner", j(rerooted)); err != nil {
		t.Fatalf("tail re-rooting after removing the unpublished turn rejected: %v", err)
	}
}

// Group 3 — the mid-drop, foreign-item (digest mismatch), stripped-reasoning
// (marker missing), and lineage-miss (source gone) shapes stay rejected with
// their UNCHANGED classes and byte-identical golden bodies.
func TestWedgePolicyRejectionClassesStayWithGoldenBodies(t *testing.T) {
	ctx := context.Background()
	t.Run("mid drop breaks the chain", func(t *testing.T) {
		const id = "70800003-0000-4708-8708-000000000003"
		h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
		turn1, _, turn3 := causeConversation(t, h, "cipher")
		messages := append(append([]any{}, turn1...), turn3...)
		messages = append(messages, map[string]any{"role": "user", "content": "summarize"})
		err := h.Check(ctx, "gpt-5.6-sol", "owner", j(messages))
		_ = goldenReplayOf(t, err, CauseChain, goldenChainBody)
	})
	t.Run("foreign item is a digest mismatch", func(t *testing.T) {
		const id = "70800004-0000-4708-8708-000000000004"
		h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
		causeConversation(t, h, "cipher")
		// Same shape, same public content, but a forged envelope this gateway
		// never issued: the digest cannot match any receipt in the recorded
		// root.
		forged := causeEnvelope(t, "rs_1", "forged")
		foreign := []any{
			map[string]any{"role": "user", "content": "hello"},
			map[string]any{"role": "assistant", "content": []any{
				map[string]any{"type": "redacted_thinking", "data": forged.Data()},
				map[string]any{"type": "text", "text": "one"},
			}},
		}
		err := h.Check(ctx, "gpt-5.6-sol", "owner", j(foreign))
		_ = goldenReplayOf(t, err, CauseChain, goldenChainBody)
	})
	t.Run("stripped reasoning with surviving marker", func(t *testing.T) {
		const id = "70800005-0000-4708-8708-000000000005"
		h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
		turn1, _, _ := causeConversation(t, h, "cipher")
		stripped := []any{
			turn1[0],
			map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "one"}}},
		}
		err := h.Check(ctx, "gpt-5.6-sol", "owner", j(stripped))
		_ = goldenReplayOf(t, err, CauseReasoning, goldenReasoningBody)
	})
	t.Run("lineage miss on an unrelated empty root", func(t *testing.T) {
		const id = "70800006-0000-4708-8708-000000000006"
		other := "70800007-0000-4708-8708-000000000007"
		h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
		turn1, turn2, _ := causeConversation(t, h, "cipher")
		// A published conversation replayed into an unrelated empty root is a
		// lineage miss: the recorded source is gone from this root.
		foreignRoot := NewGPTSubscriptionReceiptHistory(causeStore(t, other), other, id)
		messages := append(append([]any{}, turn1...), turn2...)
		err := foreignRoot.Check(ctx, "gpt-5.6-sol", "owner", j(messages))
		_ = goldenReplayOf(t, err, CauseLineage, goldenLineageBody)
	})
}

// Group 4 — an exact retry after the failed turn is accepted: re-sending the
// last published state (the bare retry, and the full recorded chain) never
// regresses behind the wedge rejection.
func TestWedgePolicyExactRetryAfterFailedTurnAccepted(t *testing.T) {
	const id = "70800008-0000-4708-8708-000000000008"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, turn2, turn3 := causeConversation(t, h, "cipher")
	ctx := context.Background()
	phantom := causeEnvelope(t, "rs_phantom", "never-published")
	wedge := append(append([]any{}, turn1...),
		map[string]any{"role": "assistant", "content": []any{
			map[string]any{"type": "redacted_thinking", "data": phantom.Data()},
			map[string]any{"type": "text", "text": "half answer"},
		}},
		map[string]any{"role": "user", "content": "continue"},
	)
	if err := h.Check(ctx, "gpt-5.6-sol", "owner", j(wedge)); err == nil {
		t.Fatal("wedge shape unexpectedly accepted before the retry cells")
	}
	if err := h.Check(ctx, "gpt-5.6-sol", "owner", j(turn1)); err != nil {
		t.Fatalf("bare exact retry after the failed turn rejected: %v", err)
	}
	full := append(append(append([]any{}, turn1...), turn2...), turn3...)
	if err := h.Check(ctx, "gpt-5.6-sol", "owner", j(full)); err != nil {
		t.Fatalf("full recorded-chain retry after the failed turn rejected: %v", err)
	}
}

// Group 5 — a truncated fork replay of published, pairing-valid boundaries is
// accepted, including across the compatible subscription domain.
func TestWedgePolicyTruncatedForkReplayAccepted(t *testing.T) {
	const id = "70800009-0000-4708-8708-000000000009"
	h := NewGPTSubscriptionReceiptHistory(causeStore(t, id), id, id)
	turn1, turn2, _ := causeConversation(t, h, "cipher")
	fork := append(append(append([]any{}, turn1...), turn2...), map[string]any{"role": "user", "content": "summarize"})
	if err := h.Check(context.Background(), "gpt-5.6-sol", "owner", j(fork)); err != nil {
		t.Fatalf("legitimate truncated fork rejected: %v", err)
	}
	if err := h.Check(context.Background(), "gpt-6-astra", "owner", j(fork)); err != nil {
		t.Fatalf("compatible-domain fork rejected: %v", err)
	}
}

// Group 6 — the golden error bodies are byte-identical over
// HistoryReplayError.Error directly: the historical guidance sentence plus
// t672's three reason clauses, including the zero-value chain default.
func TestWedgePolicyGoldenErrorBodiesByteIdentical(t *testing.T) {
	cases := []struct {
		cause ReplayCause
		body  string
	}{
		{CauseChain, goldenChainBody},
		{CauseLineage, goldenLineageBody},
		{CauseReasoning, goldenReasoningBody},
	}
	for _, c := range cases {
		if got := (HistoryReplayError{Cause: c.cause}).Error(); got != c.body {
			t.Fatalf("cause %d body drifted:\n got: %q\nwant: %q", c.cause, got, c.body)
		}
	}
	// The zero value reports the chain cause, whose message is the
	// historical guidance plus the chain clause.
	if got := (HistoryReplayError{}).Error(); got != goldenChainBody {
		t.Fatalf("zero-value body drifted:\n got: %q\nwant: %q", got, goldenChainBody)
	}
}
