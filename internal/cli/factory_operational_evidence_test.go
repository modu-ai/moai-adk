//go:build darwin || linux

package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFactoryOperationalEvidenceRejectsUnownedResult(t *testing.T) {
	call := `{"type":"response_item","payload":{"type":"function_call","name":"mcp__moai__factory_msg_status","arguments":"{\"run_id\":\"r\"}","call_id":"c"}}` + "\n"
	result := `{"type":"response_item","payload":{"type":"function_call_output","call_id":"c","output":"{\"lanes\":[{\"slot\":\"lead\"}]}"}}`
	emptyCall := strings.ReplaceAll(call, `"call_id":"c"`, `"call_id":""`)
	emptyResult := strings.ReplaceAll(result, `"call_id":"c"`, `"call_id":""`)
	for _, tc := range []struct {
		data, run string
		want      bool
	}{
		{call + result, "r", true}, {result, "r", false}, {call + result, "foreign", false},
		{emptyCall + emptyResult, "r", false},
		{emptyCall + result, "r", false},
		{call + emptyResult, "r", false},
		{strings.ReplaceAll(call, "mcp__moai__", "mcp__foreign__") + result, "r", false},
		{call + `{"type":"response_item","payload":{"type":"function_call_output","call_id":"other","output":"{\"lanes\":[]}"}}`, "r", false},
	} {
		_, ok := operationalMCPResult([]byte(tc.data), tc.run)
		if ok != tc.want {
			t.Fatalf("got %v want %v: %s", ok, tc.want, tc.data)
		}
	}
	if _, ok := operationalMCPResultCount([]byte(call+result+"\n"+result), "r", 2); ok {
		t.Fatal("duplicate old output counted as a new status call")
	}
	if _, ok := operationalOutputLanes([]byte(`{"isError":true,"lanes":[]}`)); ok {
		t.Fatal("error response accepted as evidence")
	}
}

func TestOperationalAssistantOutputRequiresOwnedExactText(t *testing.T) {
	assistant := `{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"BOUND_STABLE_1"}]}}`
	user := `{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"Reply with exactly BOUND_STABLE_1"}]}}`
	for _, tc := range []struct {
		name string
		data string
		want bool
	}{
		{name: "assistant exact", data: assistant, want: true},
		{name: "user echo", data: user, want: false},
		{name: "assistant contains", data: strings.Replace(assistant, `BOUND_STABLE_1`, `prefix BOUND_STABLE_1`, 1), want: false},
		{name: "non output content", data: strings.Replace(assistant, `output_text`, `input_text`, 1), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := operationalAssistantOutput([]byte(tc.data), "BOUND_STABLE_1"); got != tc.want {
				t.Fatalf("got %v want %v: %s", got, tc.want, tc.data)
			}
		})
	}
}

func TestFactoryOperationalEvidenceAcceptsOwnedCurrentCodexMCPShape(t *testing.T) {
	line := func(v any) string {
		t.Helper()
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(data) + "\n"
	}
	structuredOutput, err := json.Marshal(map[string]any{
		"first":  map[string]any{"structuredContent": map[string]any{"lanes": []map[string]any{{"slot": "lead"}}}},
		"second": map[string]any{"structuredContent": map[string]any{"lanes": []map[string]any{{"slot": "lead"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	current := func(turn, callID, server, tool, run, status, outputID string, includeOutput bool) []byte {
		call := map[string]any{"type": "response_item", "payload": map[string]any{
			"type": "custom_tool_call", "name": "exec", "call_id": callID,
			"internal_chat_message_metadata_passthrough": map[string]any{"turn_id": turn},
		}}
		completed := func(id string) map[string]any {
			return map[string]any{"type": "event_msg", "payload": map[string]any{
				"type": "item_completed", "turn_id": turn, "item": map[string]any{
					"id": id, "type": "McpToolCall", "server": server, "tool": tool,
					"arguments": map[string]any{"run_id": run}, "status": status,
				},
			}}
		}
		rollout := line(call) + line(completed(turn+"-exec-1")) + line(completed(turn+"-exec-2"))
		if includeOutput {
			blocks := []map[string]any{
				{"type": "input_text", "text": "Script completed\nWall time: 0.5 seconds\nOutput:\n"},
				{"type": "input_text", "text": string(structuredOutput)},
			}
			rollout += line(map[string]any{"type": "response_item", "payload": map[string]any{
				"type": "custom_tool_call_output", "call_id": outputID, "output": blocks,
				"internal_chat_message_metadata_passthrough": map[string]any{"turn_id": turn},
			}})
		}
		return []byte(rollout)
	}
	withOutputValue := func(data []byte, output any) []byte {
		t.Helper()
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		var event map[string]any
		if len(lines) == 0 || json.Unmarshal([]byte(lines[len(lines)-1]), &event) != nil {
			t.Fatal("fixture output event absent")
		}
		payload, ok := event["payload"].(map[string]any)
		if !ok {
			t.Fatal("fixture output payload absent")
		}
		payload["output"] = output
		lines[len(lines)-1] = strings.TrimSpace(line(event))
		return []byte(strings.Join(lines, "\n") + "\n")
	}

	for _, tc := range []struct {
		name                                    string
		turn, callID, server, tool, run, status string
		outputID                                string
		includeOutput, want                     bool
	}{
		{name: "owned completed calls", turn: "turn-1", callID: "call-1", server: "moai", tool: "factory_msg_status", run: "r", status: "completed", outputID: "call-1", includeOutput: true, want: true},
		{name: "empty turn", callID: "call-1", server: "moai", tool: "factory_msg_status", run: "r", status: "completed", outputID: "call-1", includeOutput: true},
		{name: "empty call id", turn: "turn-1", server: "moai", tool: "factory_msg_status", run: "r", status: "completed", includeOutput: true},
		{name: "empty output id", turn: "turn-1", callID: "call-1", server: "moai", tool: "factory_msg_status", run: "r", status: "completed", includeOutput: true},
		{name: "wrong output id", turn: "turn-1", callID: "call-1", server: "moai", tool: "factory_msg_status", run: "r", status: "completed", outputID: "other", includeOutput: true},
		{name: "wrong server", turn: "turn-1", callID: "call-1", server: "foreign", tool: "factory_msg_status", run: "r", status: "completed", outputID: "call-1", includeOutput: true},
		{name: "wrong tool", turn: "turn-1", callID: "call-1", server: "moai", tool: "factory_msg_send", run: "r", status: "completed", outputID: "call-1", includeOutput: true},
		{name: "wrong run", turn: "turn-1", callID: "call-1", server: "moai", tool: "factory_msg_status", run: "foreign", status: "completed", outputID: "call-1", includeOutput: true},
		{name: "failed status", turn: "turn-1", callID: "call-1", server: "moai", tool: "factory_msg_status", run: "r", status: "failed", outputID: "call-1", includeOutput: true},
		{name: "missing output", turn: "turn-1", callID: "call-1", server: "moai", tool: "factory_msg_status", run: "r", status: "completed", outputID: "call-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := operationalMCPResultCount(current(tc.turn, tc.callID, tc.server, tc.tool, tc.run, tc.status, tc.outputID, tc.includeOutput), "r", 2)
			if ok != tc.want {
				t.Fatalf("got %v want %v", ok, tc.want)
			}
		})
	}

	t.Run("output from another turn", func(t *testing.T) {
		data := string(current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true))
		idx := strings.LastIndex(data, `"turn_id":"turn-1"`)
		if idx < 0 {
			t.Fatal("fixture output turn absent")
		}
		data = data[:idx] + `"turn_id":"turn-2"` + data[idx+len(`"turn_id":"turn-1"`):]
		if _, ok := operationalMCPResultCount([]byte(data), "r", 2); ok {
			t.Fatal("output from another turn accepted")
		}
	})

	t.Run("non-exec custom call", func(t *testing.T) {
		data := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
		data = []byte(strings.Replace(string(data), `"name":"exec"`, `"name":"other"`, 1))
		if _, ok := operationalMCPResultCount(data, "r", 2); ok {
			t.Fatal("non-exec custom call accepted")
		}
	})

	t.Run("JSON without Output marker", func(t *testing.T) {
		data := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
		data = withOutputValue(data, []map[string]any{
			{"type": "input_text", "text": "Script completed\nWall time: 0.5 seconds\nResult:\n"},
			{"type": "input_text", "text": string(structuredOutput)},
		})
		if _, ok := operationalMCPResultCount(data, "r", 2); ok {
			t.Fatal("custom output JSON without the exact Output marker accepted")
		}
	})

	for _, tc := range []struct {
		name   string
		blocks []map[string]any
	}{
		{name: "marker only", blocks: []map[string]any{{"type": "input_text", "text": "Script completed\nOutput:\n"}}},
		{name: "extra non-empty block", blocks: []map[string]any{
			{"type": "input_text", "text": "Script completed\nOutput:\n"},
			{"type": "input_text", "text": string(structuredOutput)},
			{"type": "input_text", "text": `{"unexpected":true}`},
		}},
		{name: "non-text payload", blocks: []map[string]any{
			{"type": "input_text", "text": "Script completed\nOutput:\n"},
			{"type": "image", "text": string(structuredOutput)},
		}},
		{name: "invalid JSON payload", blocks: []map[string]any{
			{"type": "input_text", "text": "Script completed\nOutput:\n"},
			{"type": "input_text", "text": "not-json"},
		}},
		{name: "wrong wrapper", blocks: []map[string]any{
			{"type": "input_text", "text": "Script completed\nOutput:\n"},
			{"type": "input_text", "text": `{"lanes":[{"slot":"lead"}]}`},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
			if _, ok := operationalMCPResultCount(withOutputValue(data, tc.blocks), "r", 2); ok {
				t.Fatalf("accepted %s", tc.name)
			}
		})
	}

	for _, tc := range []struct {
		name   string
		output any
	}{
		{name: "scalar output", output: string(structuredOutput)},
		{name: "object output", output: map[string]any{"first": map[string]any{"structuredContent": map[string]any{"lanes": []map[string]any{{"slot": "lead"}}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
			if _, ok := operationalMCPResultCount(withOutputValue(data, tc.output), "r", 2); ok {
				t.Fatalf("accepted %s", tc.name)
			}
		})
	}

	for _, tc := range []struct {
		name, old, replacement string
	}{
		{name: "duplicate completed item id", old: `"id":"turn-1-exec-2"`, replacement: `"id":"turn-1-exec-1"`},
		{name: "empty completed item id", old: `"id":"turn-1-exec-2"`, replacement: `"id":""`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
			data = []byte(strings.Replace(string(data), tc.old, tc.replacement, 1))
			if _, ok := operationalMCPResultCount(data, "r", 2); ok {
				t.Fatalf("accepted %s", tc.name)
			}
		})
	}

	t.Run("cumulative calls across distinct turns", func(t *testing.T) {
		turn1 := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
		turn2 := current("turn-2", "call-2", "moai", "factory_msg_status", "r", "completed", "call-2", true)
		latestOutput, err := json.Marshal(map[string]any{
			"first":  map[string]any{"structuredContent": map[string]any{"lanes": []map[string]any{{"slot": "agent-2"}}}},
			"second": map[string]any{"structuredContent": map[string]any{"lanes": []map[string]any{{"slot": "agent-2"}}}},
		})
		if err != nil {
			t.Fatal(err)
		}
		turn2 = withOutputValue(turn2, []map[string]any{
			{"type": "input_text", "text": "Script completed\nOutput:\n"},
			{"type": "input_text", "text": string(latestOutput)},
		})
		lanes, ok := operationalMCPResultCount(append(turn1, turn2...), "r", 4)
		if !ok || len(lanes) != 1 || lanes[0].Slot != "agent-2" {
			t.Fatalf("cumulative result ok=%v lanes=%+v", ok, lanes)
		}
	})

	t.Run("two plus one is below four", func(t *testing.T) {
		turn1 := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
		turn2 := current("turn-2", "call-2", "moai", "factory_msg_status", "r", "completed", "call-2", true)
		lines := strings.Split(strings.TrimSpace(string(turn2)), "\n")
		kept := lines[:0]
		for _, line := range lines {
			if !strings.Contains(line, `"id":"turn-2-exec-2"`) {
				kept = append(kept, line)
			}
		}
		turn2 = []byte(strings.Join(kept, "\n") + "\n")
		oneOutput, err := json.Marshal(map[string]any{
			"first": map[string]any{"structuredContent": map[string]any{"lanes": []map[string]any{{"slot": "agent-2"}}}},
		})
		if err != nil {
			t.Fatal(err)
		}
		turn2 = withOutputValue(turn2, []map[string]any{
			{"type": "input_text", "text": "Script completed\nOutput:\n"},
			{"type": "input_text", "text": string(oneOutput)},
		})
		if _, ok := operationalMCPResultCount(append(turn1, turn2...), "r", 4); ok {
			t.Fatal("2+1 calls satisfied minimum 4")
		}
	})

	t.Run("repeated same turn and IDs do not inflate", func(t *testing.T) {
		turn := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
		if _, ok := operationalMCPResultCount(append(turn, turn...), "r", 4); ok {
			t.Fatal("repeated same-turn evidence satisfied minimum 4")
		}
	})

	t.Run("completed item IDs are unique across turns", func(t *testing.T) {
		turn1 := current("turn-1", "call-1", "moai", "factory_msg_status", "r", "completed", "call-1", true)
		turn2 := current("turn-2", "call-2", "moai", "factory_msg_status", "r", "completed", "call-2", true)
		turn2 = []byte(strings.Replace(string(turn2), `"id":"turn-2-exec-1"`, `"id":"turn-1-exec-1"`, 1))
		if _, ok := operationalMCPResultCount(append(turn1, turn2...), "r", 4); ok {
			t.Fatal("cross-turn duplicate completed item ID inflated count")
		}
	})
}
