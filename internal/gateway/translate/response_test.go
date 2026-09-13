package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func responseCtx(t *testing.T) *ResponseContext { _, c := request(t, ordinary); return c }
func alias(c *ResponseContext, name string) string {
	for k, v := range c.names {
		if v == name {
			return k
		}
	}
	panic("name")
}
func textItem(id, text string) map[string]any {
	return map[string]any{"type": "message", "id": id, "role": "assistant", "status": "completed", "content": []any{map[string]any{"type": "output_text", "text": text, "annotations": []any{}}}}
}
func toolItem(id, call, name, args string) map[string]any {
	return map[string]any{"type": "function_call", "id": id, "call_id": call, "name": name, "arguments": args, "status": "completed"}
}
func finalResponse(status string, items ...any) map[string]any {
	return map[string]any{"id": "resp-1", "model": "gpt-5.6-sol", "status": status, "output": items, "usage": map[string]any{"input_tokens": 7, "output_tokens": 3}}
}
func j(v any) []byte {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return b
}
func TestNonStreamStopsAndToolReverseMap(t *testing.T) {
	c := responseCtx(t)
	for _, tc := range []struct {
		status, reason string
		items          []any
	}{{"completed", "end_turn", []any{textItem("msg", "hello")}}, {"completed", "tool_use", []any{toolItem("item-A", "call-A", alias(c, "a.b"), `{"a":1}`)}}, {"incomplete", "max_tokens", []any{textItem("msg", "partial")}}} {
		r := finalResponse(tc.status, tc.items...)
		if tc.status == "incomplete" {
			r["incomplete_details"] = map[string]any{"reason": "max_output_tokens"}
		}
		b, e := c.Response(j(r))
		if e != nil {
			t.Fatal(e)
		}
		d := decode(t, b)
		if d["stop_reason"] != tc.reason {
			t.Fatal(d)
		}
		if tc.reason == "tool_use" {
			block := d["content"].([]any)[0].(map[string]any)
			if block["id"] != "call-A" || block["name"] != "a.b" {
				t.Fatal(block)
			}
		}
	}
}

func TestNonStreamAcceptsObservedOutputMessagePhase(t *testing.T) {
	c := responseCtx(t)
	item := textItem("msg", "hello")
	item["phase"] = "final_answer"
	if _, err := c.Response(j(finalResponse("completed", item))); err != nil {
		t.Fatalf("observed final_answer phase rejected: %v", err)
	}
	for _, phase := range []any{"analysis", 1} {
		item["phase"] = phase
		if b, err := c.Response(j(finalResponse("completed", item))); err == nil || b != nil {
			t.Fatalf("invalid output phase accepted: %v", phase)
		}
	}
	item["phase"] = nil
	if _, err := c.Response(j(finalResponse("completed", item))); err != nil {
		t.Fatalf("null output phase rejected: %v", err)
	}
}

func TestNonStreamRejectsEveryIncompleteOrOpaqueShape(t *testing.T) {
	c := responseCtx(t)
	bad := []map[string]any{finalResponse("failed", textItem("msg", "x")), finalResponse("completed"), finalResponse("incomplete", textItem("msg", "x")), finalResponse("completed", toolItem("i", "c", alias(c, "a.b"), "{")), finalResponse("completed", toolItem("i", "c", "unknown", "{}")), finalResponse("completed", map[string]any{"id": "r", "type": "reasoning", "encrypted_content": "secret"}), finalResponse("completed", textItem("same", "x"), textItem("same", "y"))}
	mixed := finalResponse("completed", textItem("m", "done"), toolItem("i", "c", alias(c, "a.b"), "{}"))
	mixed["output"].([]any)[1].(map[string]any)["status"] = "in_progress"
	bad = append(bad, mixed)
	for _, v := range bad {
		if b, e := c.Response(j(v)); e == nil || b != nil {
			t.Fatal("accepted", string(b))
		}
	}
}
func event(typ string, fields map[string]any) string {
	fields["type"] = typ
	return "event: " + typ + "\ndata: " + string(j(fields)) + "\n\n"
}
func created() string {
	return event("response.created", map[string]any{"response": map[string]any{"id": "resp-1", "model": "gpt-5.6-sol", "status": "in_progress"}})
}
func toolEvents(c *ResponseContext) []string {
	a, b := alias(c, "a.b"), alias(c, "a_b")
	return []string{created(), event("response.output_item.added", map[string]any{"output_index": 0, "item": map[string]any{"id": "item-0", "type": "function_call", "call_id": "call-0", "name": a, "arguments": "", "status": "in_progress"}}), event("response.output_item.added", map[string]any{"output_index": 1, "item": map[string]any{"id": "item-1", "type": "function_call", "call_id": "call-1", "name": b, "arguments": "", "status": "in_progress"}}), event("response.function_call_arguments.delta", map[string]any{"output_index": 1, "item_id": "item-1", "delta": "{\"b\":"}), event("response.function_call_arguments.delta", map[string]any{"output_index": 0, "item_id": "item-0", "delta": "{\"a\":1}"}), event("response.function_call_arguments.delta", map[string]any{"output_index": 1, "item_id": "item-1", "delta": "2}"}), event("response.function_call_arguments.done", map[string]any{"output_index": 0, "item_id": "item-0", "arguments": "{\"a\":1}"}), event("response.function_call_arguments.done", map[string]any{"output_index": 1, "item_id": "item-1", "arguments": "{\"b\":2}"}), event("response.output_item.done", map[string]any{"output_index": 1, "item": toolItem("item-1", "call-1", b, `{"b":2}`)}), event("response.output_item.done", map[string]any{"output_index": 0, "item": toolItem("item-0", "call-0", a, `{"a":1}`)}), event("response.completed", map[string]any{"response": finalResponse("completed", toolItem("item-0", "call-0", a, `{"a":1}`), toolItem("item-1", "call-1", b, `{"b":2}`))})}
}
func TestStreamInterleavedToolsDistinctItemAndCallIDs(t *testing.T) {
	c := responseCtx(t)
	var out bytes.Buffer
	if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(toolEvents(c), ""))), &out); e != nil {
		t.Fatal(e)
	}
	s := out.String()
	if strings.Count(s, "event: message_stop") != 1 || strings.Count(s, "event: content_block_stop") != 2 || strings.Count(s, "event: content_block_delta") != 3 {
		t.Fatal(s)
	}
	args := map[float64]string{}
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, "data: ") {
			d := decode(t, []byte(strings.TrimPrefix(line, "data: ")))
			if d["type"] == "content_block_delta" {
				args[d["index"].(float64)] += d["delta"].(map[string]any)["partial_json"].(string)
			}
		}
	}
	if args[0] != `{"a":1}` || args[1] != `{"b":2}` {
		t.Fatal(args)
	}
	if !strings.Contains(s, `"id":"call-0"`) || !strings.Contains(s, `"name":"a.b"`) {
		t.Fatal(s)
	}
}
func textEvents(status string) []string {
	r := finalResponse(status, textItem("item-t", "hello"))
	if status == "incomplete" {
		r["incomplete_details"] = map[string]any{"reason": "max_output_tokens"}
	}
	return []string{created(), event("response.output_item.added", map[string]any{"output_index": 0, "item": map[string]any{"id": "item-t", "type": "message", "role": "assistant", "status": "in_progress", "content": []any{}}}), event("response.content_part.added", map[string]any{"output_index": 0, "item_id": "item-t", "content_index": 0, "part": map[string]any{"type": "output_text", "text": "", "annotations": []any{}}}), event("response.output_text.delta", map[string]any{"output_index": 0, "item_id": "item-t", "content_index": 0, "delta": "hello"}), event("response.output_text.done", map[string]any{"output_index": 0, "item_id": "item-t", "content_index": 0, "text": "hello"}), event("response.content_part.done", map[string]any{"output_index": 0, "item_id": "item-t", "content_index": 0, "part": map[string]any{"type": "output_text", "text": "hello", "annotations": []any{}}}), event("response.output_item.done", map[string]any{"output_index": 0, "item": textItem("item-t", "hello")}), event("response."+status, map[string]any{"response": r})}
}
func TestStreamTextAndMaxOutputStops(t *testing.T) {
	for _, status := range []string{"completed", "incomplete"} {
		c := responseCtx(t)
		var out bytes.Buffer
		if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(textEvents(status), ""))), &out); e != nil {
			t.Fatal(e)
		}
		want := "end_turn"
		if status == "incomplete" {
			want = "max_tokens"
		}
		if !strings.Contains(out.String(), `"stop_reason":"`+want+`"`) || strings.Count(out.String(), `"text":"hello"`) != 1 {
			t.Fatal(out.String())
		}
	}
}
func TestStreamNeverSynthesizesSuccessOnBadInput(t *testing.T) {
	c := responseCtx(t)
	valid := toolEvents(c)
	cases := map[string]string{"EOF": strings.Join(valid[:len(valid)-1], ""), "duplicate start": valid[0] + valid[1] + valid[1], "wrong item": valid[0] + valid[1] + strings.ReplaceAll(valid[4], "item-0", "other"), "duplicate done": strings.Join(valid[:7], "") + valid[6], "error": valid[0] + event("error", map[string]any{"message": "upstream fail"}), "failed": valid[0] + event("response.failed", map[string]any{"response": finalResponse("failed")}), "bad JSON": created() + "data: {\"type\":\"x\",\"type\":\"y\"}\n\n", "done without start": valid[0] + valid[6], "no public": valid[0] + event("response.completed", map[string]any{"response": finalResponse("completed")}), "name changed": strings.Join(valid[:9], "") + strings.ReplaceAll(valid[9], alias(c, "a.b"), "wrong"), "terminal mismatch": strings.Join(valid[:10], "") + strings.ReplaceAll(valid[10], `\"a\":1`, `\"a\":9`), "event mismatch": "event: x\ndata: {\"type\":\"response.created\"}\n\n"}
	te := textEvents("completed")
	cases["mixed unfinished tool"] = strings.Join(te[:7], "") + strings.ReplaceAll(valid[1], `"output_index":0`, `"output_index":1`) + te[7]
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(s)), &out)
			if e == nil || strings.Contains(out.String(), "event: message_stop") || strings.Contains(out.String(), "event: message_delta") {
				t.Fatalf("err=%v out=%s", e, out.String())
			}
		})
	}
}
func TestStreamBoundsAndCancellationCloseReader(t *testing.T) {
	c := responseCtx(t)
	c.limits.MaxEventBytes = 64
	var out bytes.Buffer
	if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader("data: "+strings.Repeat("x", 100)+"\n\n")), &out); e == nil {
		t.Fatal("line bound")
	}
	r, w := io.Pipe()
	defer func() { _ = w.Close() }() // io.Pipe closes always return nil
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Stream(ctx, r, &out) }()
	cancel()
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel failed to close upstream")
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("client gone") }
func TestStreamWriterFailure(t *testing.T) {
	c := responseCtx(t)
	if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(created())), failWriter{}); e == nil {
		t.Fatal("write error hidden")
	}
}
