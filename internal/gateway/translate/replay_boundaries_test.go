package translate

import (
	"encoding/json"
	"testing"
)

func TestRegressionReasoningPublicItemOrderAndPhase(t *testing.T) {
	h := &testHistory{}
	_, c, e := Request("gpt-5.6-sol", []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`), Limits{History: h})
	if e != nil {
		t.Fatal(e)
	}
	first := textItem("msg_c", "commentary")
	first["phase"] = "commentary"
	last := textItem("msg_f", "final")
	last["phase"] = "final_answer"
	rs := func(id string) map[string]any {
		return map[string]any{"type": "reasoning", "id": id, "summary": []any{}, "encrypted_content": "cipher_" + id}
	}
	output, e := c.Response(j(finalResponse("completed", rs("rs_1"), first, rs("rs_2"), last)))
	if e != nil {
		t.Fatal(e)
	}
	result := decode(t, output)
	input := map[string]any{"max_tokens": 100, "messages": []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": result["content"]}, map[string]any{"role": "user", "content": "continue"}}}
	replay, _, e := Request("gpt-5.6-sol", j(input), Limits{History: h})
	if e != nil {
		t.Fatal(e)
	}
	var root map[string]any
	json.Unmarshal(replay, &root)
	rows := root["input"].([]any)
	t.Logf("observed upstream replay: %s", replay)
	if len(rows) != 6 {
		t.Fatalf("public output item boundaries lost: got %d total input items, want 6", len(rows))
	}
	if rows[2].(map[string]any)["phase"] != "commentary" || rows[4].(map[string]any)["phase"] != "final_answer" {
		t.Fatal("public message phase lost")
	}
}
func TestRegressionReasoningAfterTwoMessages(t *testing.T) {
	h := &testHistory{}
	_, c, e := Request("gpt-5.6-sol", []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`), Limits{History: h})
	if e != nil {
		t.Fatal(e)
	}
	rs := map[string]any{"type": "reasoning", "id": "rs_3", "summary": []any{}, "encrypted_content": "cipher"}
	output, e := c.Response(j(finalResponse("completed", textItem("msg_1", "a"), textItem("msg_2", "b"), rs, textItem("msg_3", "c"))))
	if e != nil {
		t.Fatal(e)
	}
	result := decode(t, output)
	input := map[string]any{"max_tokens": 100, "messages": []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": result["content"]}, map[string]any{"role": "user", "content": "continue"}}}
	_, _, e = Request("gpt-5.6-sol", j(input), Limits{History: h})
	if e != nil {
		t.Fatalf("previous successful output cannot be replayed: %v", e)
	}
}

func TestRegressionReasoningMultipartAndNullPhase(t *testing.T) {
	h := &testHistory{}
	_, c, e := Request("gpt-5.6-sol", []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`), Limits{History: h})
	if e != nil {
		t.Fatal(e)
	}
	first := textItem("msg_multi", "one")
	first["phase"] = nil
	first["content"] = append(first["content"].([]any), map[string]any{"type": "output_text", "text": "two", "annotations": []any{}})
	last := textItem("msg_last", "three")
	rs := map[string]any{"type": "reasoning", "id": "rs_null", "summary": []any{}, "encrypted_content": "synthetic"}
	out, e := c.Response(j(finalResponse("completed", first, rs, last)))
	if e != nil {
		t.Fatal(e)
	}
	response := decode(t, out)
	in := map[string]any{"max_tokens": 100, "messages": []any{map[string]any{"role": "user", "content": "hello"}, map[string]any{"role": "assistant", "content": response["content"]}, map[string]any{"role": "user", "content": "continue"}}}
	replay, _, e := Request("gpt-5.6-sol", j(in), Limits{History: h})
	if e != nil {
		t.Fatal(e)
	}
	rows := decode(t, replay)["input"].([]any)
	if len(rows) != 5 {
		t.Fatalf("item count %d", len(rows))
	}
	m := rows[1].(map[string]any)
	if len(m["content"].([]any)) != 2 {
		t.Fatal("multipart message split")
	}
	if p, ok := m["phase"]; !ok || p != nil {
		t.Fatal("explicit null phase lost")
	}
	if _, ok := rows[3].(map[string]any)["phase"]; ok {
		t.Fatal("phase invented")
	}
}
