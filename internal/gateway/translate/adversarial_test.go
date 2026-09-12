package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestRequestFieldBoundaries(t *testing.T) {
	base := `{"max_tokens":8,"messages":[{"role":"user","content":"x"}],"tools":[{"name":"t","input_schema":{"type":"object"}}]}`
	cases := []struct {
		key   string
		value any
	}{{"unknown", 1}, {"stream", nil}, {"temperature", "hot"}, {"top_p", true}, {"stop_sequences", []any{"END"}}, {"metadata", nil}, {"metadata", map[string]any{"secret": true}}, {"system", []any{map[string]any{"type": "image"}}}, {"messages", nil}, {"messages", []any{nil}}, {"messages", []any{map[string]any{"role": "alien", "content": "x"}}}, {"messages", []any{map[string]any{"role": "user", "content": nil}}}, {"messages", []any{map[string]any{"role": "user", "content": []any{nil}}}}, {"messages", []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": 12}}}}}, {"tools", nil}, {"tools", []any{nil}}, {"tools", []any{map[string]any{"name": "t", "input_schema": nil}}}, {"tools", []any{map[string]any{"name": "t", "input_schema": map[string]any{}, "description": 4}}}, {"tools", []any{map[string]any{"name": "t", "input_schema": map[string]any{}, "strict": true}}}, {"tool_choice", nil}, {"tool_choice", map[string]any{"type": "bad"}}, {"tool_choice", map[string]any{"type": "tool", "name": "missing"}}, {"tool_choice", map[string]any{"type": "auto", "disable_parallel_tool_use": nil}}, {"tool_choice", map[string]any{"type": "auto", "unknown": 1}}, {"tool_choice", map[string]any{}}}
	for i, tc := range cases {
		d := decode(t, []byte(base))
		d[tc.key] = tc.value
		if b, _, e := Request("gpt-5.6-sol", j(d), Limits{}); e == nil {
			t.Fatalf("case %d accepted %s", i, b)
		}
	}
	for _, typ := range []string{"auto", "none", "any", "tool"} {
		d := decode(t, []byte(base))
		choice := map[string]any{"type": typ, "disable_parallel_tool_use": true}
		if typ == "tool" {
			choice["name"] = "t"
		}
		d["tool_choice"] = choice
		d["metadata"] = map[string]any{"x": "y"}
		d["temperature"] = 0.2
		d["top_p"] = 0.8
		n := int64(2)
		b, _, e := Request("gpt-5.6-sol", j(d), Limits{ContextTokens: 20, InputTokens: &n})
		if e != nil {
			t.Fatal(e)
		}
		r := decode(t, b)
		if r["parallel_tool_calls"] != false || r["temperature"] != 0.2 {
			t.Fatal(r)
		}
	}
	for _, hint := range []any{nil, map[string]any{"type": "persistent"}, map[string]any{"type": "ephemeral", "ttl": "2h"}, map[string]any{"type": "ephemeral", "unknown": true}} {
		d := decode(t, []byte(base))
		d["system"] = []any{map[string]any{"type": "text", "text": "s", "cache_control": hint}}
		if _, _, e := Request("m", j(d), Limits{}); e == nil {
			t.Fatal("invalid cache hint accepted")
		}
	}
	for _, body := range []string{`{`, `{"x":`, `[]`, `{} {}`, strings.Repeat(`{"x":`, 66) + `1` + strings.Repeat("}", 66), `{"max_tokens":1,"messages":[{"role":"user","content":"x","extra":1}]}`} {
		if _, _, e := Request("m", []byte(body), Limits{}); e == nil {
			t.Fatal("invalid JSON or field accepted")
		}
	}
}
func TestRequestToolPairAdversaries(t *testing.T) {
	for _, which := range []string{"duplicate call", "duplicate result", "wrong call role", "wrong result role", "input array", "error result", "nonboolean error", "result image", "extra call", "extra result", "empty id", "empty name", "duplicate definition"} {
		d := decode(t, []byte(ordinary))
		ms := d["messages"].([]any)
		assistant := ms[2].(map[string]any)
		user := ms[3].(map[string]any)
		call := assistant["content"].([]any)[1].(map[string]any)
		result := user["content"].([]any)[0].(map[string]any)
		switch which {
		case "duplicate call":
			assistant["content"] = append(assistant["content"].([]any), call)
		case "duplicate result":
			user["content"] = append(user["content"].([]any), result)
		case "wrong call role":
			assistant["role"] = "user"
		case "wrong result role":
			user["role"] = "assistant"
		case "input array":
			call["input"] = []any{}
		case "error result":
			result["is_error"] = true
		case "nonboolean error":
			result["is_error"] = "false"
		case "result image":
			result["content"] = []any{map[string]any{"type": "image"}}
		case "extra call":
			call["extra"] = 1
		case "extra result":
			result["extra"] = 1
		case "empty id":
			call["id"] = ""
		case "empty name":
			call["name"] = ""
		case "duplicate definition":
			d["tools"] = append(d["tools"].([]any), d["tools"].([]any)[0])
		}
		if _, _, e := Request("m", j(d), Limits{}); e == nil {
			t.Fatal(which)
		}
	}
}
func TestResponseFieldAdversaries(t *testing.T) {
	c := responseCtx(t)
	for _, kind := range []string{"id", "model", "usage absent", "negative usage", "string usage", "message role", "content nil", "part null", "annotation", "logprobs", "text number", "refusal", "unknown part", "unknown item", "item null", "call duplicate", "call id", "call arguments", "call name", "error completed", "incomplete conflict", "size", "invalid JSON"} {
		r := finalResponse("completed", textItem("m", "x"))
		item := r["output"].([]any)[0].(map[string]any)
		p := item["content"].([]any)[0].(map[string]any)
		switch kind {
		case "id":
			r["id"] = nil
		case "model":
			r["model"] = "wrong"
		case "usage absent":
			delete(r, "usage")
		case "negative usage":
			r["usage"].(map[string]any)["output_tokens"] = -1
		case "string usage":
			r["usage"].(map[string]any)["input_tokens"] = "1"
		case "message role":
			item["role"] = "user"
		case "content nil":
			item["content"] = nil
		case "part null":
			item["content"] = []any{nil}
		case "annotation":
			p["annotations"] = []any{map[string]any{"type": "url_citation"}}
		case "logprobs":
			p["logprobs"] = nil
		case "text number":
			p["text"] = 3
		case "refusal":
			p["type"] = "refusal"
		case "unknown part":
			p["extra"] = "lost"
		case "unknown item":
			item["extra"] = "lost"
		case "item null":
			r["output"] = []any{nil}
		case "call duplicate":
			r["output"] = []any{toolItem("i", "c", alias(c, "a.b"), "{}"), toolItem("j", "c", alias(c, "a.b"), "{}")}
		case "call id":
			r["output"] = []any{toolItem("i", "", alias(c, "a.b"), "{}")}
		case "call arguments":
			z := toolItem("i", "c", alias(c, "a.b"), "{}")
			z["arguments"] = nil
			r["output"] = []any{z}
		case "call name":
			z := toolItem("i", "c", alias(c, "a.b"), "{}")
			z["name"] = nil
			r["output"] = []any{z}
		case "error completed":
			r["error"] = map[string]any{"code": "oops"}
		case "incomplete conflict":
			r["incomplete_details"] = map[string]any{"reason": "max_output_tokens"}
		}
		raw := j(r)
		if kind == "size" {
			c.limits.MaxOutputBytes = 1
		} else {
			c.limits.MaxOutputBytes = 0
		}
		if kind == "invalid JSON" {
			raw = []byte("bad")
		}
		if _, e := c.Response(raw); e == nil {
			t.Fatal(kind)
		}
	}
}
func changeEvent(t *testing.T, e string, mutate func(map[string]any)) string {
	t.Helper()
	line := strings.Split(e, "\n")[1]
	m := decode(t, []byte(strings.TrimPrefix(line, "data: ")))
	mutate(m)
	return event(m["type"].(string), m)
}
func TestStreamEventAdversaries(t *testing.T) {
	c := responseCtx(t)
	base := textEvents("completed")
	tests := []struct {
		at     int
		mutate func(map[string]any)
	}{
		{0, func(m map[string]any) { m["response"] = nil }}, {0, func(m map[string]any) { m["response"].(map[string]any)["model"] = "wrong" }}, {0, func(m map[string]any) { m["response"].(map[string]any)["output"] = []any{1} }},
		{1, func(m map[string]any) { m["output_index"] = -1 }}, {1, func(m map[string]any) { m["output_index"] = 0.5 }}, {1, func(m map[string]any) { m["item"] = nil }}, {1, func(m map[string]any) { m["item"].(map[string]any)["id"] = "" }}, {1, func(m map[string]any) { m["item"].(map[string]any)["type"] = "reasoning" }}, {1, func(m map[string]any) { m["item"].(map[string]any)["status"] = "completed" }}, {1, func(m map[string]any) { m["item"].(map[string]any)["role"] = "user" }}, {1, func(m map[string]any) { m["item"].(map[string]any)["extra"] = 1 }},
		{2, func(m map[string]any) { m["content_index"] = 1 }}, {2, func(m map[string]any) { m["content_index"] = "0" }}, {2, func(m map[string]any) { m["item_id"] = "wrong" }}, {2, func(m map[string]any) { m["part"] = nil }}, {2, func(m map[string]any) { m["part"].(map[string]any)["text"] = "nonempty" }},
		{3, func(m map[string]any) { m["content_index"] = 1 }}, {3, func(m map[string]any) { m["delta"] = nil }}, {4, func(m map[string]any) { m["text"] = "mismatch" }}, {5, func(m map[string]any) { m["part"] = nil }}, {5, func(m map[string]any) { m["part"].(map[string]any)["text"] = "mismatch" }}, {6, func(m map[string]any) { m["item"] = nil }}, {6, func(m map[string]any) { m["item"].(map[string]any)["id"] = "wrong" }}, {6, func(m map[string]any) { m["item"].(map[string]any)["content"] = []any{} }}, {7, func(m map[string]any) { m["response"] = nil }}, {7, func(m map[string]any) { m["response"].(map[string]any)["id"] = "wrong" }}, {7, func(m map[string]any) { m["response"].(map[string]any)["usage"] = nil }},
	}
	for i, tc := range tests {
		events := append([]string{}, base...)
		events[tc.at] = changeEvent(t, events[tc.at], tc.mutate)
		var out bytes.Buffer
		if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(strings.Join(events, ""))), &out); e == nil || strings.Contains(out.String(), "event: message_stop") {
			t.Fatalf("accepted mutation %d", i)
		}
	}
	for _, s := range []string{base[1], base[0] + base[0], base[0] + "id: resume\n", base[0] + "unknown: x\n", base[0] + "event: a\nevent: b\n", strings.Join(base[:5], "") + base[3], strings.Join(base[:3], "") + base[5], strings.Join(base[:7], "") + base[6]} {
		var out bytes.Buffer
		if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(s)), &out); e == nil {
			t.Fatal("accepted state violation")
		}
	}
	// SSE comments and an in-progress heartbeat do not create content.
	heartbeat := event("response.in_progress", map[string]any{"response": map[string]any{"id": "resp-1", "status": "in_progress"}})
	var out bytes.Buffer
	s := base[0] + ": keep alive\n\n" + heartbeat + strings.Join(base[1:], "")
	if e := c.Stream(context.Background(), io.NopCloser(strings.NewReader(s)), &out); e != nil {
		t.Fatal(e)
	}
}
func TestJSONNumbersRemainExact(t *testing.T) {
	d, _ := request(t, strings.ReplaceAll(ordinary, `"x":1`, `"x":9007199254740993`))
	arg := d["input"].([]any)[3].(map[string]any)["arguments"].(string)
	var m map[string]json.Number
	if e := json.Unmarshal([]byte(arg), &m); e != nil || m["x"] != "9007199254740993" {
		t.Fatal(arg, e)
	}
}
