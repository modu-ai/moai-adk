package gateway

import (
	"context"
	"net/http"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
)

// RoutedRequest belongs to one request and has no inbound authentication headers.
// Body has already passed size, JSON and exact-model policy. An adapter must
// normalize history, apply Credential to its fixed provider endpoint, then recheck
// Credential.Generation against Generation immediately before RoundTrip. It must
// not send if the generation changed, or choose a fallback provider on failure.
type RoutedRequest struct {
	Entry      ModelEntry
	Body       []byte
	Headers    http.Header
	Credential CredentialRef
	Generation uint64
}

// Adapter sends to its configured provider endpoint only. It returns the upstream
// response without retrying after response bytes were emitted. Implementations
// must enforce RoutedRequest's credential and history contracts before sending.
type Adapter interface {
	Send(context.Context, RoutedRequest) (*http.Response, error)
}

func (s *Server) credential(ctx context.Context, entry ModelEntry, headers http.Header) (CredentialRef, uint64, error) {
	if entry.AuthMethod == AuthOAuthPassthrough {
		if entry.Provider != ProviderAnthropic {
			return nil, 0, auth.ErrWrongProvider
		}
		ref, err := auth.NewAnthropicOAuth(ctx, headers)
		if err != nil {
			return nil, 0, err
		}
		gen, err := ref.Generation()
		return ref, gen, err
	}
	if s.resolveCredential == nil {
		return nil, 0, auth.ErrCredentialAbsent
	}
	ref, err := s.resolveCredential(ctx, entry)
	if err != nil {
		return nil, 0, err
	}
	if ref == nil {
		return nil, 0, auth.ErrCredentialAbsent
	}
	if ref.Provider() != entry.Provider {
		return nil, 0, auth.ErrWrongProvider
	}
	gen, err := ref.Generation()
	return ref, gen, err
}
