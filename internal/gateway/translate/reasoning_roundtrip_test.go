package translate

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"io"
	"strings"
	"testing"
)

type testHistory struct {
	published []byte
	checks    int
}

func (h *testHistory) Check(_ context.Context, _, _ string, _ []byte) error { h.checks++; return nil }
func (h *testHistory) Publish(_ context.Context, _, _ string, messages []byte) error {
	h.published = append([]byte(nil), messages...)
	return nil
}
func TestReasoningRoundTripUsesFinalEncryptedItem(t *testing.T) {
	h := &testHistory{}
	_, c, err := Request("gpt-5.6-sol", []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`), Limits{History: h})
	if err != nil {
		t.Fatal(err)
	}
	final := map[string]any{"type": "reasoning", "id": "rs_1", "summary": []any{}, "encrypted_content": "final_ciphertext"}
	events := []string{event("response.created", map[string]any{"response": map[string]any{"id": "resp-1", "model": "gpt-5.6-sol", "status": "in_progress", "output": []any{}}}), event("response.output_item.added", map[string]any{"output_index": 0, "item": map[string]any{"type": "reasoning", "id": "rs_1", "summary": []any{}}}), event("response.output_item.done", map[string]any{"output_index": 0, "item": final})}
	texts := textEvents("completed")
	for _, raw := range texts[1 : len(texts)-1] {
		events = append(events, strings.ReplaceAll(raw, `"output_index":0`, `"output_index":1`))
	}
	events = append(events, event("response.completed", map[string]any{"response": finalResponse("completed", final, textItem("item-t", "hello"))}))
	out, err := c.SubscriptionResponse(context.Background(), io.NopCloser(strings.NewReader(strings.Join(events, ""))))
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	json.Unmarshal(out, &response)
	if len(h.published) == 0 {
		t.Fatal("completed prefix not published")
	}
	request := map[string]any{"max_tokens": 100, "messages": []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": response["content"]}, map[string]any{"role": "user", "content": "again"}}}
	translated, _, err := Request("gpt-5.6-sol", j(request), Limits{History: h})
	if err != nil {
		t.Fatal(err)
	}
	var upstream map[string]any
	json.Unmarshal(translated, &upstream)
	input := upstream["input"].([]any)
	if input[1].(map[string]any)["encrypted_content"] != "final_ciphertext" {
		t.Fatalf("lost final reasoning: %s", translated)
	}
	if h.checks != 2 {
		t.Fatalf("history checks=%d", h.checks)
	}
}

func TestReasoningBufferedStreamUsesNativeDeltas(t *testing.T) {
	h := &testHistory{}
	_, c, err := Request("gpt-5.6-sol", []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`), Limits{History: h})
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err = c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(textEvents("completed"), ""))), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"type":"text_delta"`) {
		t.Fatal("buffered stream lacks text delta")
	}
}

func TestReasoningToolRoundTripAndPublicationFailure(t *testing.T) {
	h := &testHistory{}
	body := []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"use tool"}],"tools":[{"name":"read_file","input_schema":{"type":"object","properties":{}}}]}`)
	_, c, err := Request("gpt-5.6-sol", body, Limits{History: h})
	if err != nil {
		t.Fatal(err)
	}
	reasoning := map[string]any{"type": "reasoning", "id": "rs_tool", "summary": []any{}, "encrypted_content": "cipher"}
	out, err := c.Response(j(finalResponse("completed", reasoning, toolItem("fc_1", "call_1", "read_file", `{"path":"fixture"}`))))
	if err != nil {
		t.Fatal(err)
	}
	response := decode(t, out)
	blocks := response["content"].([]any)
	tool := blocks[1].(map[string]any)
	var root map[string]any
	json.Unmarshal(body, &root)
	root["messages"] = []any{map[string]any{"role": "user", "content": "use tool"}, map[string]any{"role": "assistant", "content": blocks}, map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": tool["id"], "content": "file data"}}}}
	translated, _, err := Request("gpt-5.6-sol", j(root), Limits{History: h})
	if err != nil {
		t.Fatal(err)
	}
	upstream := decode(t, translated)
	input := upstream["input"].([]any)
	if input[2].(map[string]any)["call_id"] != "call_1" || input[3].(map[string]any)["call_id"] != "call_1" {
		t.Fatalf("tool IDs not restored: %s", translated)
	}
	denied := &rejectPublication{}
	_, c, err = Request("gpt-5.6-sol", body, Limits{History: denied})
	if err != nil {
		t.Fatal(err)
	}
	var streamed strings.Builder
	err = c.SubscriptionStream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(textEvents("completed"), ""))), &streamed)
	if err == nil || strings.Contains(streamed.String(), "message_stop") {
		t.Fatal("receipt failure emitted success")
	}
}

type rejectPublication struct{ testHistory }

func (*rejectPublication) Publish(context.Context, string, string, []byte) error {
	return errors.New("fixture receipt write failed")
}

func TestReasoningCarrierPreservesOriginalItemBytes(t *testing.T) {
	_, c, err := Request("gpt-5.6-sol", []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`), Limits{History: &testHistory{}})
	if err != nil {
		t.Fatal(err)
	}
	item := `{ "type":"reasoning", "id":"rs_raw", "summary":[], "encrypted_content":"cipher" }`
	raw := `{"id":"resp-1","model":"gpt-5.6-sol","status":"completed","output":[` + item + `,` + string(j(textItem("msg_1", "hello"))) + `],"usage":{"input_tokens":1,"output_tokens":1}}`
	response, err := c.Response([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	blocks := decode(t, response)["content"].([]any)
	envelope, err := opaque.Decode(blocks[0].(map[string]any)["data"].(string))
	if err != nil {
		t.Fatal(err)
	}
	if string(envelope.Items()[0].Raw) != item {
		t.Fatal("opaque item bytes were reserialized")
	}
}

func TestReasoningWithoutPublicCompletionFailsClosed(t *testing.T) {
	h := &testHistory{}
	_, c, err := Request("gpt-5.6-sol", []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`), Limits{History: h})
	if err != nil {
		t.Fatal(err)
	}
	raw := j(finalResponse("completed", map[string]any{"type": "reasoning", "id": "rs_only", "summary": []any{}, "encrypted_content": "cipher"}))
	if _, err = c.Response(raw); err == nil {
		t.Fatal("reasoning-only output reported success")
	}
	if h.published != nil {
		t.Fatal("reasoning-only receipt published")
	}
}
