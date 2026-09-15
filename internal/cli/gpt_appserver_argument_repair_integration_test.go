package cli

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProductionGatewayCorrectsInvalidAgentAfterStreamingText(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("local shared peer requires Unix socket")
	}
	home, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	binDir, _ := installProductionWiringFakeCodex(t)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	peer := startClaudeGatewayPeer(t, home, true)
	payload, err := marshalGatewayPrivatePayload("argument-repair-token", gatewayGPTModels())
	if err != nil {
		t.Fatal(err)
	}
	handler, err := productionGatewayHandlerFactory(payload)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := handler.(io.Closer).Close(); err != nil {
			t.Error(err)
		}
	})
	tools := []any{map[string]any{"name": "Agent", "input_schema": map[string]any{"type": "object", "properties": map[string]any{"description": map[string]string{"type": "string"}, "prompt": map[string]string{"type": "string"}}, "required": []string{"description", "prompt"}, "additionalProperties": false}}}
	messages := []any{map[string]string{"role": "user", "content": "Delegate once"}}
	send := func() *httptest.ResponseRecorder {
		body, e := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 100, "stream": true, "tools": tools, "messages": messages})
		if e != nil {
			t.Fatal(e)
		}
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
		req.Header.Set("Authorization", "Bearer argument-repair-token")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	first := send()
	if first.Code != 200 || !strings.Contains(first.Body.String(), "PRE_TOOL_PROGRESS") || strings.Contains(first.Body.String(), "appserver_") {
		t.Fatalf("first stream status=%d body=%s", first.Code, first.Body.String())
	}
	toolID := ""
	toolCount := 0
	for _, line := range strings.Split(first.Body.String(), "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var row map[string]any
		if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &row) != nil {
			continue
		}
		block, ok := row["content_block"].(map[string]any)
		if !ok || block["type"] != "tool_use" {
			continue
		}
		toolID, _ = block["id"].(string)
		toolCount++
	}
	if toolID == "" || toolCount != 1 || !strings.Contains(first.Body.String(), "Inspect repository") {
		t.Fatalf("invalid or duplicate public call count=%d body=%s", toolCount, first.Body.String())
	}
	messages = append(messages, map[string]any{"role": "assistant", "content": []any{map[string]string{"type": "text", "text": "PRE_TOOL_PROGRESS"}, map[string]any{"type": "tool_use", "id": toolID, "name": "Agent", "input": map[string]string{"description": "Inspect repository", "prompt": "Reply CHILD_GATEWAY_OK"}}}}, map[string]any{"role": "user", "content": []any{map[string]string{"type": "tool_result", "tool_use_id": toolID, "content": "CHILD_GATEWAY_OK"}}})
	second := send()
	if second.Code != 200 || !strings.Contains(second.Body.String(), "MAIN_GATEWAY_OK") || strings.Contains(second.Body.String(), "appserver_") {
		t.Fatalf("completion status=%d body=%s", second.Code, second.Body.String())
	}
	peer.mu.Lock()
	threads, executions, rejected := peer.threads, peer.toolResults, peer.rejectedArguments
	peer.mu.Unlock()
	if threads != 1 || executions != 1 || rejected != 1 {
		t.Fatalf("thread=%d executions=%d rejected=%d", threads, executions, rejected)
	}
	t.Logf("production streaming progress -> negative schema feedback=%d -> corrected public Agent=%d -> HTTP200 done; thread=%d", rejected, executions, threads)
}
