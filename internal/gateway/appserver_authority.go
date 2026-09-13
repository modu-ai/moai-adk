package gateway

import (
	"context"
	"errors"
	"github.com/modu-ai/moai-adk/internal/codexapp"
)

var ErrManagedAuthority = errors.New("managed App Server authority unavailable or changed")

type AccountReader interface {
	Account(context.Context) (codexapp.AccountResult, error)
}

// AppServerAuthority uses official account/read and a trusted profile/account
// generation. Scope is local ownership metadata, not a claimed upstream account ID.
// The owner must change generation for every login/logout/account replacement.
type AppServerAuthority struct {
	account    AccountReader
	kind       string
	generation func(context.Context) (string, error)
}
type ManagedGrant struct {
	authority *AppServerAuthority
	scope     string
}

func (g *ManagedGrant) Scope() string {
	if g == nil {
		return ""
	}
	return g.scope
}
func NewAppServerAuthority(account AccountReader, kind string, generation func(context.Context) (string, error)) (*AppServerAuthority, error) {
	if account == nil || generation == nil || (kind != "chatgpt" && kind != "apiKey") {
		return nil, ErrManagedAuthority
	}
	return &AppServerAuthority{account: account, kind: kind, generation: generation}, nil
}
func (a *AppServerAuthority) Authorize(ctx context.Context, entry ModelEntry) (*ManagedGrant, error) {
	if a == nil || entry.Provider != ProviderOpenAI || entry.AuthMethod != AuthAppServer {
		return nil, ErrManagedAuthority
	}
	scope, err := a.generation(ctx)
	if err != nil || scope == "" || len(scope) > 256 {
		return nil, ErrManagedAuthority
	}
	result, err := a.account.Account(ctx)
	if err != nil || result.Account == nil || result.Account.Type != a.kind {
		return nil, ErrManagedAuthority
	}
	current, err := a.generation(ctx)
	if err != nil || current != scope {
		return nil, ErrManagedAuthority
	}
	return &ManagedGrant{authority: a, scope: scope}, nil
}
func (a *AppServerAuthority) Check(ctx context.Context, grant *ManagedGrant) error {
	if grant == nil || grant.authority != a {
		return ErrManagedAuthority
	}
	current, err := a.Authorize(ctx, ModelEntry{Provider: ProviderOpenAI, AuthMethod: AuthAppServer})
	if err != nil || current.scope != grant.scope {
		return ErrManagedAuthority
	}
	return nil
}
