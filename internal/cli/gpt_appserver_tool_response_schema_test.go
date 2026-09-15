package cli

import (
	"encoding/json"
	"testing"
)

func TestToolSearchReferenceOnlyReplyKeepsRequiredNativeTextField(t *testing.T) {
	content, refs, err := managedGPTToolContent(json.RawMessage(`[{"type":"tool_reference","tool_name":"Write"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].ToolName != "Write" {
		t.Fatal("reference authority lost")
	}
	raw, err := json.Marshal(map[string]any{"contentItems": content, "success": true})
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		ContentItems []map[string]json.RawMessage `json:"contentItems"`
	}
	if err = json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.ContentItems) != 1 {
		t.Fatal(string(raw))
	}
	text, ok := response.ContentItems[0]["text"]
	if !ok || string(text) != `""` {
		t.Fatalf("App Server inputText schema requires text even when empty: %s", raw)
	}
}
