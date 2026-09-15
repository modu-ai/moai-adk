package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func TestManagedGPTModelSwitchDefersUntilPendingToolsComplete(t *testing.T) {
	ctx := context.Background()
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	dir, cwd := t.TempDir(), t.TempDir()
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := codexbridge.OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("family", cwd, translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)
	result := map[string]any{"type": "tool_result", "tool_use_id": "pending-call", "content": "result"}
	messages := []any{map[string]any{"role": "user", "content": "work"}, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": "pending-call", "name": "Read", "input": map[string]any{}}}}, map[string]any{"role": "user", "content": []any{result}}}
	r := gateway.RoutedRequest{Entry: entry, Managed: grant, Headers: http.Header{"X-Claude-Code-Session-Id": {"family"}}}
	encode := func() { r.Body, _ = json.Marshal(map[string]any{"max_tokens": 100, "messages": messages}) }
	encode()
	key := sha256.Sum256([]byte(grant.Scope() + "\x00family"))
	path := filepath.Join(dir, fmt.Sprintf("%x.json", key))
	seed := func(phase, model, workdir string) {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"Schema": 2, "Owner": map[string]string{"ConversationID": "family", "AccountScope": grant.Scope(), "ThreadID": "thread"}, "Phase": phase, "Prefix": "accepted-prefix", "Model": model, "CWD": workdir})
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	seed("waiting", "gpt-5.6-terra", cwd)
	q, err := prepare(ctx, r)
	if err != nil {
		t.Fatalf("pure continuation rejected during model selection change: %v", err)
	}
	if q.Model != "gpt-5.6-terra" || len(q.Input) != 0 || len(q.Results) != 1 || q.ExpectedPrefix != "accepted-prefix" || q.Resume {
		t.Fatalf("wrong continuation: %+v", q)
	}
	for _, phase := range []string{"active", "failed"} {
		seed(phase, "gpt-5.6-terra", cwd)
		if _, err = prepare(ctx, r); !errors.Is(err, codexbridge.ErrScope) {
			t.Fatalf("%s model change accepted: %v", phase, err)
		}
	}
	seed("waiting", "gpt-5.6-terra", cwd+"/foreign")
	if _, err = prepare(ctx, r); !errors.Is(err, codexbridge.ErrScope) {
		t.Fatalf("foreign cwd accepted: %v", err)
	}
	seed("waiting", "gpt-5.6-terra", cwd)
	messages[2].(map[string]any)["content"] = []any{result, map[string]string{"type": "text", "text": "new input"}}
	encode()
	if _, err = prepare(ctx, r); !errors.Is(err, codexbridge.ErrScope) {
		t.Fatalf("mixed new input accepted: %v", err)
	}
	messages[2].(map[string]any)["content"] = []any{result}
	messages = append(messages, map[string]string{"role": "user", "content": "separate followup"})
	encode()
	if _, err = prepare(ctx, r); !errors.Is(err, codexbridge.ErrScope) {
		t.Fatalf("separate new input accepted: %v", err)
	}
	seed("idle", "gpt-5.6-terra", cwd)
	messages = append(messages, map[string]string{"role": "assistant", "content": "done"}, map[string]string{"role": "user", "content": "next task"})
	encode()
	q, err = prepare(ctx, r)
	if err != nil || q.Model != entry.UpstreamID || !q.Resume || len(q.Results) != 0 {
		t.Fatalf("selected model not applied to next idle turn: %+v %v", q, err)
	}
}

func TestProductionGatewayModelChangeCompletesPendingTool(t *testing.T) {
	home, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	binDir, _ := installProductionWiringFakeCodex(t)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	peer := startClaudeGatewayPeer(t, home)
	payload, err := marshalGatewayPrivatePayload("model-change-token", gatewayGPTModels())
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
	tools := []any{map[string]any{"name": "Agent", "input_schema": map[string]any{"type": "object"}}}
	messages := []any{map[string]string{"role": "user", "content": "delegate"}}
	send := func(model string) *httptest.ResponseRecorder {
		t.Helper()
		raw, err := json.Marshal(map[string]any{"model": model, "max_tokens": 100, "tools": tools, "messages": messages})
		if err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(raw)))
		r.Header.Set("Authorization", "Bearer model-change-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	first := send("gpt-5.6-sol")
	if first.Code != 200 {
		t.Fatalf("first=%d %s", first.Code, first.Body.String())
	}
	var response struct {
		Content []map[string]any `json:"content"`
	}
	if err = json.Unmarshal(first.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Content) != 1 || response.Content[0]["type"] != "tool_use" {
		t.Fatal(first.Body.String())
	}
	toolID, _ := response.Content[0]["id"].(string)
	messages = append(messages, map[string]any{"role": "assistant", "content": response.Content}, map[string]any{"role": "user", "content": []any{map[string]string{"type": "tool_result", "tool_use_id": toolID, "content": "CHILD_GATEWAY_OK"}}})
	second := send("gpt-5.6-terra")
	if second.Code != 200 || !strings.Contains(second.Body.String(), "MAIN_GATEWAY_OK") {
		t.Fatalf("pending result model change=%d %s", second.Code, second.Body.String())
	}
	peer.mu.Lock()
	threads, results := peer.threads, peer.toolResults
	peer.mu.Unlock()
	if threads != 1 || results != 1 {
		t.Fatalf("continuation moved thread or replayed: threads=%d results=%d", threads, results)
	}
	t.Log("production gateway → prepare → adapter → Engine: selected terra while sol tool pending; same thread completed once with HTTP 200")
}

func TestManagedGPTPureToolResultRequestFailsClosed(t *testing.T) {
	for _, body := range []string{
		`{`, `{"messages":[]}`,
		`{"messages":[{"role":"assistant","content":"done"}]}`,
		`{"messages":[{"role":"system","content":[]}]}`,
		`{"messages":[{"role":"user","content":[]}]}`,
		`{"messages":[{"role":"user","content":"new input"}]}`,
		`{"messages":[{"role":"user","content":[{"type":"image"}]}]}`,
	} {
		if managedGPTPureToolResultRequest([]byte(body)) {
			t.Errorf("non-continuation accepted: %s", body)
		}
	}
}
