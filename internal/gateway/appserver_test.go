package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type managedAccount struct{ kind string }

func (a *managedAccount) Account(context.Context) (codexapp.AccountResult, error) {
	return codexapp.AccountResult{Account: &codexapp.Account{Type: a.kind}}, nil
}

type managedProbeAdapter struct {
	authority *AppServerAuthority
	called    int
}

func (a *managedProbeAdapter) Send(ctx context.Context, r RoutedRequest) (*http.Response, error) {
	if r.Credential != nil || r.Managed == nil {
		return nil, ErrManagedAuthority
	}
	if err := a.authority.Check(ctx, r.Managed); err != nil {
		return nil, err
	}
	a.called++
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}, nil
}
func TestManagedAuthorityBypassesCredentialStoreAndRejectsLogout(t *testing.T) {
	account := &managedAccount{kind: "chatgpt"}
	authority, err := NewAppServerAuthority(account, "chatgpt", func(context.Context) (string, error) { return "profile-generation-1", nil })
	if err != nil {
		t.Fatal(err)
	}
	adapter := &managedProbeAdapter{authority: authority}
	catalog, err := NewCatalog([]ModelEntry{{RouteID: "gpt-test", UpstreamID: "gpt-test", Provider: ProviderOpenAI, AuthMethod: AuthAppServer}})
	if err != nil {
		t.Fatal(err)
	}
	storeCalls := 0
	server, err := NewServer(ServerConfig{SessionHeader: "X-Test-Session", SessionToken: "local-test", MaxBodyBytes: 4096, Catalog: catalog, ManagedAuthority: authority, ResolveCredential: func(context.Context, ModelEntry) (CredentialRef, error) { storeCalls++; return nil, nil }, Adapters: map[ProviderID]Adapter{ProviderOpenAI: adapter}})
	if err != nil {
		t.Fatal(err)
	}
	send := func() int {
		r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"gpt-test","max_tokens":100,"messages":[{"role":"user","content":"hello"}]}`))
		r.Header.Set("X-Test-Session", "local-test")
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w.Code
	}
	if got := send(); got != 200 {
		t.Fatal(got)
	}
	account.kind = ""
	if got := send(); got != 401 {
		t.Fatal(got)
	}
	if storeCalls != 0 || adapter.called != 1 {
		t.Fatal(storeCalls, adapter.called)
	}
}
func TestManagedGenerationCannotBeRebound(t *testing.T) {
	generation := "one"
	authority, err := NewAppServerAuthority(&managedAccount{kind: "chatgpt"}, "chatgpt", func(context.Context) (string, error) { return generation, nil })
	if err != nil {
		t.Fatal(err)
	}
	grant, err := authority.Authorize(context.Background(), ModelEntry{Provider: ProviderOpenAI, AuthMethod: AuthAppServer})
	if err != nil {
		t.Fatal(err)
	}
	generation = "two"
	if authority.Check(context.Background(), grant) == nil {
		t.Fatal("old generation accepted")
	}
}

func TestAppServerAdapterRejectsUntrustedBindingsBeforeTurn(t *testing.T) {
	for _, kind := range []string{"wrong-provider", "wrong-auth", "missing-grant", "foreign-grant", "prepare-error", "wrong-scope", "wrong-model", "invalid-stream"} {
		t.Run(kind, func(t *testing.T) {
			authority, _ := NewAppServerAuthority(&managedAccount{kind: "chatgpt"}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
			entry := ModelEntry{Provider: ProviderOpenAI, AuthMethod: AuthAppServer, UpstreamID: "gpt-test"}
			grant, err := authority.Authorize(context.Background(), entry)
			if err != nil {
				t.Fatal(err)
			}
			request := RoutedRequest{Entry: entry, Managed: grant, Body: []byte(`{"stream":false}`)}
			prepareCalls := 0
			adapter, err := NewAppServerAdapter(AppServerAdapterConfig{Engine: &codexbridge.Engine{}, Authority: authority, Prepare: func(context.Context, RoutedRequest) (codexbridge.Request, error) {
				prepareCalls++
				q := codexbridge.Request{Owner: codextools.Binding{ConversationID: "owned", AccountScope: "scope"}, Model: "gpt-test"}
				switch kind {
				case "prepare-error":
					return q, errors.New("invalid authenticated history")
				case "wrong-scope":
					q.Owner.AccountScope = "foreign"
				case "wrong-model":
					q.Model = "other"
				}
				return q, nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "wrong-provider":
				request.Entry.Provider = ProviderZAI
			case "wrong-auth":
				request.Entry.AuthMethod = AuthPKCE
			case "missing-grant":
				request.Managed = nil
			case "foreign-grant":
				other, _ := NewAppServerAuthority(&managedAccount{kind: "chatgpt"}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
				request.Managed, _ = other.Authorize(context.Background(), entry)
			case "invalid-stream":
				request.Body = []byte(`{"stream":"yes"}`)
			}
			if response, err := adapter.Send(context.Background(), request); err == nil || response != nil {
				t.Fatal("untrusted turn reached engine", response, err)
			}
			if (kind == "wrong-provider" || kind == "wrong-auth" || kind == "missing-grant" || kind == "foreign-grant") && prepareCalls != 0 {
				t.Fatal("invalid authority reached history resolver")
			}
		})
	}
}

func TestAppServerTextAndToolSSEBoundaries(t *testing.T) {
	for _, stream := range []bool{false, true} {
		response, err := appServerResponse("gpt-test", stream, codexbridge.Segment{Text: "before tool", Tool: &codexbridge.Tool{ID: "tool-test", Name: "echo", Arguments: json.RawMessage(`{"text":"hello"}`)}})
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(response.Body)
		if cerr := response.Body.Close(); cerr != nil {
			t.Fatal(cerr)
		}
		if err != nil {
			t.Fatal(err)
		}
		if stream {
			if strings.Count(string(raw), "event: content_block_start") != 2 || strings.Count(string(raw), "event: content_block_stop") != 2 || strings.Count(string(raw), "event: message_stop") != 1 || !strings.Contains(string(raw), "input_json_delta") {
				t.Fatal(string(raw))
			}
		} else {
			var message map[string]any
			if json.Unmarshal(raw, &message) != nil || message["stop_reason"] != "tool_use" {
				t.Fatal(string(raw))
			}
		}
	}
	if response, err := appServerResponse("gpt-test", true, codexbridge.Segment{}); err == nil || response != nil {
		t.Fatal("empty nonterminal segment succeeded")
	}
}

func TestManagedAuthorityRejectsModeFallbackAndGenerationRace(t *testing.T) {
	ctx := context.Background()
	entry := ModelEntry{Provider: ProviderOpenAI, AuthMethod: AuthAppServer}
	authority, err := NewAppServerAuthority(&managedAccount{kind: "apiKey"}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	if _, err = authority.Authorize(ctx, entry); err == nil {
		t.Fatal("subscription silently fell back to API mode")
	}
	calls := 0
	authority, err = NewAppServerAuthority(&managedAccount{kind: "chatgpt"}, "chatgpt", func(context.Context) (string, error) {
		calls++
		if calls == 1 {
			return "before", nil
		}
		return "after", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = authority.Authorize(ctx, entry); err == nil {
		t.Fatal("generation changed during account/read")
	}
	if _, err = NewAppServerAuthority(nil, "chatgpt", nil); err == nil {
		t.Fatal("missing authority accepted")
	}
	if _, err = NewAppServerAdapter(AppServerAdapterConfig{}); err == nil {
		t.Fatal("missing engine/resolver accepted")
	}
}
