package auth

import (
	"net/http"
	"strings"
)

// NewAPIKey requires an explicit key choice. It never reads the environment or
// acts as fallback for a missing subscription. Its immutable generation lasts
// only as long as this reference; no persistent API-key login is implied.
func NewAPIKey(key string) (CredentialRef, error) {
	if key == "" || strings.ContainsAny(key, "\r\n\x00") {
		return nil, ErrCredentialAbsent
	}
	return &apiKeyRef{key: key}, nil
}

type apiKeyRef struct{ key string }

func (r *apiKeyRef) Provider() ProviderID        { return ProviderOpenAI }
func (r *apiKeyRef) Generation() (uint64, error) { return 1, nil }
func (r *apiKeyRef) Redacted() string            { return "openai:api-key" }
func (r *apiKeyRef) Apply(req *http.Request) error {
	if req == nil || req.URL == nil || req.URL.String() != APIEndpoint || req.URL.User != nil || req.Host != "" && req.Host != "api.openai.com" {
		return ErrWrongProvider
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Authorization", "Bearer "+r.key)
	req.Header.Del("ChatGPT-Account-Id")
	return nil
}
