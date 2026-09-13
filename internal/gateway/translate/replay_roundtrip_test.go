package translate

// Card t707 regression: the full publish→check round trip across the
// subscription stream boundary. The client model assembles the replayed
// assistant message exactly as the public Messages client does: one assistant
// message per response, content blocks merged from content_block events, tool
// inputs parsed from input_json_delta, followed by a tool_result user message.
// Both observed tool ID forms (opaque-wrapped marker and raw call ID) and the
// Edit replace_all shape are pinned here.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

func newReceiptStore(t *testing.T, id string) *receipt.Store {
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

// replayEvent builds one upstream subscription SSE frame.
func replayEvent(typ string, fields map[string]any) string {
	fields["type"] = typ
	return "event: " + typ + "\ndata: " + string(j(fields)) + "\n\n"
}

func replayReasoningItem(id, cipher string) map[string]any {
	return map[string]any{"id": id, "type": "reasoning", "summary": []any{}, "content": []any{}, "encrypted_content": cipher, "status": "completed"}
}

// replayStream builds a subscription-style upstream stream for the given
// completed output items with a sparse terminal (empty output array), plus the
// in-progress added events and deltas a real stream carries.
func replayStream(t *testing.T, c *ResponseContext, items []map[string]any) []string {
	t.Helper()
	events := []string{replayEvent("response.created", map[string]any{"response": map[string]any{"id": "resp-1", "model": "gpt-5.6-sol", "status": "in_progress"}})}
	for i, item := range items {
		added := map[string]any{}
		for k, v := range item {
			added[k] = v
		}
		added["status"] = "in_progress"
		if item["type"] == "function_call" {
			added["arguments"] = ""
		}
		if item["type"] == "message" {
			added["content"] = []any{}
		}
		events = append(events, replayEvent("response.output_item.added", map[string]any{"output_index": i, "item": added}))
		if item["type"] == "function_call" {
			args := item["arguments"].(string)
			events = append(events,
				replayEvent("response.function_call_arguments.delta", map[string]any{"output_index": i, "item_id": item["id"], "delta": args}),
				replayEvent("response.function_call_arguments.done", map[string]any{"output_index": i, "item_id": item["id"], "arguments": args}))
		}
		if item["type"] == "message" {
			parts := item["content"].([]any)
			for ci, p := range parts {
				part := map[string]any{}
				for k, v := range p.(map[string]any) {
					part[k] = v
				}
				text := part["text"].(string)
				part["text"] = ""
				events = append(events,
					replayEvent("response.content_part.added", map[string]any{"output_index": i, "item_id": item["id"], "content_index": ci, "part": part}),
					replayEvent("response.output_text.delta", map[string]any{"output_index": i, "item_id": item["id"], "content_index": ci, "delta": text}),
					replayEvent("response.output_text.done", map[string]any{"output_index": i, "item_id": item["id"], "content_index": ci, "text": text}))
				part["text"] = text
				events = append(events, replayEvent("response.content_part.done", map[string]any{"output_index": i, "item_id": item["id"], "content_index": ci, "part": part}))
			}
		}
		events = append(events, replayEvent("response.output_item.done", map[string]any{"output_index": i, "item": item}))
	}
	events = append(events, replayEvent("response.completed", map[string]any{"response": map[string]any{"id": "resp-1", "model": "gpt-5.6-sol", "status": "completed", "output": []any{}, "usage": map[string]any{"input_tokens": 7, "output_tokens": 3}}}))
	return events
}

// replayAssistantMessage reconstructs the public client's view of the emitted
// Anthropic-side stream: one assistant message holding every content block.
func replayAssistantMessage(t *testing.T, emitted string) map[string]any {
	t.Helper()
	var blocks []any
	for _, line := range strings.Split(emitted, "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &m); err != nil {
			t.Fatal(err)
		}
		switch m["type"] {
		case "content_block_start":
			blocks = append(blocks, m["content_block"].(map[string]any))
		case "content_block_delta":
			delta := m["delta"].(map[string]any)
			block := blocks[len(blocks)-1].(map[string]any)
			switch delta["type"] {
			case "text_delta":
				block["text"] = block["text"].(string) + delta["text"].(string)
			case "input_json_delta":
				acc, _ := block["__args"].(string)
				block["__args"] = acc + delta["partial_json"].(string)
			}
		}
	}
	for _, v := range blocks {
		block := v.(map[string]any)
		if raw, ok := block["__args"]; ok {
			var input map[string]any
			if err := json.Unmarshal([]byte(raw.(string)), &input); err != nil {
				t.Fatal(err)
			}
			delete(block, "__args")
			block["input"] = input
		}
	}
	return map[string]any{"role": "assistant", "content": blocks}
}

// replayToolResult is the client's tool_result user message for a call ID.
func replayToolResult(id string) map[string]any {
	return map[string]any{"type": "tool_result", "tool_use_id": id, "content": "done"}
}

// runReplayRoundTrip streams one response through the subscription path and
// re-submits the replayed history. Every step must succeed: publish, restore,
// canonical replay and check.
func runReplayRoundTrip(t *testing.T, name string, buildItems func(c *ResponseContext) []map[string]any, toolResults []any) {
	t.Helper()
	const id = "44444444-4444-4444-8444-444444444444"
	store := newReceiptStore(t, id)
	limits := Limits{History: NewReceiptHistory(store, id, id), CredentialScope: "scope"}
	body := `{"model":"ignored","max_tokens":123,"stream":true,"messages":[{"role":"user","content":"hello"}],"tools":[{"name":"a.b","input_schema":{"type":"object"}}]}`
	_, c, err := Request("gpt-5.6-sol", []byte(body), limits)
	if err != nil {
		t.Fatalf("%s: request 1: %v", name, err)
	}
	var out bytes.Buffer
	if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(replayStream(t, c, buildItems(c)), ""))), &out); err != nil {
		t.Fatalf("%s: stream: %v", name, err)
	}
	if !strings.Contains(out.String(), "event: message_stop") {
		t.Fatalf("%s: stream did not complete: %s", name, out.String())
	}
	messages := []any{
		map[string]any{"role": "user", "content": "hello"},
		replayAssistantMessage(t, out.String()),
	}
	if toolResults != nil {
		messages = append(messages, map[string]any{"role": "user", "content": toolResults})
	}
	next, err := json.Marshal(map[string]any{"model": "ignored", "max_tokens": 123, "stream": true, "messages": messages, "tools": []any{map[string]any{"name": "a.b", "input_schema": map[string]any{"type": "object"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = Request("gpt-5.6-sol", next, limits); err != nil {
		t.Fatalf("%s: replayed history rejected: %v", name, err)
	}
}

func TestReplayRoundTripMarkerToolID(t *testing.T) {
	runReplayRoundTrip(t, "marker tool id", func(c *ResponseContext) []map[string]any {
		return []map[string]any{
			replayReasoningItem("rs_1", "cipher-1"),
			{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"), "arguments": `{"a":1}`, "status": "completed"},
		}
	}, []any{replayToolResult("call_1")})
}

// TestReplayRoundTripConsecutiveReasoningItems pins the live failure shape
// (family 1f14d174, response 11): two consecutive opaque reasoning items
// followed by one function_call, all carried in one v2 envelope.
func TestReplayRoundTripConsecutiveReasoningItems(t *testing.T) {
	runReplayRoundTrip(t, "consecutive reasoning", func(c *ResponseContext) []map[string]any {
		return []map[string]any{
			replayReasoningItem("rs_a", "cipher-a"),
			replayReasoningItem("rs_b", "cipher-b"),
			{"id": "item-2", "type": "function_call", "call_id": "call_Gpe6", "name": alias(c, "a.b"),
				"arguments": `{"replace_all":false,"file_path":"x.md","old_string":"a","new_string":"b"}`, "status": "completed"},
		}
	}, []any{replayToolResult("call_Gpe6")})
}

func TestReplayRoundTripInterleavedReasoning(t *testing.T) {
	runReplayRoundTrip(t, "interleaved reasoning", func(c *ResponseContext) []map[string]any {
		return []map[string]any{
			replayReasoningItem("rs_1", "cipher-1"),
			{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"), "arguments": `{"a":1}`, "status": "completed"},
			replayReasoningItem("rs_2", "cipher-2"),
			{"id": "item-3", "type": "function_call", "call_id": "call_2", "name": alias(c, "a.b"), "arguments": `{"b":2}`, "status": "completed"},
		}
	}, []any{replayToolResult("call_1"), replayToolResult("call_2")})
}

func TestReplayRoundTripRawCallIDWithoutReasoning(t *testing.T) {
	runReplayRoundTrip(t, "raw call id", func(c *ResponseContext) []map[string]any {
		return []map[string]any{
			{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"), "arguments": `{"a":1}`, "status": "completed"},
		}
	}, []any{replayToolResult("call_1")})
}

func TestReplayRoundTripEditReplaceAllUnicode(t *testing.T) {
	runReplayRoundTrip(t, "edit replace_all unicode", func(c *ResponseContext) []map[string]any {
		return []map[string]any{
			replayReasoningItem("rs_1", "cipher-1"),
			{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"),
				"arguments": `{"replace_all":true,"file_path":"/tmp/x.md","old_string":"낡은 문장","new_string":"새 문장 <>&\"quote\""}`, "status": "completed"},
		}
	}, []any{replayToolResult("call_1")})
}

func TestReplayRoundTripTextAnswer(t *testing.T) {
	runReplayRoundTrip(t, "text answer", func(c *ResponseContext) []map[string]any {
		return []map[string]any{
			{"id": "msg-1", "type": "message", "role": "assistant", "status": "completed",
				"content": []any{map[string]any{"type": "output_text", "text": "정답입니다", "annotations": []any{}}}},
		}
	}, nil)
}

func TestReplayRoundTripReasoningWithSummary(t *testing.T) {
	runReplayRoundTrip(t, "reasoning summary", func(c *ResponseContext) []map[string]any {
		item := replayReasoningItem("rs_1", "cipher-1")
		item["summary"] = []any{map[string]any{"type": "summary_text", "text": "thinking"}}
		return []map[string]any{
			item,
			{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"), "arguments": `{}`, "status": "completed"},
		}
	}, []any{replayToolResult("call_1")})
}

func TestReplayRoundTripTwoSerialTurns(t *testing.T) {
	// Two consecutive publish/check cycles: turn 2 replays turn 1's message,
	// turn 3 replays both. Fast serial requests must stay chain-consistent.
	const id = "44444444-4444-4444-8444-444444444445"
	store := newReceiptStore(t, id)
	limits := Limits{History: NewReceiptHistory(store, id, id), CredentialScope: "scope"}
	tools := []any{map[string]any{"name": "a.b", "input_schema": map[string]any{"type": "object"}}}
	first := `{"model":"ignored","max_tokens":123,"stream":true,"messages":[{"role":"user","content":"hello"}],"tools":[{"name":"a.b","input_schema":{"type":"object"}}]}`
	_, c, err := Request("gpt-5.6-sol", []byte(first), limits)
	if err != nil {
		t.Fatal(err)
	}
	items := []map[string]any{
		replayReasoningItem("rs_1", "cipher-1"),
		{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"), "arguments": `{"a":1}`, "status": "completed"},
	}
	var out bytes.Buffer
	if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(replayStream(t, c, items), ""))), &out); err != nil {
		t.Fatal(err)
	}
	assistant1 := replayAssistantMessage(t, out.String())
	second, err := json.Marshal(map[string]any{"model": "ignored", "max_tokens": 123, "stream": true,
		"messages": []any{
			map[string]any{"role": "user", "content": "hello"},
			assistant1,
			map[string]any{"role": "user", "content": []any{replayToolResult("call_1")}},
		},
		"tools": tools})
	if err != nil {
		t.Fatal(err)
	}
	_, c2, err := Request("gpt-5.6-sol", second, limits)
	if err != nil {
		t.Fatalf("second request rejected: %v", err)
	}
	items2 := []map[string]any{
		{"id": "msg-2", "type": "message", "role": "assistant", "status": "completed",
			"content": []any{map[string]any{"type": "output_text", "text": "완료", "annotations": []any{}}}},
	}
	var out2 bytes.Buffer
	if err := c2.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(replayStream(t, c2, items2), ""))), &out2); err != nil {
		t.Fatal(err)
	}
	assistant2 := replayAssistantMessage(t, out2.String())
	third, err := json.Marshal(map[string]any{"model": "ignored", "max_tokens": 123, "stream": true,
		"messages": []any{
			map[string]any{"role": "user", "content": "hello"},
			assistant1,
			map[string]any{"role": "user", "content": []any{replayToolResult("call_1")}},
			assistant2,
			map[string]any{"role": "user", "content": "thanks"},
		},
		"tools": tools})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = Request("gpt-5.6-sol", third, limits); err != nil {
		t.Fatalf("third request rejected: %v", err)
	}
}

// TestReplayRoundTripStillRejectsTampering proves the checks above did not
// loosen the anti-replay chain: both observed tool ID forms must still reject
// an edited replay.
func TestReplayRoundTripStillRejectsTampering(t *testing.T) {
	for _, tc := range []struct {
		name    string
		corrupt func(messages []any)
	}{
		{"edited tool input", func(ms []any) {
			a := ms[1].(map[string]any)["content"].([]any)
			for _, b := range a {
				if block, ok := b.(map[string]any); ok && block["type"] == "tool_use" {
					block["input"].(map[string]any)["a"] = 999
				}
			}
		}},
		{"renamed tool", func(ms []any) {
			a := ms[1].(map[string]any)["content"].([]any)
			for _, b := range a {
				if block, ok := b.(map[string]any); ok && block["type"] == "tool_use" {
					block["name"] = "other"
				}
			}
		}},
		{"dropped tool result", func(ms []any) {
			ms[2].(map[string]any)["content"] = []any{}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const id = "44444444-4444-4444-8444-444444444446"
			store := newReceiptStore(t, id)
			limits := Limits{History: NewReceiptHistory(store, id, id), CredentialScope: "scope"}
			body := `{"model":"ignored","max_tokens":123,"stream":true,"messages":[{"role":"user","content":"hello"}],"tools":[{"name":"a.b","input_schema":{"type":"object"}}]}`
			_, c, err := Request("gpt-5.6-sol", []byte(body), limits)
			if err != nil {
				t.Fatal(err)
			}
			items := []map[string]any{
				replayReasoningItem("rs_1", "cipher-1"),
				{"id": "item-1", "type": "function_call", "call_id": "call_1", "name": alias(c, "a.b"), "arguments": `{"a":1}`, "status": "completed"},
			}
			var out bytes.Buffer
			if err := c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(replayStream(t, c, items), ""))), &out); err != nil {
				t.Fatal(err)
			}
			messages := []any{
				map[string]any{"role": "user", "content": "hello"},
				replayAssistantMessage(t, out.String()),
				map[string]any{"role": "user", "content": []any{replayToolResult("call_1")}},
			}
			tc.corrupt(messages)
			next, err := json.Marshal(map[string]any{"model": "ignored", "max_tokens": 123, "stream": true, "messages": messages,
				"tools": []any{map[string]any{"name": "a.b", "input_schema": map[string]any{"type": "object"}}}})
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err = Request("gpt-5.6-sol", next, limits); err == nil {
				t.Fatalf("%s: tampered replay authorized", tc.name)
			}
			if tc.name != "dropped tool result" {
				var replay HistoryReplayError
				if !errors.As(err, &replay) {
					t.Fatalf("%s: tamper rejected outside the chain check: %v", tc.name, err)
				}
			}
		})
	}
}
