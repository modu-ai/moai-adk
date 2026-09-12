package translate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func orderedStreamTexts(t *testing.T, raw string) []string {
	t.Helper()
	texts := map[int]string{}
	for _, line := range strings.Split(raw, "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		v := decode(t, []byte(strings.TrimPrefix(line, "data: ")))
		if v["type"] != "content_block_delta" {
			continue
		}
		d := v["delta"].(map[string]any)
		if d["type"] == "text_delta" {
			texts[int(v["index"].(float64))] += d["text"].(string)
		}
	}
	result := make([]string, len(texts))
	for i := range result {
		result[i] = texts[i]
	}
	return result
}
func TestStreamCanonicalOutputOrderRegression(t *testing.T) {
	raw, e := os.ReadFile("testdata/interleaved-text.sse")
	if e != nil {
		t.Fatal(e)
	}
	terminal, e := os.ReadFile("testdata/interleaved-text-terminal.json")
	if e != nil {
		t.Fatal(e)
	}
	c := responseCtx(t)
	complete, e := c.Response(terminal)
	if e != nil {
		t.Fatal(e)
	}
	want := []string{}
	for _, block := range decode(t, complete)["content"].([]any) {
		want = append(want, block.(map[string]any)["text"].(string))
	}
	var out bytes.Buffer
	if e = c.Stream(context.Background(), io.NopCloser(bytes.NewReader(raw)), &out); e != nil {
		t.Fatalf("valid interleaving rejected: %v", e)
	}
	got := orderedStreamTexts(t, out.String())
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stream=%v nonstream=%v", got, want)
	}
}

// boundaryReader does not release the next upstream chunk until prior output is
// observed. This fails a converter that buffers all output until terminal.
type boundaryReader struct {
	first, rest *strings.Reader
	check       func() error
	checked     bool
}

func (r *boundaryReader) Read(p []byte) (int, error) {
	if r.first.Len() > 0 {
		return r.first.Read(p)
	}
	if !r.checked {
		r.checked = true
		if e := r.check(); e != nil {
			return 0, e
		}
	}
	return r.rest.Read(p)
}
func TestStreamReadyPrefixDoesNotWaitForTerminal(t *testing.T) {
	events := textEvents("completed")
	var out bytes.Buffer
	r := &boundaryReader{first: strings.NewReader(strings.Join(events[:4], "")), rest: strings.NewReader(strings.Join(events[4:], "")), check: func() error {
		if !strings.Contains(out.String(), `"text":"hello"`) {
			return fmt.Errorf("ready prefix withheld")
		}
		if strings.Contains(out.String(), "message_stop") {
			return fmt.Errorf("premature terminal")
		}
		return nil
	}}
	if e := responseCtx(t).Stream(context.Background(), io.NopCloser(r), &out); e != nil {
		t.Fatal(e)
	}
}
func TestStreamBufferedSuccessorFlushesBeforeTerminal(t *testing.T) {
	raw, e := os.ReadFile("testdata/interleaved-text.sse")
	if e != nil {
		t.Fatal(e)
	}
	at := strings.LastIndex(string(raw), "event: response.completed")
	if at < 0 {
		t.Fatal("fixture terminal missing")
	}
	var out bytes.Buffer
	r := &boundaryReader{first: strings.NewReader(string(raw[:at])), rest: strings.NewReader(string(raw[at:])), check: func() error {
		if got := orderedStreamTexts(t, out.String()); !reflect.DeepEqual(got, []string{"first", "second"}) {
			return fmt.Errorf("successor not released after predecessor width known: %v", got)
		}
		return nil
	}}
	if e = responseCtx(t).Stream(context.Background(), io.NopCloser(r), &out); e != nil {
		t.Fatal(e)
	}
}

func TestStreamCanonicalOffsetIncludesEveryPrecedingContentPart(t *testing.T) {
	c := responseCtx(t)
	name := alias(c, "a.b")
	parts := []string{created(), event("response.output_item.added", map[string]any{"output_index": 0, "item": map[string]any{"id": "text-item", "type": "message", "role": "assistant", "status": "in_progress", "content": []any{}}}), event("response.output_item.added", map[string]any{"output_index": 1, "item": map[string]any{"id": "tool-item", "type": "function_call", "status": "in_progress", "name": name, "call_id": "call-original", "arguments": ""}}), event("response.function_call_arguments.delta", map[string]any{"output_index": 1, "item_id": "tool-item", "delta": "{}"}), event("response.function_call_arguments.done", map[string]any{"output_index": 1, "item_id": "tool-item", "arguments": "{}"}), event("response.output_item.done", map[string]any{"output_index": 1, "item": toolItem("tool-item", "call-original", name, "{}")})}
	for i := 0; i < 2; i++ {
		parts = append(parts, event("response.content_part.added", map[string]any{"output_index": 0, "item_id": "text-item", "content_index": i, "part": map[string]any{"type": "output_text", "text": ""}}))
	}
	texts := []string{"first", "second"}
	for _, i := range []int{1, 0} {
		base := func() map[string]any {
			return map[string]any{"output_index": 0, "item_id": "text-item", "content_index": i}
		}
		m := base()
		m["delta"] = texts[i]
		parts = append(parts, event("response.output_text.delta", m))
		m = base()
		m["text"] = texts[i]
		parts = append(parts, event("response.output_text.done", m))
		m = base()
		m["part"] = map[string]any{"type": "output_text", "text": texts[i]}
		parts = append(parts, event("response.content_part.done", m))
	}
	item := map[string]any{"id": "text-item", "type": "message", "role": "assistant", "status": "completed", "content": []any{map[string]any{"type": "output_text", "text": "first"}, map[string]any{"type": "output_text", "text": "second"}}}
	parts = append(parts, event("response.output_item.done", map[string]any{"output_index": 0, "item": item}), event("response.completed", map[string]any{"response": finalResponse("completed", item, toolItem("tool-item", "call-original", name, "{}"))}))
	var out bytes.Buffer
	if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(parts, ""))), &out); e != nil {
		t.Fatal(e)
	}
	if got := orderedStreamTexts(t, out.String()); !reflect.DeepEqual(got, texts) {
		t.Fatal(got)
	}
	starts := map[int]string{}
	for _, line := range strings.Split(out.String(), "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		v := decode(t, []byte(strings.TrimPrefix(line, "data: ")))
		if v["type"] == "content_block_start" {
			block := v["content_block"].(map[string]any)
			starts[int(v["index"].(float64))] = block["type"].(string)
			if block["type"] == "tool_use" && (block["id"] != "call-original" || block["name"] != "a.b") {
				t.Fatal(block)
			}
		}
	}
	if !reflect.DeepEqual(starts, map[int]string{0: "text", 1: "text", 2: "tool_use"}) {
		t.Fatal(starts)
	}
}
func TestStreamBufferedOrderingErrorsNeverFinish(t *testing.T) {
	raw, e := os.ReadFile("testdata/interleaved-text.sse")
	if e != nil {
		t.Fatal(e)
	}
	parts := strings.SplitAfter(string(raw), "\n\n")
	for i := 0; i < len(parts)-1; i++ {
		var out bytes.Buffer
		e := responseCtx(t).Stream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(parts[:i], ""))), &out)
		if e == nil || strings.Contains(out.String(), "message_stop") {
			t.Fatalf("truncated prefix %d: %v", i, e)
		}
	}
}

type rejectDeferredWriter struct{ bytes.Buffer }

func (w *rejectDeferredWriter) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte(`"text":"second"`)) {
		return 0, fmt.Errorf("deferred client write failed")
	}
	return w.Buffer.Write(p)
}
func TestStreamDeferredWriterFailureHasNoTerminal(t *testing.T) {
	raw, e := os.ReadFile("testdata/interleaved-text.sse")
	if e != nil {
		t.Fatal(e)
	}
	var out rejectDeferredWriter
	e = responseCtx(t).Stream(context.Background(), io.NopCloser(bytes.NewReader(raw)), &out)
	if e == nil || strings.Contains(out.String(), "message_stop") || strings.Contains(out.String(), "message_delta") {
		t.Fatal(e, out.String())
	}
}
