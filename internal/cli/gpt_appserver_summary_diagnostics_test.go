package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func TestSummaryDiagnosticPrepareToPrivateRejection(t *testing.T) {
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	catalog, err := gateway.NewCatalog([]gateway.ModelEntry{entry})
	if err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	logger, err := newManagedGPTRejectionLogger(dir, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("family", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative}, managedGPTTestStore(t))
	adapter, err := gateway.NewAppServerAdapter(gateway.AppServerAdapterConfig{Engine: &codexbridge.Engine{}, Authority: authority, Prepare: func(ctx context.Context, r gateway.RoutedRequest) (codexbridge.Request, error) {
		q, e := prepare(ctx, r)
		if e != nil {
			return q, e
		}
		return q, fmt.Errorf("%w: unknown or duplicate tool result", codexbridge.ErrScope)
	}})
	if err != nil {
		t.Fatal(err)
	}
	server, err := gateway.NewServer(gateway.ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "token", MaxBodyBytes: 1 << 20, Catalog: catalog, ManagedAuthority: authority, Adapters: map[gateway.ProviderID]gateway.Adapter{gateway.ProviderOpenAI: adapter}, RejectionLogger: logger})
	if err != nil {
		t.Fatal(err)
	}
	for _, withAgent := range []bool{false, true} {
		body, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 100, "messages": []any{map[string]string{"role": "user", "content": observedAgentSummaryPrompt}}})
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
		req.Header.Set("X-Test-Session", "token")
		if withAgent {
			req.Header.Set("X-Claude-Code-Agent-Id", "CANARY-PRIVATE-HEADER")
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		if w.Code != 400 || strings.TrimSpace(w.Body.String()) != `{"error":{"message":"appserver_scope_mismatch","type":"invalid_request_error"},"type":"error"}` {
			t.Fatalf("public changed: %d %s", w.Code, w.Body.String())
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, logger.name))
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(rows) != 2 {
		t.Fatal(string(raw))
	}
	for i, line := range rows {
		var row map[string]string
		if err = json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		want := "false"
		if i == 1 {
			want = "true"
		}
		if row["summary_classified"] != want || row["agent_header_present"] != want || row["summary_prompt_match"] != "true" || row["last_message_role"] != "user" || row["last_content_type"] != "string" {
			t.Fatalf("diagnostic missing or wrong: %s", line)
		}
	}
	if strings.Contains(string(raw), "CANARY") || strings.Contains(string(raw), "Describe your") {
		t.Fatal("private data leaked")
	}
}

func TestSummaryDiagnosticLoggerRejectsUnlistedValues(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	logger, err := newManagedGPTRejectionLogger(dir, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{"cause": "appserver_scope_mismatch", "route": "gpt-5.6-sol", "digest": strings.Repeat("a", 64)}
	for _, key := range []string{"summary_classified", "agent_header_present", "summary_prompt_match", "last_message_role", "last_content_type"} {
		fields[key] = "CANARY-PRIVATE"
	}
	logger.RecordGatewayRejection(fields)
	raw, err := os.ReadFile(filepath.Join(dir, logger.name))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "CANARY") {
		t.Fatal(string(raw))
	}
}

func TestSummaryDiagnosticShapeNeverCopiesUnknownRoleOrType(t *testing.T) {
	for _, tt := range []struct{ body, role, content string }{
		{`{`, "missing", "missing"},
		{`{"messages":[]}`, "missing", "missing"},
		{`{"messages":[{"role":"CANARY-ROLE","content":"CANARY-TEXT"}]}`, "unknown", "string"},
		{`{"messages":[{"role":"user","content":[{"type":"CANARY-TYPE","text":"CANARY-TEXT"}]}]}`, "user", "unknown"},
		{`{"messages":[{"role":"assistant","content":[{"type":"text","text":"CANARY-TEXT"}]}]}`, "assistant", "text"},
		{`{"messages":[{"role":"user","content":[{"type":"tool_result","content":"CANARY-RESULT"}]}]}`, "user", "tool_result"},
		{`{"messages":[{"role":"user","content":[]}]}`, "user", "unknown"},
	} {
		d := &gateway.AppServerSummaryDiagnostic{}
		matched := managedGPTAgentSummaryShape([]byte(tt.body), d)
		if matched || d.LastRole != tt.role || d.LastContent != tt.content {
			t.Fatalf("unexpected shape: %+v match=%v", d, matched)
		}
		raw, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "CANARY") {
			t.Fatal("untrusted shape leaked")
		}
	}
}
