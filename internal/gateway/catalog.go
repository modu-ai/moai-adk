// Package gateway provides session-local model routing contracts.
package gateway

import (
	"errors"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
)

type ProviderID = auth.ProviderID

const (
	ProviderAnthropic = auth.ProviderAnthropic
	ProviderOpenAI    = auth.ProviderOpenAI
	ProviderZAI       = auth.ProviderZAI
)

type CredentialRef = auth.CredentialRef
type AuthMethod string

const (
	AuthOAuthPassthrough AuthMethod = "oauth-passthrough"
	AuthAPIKey           AuthMethod = "api-key"
	AuthPKCE             AuthMethod = "pkce"
	AuthExistingGLM      AuthMethod = "existing-glm"
)

// Capabilities contains only verified catalog capabilities. Zero values do not
// promise support; transport adapters must enforce the selected entry's policy.
type Capabilities struct {
	ContextTokens int
	Images        bool
	PDF           bool
	Tools         bool
	Streaming     bool
}
type ModelEntry struct {
	RouteID      string
	Provider     ProviderID
	UpstreamID   string
	AuthMethod   AuthMethod
	Capabilities Capabilities
}
type LaunchPlan struct {
	InitialModel    string
	InitialProvider ProviderID
	Catalog         CatalogSnapshot
	ChildEnv        []string
	ChildSettings   string
	// Args is the validated native-session descriptor. A nil value preserves
	// the caller's pass-through arguments; a non-nil value replaces them.
	// Conversation ownership is established before the child is started.
	Args []string
}
type RequestContext struct {
	RequestID     string
	Entry         ModelEntry
	CredentialRef CredentialRef
}

var ErrUnknownModel = errors.New("gateway model not registered")
var ErrConflictingModel = errors.New("gateway model registration conflict")

// CatalogSnapshot owns its input and exposes only value copies. Concurrent
// requests can resolve entries without mutable current-provider state.
type CatalogSnapshot struct {
	entries []ModelEntry
	byID    map[string]ModelEntry
}

// NewCatalog copies entries, deduplicating identical rows and rejecting conflicts.
func NewCatalog(entries []ModelEntry) (CatalogSnapshot, error) {
	c := CatalogSnapshot{byID: make(map[string]ModelEntry, len(entries))}
	for _, e := range entries {
		if e.RouteID == "" || e.UpstreamID == "" {
			return CatalogSnapshot{}, errors.New("gateway model IDs must not be empty")
		}
		switch e.Provider {
		case ProviderAnthropic, ProviderOpenAI, ProviderZAI:
		default:
			return CatalogSnapshot{}, errors.New("gateway provider not supported")
		}
		if prior, ok := c.byID[e.RouteID]; ok {
			if prior != e {
				return CatalogSnapshot{}, fmt.Errorf("%w: %s", ErrConflictingModel, e.RouteID)
			}
			continue
		}
		c.byID[e.RouteID] = e
		c.entries = append(c.entries, e)
	}
	return c, nil
}

// NewSessionCatalog registers exact full model IDs plus configured GLM tiers.
// Claude subscription rows follow the positive M0 refresh observation recorded
// on 2026-09-11. Zero capabilities still require an explicit verified factory.
func NewSessionCatalog(glmTierIDs []string) (CatalogSnapshot, error) {
	var entries []ModelEntry
	for _, id := range []string{"claude-opus-5", "claude-sonnet-5"} {
		entries = append(entries, ModelEntry{RouteID: id, UpstreamID: id, Provider: ProviderAnthropic, AuthMethod: AuthOAuthPassthrough})
		entries = append(entries, ModelEntry{RouteID: id + "[1m]", UpstreamID: id, Provider: ProviderAnthropic, AuthMethod: AuthOAuthPassthrough})
	}
	for _, id := range []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"} {
		entries = append(entries, ModelEntry{RouteID: id, UpstreamID: id, Provider: ProviderOpenAI, AuthMethod: AuthPKCE})
	}
	for _, id := range glmTierIDs {
		entries = append(entries, ModelEntry{RouteID: id, UpstreamID: id, Provider: ProviderZAI, AuthMethod: AuthExistingGLM})
	}
	return NewCatalog(entries)
}

func (c CatalogSnapshot) Resolve(id string) (ModelEntry, error) {
	e, ok := c.byID[id]
	if !ok {
		return ModelEntry{}, fmt.Errorf("%w: %s", ErrUnknownModel, id)
	}
	return e, nil
}
func (c CatalogSnapshot) Entries() []ModelEntry { return append([]ModelEntry(nil), c.entries...) }
