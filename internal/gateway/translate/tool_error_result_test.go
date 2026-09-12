package translate

import (
	"encoding/json"
	"testing"
)

// A failed Bash/MCP operation is a valid tool result, not a malformed request.
func TestRequestErrorToolResultPreservesFailureAndContinuation(t *testing.T) {
	for _, content := range []any{"Exit code 5: synthetic failure", []any{map[string]any{"type": "text", "text": "Exit code 5: synthetic failure"}}, ""} {
		d := decode(t, []byte(ordinary))
		messages := d["messages"].([]any)
		result := messages[3].(map[string]any)["content"].([]any)[0].(map[string]any)
		result["is_error"] = true
		result["content"] = content
		messages = append(messages, map[string]any{"role": "user", "content": "Continue after the failure"})
		d["messages"] = messages
		body, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _, err := Request("gpt-5.6-sol", body, Limits{})
		if err != nil {
			t.Fatalf("valid failed tool result rejected: %v", err)
		}
		output := decode(t, encoded)
		found := false
		for _, v := range output["input"].([]any) {
			item := v.(map[string]any)
			if item["type"] != "function_call_output" {
				continue
			}
			found = true
			blocks := item["output"].([]any)
			if item["call_id"] != "call-A" || len(blocks) == 0 || blocks[0].(map[string]any)["text"] != "Tool execution failed (is_error=true)." {
				t.Fatalf("failure status lost: %#v", item)
			}
			if text, ok := content.(string); !ok || text != "" {
				if len(blocks) != 2 || blocks[1].(map[string]any)["text"] != "Exit code 5: synthetic failure" {
					t.Fatalf("failure detail lost: %#v", item)
				}
			}
		}
		if !found {
			t.Fatal("missing paired result")
		}
	}
}
