package auth

import (
	"errors"
	"github.com/modu-ai/moai-adk/internal/glmcred"
	"net/http"
	"strings"
	"testing"
)

func TestGLMStoredCredentialRotationAndProviderBoundary(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv(glmcred.EnvTestGLMKey, "")
	t.Setenv("Z_AI_API_KEY", "inherited-not-selected")
	if _, e := NewGLMCredential(); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
	if e := glmcred.Save(`synthetic-$key\value`); e != nil {
		t.Fatal(e)
	}
	ref, e := NewGLMCredential()
	if e != nil {
		t.Fatal(e)
	}
	if ref.Provider() != ProviderZAI || strings.Contains(ref.Redacted(), "synthetic") {
		t.Fatal("provider")
	}
	req, _ := http.NewRequest("POST", GLMEndpoint, nil)
	if e = ref.Apply(req); e != nil {
		t.Fatal(e)
	}
	if req.Header.Get("Authorization") != `Bearer synthetic-$key\value` {
		t.Fatal("storage roundtrip")
	}
	wrong, _ := http.NewRequest("POST", AnthropicEndpoint, nil)
	if e = ref.Apply(wrong); !errors.Is(e, ErrWrongProvider) {
		t.Fatal(e)
	}
	if e = glmcred.Save("rotated"); e != nil {
		t.Fatal(e)
	}
	if _, e = ref.Generation(); !errors.Is(e, ErrCredentialChanged) {
		t.Fatal(e)
	}
	req, _ = http.NewRequest("POST", GLMEndpoint, nil)
	if e = ref.Apply(req); !errors.Is(e, ErrCredentialChanged) || req.Header.Get("Authorization") != "" {
		t.Fatal("old reference sent")
	}
	next, e := NewGLMCredential()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = next.Generation(); e != nil {
		t.Fatal(e)
	}
	if e = glmcred.Save(""); e != nil {
		t.Fatal(e)
	}
	if _, e = next.Generation(); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
}
