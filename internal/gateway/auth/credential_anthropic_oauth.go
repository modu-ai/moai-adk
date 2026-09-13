package auth

import (
	"context"
	"net/http"
	"strings"
)

// NewAnthropicOAuth captures exactly one request's native subscription Bearer.
// M0 refresh transport observation on 2026-09-11 established the separate session
// header contract. This reference never reads or refreshes stored credentials.
func NewAnthropicOAuth(ctx context.Context, headers http.Header) (CredentialRef, error) {
	if ctx == nil {
		return nil, ErrCredentialAbsent
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var values []string
	for key, items := range headers {
		if strings.EqualFold(key, "Authorization") {
			values = append(values, items...)
		}
		if strings.EqualFold(key, "X-Api-Key") && len(items) > 0 {
			return nil, ErrCredentialAbsent
		}
	}
	if len(values) != 1 || len(values[0]) > 16384 {
		return nil, ErrCredentialAbsent
	}
	scheme, token, ok := strings.Cut(values[0], " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return nil, ErrCredentialAbsent
	}
	padding := false
	for _, c := range token {
		if c == '=' {
			padding = true
			continue
		}
		tokenRune := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("-._~+/", c)
		if padding || !tokenRune {
			return nil, ErrCredentialAbsent
		}
	}
	if strings.Trim(token, "=") == "" {
		return nil, ErrCredentialAbsent
	}
	return &anthropicOAuth{ctx: ctx, bearer: "Bearer " + token}, nil
}

type anthropicOAuth struct {
	ctx    context.Context
	bearer string
}

func (r *anthropicOAuth) Provider() ProviderID { return ProviderAnthropic }

// Generation is constant for this immutable request snapshot, not a global
// account generation. A subsequent native refresh creates a different reference.
func (r *anthropicOAuth) Generation() (uint64, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return 1, nil
}
func (r *anthropicOAuth) Redacted() string { return "anthropic:oauth-passthrough" }
func (r *anthropicOAuth) String() string   { return r.Redacted() }
func (r *anthropicOAuth) GoString() string { return r.Redacted() }
func (r *anthropicOAuth) Apply(req *http.Request) error {
	if err := r.ctx.Err(); err != nil {
		return err
	}
	if !messagesDestination(req, AnthropicEndpoint, "api.anthropic.com") {
		return ErrWrongProvider
	}
	if err := req.Context().Err(); err != nil {
		return err
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	for key := range req.Header {
		if strings.EqualFold(key, "Authorization") || strings.EqualFold(key, "X-Api-Key") {
			delete(req.Header, key)
		}
	}
	req.Header.Set("Authorization", r.bearer)
	return nil
}
