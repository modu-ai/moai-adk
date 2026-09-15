package gateway

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
)

func TestAppServerSummaryDiagnosticsAreRequestLocalAndOptional(t *testing.T) {
	authority, err := NewAppServerAuthority(&managedAccount{kind: "chatgpt"}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := ModelEntry{Provider: ProviderOpenAI, AuthMethod: AuthAppServer, UpstreamID: "gpt-test"}
	grant, err := authority.Authorize(context.Background(), entry)
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := NewAppServerAdapter(AppServerAdapterConfig{Engine: &codexbridge.Engine{}, Authority: authority, Prepare: func(_ context.Context, r RoutedRequest) (codexbridge.Request, error) {
		if string(r.Body) == "observed" {
			r.SummaryDiagnostic.LastRole = "user"
			r.SummaryDiagnostic.Classified = true
		}
		return codexbridge.Request{}, codexbridge.ErrScope
	}})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, body := range []string{"observed", "unobserved"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := adapter.Send(context.Background(), RoutedRequest{Entry: entry, Managed: grant, Body: []byte(body), SummaryDiagnostic: &AppServerSummaryDiagnostic{LastRole: "CANARY"}})
			if !errors.Is(e, codexbridge.ErrScope) {
				t.Errorf("error identity changed: %v", e)
			}
			if e.Error() != codexbridge.ErrScope.Error() {
				t.Errorf("error text changed: %v", e)
			}
			var d *appServerDiagnosticError
			if errors.As(e, &d) != (body == "observed") {
				t.Errorf("observability leaked across requests: %s", body)
			}
			if d != nil && (!d.diagnostic.Classified || d.diagnostic.LastRole != "user") {
				t.Errorf("wrong diagnostic: %+v", d.diagnostic)
			}
		}()
	}
	wg.Wait()
}

func TestSummaryDiagnosticSuccessDoesNotWriteRejection(t *testing.T) {
	authority, err := NewAppServerAuthority(&managedAccount{kind: "chatgpt"}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCatalog([]ModelEntry{{RouteID: "gpt-test", UpstreamID: "gpt-test", Provider: ProviderOpenAI, AuthMethod: AuthAppServer}})
	if err != nil {
		t.Fatal(err)
	}
	logger := &lifecycleRejectionRecorder{}
	server, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "token", MaxBodyBytes: 4096, Catalog: catalog, ManagedAuthority: authority, Adapters: map[ProviderID]Adapter{ProviderOpenAI: &managedProbeAdapter{authority: authority}}, RejectionLogger: logger})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"gpt-test","max_tokens":100,"messages":[{"role":"user","content":"CANARY-TEXT"}]}`))
	r.Header.Set("X-Test-Session", "token")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	if w.Code != 200 || len(logger.rows) != 0 {
		t.Fatalf("successful request wrote diagnostics: status=%d rows=%d", w.Code, len(logger.rows))
	}
}
