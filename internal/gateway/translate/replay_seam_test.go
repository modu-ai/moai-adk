package translate

// Card t846: the three seam shapes the t707 follow-up RED design names
// (.moai/reports/t707/followup-red.md), exercised on the current code. These
// are measurement tests: PASS proves the seam innocent, an unexpected FAIL is
// the defect reproduction. The shapes ride the same public client flow as
// replay_roundtrip_test.go — subscription stream publish, client-model
// reconstruction, replayed-history Check.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// seamDeadWriter is the client whose pipe died: every write fails with
// nothing delivered. The terminal publish happens before the first emit, so a
// stream into this writer lands its candidate while the client keeps nothing.
type seamDeadWriter struct{}

func (seamDeadWriter) Write(p []byte) (int, error) { return 0, errors.New("client connection gone") }

// seamCandidates snapshots the manifest and returns its candidates.
func seamCandidates(t *testing.T, store *receipt.Store) []receipt.Candidate {
	t.Helper()
	m, err := store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return m.Candidates()
}

// seamBody assembles the subscription request body the client sends.
func seamBody(t *testing.T, messages []any) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"model": "ignored", "max_tokens": 123, "stream": true,
		"messages": messages,
		"tools":    []any{map[string]any{"name": "a.b", "input_schema": map[string]any{"type": "object"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// seamPublishTurn streams one response through the subscription path and lets
// the terminal publish the turn. It returns the client-visible assistant
// message exactly as replayAssistantMessage reconstructs it.
func seamPublishTurn(t *testing.T, name string, limits Limits, messages []any, buildItems func(c *ResponseContext) []map[string]any) map[string]any {
	t.Helper()
	_, c, err := Request("gpt-5.6-sol", seamBody(t, messages), limits)
	if err != nil {
		t.Fatalf("%s: request: %v", name, err)
	}
	var out strings.Builder
	if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(replayStream(t, c, buildItems(c)), ""))), &out); err != nil {
		t.Fatalf("%s: stream: %v", name, err)
	}
	if !strings.Contains(out.String(), "event: message_stop") {
		t.Fatalf("%s: stream did not complete: %s", name, out.String())
	}
	return replayAssistantMessage(t, out.String())
}

// seamReplayCheck re-submits a replayed history and requires the Check to pass.
func seamReplayCheck(t *testing.T, name string, limits Limits, messages []any) {
	t.Helper()
	if _, _, err := Request("gpt-5.6-sol", seamBody(t, messages), limits); err != nil {
		t.Logf("%s: dump: %s", name, seamDump(messages))
		t.Fatalf("%s: replayed history rejected: %v", name, err)
	}
}

func seamDump(messages []any) string {
	b, err := json.Marshal(messages)
	if err != nil {
		return fmt.Sprintf("unmarshalable: %v", err)
	}
	return string(b)
}

// TestReplayFallbackRetryAfterAbandonedAttempt pins the harmless shape behind
// the live fallback: the first attempt completes and its candidate lands, but
// the client's pipe dies before the first emitted byte, so the client never
// records the turn. The identical request is retried, the retry turn publishes
// its own candidate, and both the retry replay and the two-candidate manifest
// stay clean.
func TestReplayFallbackRetryAfterAbandonedAttempt(t *testing.T) {
	const id = "44444444-4444-4444-8444-444444444447"
	store := newReceiptStore(t, id)
	limits := Limits{History: NewReceiptHistory(store, id, id), CredentialScope: "scope"}
	user := []any{map[string]any{"role": "user", "content": "hello"}}

	// Attempt 1: the upstream response completes and publishes, then the very
	// first client-visible emit fails on the dead pipe.
	_, c1, err := Request("gpt-5.6-sol", seamBody(t, user), limits)
	if err != nil {
		t.Fatal(err)
	}
	attempt1 := []map[string]any{
		replayReasoningItem("rs_1", "cipher-1"),
		{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c1, "a.b"), "arguments": `{"a":1}`, "status": "completed"},
	}
	emitErr := c1.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(replayStream(t, c1, attempt1), ""))), seamDeadWriter{})
	if emitErr == nil {
		t.Fatal("first attempt: expected the dead client pipe to fail the stream")
	}
	got := seamCandidates(t, store)
	if len(got) != 1 {
		t.Fatalf("first attempt: expected the abandoned candidate to be published, got %d candidates", len(got))
	}

	// Retry: the identical request, a fresh upstream response, a live client.
	assistant := seamPublishTurn(t, "retry", limits, user, func(c *ResponseContext) []map[string]any {
		return []map[string]any{
			replayReasoningItem("rs_retry", "cipher-retry"),
			{"id": "item-retry", "type": "function_call", "call_id": "call_retry", "name": alias(c, "a.b"), "arguments": `{"a":2}`, "status": "completed"},
		}
	})

	got = seamCandidates(t, store)
	if len(got) != 2 {
		t.Fatalf("retry: expected exactly 2 candidates (abandoned + retry), got %d", len(got))
	}

	// The retry turn replays clean: the abandoned candidate at a different
	// prefix must not interfere.
	seamReplayCheck(t, "retry replay", limits, []any{
		map[string]any{"role": "user", "content": "hello"},
		assistant,
		map[string]any{"role": "user", "content": []any{replayToolResult("call_retry")}},
	})
}

// TestReplayConcurrentPublishDistinctTurns pins the flock+CAS contract: two
// concurrent publish paths on one store root both land, no lost update, and
// each branch replays clean afterwards.
func TestReplayConcurrentPublishDistinctTurns(t *testing.T) {
	user := []any{map[string]any{"role": "user", "content": "hello"}}
	for round := 0; round < 3; round++ {
		t.Run(fmt.Sprintf("round-%d", round), func(t *testing.T) {
			const id = "44444444-4444-4444-8444-444444444448"
			store := newReceiptStore(t, id)
			limits := Limits{History: NewReceiptHistory(store, id, id), CredentialScope: "scope"}
			bodies := [2]func(c *ResponseContext) []map[string]any{
				func(c *ResponseContext) []map[string]any {
					return []map[string]any{
						replayReasoningItem("rs_a", "cipher-a"),
						{"id": "item-a", "type": "function_call", "call_id": "call_a", "name": alias(c, "a.b"), "arguments": `{"branch":"a"}`, "status": "completed"},
					}
				},
				func(c *ResponseContext) []map[string]any {
					return []map[string]any{
						replayReasoningItem("rs_b", "cipher-b"),
						{"id": "item-b", "type": "function_call", "call_id": "call_b", "name": alias(c, "a.b"), "arguments": `{"branch":"b"}`, "status": "completed"},
					}
				},
			}
			// Build both request contexts up front, then release the streams
			// together so the two publishes race on the same manifest.
			contexts := [2]*ResponseContext{}
			for g := range contexts {
				_, c, err := Request("gpt-5.6-sol", seamBody(t, user), limits)
				if err != nil {
					t.Fatal(err)
				}
				contexts[g] = c
			}
			type outcome struct {
				branch  int
				emitted string
				err     error
			}
			results := make(chan outcome, 2)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for g := range contexts {
				wg.Add(1)
				go func(g int) {
					defer wg.Done()
					<-start
					var out strings.Builder
					err := contexts[g].SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(replayStream(t, contexts[g], bodies[g](contexts[g])), ""))), &out)
					results <- outcome{branch: g, emitted: out.String(), err: err}
				}(g)
			}
			close(start)
			wg.Wait()
			close(results)
			assistants := [2]map[string]any{}
			for i := 0; i < 2; i++ {
				r := <-results
				if r.err != nil {
					t.Fatalf("branch %d: stream: %v", r.branch, r.err)
				}
				if !strings.Contains(r.emitted, "event: message_stop") {
					t.Fatalf("branch %d: stream did not complete: %s", r.branch, r.emitted)
				}
				assistants[r.branch] = replayAssistantMessage(t, r.emitted)
			}
			got := seamCandidates(t, store)
			if len(got) != 2 {
				t.Fatalf("expected both concurrent publishes to land (2 candidates), got %d", len(got))
			}
			// Both branches replay clean against the shared manifest.
			for g := 0; g < 2; g++ {
				callID := "call_a"
				if g == 1 {
					callID = "call_b"
				}
				seamReplayCheck(t, "branch replay", limits, []any{
					map[string]any{"role": "user", "content": "hello"},
					assistants[g],
					map[string]any{"role": "user", "content": []any{replayToolResult(callID)}},
				})
			}
		})
	}
}

// TestReplayForkChainCrossing pins the fork-on-subagent-chain shape from the
// live incident window: a summary fork publishes its turn on top of a finished
// subagent chain on the same store root, and afterwards the subagent's own
// replay — which never contained the fork turn — must still pass. A fork
// candidate carries fork-specific content, so its (prefix, previous) pair
// cannot collide with a subagent boundary in either compatible domain; both
// the required (opaque reasoning) and the empty (no reasoning) fork turn are
// measured, since checkObserved's required-collision rule only inspects empty
// replay boundaries.
func TestReplayForkChainCrossing(t *testing.T) {
	for _, tc := range []struct {
		name          string
		forkReasoning bool
	}{
		{"required fork turn", true},
		{"empty fork turn", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const id = "44444444-4444-4444-8444-444444444449"
			store := newReceiptStore(t, id)
			limits := Limits{History: NewGPTSubscriptionReceiptHistory(store, id, id), CredentialScope: "scope"}

			// Subagent chain: two completed opaque turns.
			user := []any{map[string]any{"role": "user", "content": "hello"}}
			assistant1 := seamPublishTurn(t, "chain turn 1", limits, user, func(c *ResponseContext) []map[string]any {
				return []map[string]any{
					replayReasoningItem("rs_1", "cipher-1"),
					{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"), "arguments": `{"step":1}`, "status": "completed"},
				}
			})
			chain1 := []any{
				map[string]any{"role": "user", "content": "hello"},
				assistant1,
				map[string]any{"role": "user", "content": []any{replayToolResult("call_1")}},
			}
			assistant2 := seamPublishTurn(t, "chain turn 2", limits, chain1, func(c *ResponseContext) []map[string]any {
				return []map[string]any{
					replayReasoningItem("rs_2", "cipher-2"),
					{"id": "item-2", "type": "function_call", "call_id": "call_2", "name": alias(c, "a.b"), "arguments": `{"step":2}`, "status": "completed"},
				}
			})
			chain2 := append(append([]any{}, chain1...),
				assistant2,
				map[string]any{"role": "user", "content": []any{replayToolResult("call_2")}},
			)

			// The summary fork: subagent history [0..k] plus its own prompt.
			forkMessages := append(append([]any{}, chain2...), map[string]any{"role": "user", "content": "summarize the findings"})
			var forkAssistant map[string]any
			if tc.forkReasoning {
				forkAssistant = seamPublishTurn(t, "fork turn", limits, forkMessages, func(c *ResponseContext) []map[string]any {
					return []map[string]any{
						replayReasoningItem("rs_fork", "cipher-fork"),
						{"id": "item-fork", "type": "function_call", "call_id": "call_fork", "name": alias(c, "a.b"), "arguments": `{"summary":true}`, "status": "completed"},
					}
				})
			} else {
				forkAssistant = seamPublishTurn(t, "fork turn", limits, forkMessages, func(c *ResponseContext) []map[string]any {
					return []map[string]any{
						{"id": "msg-fork", "type": "message", "role": "assistant", "status": "completed",
							"content": []any{map[string]any{"type": "output_text", "text": "요약 완료", "annotations": []any{}}}},
					}
				})
			}
			if got := len(seamCandidates(t, store)); got != 3 {
				t.Fatalf("after fork: expected 3 candidates (2 chain + 1 fork), got %d", got)
			}

			// The fork's own replay survives.
			forkReplay := append(append([]any{}, forkMessages...), forkAssistant)
			if tc.forkReasoning {
				forkReplay = append(forkReplay, map[string]any{"role": "user", "content": []any{replayToolResult("call_fork")}})
			}
			seamReplayCheck(t, "fork replay", limits, forkReplay)

			// The subagent's own replay — no fork turn in it — must survive the
			// fork candidate sitting on the same root.
			seamReplayCheck(t, "subagent replay", limits, append(append([]any{}, chain2...), map[string]any{"role": "user", "content": "continue the work"}))

			// The subagent keeps publishing past the fork point.
			assistant3 := seamPublishTurn(t, "chain turn 3", limits, append(append([]any{}, chain2...), map[string]any{"role": "user", "content": "continue the work"}), func(c *ResponseContext) []map[string]any {
				return []map[string]any{
					replayReasoningItem("rs_3", "cipher-3"),
					{"id": "item-3", "type": "function_call", "call_id": "call_3", "name": alias(c, "a.b"), "arguments": `{"step":3}`, "status": "completed"},
				}
			})
			got := seamCandidates(t, store)
			if len(got) != 4 {
				t.Fatalf("after continuation: expected 4 candidates (2 chain + 1 fork + 1 continuation), got %d", len(got))
			}
			seamReplayCheck(t, "full subagent replay", limits, append(append([]any{}, chain2...),
				map[string]any{"role": "user", "content": "continue the work"},
				assistant3,
				map[string]any{"role": "user", "content": []any{replayToolResult("call_3")}},
			))
		})
	}
}
