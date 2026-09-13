package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
)

// AS-011 rejection groups on the gateway surface: a bare "gpt-6" is not an
// exact selection ID and never resolves, and a foreign-provider entry never
// reaches the App Server bridge even when a request forges its auth method —
// an idle model switch can only land on a model the catalog carries for the
// App Server provider.
func TestAppServerModelSwitchRejectsBareAndForeignIDs(t *testing.T) {
	catalog, err := NewSessionCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = catalog.Resolve("gpt-6"); !errors.Is(err, ErrUnknownModel) {
		t.Fatal("bare model id accepted", err)
	}
	entry, err := catalog.Resolve("gpt-6-astra")
	if err != nil || entry.Provider != ProviderOpenAI {
		t.Fatal("exact selection id missing from the session catalog", entry, err)
	}

	account := &managedAccount{kind: "chatgpt"}
	authority, err := NewAppServerAuthority(account, "chatgpt", func(context.Context) (string, error) { return "profile-generation-1", nil })
	if err != nil {
		t.Fatal(err)
	}
	grant, err := authority.Authorize(context.Background(), ModelEntry{Provider: ProviderOpenAI, AuthMethod: AuthAppServer})
	if err != nil {
		t.Fatal(err)
	}
	stepped := 0
	adapter, err := NewAppServerAdapter(AppServerAdapterConfig{
		Engine:    &codexbridge.Engine{},
		Authority: authority,
		Prepare: func(context.Context, RoutedRequest) (codexbridge.Request, error) {
			stepped++
			return codexbridge.Request{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	foreign := ModelEntry{RouteID: "claude-opus-5", UpstreamID: "claude-opus-5", Provider: ProviderAnthropic, AuthMethod: AuthAppServer}
	request := RoutedRequest{Entry: foreign, Managed: grant, Body: []byte(`{"stream":false}`)}
	if _, err = adapter.Send(context.Background(), request); !errors.Is(err, ErrManagedAuthority) {
		t.Fatal("foreign provider reached the App Server bridge", err)
	}
	if stepped != 0 {
		t.Fatal("foreign provider entered the bridge prepare boundary")
	}
}
