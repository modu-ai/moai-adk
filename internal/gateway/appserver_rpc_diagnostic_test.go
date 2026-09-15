package gateway

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

func TestRPCCodeDiagnosticNeverChangesPublicResponse(t *testing.T) {
	authority, err := NewAppServerAuthority(&managedAccount{kind: "chatgpt"}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog([]ModelEntry{{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: ProviderOpenAI, AuthMethod: AuthAppServer}})
	if err != nil {
		t.Fatal(err)
	}
	logger := &lifecycleRejectionRecorder{}
	adapter := &testAdapter{err: fmt.Errorf("PRIVATE-PROVIDER-DETAIL: %w", &codexapp.RPCError{Code: -32603})}
	server, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "token", MaxBodyBytes: 4096, Catalog: catalog, ManagedAuthority: authority, Adapters: map[ProviderID]Adapter{ProviderOpenAI: adapter}, RejectionLogger: logger})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"gpt-5.6-sol","max_tokens":64,"messages":[{"role":"user","content":"PRIVATE-REQUEST"}]}`))
	r.Header.Set("X-Test-Session", "token")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	if w.Code != 502 || w.Body.String() != `{"error":{"message":"appserver_transport_error","type":"api_error"},"type":"error"}`+"\n" {
		t.Fatalf("public contract changed: status=%d body=%q", w.Code, w.Body.String())
	}
	if len(logger.rows) != 1 || logger.rows[0]["rpc_code"] != "-32603" {
		t.Fatalf("missing private RPC code: %+v", logger.rows)
	}
	if strings.Contains(fmt.Sprint(logger.rows), "PRIVATE") {
		t.Fatal("private metadata leaked body/provider details")
	}
}
