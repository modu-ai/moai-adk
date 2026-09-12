package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
)

func TestOAuthRequestIsolationAnd401(t *testing.T) {
	catalog, err := NewCatalog([]ModelEntry{{RouteID: "claude-opus-5", UpstreamID: "claude-opus-5", Provider: ProviderAnthropic, AuthMethod: AuthOAuthPassthrough}})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	seen := map[string]bool{}
	adapter := &testAdapter{status: 401, check: func(q RoutedRequest) {
		if q.Headers.Get("Authorization") != "" || q.Headers.Get("X-MoAI-Session-Token") != "" {
			t.Error("inbound secret forwarded")
		}
		out, _ := http.NewRequest("POST", auth.AnthropicEndpoint, nil)
		if err := q.Credential.Apply(out); err != nil {
			t.Error(err)
			return
		}
		mu.Lock()
		seen[out.Header.Get("Authorization")] = true
		mu.Unlock()
	}}
	cfg := ServerConfig{SessionHeader: "X-MoAI-Session-Token", SessionToken: "local", MaxBodyBytes: 4096, Catalog: catalog, Adapters: map[ProviderID]Adapter{ProviderAnthropic: adapter}, ResolveCredential: func(context.Context, ModelEntry) (CredentialRef, error) {
		t.Error("OAuth reached storage resolver")
		return nil, nil
	}}
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	send := func(value []string, ctx context.Context) int {
		r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-opus-5","max_tokens":2,"messages":[{"role":"user","content":"hello"}]}`)).WithContext(ctx)
		r.Header.Set("X-MoAI-Session-Token", "local")
		r.Header["Authorization"] = value
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w.Code
	}
	var wg sync.WaitGroup
	for _, token := range []string{"Bearer request-one", "Bearer request-two"} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			if send([]string{token}, context.Background()) != 401 {
				t.Error("upstream401 lost")
			}
		}(token)
	}
	wg.Wait()
	if len(seen) != 2 || !seen["Bearer request-one"] || !seen["Bearer request-two"] {
		t.Fatal("request credentials crossed")
	}
	for _, values := range [][]string{nil, {"invalid"}, {"Bearer one", "Bearer two"}} {
		if send(values, context.Background()) != 401 {
			t.Fatal("bad credential accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	send([]string{"Bearer cancelled"}, ctx)
	if adapter.calls.Load() != 2 {
		t.Fatal("invalid/cancelled request sent")
	}
	cfg.SessionHeader = "Authorization"
	if _, err := NewServer(cfg); err == nil {
		t.Fatal("OAuth reused session header")
	}
}

func TestOAuthNative401AndHeaders(t *testing.T) {
	tr, calls := nativeTLS(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer request-oauth" || r.Header.Get("X-Api-Key") != "" || r.Header.Get("X-MoAI-Session-Token") != "" || r.Header.Get("Anthropic-Beta") != "oauth-2025-04-20" {
			t.Error("OAuth headers")
		}
		w.WriteHeader(401)
	})
	cfg := nativeConfig(tr)
	cfg.AllowedBetas = []string{"oauth-2025-04-20"}
	a, err := NewAnthropicOAuthAdapter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	q := nativeQ(t)
	q.Entry.AuthMethod = AuthOAuthPassthrough
	q.Headers = http.Header{"Anthropic-Beta": {"oauth-2025-04-20"}, "X-MoAI-Session-Token": {"must-not-leave"}}
	q.Credential, err = auth.NewAnthropicOAuth(context.Background(), http.Header{"Authorization": {"Bearer request-oauth"}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := a.Send(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	body, readErr := io.ReadAll(r.Body)
	if readErr != nil || !strings.Contains(string(body), `"authentication_error"`) {
		t.Fatal("native refresh error type lost")
	}
	if r.StatusCode != 401 || calls.Load() != 1 {
		t.Fatal("401 retry/translation")
	}
}

func TestOAuthDefaultCatalogAfterMeasuredM0(t *testing.T) {
	c, err := NewSessionCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"claude-opus-5", "claude-sonnet-5"} {
		e, err := c.Resolve(id)
		if err != nil || e.AuthMethod != AuthOAuthPassthrough || e.Capabilities.ContextTokens != 0 {
			t.Fatal("M0 catalog contract")
		}
	}
}
