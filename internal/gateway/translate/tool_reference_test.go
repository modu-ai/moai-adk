package translate

import (
	"encoding/json"
	"strings"
	"testing"
)

const referenceRequest = `{"max_tokens":32,"messages":[{"role":"user","content":"search"},{"role":"assistant","content":[{"type":"tool_use","id":"call_search","name":"ToolSearch","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_search","content":[{"type":"text","text":"Found tools"},{"type":"tool_reference","tool_name":"web.search"}]}]}],"tools":[{"name":"ToolSearch","input_schema":{"type":"object"}},{"name":"web.search","defer_loading":true,"input_schema":{"type":"object"}},{"name":"unselected","defer_loading":true,"input_schema":{"type":"object"}}]}`

func TestToolReferenceActivatesOnlySelectedSchema(t *testing.T) {
	out, c, err := Request("gpt-5.6-sol", []byte(referenceRequest), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	d := decode(t, out)
	tools := d["tools"].([]any)
	if len(tools) != 2 {
		t.Fatalf("tools: %v", tools)
	}
	alias := tools[1].(map[string]any)["name"].(string)
	if c.names[alias] != "web.search" {
		t.Fatal("lost alias mapping")
	}
	inputs := d["input"].([]any)
	result := inputs[len(inputs)-1].(map[string]any)["output"].([]any)
	if len(result) != 2 || result[0].(map[string]any)["text"] != "Found tools" || !strings.Contains(result[1].(map[string]any)["text"].(string), alias) {
		t.Fatal(result)
	}
}

func TestToolReferenceRejectsUnknownMalformedAndWrongPlacement(t *testing.T) {
	for _, replacement := range []string{`{"type":"tool_reference","tool_name":"missing"}`, `{"type":"tool_reference","tool_name":""}`, `{"type":"tool_reference","tool_name":42}`, `{"type":"tool_reference","tool_name":"web.search","extra":true}`} {
		raw := strings.Replace(referenceRequest, `{"type":"tool_reference","tool_name":"web.search"}`, replacement, 1)
		if _, _, err := Request("gpt-5.6-sol", []byte(raw), Limits{}); err == nil {
			t.Fatal("accepted", replacement)
		}
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(referenceRequest), &root); err != nil {
		t.Fatal(err)
	}
	root["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_reference", "tool_name": "web.search"}}}}
	raw, _ := json.Marshal(root)
	if _, _, err := Request("gpt-5.6-sol", raw, Limits{}); err == nil {
		t.Fatal("accepted misplaced reference")
	}
}
