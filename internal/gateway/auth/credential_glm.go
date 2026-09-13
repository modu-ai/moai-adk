package auth

import (
	"crypto/subtle"
	"github.com/modu-ai/moai-adk/internal/glmcred"
	"net/http"
	"strings"
)

// GLMEndpoint matches the existing default GLM Messages route. It cannot be
// replaced by an inherited base URL or inferred from a requested model name.
const GLMEndpoint = "https://api.z.ai/api/anthropic/v1/messages"

// NewGLMCredential uses the existing glmcred reader (including its documented
// test seam). It never consumes inherited Z_AI_API_KEY as an alternative source.
func NewGLMCredential() (CredentialRef, error) {
	key := glmcred.Load()
	if !validGLMKey(key) {
		return nil, ErrCredentialAbsent
	}
	return &glmKey{key: key}, nil
}

type glmKey struct{ key string }

func (r *glmKey) Provider() ProviderID { return ProviderZAI }
func (r *glmKey) Redacted() string     { return "zai:existing-glm" }

// Generation detects a stored key change for this reference. This reader-only
// check is NOT an atomic barrier against concurrent glmcred.Save or key restore.
func (r *glmKey) Generation() (uint64, error) {
	current := glmcred.Load()
	if !validGLMKey(current) {
		return 0, ErrCredentialAbsent
	}
	if subtle.ConstantTimeCompare([]byte(current), []byte(r.key)) != 1 {
		return 0, ErrCredentialChanged
	}
	return 1, nil
}
func (r *glmKey) Apply(req *http.Request) error {
	if !messagesDestination(req, GLMEndpoint, "api.z.ai") {
		return ErrWrongProvider
	}
	if _, e := r.Generation(); e != nil {
		return e
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Del("X-Api-Key")
	req.Header.Set("Authorization", "Bearer "+r.key)
	return nil
}
func validGLMKey(key string) bool { return key != "" && !strings.ContainsAny(key, "\r\n\x00") }
