package auth

import (
	"net/http"
	"strings"
)

const AnthropicEndpoint = "https://api.anthropic.com/v1/messages"

// NewAnthropicAPIKey accepts an explicit key only. Native subscription Bearers
// use NewAnthropicOAuth separately. Generation identifies this immutable reference.
func NewAnthropicAPIKey(key string) (CredentialRef, error) {
	if key == "" || strings.ContainsAny(key, "\r\n\x00") {
		return nil, ErrCredentialAbsent
	}
	return &anthropicKey{key: key}, nil
}

type anthropicKey struct{ key string }

func (r *anthropicKey) Provider() ProviderID        { return ProviderAnthropic }
func (r *anthropicKey) Generation() (uint64, error) { return 1, nil }
func (r *anthropicKey) Redacted() string            { return "anthropic:api-key" }
func (r *anthropicKey) Apply(req *http.Request) error {
	if !messagesDestination(req, AnthropicEndpoint, "api.anthropic.com") {
		return ErrWrongProvider
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Del("Authorization")
	req.Header.Set("X-Api-Key", r.key)
	return nil
}
func messagesDestination(req *http.Request, endpoint, host string) bool {
	return req != nil && req.Method == http.MethodPost && req.URL != nil && req.URL.String() == endpoint && req.URL.User == nil && (req.Host == "" || req.Host == host)
}
