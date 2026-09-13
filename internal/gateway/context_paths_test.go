package gateway

import (
	"context"
	"testing"
)

// glmImageBody is a Messages request whose user turn carries a text block and
// an image block — the input shape the AS-020 re-verdict refuses on GLM
// routes.
const glmImageBody = `{"model":"glm-tier-x","max_tokens":8,"messages":[{"role":"user","content":[{"type":"text","text":"what is in this picture"},{"type":"image","source":{"type":"base64","media_type":"image/png","data":"AAAA"}}]}]}`

// TestGatewayContextPathCapabilityGate pins the AS-020 provider capability
// re-verdict at the request boundary: a route whose declared capabilities are
// text-only rejects image input explicitly before any credential resolution
// or upstream send (upstream request count 0), while an image-accepting route
// passes the same input through, and a text-only route still serves text.
// The common declaration in gateway_product_binding is a source claim, never
// an acceptance guarantee.
func TestGatewayContextPathCapabilityGate(t *testing.T) {
	catalog, err := NewCatalog([]ModelEntry{
		{RouteID: "glm-tier-x", UpstreamID: "glm-tier-x", Provider: ProviderZAI, AuthMethod: AuthAPIKey, Capabilities: Capabilities{ContextTokens: 200000, Images: false, Tools: true, Streaming: true}},
		{RouteID: "claude-opus-5", UpstreamID: "claude-opus-5", Provider: ProviderAnthropic, AuthMethod: AuthAPIKey, Capabilities: Capabilities{ContextTokens: 1000000, Images: true, Tools: true, Streaming: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	a := &testAdapter{body: "{}"}
	s, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "local-secret", MaxBodyBytes: 4096, Catalog: catalog, ResolveCredential: func(_ context.Context, e ModelEntry) (CredentialRef, error) { return &testCredential{provider: e.Provider}, nil }, Adapters: map[ProviderID]Adapter{ProviderZAI: a, ProviderAnthropic: a}})
	if err != nil {
		t.Fatal(err)
	}

	// GLM text-only refusal: image input is an explicit 400 and the upstream
	// request count stays 0.
	w := request(s, "POST", "/v1/messages", glmImageBody)
	if w.Code != 400 {
		t.Fatalf("image input on a text-only route answered %d, want explicit 400", w.Code)
	}
	if got := a.calls.Load(); got != 0 {
		t.Fatalf("rejected image input reached the upstream %d times", got)
	}

	// The same route still serves text-only input.
	w = request(s, "POST", "/v1/messages", `{"model":"glm-tier-x","max_tokens":8,"messages":[{"role":"user","content":"hello"}]}`)
	if w.Code != 200 {
		t.Fatalf("text input on a text-only route answered %d", w.Code)
	}
	if got := a.calls.Load(); got != 1 {
		t.Fatalf("text input upstream calls = %d, want 1", got)
	}

	// Claude's declared image capability passes the identical input through.
	w = request(s, "POST", "/v1/messages", `{"model":"claude-opus-5","max_tokens":8,"messages":[{"role":"user","content":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"AAAA"}}]}]}`)
	if w.Code != 200 {
		t.Fatalf("image input on an image-accepting route answered %d", w.Code)
	}
	if got := a.calls.Load(); got != 2 {
		t.Fatalf("image-accepting route upstream calls = %d, want 2", got)
	}
}
