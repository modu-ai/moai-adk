package gateway

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
)

// TestGatewaySubscriptionFailureNeverSwitchesToAPI pins the automation-only
// half of AS-021: when the subscription credential cannot be resolved, the
// client gets an explicit 401 and the upstream request count stays 0 — no
// API-billing route is attempted, because the GPT session catalog declares no
// API-key route at all. The valid-credential contrast proves the fixture
// reaches the resolver and the upstream, so the absent-credential zero is a
// measured refusal, not a vacuous pass.
func TestGatewaySubscriptionFailureNeverSwitchesToAPI(t *testing.T) {
	catalog, err := NewCatalog([]ModelEntry{
		{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: ProviderOpenAI, AuthMethod: AuthPKCE, Capabilities: Capabilities{ContextTokens: 872000, Images: true, Tools: true, Streaming: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	a := &testAdapter{body: "{}"}
	cases := []struct {
		name      string
		ref       CredentialRef
		wantCode  int
		wantCalls int64
	}{
		{name: "valid subscription credential reaches the upstream", ref: &testCredential{provider: ProviderOpenAI}, wantCode: 200, wantCalls: 1},
		{name: "absent subscription credential is a refusal with zero upstream requests", ref: &testCredential{provider: ProviderOpenAI, absent: true}, wantCode: 401, wantCalls: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a.calls.Store(0)
			s, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "local-secret", MaxBodyBytes: 4096, Catalog: catalog, ResolveCredential: func(context.Context, ModelEntry) (CredentialRef, error) {
				if tc.ref.(*testCredential).absent {
					return nil, auth.ErrCredentialAbsent
				}
				return tc.ref, nil
			}, Adapters: map[ProviderID]Adapter{ProviderOpenAI: a}})
			if err != nil {
				t.Fatal(err)
			}
			w := request(s, "POST", "/v1/messages", `{"model":"gpt-5.6-sol","max_tokens":2,"messages":[{"role":"user","content":"hello"}]}`)
			if w.Code != tc.wantCode {
				t.Fatalf("subscription failure answered %d, want %d", w.Code, tc.wantCode)
			}
			if got := a.calls.Load(); got != tc.wantCalls {
				t.Fatalf("upstream requests = %d, want %d", got, tc.wantCalls)
			}
		})
	}
}
