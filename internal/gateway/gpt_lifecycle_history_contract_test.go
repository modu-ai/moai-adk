package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

type lifecycleContractRPC struct {
	mu          sync.Mutex
	events      chan codexapp.Message
	calls       map[string]int
	models      []string
	currentTurn map[string]string
	turnSeq     map[string]int
	lastErr     error
}

func (f *lifecycleContractRPC) Call(_ context.Context, method string, params any, out any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[method]++
	m, _ := params.(map[string]any)
	var result any
	switch method {
	case "thread/start":
		result = map[string]any{"thread": map[string]string{"id": "thread-contract"}}
	case "thread/resume":
		thread, _ := m["threadId"].(string)
		result = map[string]any{"thread": map[string]string{"id": thread}}
	case "turn/start":
		model, _ := m["model"].(string)
		f.models = append(f.models, model)
		thread, _ := m["threadId"].(string)
		f.turnSeq[thread]++
		turn := fmt.Sprintf("turn-%s-%d", thread, f.turnSeq[thread])
		f.currentTurn[thread] = turn
		result = map[string]any{"turn": map[string]string{"id": turn}}
		params, _ := json.Marshal(map[string]any{"threadId": thread, "turnId": turn, "callId": "call-" + turn, "tool": "echo", "arguments": map[string]string{"text": "hello"}})
		f.events <- codexapp.Message{ID: json.RawMessage(fmt.Sprintf("%q", thread)), Method: "item/tool/call", Params: params}
	case "turn/interrupt":
		result = map[string]any{}
	default:
		return fmt.Errorf("unexpected lifecycle RPC %s", method)
	}
	raw, _ := json.Marshal(result)
	return json.Unmarshal(raw, out)
}

func (f *lifecycleContractRPC) Respond(_ context.Context, id json.RawMessage, _ any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var thread string
	_ = json.Unmarshal(id, &thread)
	turn := f.currentTurn[thread]
	delta, _ := json.Marshal(map[string]any{"threadId": thread, "turnId": turn, "itemId": "text", "delta": "done"})
	completed, _ := json.Marshal(map[string]any{"threadId": thread, "turn": map[string]string{"id": turn, "status": "completed"}})
	f.events <- codexapp.Message{Method: "item/agentMessage/delta", Params: delta}
	f.events <- codexapp.Message{Method: "turn/completed", Params: completed}
	return nil
}
func (f *lifecycleContractRPC) DiscardRequest(json.RawMessage) error { return nil }
func (f *lifecycleContractRPC) Events() <-chan codexapp.Message      { return f.events }
func (f *lifecycleContractRPC) Err() error                           { return f.lastErr }

func TestGPTAppServerLifecycleAndHistoryAttribution(t *testing.T) {
	t.Run("resume", func(t *testing.T) {
		engine, _, store := lifecycleContractEngine(t)
		completeLifecycleConversation(t, engine, "resume")
		engine.Close()
		fresh, rpc := lifecycleContractEngineOn(t, store)
		q := lifecycleContractRequest("resume")
		q.Resume = true
		q.ExpectedPrefix = "prefix-two"
		q.PrefixDigest = "prefix-three"
		if _, err := fresh.Step(context.Background(), q); err != nil {
			t.Fatalf("durable idle resume: %v", err)
		}
		got := rpc.calls["turn/start"]
		if got != 1 {
			t.Errorf("resume turn/start count = %d, want 1", got)
		}
	})
	t.Run("model", func(t *testing.T) {
		engine, rpc, _ := lifecycleContractEngine(t)
		completeLifecycleConversation(t, engine, "model")
		q := lifecycleContractRequest("model")
		q.Model = "gpt-6-astra"
		q.ExpectedPrefix = "prefix-two"
		q.PrefixDigest = "prefix-three"
		if _, err := engine.Step(context.Background(), q); err != nil {
			t.Fatalf("idle model change: %v", err)
		}
		if got := strings.Join(rpc.models, ","); got != "gpt-5.6-sol,gpt-6-astra" {
			t.Errorf("turn model sequence = %q, want gpt-5.6-sol,gpt-6-astra", got)
		}
	})
	t.Run("compact", func(t *testing.T) {
		engine, rpc, _ := lifecycleContractEngine(t)
		completeLifecycleConversation(t, engine, "compact")
		if got := rpc.calls["thread/compact/start"]; got != 0 {
			t.Errorf("Claude summary compaction RPC count = %d, want 0", got)
		}
	})
	t.Run("fork", func(t *testing.T) {
		parent, _, store := lifecycleContractEngine(t)
		completeLifecycleConversation(t, parent, "parent")
		parent.Close()
		engine, rpc := lifecycleContractEngineOn(t, store)
		q := lifecycleContractRequest("fork")
		q.Fork = true
		q.ExpectedPrefix = "prefix-two"
		q.PrefixDigest = "prefix-three"
		if _, err := engine.Step(context.Background(), q); err != nil {
			t.Fatalf("fork at durable prefix: %v", err)
		}
		if got := rpc.calls["thread/start"]; got != 1 {
			t.Errorf("fork thread/start count = %d, want 1", got)
		}
	})

	t.Run("same-prefix-retry", func(t *testing.T) {
		history, base := lifecycleHistoryFixture(t)
		first := runLifecycleHistoryProbe(t, history, base, "")
		second := runLifecycleHistoryProbe(t, history, base, "")
		if first.status != http.StatusOK || second.status != http.StatusOK || first.external+second.external != 2 || first.rpc+second.rpc != 2 || first.upstreamHTTP+second.upstreamHTTP != 2 {
			t.Errorf("same idempotency/prefix retry statuses=%d/%d generation=%d rpc=%d http=%d, want 200/200/2/2/2", first.status, second.status, first.external+second.external, first.rpc+second.rpc, first.upstreamHTTP+second.upstreamHTTP)
		}
	})
	t.Run("completed-prefix-new-user-input", func(t *testing.T) {
		history, base := lifecycleHistoryFixture(t)
		continued := append(append([]map[string]any(nil), base...), map[string]any{"role": "user", "content": "new user input after completed prefix"})
		probe := runLifecycleHistoryProbe(t, history, continued, "")
		if probe.status != http.StatusOK || probe.external != 1 || probe.rpc != 1 || probe.upstreamHTTP != 1 {
			t.Errorf("completed prefix plus new user status=%d generation=%d rpc=%d http=%d, want 200/1/1/1", probe.status, probe.external, probe.rpc, probe.upstreamHTTP)
		}
	})
	t.Run("changed-history", func(t *testing.T) {
		history, base := lifecycleHistoryFixture(t)
		changed := append([]map[string]any(nil), base...)
		changed[1] = map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "changed"}}}
		probe := runLifecycleHistoryProbe(t, history, withLifecycleCanary(changed, "CANARY-CHANGED-91"), "CANARY-CHANGED-91")
		assertLifecycleHistoryBlocked(t, "history_changed", probe)
	})
	t.Run("agent-summary", func(t *testing.T) {
		history, base := lifecycleHistoryFixture(t)
		summary := append([]map[string]any(nil), base...)
		summary[1] = map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "agent_summary", "summary": "CANARY-SUMMARY-92", "source": "untrusted"}, map[string]any{"type": "text", "text": "one"}}}
		probe := runLifecycleHistoryProbe(t, history, withLifecycleCanary(summary, "CANARY-SUMMARY-92"), "CANARY-SUMMARY-92")
		assertLifecycleHistoryBlocked(t, "agent_summary_untrusted", probe)
	})
	t.Run("receipt-manifest-mismatch", func(t *testing.T) {
		_, base := lifecycleHistoryFixture(t)
		empty := lifecycleHistoryAuthority(t, "88888888-8888-4888-8888-888888888888")
		probe := runLifecycleHistoryProbe(t, empty, withLifecycleCanary(base, "CANARY-MANIFEST-93"), "CANARY-MANIFEST-93")
		assertLifecycleHistoryBlocked(t, "receipt_manifest_mismatch", probe)
	})
	t.Run("resume-duplicate", func(t *testing.T) {
		history, base := lifecycleHistoryFixture(t)
		duplicate := append(append([]map[string]any(nil), base...), base[1])
		probe := runLifecycleHistoryProbe(t, history, withLifecycleCanary(duplicate, "CANARY-DUPLICATE-94"), "CANARY-DUPLICATE-94")
		assertLifecycleHistoryBlocked(t, "resume_duplicate_input", probe)
	})
	t.Run("structured-rejection-logger", func(t *testing.T) {
		history, base := lifecycleHistoryFixture(t)
		changed := append([]map[string]any(nil), base...)
		changed[1] = map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "CANARY-LOGGER-95"}}}
		probe := lifecycleHistoryProbe{canary: "CANARY-LOGGER-95"}
		catalog, err := NewCatalog([]ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: ProviderAnthropic, AuthMethod: AuthOAuthPassthrough, Capabilities: Capabilities{ContextTokens: 100}}})
		if err != nil {
			t.Fatal(err)
		}
		recorder := &lifecycleRejectionRecorder{}
		cfg := ServerConfig{SessionHeader: "X-MoAI-Session-Token", SessionToken: "fixture-session", MaxBodyBytes: 1 << 20, Catalog: catalog, Adapters: map[ProviderID]Adapter{ProviderAnthropic: lifecycleRejectingAdapter{history: history, probe: &probe}}}
		field := reflect.ValueOf(&cfg).Elem().FieldByName("RejectionLogger")
		value := reflect.ValueOf(recorder)
		if !field.IsValid() || !field.CanSet() || !value.Type().AssignableTo(field.Type()) {
			t.Errorf("production ServerConfig RejectionLogger seam available=%v settable=%v type=%v, want injectable structured rejection recorder", field.IsValid(), field.IsValid() && field.CanSet(), func() any {
				if field.IsValid() {
					return field.Type()
				}
				return nil
			}())
			return
		}
		field.Set(value)
		server, err := NewServer(cfg)
		if err != nil {
			t.Fatal(err)
		}
		serveLifecycleHistoryRequest(server, changed)
		recorder.assertSafe(t, "history_changed", "gpt-5.6-sol", "CANARY-LOGGER-95")
	})
}

type lifecycleRejectionRecorder struct {
	mu   sync.Mutex
	rows []map[string]string
}

func (r *lifecycleRejectionRecorder) RecordGatewayRejection(fields map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := map[string]string{}
	for k, v := range fields {
		clone[k] = v
	}
	r.rows = append(r.rows, clone)
}
func (r *lifecycleRejectionRecorder) assertSafe(t *testing.T, cause, route, canary string) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.rows) != 1 {
		t.Errorf("structured rejection rows=%d want=1", len(r.rows))
		return
	}
	row := r.rows[0]
	if len(row) != 3 || row["cause"] != cause || row["route"] != route || len(row["digest"]) != 64 {
		t.Errorf("structured rejection fields=%v want exact cause/route/nonidentifying digest", row)
	}
	for key, value := range row {
		lower := strings.ToLower(key + "=" + value)
		for _, forbidden := range []string{"prompt", "summary", "tool", "reasoning", "token", "credential", strings.ToLower(canary)} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("structured rejection leaked forbidden %q in %s", forbidden, lower)
			}
		}
	}
}

type lifecycleHistoryProbe struct {
	status                                int
	external, rpc, upstreamHTTP, redacted int
	body, cause, canary                   string
}

type lifecycleRejectingAdapter struct {
	history translate.HistoryAuthority
	probe   *lifecycleHistoryProbe
}

func (a lifecycleRejectingAdapter) Send(ctx context.Context, r RoutedRequest) (*http.Response, error) {
	var body struct {
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(body.Messages)
	if err := a.history.Check(ctx, r.Entry.RouteID, "owner", raw); err != nil {
		var classified interface{ CauseCode() string }
		if errors.As(err, &classified) {
			a.probe.cause = classified.CauseCode()
		}
		if strings.Contains(err.Error(), a.probe.canary) {
			a.probe.redacted++
		}
		return nil, err
	}
	a.probe.external++
	a.probe.rpc++
	a.probe.upstreamHTTP++
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"type":"message","content":[]}`))}, nil
}

func runLifecycleHistoryProbe(t *testing.T, history translate.HistoryAuthority, messages []map[string]any, canary string) lifecycleHistoryProbe {
	t.Helper()
	observed := lifecycleHistoryProbe{canary: canary}
	catalog, err := NewCatalog([]ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: ProviderAnthropic, AuthMethod: AuthOAuthPassthrough, Capabilities: Capabilities{ContextTokens: 100}}})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServer(ServerConfig{SessionHeader: "X-MoAI-Session-Token", SessionToken: "fixture-session", MaxBodyBytes: 1 << 20, Catalog: catalog, Adapters: map[ProviderID]Adapter{ProviderAnthropic: lifecycleRejectingAdapter{history: history, probe: &observed}}})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 10, "messages": messages})
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(string(raw)))
	req.Header.Set("X-MoAI-Session-Token", "fixture-session")
	req.Header.Set("Authorization", "Bearer fixture-oauth")
	req.Header.Set("Idempotency-Key", "history-contract-idempotency")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	observed.status = res.Code
	observed.body = res.Body.String()
	return observed
}

func serveLifecycleHistoryRequest(server *Server, messages []map[string]any) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 10, "messages": messages})
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(string(raw)))
	req.Header.Set("X-MoAI-Session-Token", "fixture-session")
	req.Header.Set("Authorization", "Bearer fixture-oauth")
	req.Header.Set("Idempotency-Key", "history-contract-idempotency")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	return res
}

func assertLifecycleHistoryBlocked(t *testing.T, wantCause string, got lifecycleHistoryProbe) {
	t.Helper()
	if got.status != http.StatusBadRequest || got.external != 0 || got.rpc != 0 || got.upstreamHTTP != 0 || got.redacted != 0 {
		t.Errorf("%s boundary status=%d generation=%d rpc=%d http=%d redacted-log=%d, want 400/0/0/0/0", wantCause, got.status, got.external, got.rpc, got.upstreamHTTP, got.redacted)
	}
	if got.cause != wantCause {
		t.Errorf("history error CauseCode=%q, want %q", got.cause, wantCause)
	}
	if !strings.Contains(got.body, `"type":"invalid_request_error"`) || !strings.Contains(got.body, wantCause) {
		t.Errorf("history HTTP envelope=%q, want invalid_request_error with cause %q", got.body, wantCause)
	}
	if got.canary != "" && strings.Contains(got.body, got.canary) {
		t.Errorf("history HTTP envelope leaked canary %q: %q", got.canary, got.body)
	}
}

func withLifecycleCanary(messages []map[string]any, canary string) []map[string]any {
	result := append([]map[string]any(nil), messages...)
	return append(result, map[string]any{"role": "user", "content": canary})
}

func lifecycleHistoryFixture(t *testing.T) (translate.HistoryAuthority, []map[string]any) {
	t.Helper()
	history := lifecycleHistoryAuthority(t, "77777777-7777-4777-8777-777777777777")
	envelope, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_contract","summary":[],"encrypted_content":"cipher-contract"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	base := []map[string]any{
		{"role": "user", "content": "hello"},
		{"role": "assistant", "content": []any{map[string]any{"type": "redacted_thinking", "data": envelope.Data()}, map[string]any{"type": "text", "text": "one"}}},
	}
	raw, _ := json.Marshal(base)
	if err := history.Publish(context.Background(), "gpt-5.6-sol", "owner", raw); err != nil {
		t.Fatalf("publish receipt history fixture: %v", err)
	}
	return history, base
}

func lifecycleHistoryAuthority(t *testing.T, session string) translate.HistoryAuthority {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := receipt.OpenStore(context.Background(), dir, session, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return translate.NewGPTSubscriptionReceiptHistory(store, session, session)
}

func lifecycleContractEngine(t *testing.T) (*codexbridge.Engine, *lifecycleContractRPC, *codexbridge.FileStore) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := codexbridge.OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	engine, rpc := lifecycleContractEngineOn(t, store)
	return engine, rpc, store
}

func lifecycleContractEngineOn(t *testing.T, store *codexbridge.FileStore) (*codexbridge.Engine, *lifecycleContractRPC) {
	t.Helper()
	rpc := &lifecycleContractRPC{events: make(chan codexapp.Message, 16), calls: map[string]int{}, currentTurn: map[string]string{}, turnSeq: map[string]int{}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(engine.Close)
	return engine, rpc
}

func completeLifecycleConversation(t *testing.T, engine *codexbridge.Engine, id string) {
	t.Helper()
	q := lifecycleContractRequest(id)
	segment, err := engine.Step(context.Background(), q)
	if err != nil || segment.Tool == nil {
		t.Fatalf("start lifecycle conversation: segment=%+v err=%v", segment, err)
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-one"
	q.PrefixDigest = "prefix-two"
	q.Results = []codexbridge.ToolResult{{ID: segment.Tool.ID, Content: []codexbridge.Content{{Type: "inputText", Text: "done"}}, Success: true}}
	last, err := engine.Step(context.Background(), q)
	if err != nil || !last.Done {
		t.Fatalf("complete lifecycle conversation: segment=%+v err=%v", last, err)
	}
}

func lifecycleContractRequest(id string) codexbridge.Request {
	return codexbridge.Request{
		Owner:        codextools.Binding{ConversationID: id, AccountScope: "profile-generation-contract"},
		Model:        "gpt-5.6-sol",
		CWD:          "/tmp",
		Tools:        []codextools.Definition{{Name: "echo", InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)}},
		Input:        []any{map[string]string{"type": "text", "text": "hello"}},
		PrefixDigest: "prefix-one",
	}
}
